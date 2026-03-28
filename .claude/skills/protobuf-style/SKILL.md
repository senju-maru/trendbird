---
name: protobuf-style
description: TrendBirdのProtobufファイルにおける日本語コメント規約。proto定義の新規作成・修正時に自動参照される。protoファイルをサービス間の契約書として機能させる。
---

# TrendBird Protobuf コメント規約 v1.0

最終更新: 2026-02-17

---

## 1. 概要・目的

protoファイルは **サービス間の契約書** である。
コードを読むだけで API の目的・構造・制約・前提条件を理解できる状態を目指す。

### 原則

- コメントは **日本語** で書く（識別子・技術用語はそのまま英語可）
- 「何をするか」だけでなく「なぜそうなっているか」「どんな制約があるか」を書く
- バリデーションルール（`buf.validate`）だけでは伝わらないビジネス上の意図を補足する
- 自明なこと（フィールド名から明らかな内容）は繰り返さない

---

## 2. ファイル冒頭コメント

各 proto ファイルの先頭行に、そのファイルが担当する **領域の要約** を1行コメントで記述する。

### ルール

- `syntax` 宣言の直前に置く
- 1行で完結させる（詳細は service / message コメントに委ねる）
- 認証要件・外部依存がある場合は明示する

### 具体例

```protobuf
// X(Twitter) OAuth 認証とユーザー管理
syntax = "proto3";
```

```protobuf
// 複数サービスで共有される汎用メッセージ定義
syntax = "proto3";
```

---

## 3. service コメント規約

### ルール

- サービス宣言の直前に **1行コメント** を置く
- 内容: サービス全体の責務 + 認証要件
- 形式: `// <責務の要約>（<認証要件>）`

### 具体例

```protobuf
// 認証サービス
service AuthService {
```

```protobuf
// トピック管理サービス（認証必須）
service TopicService {
```

```protobuf
// 投稿管理サービス（認証必須）
service PostService {
```

---

## 4. rpc コメント規約

### ルール

- 各 rpc の直前に **1行コメント** を置く
- 内容: API の目的を簡潔に記述する
- 認証がサービスレベルと異なる場合は括弧書きで補足する
- 副作用・重要なエラーケースがある場合は積極的に記述する（複数行可）

### 具体例（基本）

```protobuf
service TopicService {
  // 監視中の全トピック一覧を取得
  rpc ListTopics(ListTopicsRequest) returns (ListTopicsResponse);
  // トピックの詳細を取得
  rpc GetTopic(GetTopicRequest) returns (GetTopicResponse);
  // 新しいトピックを作成して監視を開始
  rpc CreateTopic(CreateTopicRequest) returns (CreateTopicResponse);
  // トピックを削除して監視を停止
  rpc DeleteTopic(DeleteTopicRequest) returns (DeleteTopicResponse);
}
```

### 具体例（認証レベルの違いを明示）

```protobuf
service AuthService {
  // X OAuth 認証（認証不要・パブリック）
  rpc XAuth(XAuthRequest) returns (XAuthResponse);
  // ログアウト（認証必須）
  rpc Logout(LogoutRequest) returns (LogoutResponse);
}
```

### 具体例（副作用・制限の記述）

```protobuf
service PostService {
  // トピックに基づいて投稿文を生成
  // - スタイル未指定時は全スタイル（casual, breaking, analysis）で生成
  rpc GeneratePosts(GeneratePostsRequest) returns (GeneratePostsResponse);
  // 即時投稿
  // - X API を呼び出して投稿を実行する（外部副作用あり）
  // - 投稿失敗時は PostStatus.FAILED となり error_message にエラー詳細が入る
  rpc PublishPost(PublishPostRequest) returns (PublishPostResponse);
}
```

---

## 5. message コメント規約

### ルール

- message 宣言の直前に **1行コメント** を置く
- 内容: そのメッセージが表すデータの概要・主な用途
- Request/Response メッセージは、自明な場合はコメント省略可（ただし特殊な挙動がある場合は記述する）
- optional フィールドの出現条件やフィールド間の組み合わせルールがある場合は、message コメントに記述する

### 具体例（基本）

```protobuf
// ユーザープロフィール情報
message User {
```

```protobuf
// 監視対象トピック
message Topic {
```

```protobuf
// AI が生成した投稿文
message GeneratedPost {
```

