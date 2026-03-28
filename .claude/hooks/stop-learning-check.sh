#!/bin/bash
# Stop フック: セッション終了時に会話内容を分析し、知識ベースをバックグラウンドで自動更新する
#
# 動作:
#   1. stop_hook_active チェック（再帰防止）
#   2. トリガー条件: pending-learning.txt に内容がある OR ユーザーメッセージが3件以上
#   3. バックグラウンドで claude --print を起動
#      - transcript_path を Read して会話内容を把握
#      - 記録すべき知識を判断・抽出
#      - MEMORY.md と memory/*.md を更新
#      - pending-learning.txt をクリア
#   4. ロックファイルで多重起動を防止
#   5. 実行ログを .claude/retro/auto-update.log に記録

PROJECT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
PENDING_FILE="${PROJECT_DIR}/.claude/retro/pending-learning.txt"
LOCK_FILE="${PROJECT_DIR}/.claude/retro/.update-lock"
LOG_FILE="${PROJECT_DIR}/.claude/retro/auto-update.log"
SKILL_FILE="${PROJECT_DIR}/.claude/skills/domain-knowledge/SKILL.md"

# メモリファイルのパス計算（/path/to/project → -path-to-project）
MEMORY_KEY=$(echo "$PROJECT_DIR" | tr '/' '-')
MEMORY_DIR="${HOME}/.claude/projects/${MEMORY_KEY}/memory"
MEMORY_FILE="${MEMORY_DIR}/MEMORY.md"

# ------ stop_hook_active チェック（バックグラウンド claude からの再帰起動を防ぐ） ------
INPUT=$(cat /dev/stdin 2>/dev/null || echo "{}")
STOP_HOOK_ACTIVE=$(echo "$INPUT" | python3 -c \
  "import sys,json; d=json.load(sys.stdin); print(d.get('stop_hook_active', False))" \
  2>/dev/null || echo "False")

if [ "$STOP_HOOK_ACTIVE" = "True" ]; then
  exit 0
fi

# ------ transcript_path の抽出 ------
TRANSCRIPT_PATH=$(echo "$INPUT" | python3 -c \
  "import sys,json; d=json.load(sys.stdin); print(d.get('transcript_path', ''))" \
  2>/dev/null || echo "")

