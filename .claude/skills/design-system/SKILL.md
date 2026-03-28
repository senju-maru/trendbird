---
name: design-system
description: TrendBirdのフロントエンドUIを実装する際に使用するデザインシステム。コンポーネント作成、スタイリング、レイアウト設計時に自動参照される。Sky Blue Neumorphismテーマに基づく。
---

# TrendBird デザインシステム v1.0

**Sky Blue Neumorphism Theme**
最終更新: 2026-02-14

---

## 1. コンセプト

TrendBirdのUIは **ライトニューモーフィズム** をベースにしている。
背景と要素が同じ色で、影の明暗だけで「浮き（raised）」と「凹み（pressed）」を表現するスタイル。

アクセントカラーは **スカイブルー1色のみ** に統一し、濃淡で情報の強弱をつける。
蛍光色・多色使い・絵文字は使わない。

---

## 2. デザイントークン

### 2.1 カラー

```
背景（全要素共通）   #e4eaf1
影・暗い側           #becad6
影・明るい側         #ffffff

アクセント（メイン）  #41b1e1
アクセント（淡い）    #7acbee
アクセント（濃い）    #2a8fbd

テキスト・主要       #2c3e50
テキスト・補助       #5a7184
テキスト・淡い       #99aab5

オレンジ（急上昇）   #e67e22
```

**CSS変数として定義する場合:**

```css
:root {
  --bg:         #e4eaf1;
  --shadow-d:   #becad6;
  --shadow-l:   #ffffff;
  --blue:       #41b1e1;
  --blue-light: #7acbee;
  --blue-dark:  #2a8fbd;
  --orange:     #e67e22;
  --text:       #2c3e50;
  --text-sub:   #5a7184;
  --text-muted: #99aab5;
}
```

**使い分けルール:**