### 具体例（optional 条件の説明）

```protobuf
// 通知メッセージ
// - type=TREND の場合: topic_id, topic_name, topic_status が設定される
// - type=SYSTEM の場合: action_url, action_label が設定されることがある
message Notification {
```

### 具体例（特殊な挙動を持つリクエスト）

```protobuf
// プロフィール更新リクエスト（optional フィールドのみ部分更新）
message UpdateProfileRequest {
```

```protobuf
// 通知設定の更新リクエスト（optional フィールドのみ部分更新）
message UpdateNotificationsRequest {
```

---

## 6. enum コメント規約

### ルール

- enum 宣言の直前に **1行コメント** を置く
- 各値にも **インラインコメント** を付ける（`UNSPECIFIED` を除く）
- 状態遷移がある enum は、遷移の流れを message レベルまたは enum レベルのコメントで補足する

### 具体例（状態遷移あり）

```protobuf
// 投稿のステータス
// 遷移: DRAFT → SCHEDULED → PUBLISHED
//                         → FAILED（投稿失敗時）
enum PostStatus {
  POST_STATUS_UNSPECIFIED = 0;
  // 下書き
  POST_STATUS_DRAFT = 1;
  // 予約投稿
  POST_STATUS_SCHEDULED = 2;
  // 投稿済み
  POST_STATUS_PUBLISHED = 3;
  // 投稿失敗
  POST_STATUS_FAILED = 4;
}
```

---

## 7. field コメント規約

### ルール

- フィールド宣言の **直前の行** にコメントを置く（インラインコメント `// ...` も可だが、長い場合は前行に）
- 以下の場合に記述する:
  - フィールド名だけでは意味が不明確なとき
  - 単位がある場合（円、分、%、0-23 等）
  - optional の出現条件がある場合
  - バリデーションルールだけでは伝わらないビジネス上の意図がある場合
- 自明なフィールド（`id`, `name` 等で用途が明確）にはコメント不要

### 具体例

```protobuf
message Topic {
  string id = 1;
  // トピック表示名
  string name = 2;
  // 検索に使用するキーワード群
  repeated string keywords = 3;
  // ベースラインからの変化率（%）
  double change_percent = 6;
  // 統計的異常度を示す z-score（高いほどスパイクの可能性大）
  optional double z_score = 7;
  // 直近24時間のスパークラインデータ
  repeated SparklineDataPoint sparkline_data = 10;
  // トピックの補足コンテキスト
  optional string context = 11;
  // 現在のスパイク開始日時（ISO 8601、スパイク中のみ）
  optional string spike_started_at = 15;
}
```

```protobuf
message SparklineDataPoint {
  // ISO 8601 形式のタイムスタンプ
  string timestamp = 1;
  // その時点での言及量
  int32 value = 2;
}
```

### 記述パターン早見表

| パターン | 例 |
|---|---|
| 単位 | `// 支払い金額（円）` |
| 時刻形式 | `// 支払い日（ISO 8601）` |
| 範囲 | `// ピーク時間帯の開始時（0-23）` |
| optional 条件 | `// 現在のスパイク開始日時（ISO 8601、スパイク中のみ）` |
| ビジネス意図 | `// 自動投稿の有効フラグ（X連携済みユーザーのみ変更可）` |
| 用途の補足 | `// UIに表示するスタイル名（例: "カジュアル"）` |

---

## 8. チェックリスト

proto ファイルを新規作成・修正した際は、以下を確認すること。

### 新規ファイル作成時

- [ ] ファイル冒頭に領域の要約コメントがあるか
- [ ] すべての service にコメント（責務 + 認証要件）があるか
- [ ] すべての rpc にコメント（目的）があるか
- [ ] すべての message にコメント（概要・用途）があるか（自明な Request/Response は省略可）
- [ ] すべての enum にコメントがあり、各値にも説明があるか
- [ ] 状態遷移のある enum に遷移図コメントがあるか
- [ ] optional フィールドの出現条件が記述されているか
- [ ] 単位・形式のあるフィールドに補足があるか

### 既存ファイル修正時

- [ ] 追加した service / rpc / message / enum / field にコメントがあるか
- [ ] 既存コメントが変更内容と矛盾していないか（コメントの更新漏れがないか）
- [ ] 状態遷移に変更があった場合、遷移図コメントが更新されているか