# ------ ユーザーメッセージ数のカウント ------
USER_MSG_COUNT=0
if [ -n "$TRANSCRIPT_PATH" ] && [ -f "$TRANSCRIPT_PATH" ]; then
  USER_MSG_COUNT=$(python3 -c "
import json
count = 0
try:
    with open('$TRANSCRIPT_PATH') as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                d = json.loads(line)
                if d.get('message', {}).get('role') == 'user':
                    count += 1
            except: pass
except: pass
print(count)
" 2>/dev/null || echo "0")
fi

# ------ トリガー条件チェック ------
HAS_PENDING=false
if [ -f "$PENDING_FILE" ] && [ -s "$PENDING_FILE" ]; then
  HAS_PENDING=true
fi

HAS_ENOUGH_MSGS=false
if [ "$USER_MSG_COUNT" -ge 3 ] 2>/dev/null; then
  HAS_ENOUGH_MSGS=true
fi

# どちらの条件も満たさなければスキップ
if [ "$HAS_PENDING" = "false" ] && [ "$HAS_ENOUGH_MSGS" = "false" ]; then
  exit 0
fi

# ロック中なら別プロセスが実行中 → スキップ
if [ -f "$LOCK_FILE" ]; then
  exit 0
fi

# ------ バックグラウンドで知識ベース更新 ------
mkdir -p "${PROJECT_DIR}/.claude/retro"

(
  touch "$LOCK_FILE"
  TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
  ITEM_COUNT=$(grep -c '' "$PENDING_FILE" 2>/dev/null || echo "0")

  {
    echo ""
    echo "=== [$TIMESTAMP] 自動更新開始 (pending: ${ITEM_COUNT} 件, user_msgs: ${USER_MSG_COUNT}) ==="
  } >> "$LOG_FILE"

  # プロンプトをファイルに書き出す（特殊文字を安全に扱うため）
  PROMPT_FILE=$(mktemp)
  {
    echo "あなたは TrendBird プロジェクトの知識ベース管理者です。"
    echo "セッションの会話内容を分析し、将来のセッションで役立つ知識を抽出・保存してください。"
    echo ""

    if [ -n "$TRANSCRIPT_PATH" ] && [ -f "$TRANSCRIPT_PATH" ]; then
      echo "## 会話トランスクリプト"
      echo "ファイルパス: ${TRANSCRIPT_PATH}"
      echo "Read ツールで末尾 2000 行程度を読み込み、何が議論・実装されたかを把握してください。"
      echo "（ファイルが大きい場合は offset を指定して末尾から読む）"
      echo ""
    fi

    if [ -f "$PENDING_FILE" ] && [ -s "$PENDING_FILE" ]; then
      echo "## 明示的な学習事項（pending-learning.txt の内容）"
      cat "$PENDING_FILE"
      echo ""
    fi

    echo "## ナレッジ抽出基準"
    echo ""
    echo "### 記録すべき内容:"
    echo "1. 技術仕様・実装ロジックの詳細（DBスキーマ、API設計、フロー、アルゴリズム）"
    echo "2. ユーザーの修正・好み・作業スタイル（コーディング規約、コミット規約など）"
    echo "3. バグの原因と解決パターン"
    echo "4. 設計上の重要な決定と背景"
    echo "5. 外部サービスの挙動・制約・コスト情報"
    echo ""
    echo "### 記録不要な内容:"
    echo "- ファイル読み込み・検索・起動など単純な操作のみのセッション"
    echo "- 一時的な状態確認（DBの件数確認など）"
    echo "- すでに MEMORY.md に記載済みの内容（重複は追記しない）"
    echo "- 特定の一度きりのタスクの作業手順"
    echo ""
    echo "## 更新手順"
    echo ""
    echo "### ステップ 1: 現在の MEMORY.md を確認（必須）"
    echo "- 対象ファイル: ${MEMORY_FILE}"
    echo "- Read して内容を把握する"
    echo ""
    echo "### ステップ 2: 記録すべき知識があるか判断"
    echo "- なければ「スキップ: 記録すべき新規知識なし」と 1 行で述べて終了"
    echo "- あれば以下のステップを続ける"
    echo ""
    echo "### ステップ 3: 詳細ナレッジは専用ファイルに保存"
    echo "- MEMORY.md はサマリーのみ（200 行以内）"
    echo "- 詳細なロジック・仕様・フロー図は ${MEMORY_DIR}/ 配下の専用ファイルに保存"
    echo "  例: ${MEMORY_DIR}/ai-post-generation.md, ${MEMORY_DIR}/stripe-webhook.md"
    echo "- 専用ファイルを作成したら MEMORY.md からリンクを張る"
    echo "  例: 「詳細は \`memory/ai-post-generation.md\` を参照」"
    echo ""
    echo "### ステップ 4: MEMORY.md を更新（必須）"
    echo "- 学習事項の各カテゴリに対応するセクションを見つけて更新する"
    echo "- 重複する内容は追記しない（既に同じ趣旨が書いてあればスキップ）"
    echo "- 全体を 200 行以内に収める（冗長な箇所は短縮する）"
    echo ""
    echo "### ステップ 5: domain-knowledge/SKILL.md を更新（条件付き）"
    echo "- 対象ファイル: ${SKILL_FILE}"
    echo "- 学習事項に「[スキル:domain-knowledge]」が含まれる場合のみ更新する"
    echo "- 含まれない場合はスキップしてよい"
    echo ""
    echo "### ステップ 6: pending-learning.txt をクリア（必須・最後に実行）"
    echo "- 対象ファイル: ${PENDING_FILE}"
    echo "- 更新が完了したら、このファイルの内容を空にする"
    echo "- ファイル自体は残す（内容だけ消す）"
    echo ""
    echo "作業完了後に「完了: X 件の学習事項を処理しました（新規: Y 件、スキップ: Z 件）」と 1 行で述べてください。"
  } > "$PROMPT_FILE"

  # CLAUDECODE を unset して実行（Claude Code セッション内からのネスト起動を許可）
  env -u CLAUDECODE claude --print \
    --allowedTools "Read,Edit,Write,Glob,Grep" \
    --max-turns 20 \
    "$(cat "$PROMPT_FILE")" >> "$LOG_FILE" 2>&1

  UPDATE_EXIT=$?
  rm -f "$PROMPT_FILE"

  {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] 自動更新完了 (exit: ${UPDATE_EXIT})"
  } >> "$LOG_FILE"

  rm -f "$LOCK_FILE"
) &

exit 0