| 用途 | 色 |
|------|-----|
| 急上昇・強い強調 | `--orange` (#e67e22) |
| 上昇中・中程度の強調 | `--blue` (#41b1e1) |
| 安定・無効状態 | `--text-muted` (#99aab5) |
| 塗りつぶしボタン背景 | `--blue` → `--blue-light` のグラデーション |
| リンクテキスト | `--blue` |
| 見出し・本文 | `--text` (#2c3e50) |
| 説明・補足 | `--text-sub` (#5a7184) |
| ラベル・プレースホルダー | `--text-muted` (#99aab5) |

### 2.2 タイポグラフィ

```
フォント: 'Zen Maru Gothic', -apple-system, BlinkMacSystemFont, sans-serif

太さの使い分け:
  700  大きな数値、ブランド名、見出し、バッジ、ボタン
  500  ハンドル名、小ボタン
  400  本文、説明テキスト
  300  軽い補助テキスト

※ Zen Maru Gothic は 300/400/500/700/900 のみ。600 は存在しないため
  font-semibold(600) はブラウザが 700 にフォールバックする

サイズの目安:
  28-34px  ヒーロー数値（盛り上がり度など）
  22px     モーダル見出し
  18-19px  カード見出し（強調時）
  16px     カード見出し（通常）
  13-14px  本文
  12px     補足テキスト
  11px     バッジ、ラベル
  10px     キャプション、セクションラベル
```

### 2.3 角丸（border-radius）

```
ピル型バッジ      20px
カード            18-20px
モーダル          22-24px
ボタン（大）      14-16px
ボタン（小）      10-12px
インプット        12px
インセット領域    14-18px
```

### 2.4 スペーシング

```
カード内パディング    18px 20px
モーダル内パディング  26px 30px
カード間のギャップ    20px
セクション間マージン  18-26px
バッジ内パディング    4px 14px
ボタン内パディング    5px 16px（小）/ 14px 0（大・全幅）
```

---

## 3. シャドウシステム（核心）

このデザインの根幹。全ての要素は **raised（浮き出し）** か **pressed（凹み）** のどちらか。

### 3.1 基本公式

```
raised(size):
  box-shadow:
    {size}px {size}px {size*2}px #becad6,
    -{size}px -{size}px {size*2}px #ffffff;

pressed(size):
  box-shadow:
    inset {size}px {size}px {size*2}px #becad6,
    inset -{size}px -{size}px {size*2}px #ffffff;
```

**sizeの指針:**

| size | 用途 |
|------|------|
| 2-3 | 小ボタン、バッジ、微細な凹凸 |
| 4-5 | 中ボタン、入力フィールド、カード内のウェル |
| 6 | カード（デフォルト） |
| 8-9 | カード（ホバー時） |
| 10-14 | モーダル、大きなパネル |

### 3.2 JS/TSでの実装

```typescript
const SHADOW = {
  bg: "#e4eaf1",
  dark: "#becad6",
  light: "#ffffff",
};

export const raised = (size: number = 6): string =>
  `${size}px ${size}px ${size * 2}px ${SHADOW.dark}, ` +
  `-${size}px -${size}px ${size * 2}px ${SHADOW.light}`;

export const pressed = (size: number = 4): string =>
  `inset ${size}px ${size}px ${size * 2}px ${SHADOW.dark}, ` +
  `inset -${size}px -${size}px ${size * 2}px ${SHADOW.light}`;
```

### 3.3 CSSでの実装

```css
/* ユーティリティクラス */
.neu-raised-sm  { box-shadow: 3px 3px 6px #becad6, -3px -3px 6px #fff; }
.neu-raised     { box-shadow: 6px 6px 12px #becad6, -6px -6px 12px #fff; }
.neu-raised-lg  { box-shadow: 9px 9px 18px #becad6, -9px -9px 18px #fff; }
.neu-raised-xl  { box-shadow: 12px 12px 24px #becad6, -12px -12px 24px #fff; }

.neu-pressed-sm { box-shadow: inset 2px 2px 4px #becad6, inset -2px -2px 4px #fff; }
.neu-pressed    { box-shadow: inset 4px 4px 8px #becad6, inset -4px -4px 8px #fff; }
.neu-pressed-lg { box-shadow: inset 5px 5px 10px #becad6, inset -5px -5px 10px #fff; }
```

### 3.4 背景色のルール

**最重要ルール: 全ての要素の background は `#e4eaf1` と同じにする。**
ニューモーフィズムは背景と要素が同色であることで影の立体感が成立する。
要素に白や別の色を使うとデザインが崩壊する。

例外は以下のみ:
- 塗りつぶしボタン → `linear-gradient(135deg, #41b1e1, #7acbee)`
- トースト通知 → 同上のグラデーション

---

## 4. コンポーネントパターン

### 4.1 カード

ページ上の主要なコンテンツ容器。raised で浮き出す。

```
デフォルト:
  background: #e4eaf1
  border-radius: 20px
  box-shadow: raised(6)

ホバー:
  box-shadow: raised(9)
  transform: translateY(-2px)

プレス:
  box-shadow: pressed(5)
  transform: scale(0.99)

transition: all 0.22s ease
```

**カード上部のアクセントバー（オプション）:**
```
height: 3px
background: linear-gradient(90deg, #41b1e1, #7acbee)
opacity: 状態に応じて 0.5〜0.9
border-radius: カードと同じ値を上部のみに適用
```

### 4.2 インセット（凹み領域）

グラフ、テキスト表示領域など、情報を「くぼみ」に置く。

```
background: #e4eaf1
border-radius: 14-18px
padding: 8px 10px 〜 14px 18px
box-shadow: pressed(3)
```

### 4.3 バッジ（ステータス表示）

ピル型のインセットバッジ。

```
display: inline-flex
align-items: center
gap: 5px
font-size: 11px
font-weight: 600
color: 状態色（#e67e22 / #41b1e1 / #99aab5）
padding: 4px 14px
border-radius: 20px
background: #e4eaf1
box-shadow: pressed(2)
```

ドットインジケーター（オプション）:
```
width: 6px, height: 6px
border-radius: 50%
background: 状態色
```

### 4.4 ボタン

**通常ボタン（raised）:**
```
background: #e4eaf1
border: none
border-radius: 10-12px
padding: 5px 16px
font-size: 11px
font-weight: 500
color: --text-muted（デフォルト）
box-shadow: raised(3)

ホバー: 色が --blue に変わる
プレス: box-shadow を pressed(2) に切替
```

**塗りつぶしボタン（CTA）:**
```
background: linear-gradient(135deg, #41b1e1, #7acbee)
border: none
border-radius: 14-16px
padding: 14px 0（全幅）/ 9px 22px
color: #ffffff
font-size: 12-14px
font-weight: 600
box-shadow: 3px 3px 8px #becad6

プレス: pressed(3) に切替
```

**全幅ボタン（モーダル内のAI生成ボタンなど）:**
```
width: 100%
上記の塗りつぶしボタンのスタイルを適用
display: flex, justify-content: center
```

### 4.5 ヘッダー

```
position: sticky
top: 0
z-index: 100
background: #e4eaf1
box-shadow: 0 3px 12px rgba(190, 202, 214, 0.38)
height: 54px
padding: 0 28px
```

ロゴ: `font-size: 16px, font-weight: 700, color: #41b1e1`
ナビ（アクティブ）: `pressed(2)`, `color: --blue`, `font-weight: 600`
ナビ（非アクティブ）: `shadow: none`, `color: --text-muted`

### 4.6 モーダル

```
オーバーレイ:
  background: rgba(190, 202, 214, 0.45)
  backdrop-filter: blur(6px)

モーダル本体:
  background: #e4eaf1
  border-radius: 24px
  box-shadow: raised(14)
  max-width: 660px

閉じるボタン:
  32x32px, border-radius: 12px
  raised(3) → pressed(2)

アニメーション:
  overlay → fadeIn 0.2s
  modal → scaleIn 0.25s (scale 0.96 → 1, translateY 8px → 0)
```

### 4.7 フッター（固定バー）

```
position: fixed
bottom: 0
background: #e4eaf1
box-shadow: 0 -3px 12px rgba(190, 202, 214, 0.38)
padding: 12px 28px
```

### 4.8 トースト通知

```
position: fixed
bottom: 32px
left: 50%, transform: translateX(-50%)
background: linear-gradient(135deg, #41b1e1, #7acbee)
color: #ffffff
border-radius: 16px
padding: 10px 24px
font-size: 13px
font-weight: 500
box-shadow: 3px 3px 10px #becad6
```

### 4.9 入力フィールド（新規画面向け）

```
background: #e4eaf1
border: none
border-radius: 12px
padding: 12px 16px
font-size: 14px
color: --text
box-shadow: pressed(3)

フォーカス時:
  box-shadow: pressed(3), 0 0 0 2px #41b1e1（アウトライン追加）

プレースホルダー:
  color: --text-muted
```

### 4.10 トグルスイッチ（新規画面向け）

```
トラック（OFF）:
  width: 44px, height: 24px
  border-radius: 12px
  background: #e4eaf1
  box-shadow: pressed(2)

トラック（ON）:
  background: linear-gradient(135deg, #41b1e1, #7acbee)

サム（つまみ）:
  width: 20px, height: 20px
  border-radius: 50%
  background: #e4eaf1
  box-shadow: raised(2)
```

### 4.11 プログレスバー（新規画面向け）

```
トラック:
  height: 8px
  border-radius: 4px
  background: #e4eaf1
  box-shadow: pressed(2)

バー:
  height: 8px
  border-radius: 4px
  background: linear-gradient(90deg, #41b1e1, #7acbee)
```

### 4.12 アバター / アイコンボタン

```
width: 34px, height: 34px
border-radius: 12px（角丸）/ 50%（丸）
background: #e4eaf1
box-shadow: raised(3)
display: flex, align-items: center, justify-content: center

アイコン:
  SVGの stroke: --text-muted
  strokeWidth: 2
```

---

## 5. インタラクション

### 5.1 基本トランジション

全てのインタラクティブ要素に適用:
```
transition: all 0.22s ease
```

もしくは、より滑らかな:
```
transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1)
```

### 5.2 3段階ステート

全てのクリッカブル要素は以下の3段階で変化する:

| ステート | box-shadow | transform | 補足 |
|----------|------------|-----------|------|
| デフォルト | raised(適切なsize) | none | — |
| ホバー | raised(やや大きいsize) | translateY(-2px) | 影が強くなり浮き上がる |
| プレス | pressed(size) | scale(0.99) | 凹んで沈み込む |

### 5.3 登場アニメーション

カード等のリスト要素:
```css
@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

/* i番目の要素に delay を付与 */
animation: fadeUp 0.4s ease {i * 0.05}s both;
```

モーダル:
```css
@keyframes scaleIn {
  from { opacity: 0; transform: scale(0.96) translateY(8px); }
  to { opacity: 1; transform: scale(1) translateY(0); }
}
```

### 5.4 非アクティブ状態の表現

優先度の低いアイテムは opacity を下げる:
```
通常: opacity 0.58
ホバー: opacity 0.88
```

---

## 6. レイアウト

### 6.1 全体構造

```
max-width: 1200px
margin: 0 auto
padding: 24px 28px

ヘッダー: sticky top, 54px高
メイン: 上記padding
フッター: fixed bottom
```

### 6.2 グリッド

カード一覧:
```css
display: grid;
grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
gap: 20px;
```

モーダル内2カラム:
```css
display: flex;
gap: 14px;
flex-wrap: wrap;

/* 各カラム */
flex: 1;
min-width: 140px;
```

---

## 7. 画面ごとの適用ガイド

### ダッシュボード（実装済み）
- ヘッダー + ステータスバー + カードグリッド + フッター
- カード: raised(6) → hover raised(9) → press pressed(5)
- バッジ: pressed(2) のピル型

### トピック管理画面
- カードリスト or テーブル形式
- 追加ボタン: 塗りつぶしCTA
- 各トピック行: raised(4) のカード
- 編集モード: pressed(3) のインセット入力欄
- 削除: テキストボタン（色: --text-muted, ホバーで赤みを帯びさせない。同じブルー系で）

### 設定画面
- セクション: raised(5) のカード
- トグル: 4.10のパターン
- 入力欄: pressed(3)
- 保存ボタン: 塗りつぶしCTA

### オンボーディング / ログイン
- 中央配置のカード: raised(10)
- フォーム入力: pressed(3)
- ログインボタン: 塗りつぶしCTA（全幅）
- ソーシャルログイン: raised(4) のアイコンボタン

### プラン選択画面
- 3カラムのプランカード: raised(6)
- おすすめプラン: raised(8) + アクセントバー（opacity 0.9）
- 価格数値: font-size 34px, font-weight 700, color --blue
- 選択ボタン: おすすめのみ塗りつぶし、他は raised

### 通知設定画面
- チャネルごとのカード: raised(5)
- オン/オフ: トグルスイッチ
- 時間帯選択: pressed(3) のセレクトボックス

---

## 8. やること・やらないこと

### やること
- 全要素の background を `#e4eaf1` にする
- アクセントは `#41b1e1` の濃淡のみ使う
- クリッカブルな要素は3段階ステート（default → hover → press）を実装する
- 影の size を要素の大きさに合わせて調整する
- 凹み領域（グラフ、入力欄）には pressed を使う
- アニメーションは fadeUp + staggered delay で統一する

### やらないこと
- 蛍光色、ネオンカラーを使わない
- 絵文字をUI要素として使わない
- border で区切りを作らない（影で十分）
- `borderLeft` や `borderTop` にアクセントカラーを付けてアクティブ状態や強調を表現しない（AIが生成しがちなパターンだが、ニューモーフィズムでは影の凹凸で状態を表現する）
- 白背景(#fff)の要素を作らない（ニューモーフィズムが崩壊する）
- 影の size を大きくしすぎない（14以上は避ける）
- 1画面にニューモーフィズム要素を15個以上置かない（視覚的にうるさくなる）

---

## 9. コピペ用テンプレート

### React コンポーネントの起点

```jsx
// tokens.js
export const C = {
  bg: "#e4eaf1",
  shD: "#becad6",
  shL: "#ffffff",
  blue: "#41b1e1",
  blueLight: "#7acbee",
  blueDark: "#2a8fbd",
  text: "#2c3e50",
  textSub: "#5a7184",
  textMuted: "#99aab5",
};

export const up = (s = 6) =>
  `${s}px ${s}px ${s*2}px ${C.shD}, -${s}px -${s}px ${s*2}px ${C.shL}`;

export const dn = (s = 4) =>
  `inset ${s}px ${s}px ${s*2}px ${C.shD}, inset -${s}px -${s}px ${s*2}px ${C.shL}`;
```

### Tailwind CSS カスタム設定（参考）

```js
// tailwind.config.js
module.exports = {
  theme: {
    extend: {
      colors: {
        'neu-bg': '#e4eaf1',
        'neu-shadow-d': '#becad6',
        'accent': '#41b1e1',
        'accent-light': '#7acbee',
        'accent-dark': '#2a8fbd',
      },
      boxShadow: {
        'neu-sm': '3px 3px 6px #becad6, -3px -3px 6px #fff',
        'neu': '6px 6px 12px #becad6, -6px -6px 12px #fff',
        'neu-lg': '9px 9px 18px #becad6, -9px -9px 18px #fff',
        'neu-inset-sm': 'inset 2px 2px 4px #becad6, inset -2px -2px 4px #fff',
        'neu-inset': 'inset 4px 4px 8px #becad6, inset -4px -4px 8px #fff',
      },
    },
  },
};
```
