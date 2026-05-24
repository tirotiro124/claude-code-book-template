# ユビキタス言語定義

本ドキュメント・コード・レビュー・会話において共通の用語として使用する。
新しい概念を導入するときは必ずここに定義を追加してから実装する。

## 禁止用語

以下の用語は使わない。括弧内の用語を使うこと。

| 使ってはいけない | 代わりに使う | 理由 |
|----------------|------------|------|
| 優先度・優先順位 | スコア | 「優先度」は手動設定のイメージを持つ。このツールは自動計算のため |
| 期限・期日 | 締め切り（due date） | 表現を統一する |
| 依存 | ブロック依存 | 「依存」は方向が曖昧。「AがBをブロックする」の形で表現する |
| タスクID | `#<数字>` | UIおよびドキュメントでは `#42` の形式で表記する |

---

## 1. ドメイン用語

### タスク（Task）

開発者が管理する作業の最小単位。「何をするか」を表す1件のレコード。
コードでは `Task` 構造体として表現する。

サブタスクの概念はない。1件のタスクをそれ以上細かく分割できない。
複数の作業をまとめて管理する場合は、タグやブランチで関連付ける。

**含むもの：** タイトル・ステータス・締め切り・ブランチ・タグ・依存関係
**含まないもの：** スコア（スコアはタスクの属性ではなく、実行時に計算される値）

---

### ステータス（Status）

タスクが現在どの状態にあるかを表す。以下の4種類のみ存在する。

| 日本語 | 英語（コード上） | 意味 |
|--------|----------------|------|
| 未着手 | `todo` | 作成されたが、まだ着手していない |
| 進行中 | `in_progress` | 現在作業中（`task edit` で手動設定） |
| 完了 | `done` | 作業が終わり、一覧に表示されなくなる |
| スヌーズ中 | `snoozed` | 一時的に除外されている。期限が来たら `todo` に戻る |

**ステータス遷移：**

```
作成
  ↓
todo ──── task edit ────→ in_progress
  │                           │
  │ task snooze               │ task done / task edit
  ↓                           ↓
snoozed ──期限到来──→ todo     done
```

- `todo` → `in_progress`：`task edit` で手動変更のみ。自動遷移はない
- `todo`／`in_progress` → `done`：`task done <id>` コマンド
- `todo`／`in_progress` → `snoozed`：`task snooze <id> <期間>` コマンド
- `snoozed` → `todo`：`snoozed_until` の翌日以降、`task list`／`task next` 実行時に自動復帰（→ [スヌーズ](#スヌーズsnooze)）

---

### スコア（Score）

タスクの「今やるべき度」を表す数値。高いほど優先して取り組むべきタスクであることを意味する。
スコアはタスクの属性ではなく、実行時に5つのシグナルから計算される。

**取りうる範囲（balanced プリセット時）：**

| 状況 | スコア |
|------|--------|
| 最小値 | −10（全シグナルが0点 + 疲労ペナルティ最大） |
| 最大値 | 105（全シグナルが満点 + 疲労ペナルティなし） |

プリセットの重み係数によって最大値は変動する（例：deadline プリセット時の最大値は135）。

コードでは `ScoreResult` として各要素の内訳とともに返す。

---

### シグナル（Signal）

スコア計算に使われるインプット要素。現在5種類定義されている。

| シグナル名 | 日本語 | コード上のフィールド | 最大点 |
|-----------|--------|------------------|--------|
| Urgency | 緊急度 | `ScoreResult.Urgency` | 40点 |
| Importance | 重要度 | `ScoreResult.Importance` | 30点 |
| Context | コンテキスト加点 | `ScoreResult.Context` | 20点 |
| Aging | 経年加点 | `ScoreResult.Aging` | 15点 |
| Fatigue | 疲労ペナルティ | `ScoreResult.Fatigue` | −10点 |

---

### プリセット（Preset）

スコア計算における各シグナルの重み係数セット。
ユーザーの作業スタイルに合わせて `.taskrc` で選択する。

| プリセット名 | 日本語説明 | コード上の定数 |
|------------|-----------|-------------|
| `deadline` | 締め切り重視 | `PresetDeadline` |
| `balanced` | バランス型（デフォルト） | `PresetBalanced` |
| `flow` | フロー集中型 | `PresetFlow` |

---

### ブロック依存（Block）

「タスクAを始めるには、タスクBを先に完了させる必要がある」という前提条件の関係。

```
表現の統一:
「BはAをブロックしている」= 「AはBの完了を待っている」= 「BはAの前提条件である」

例:
#38「パスワードリセット実装」は #42「ログインバリデーション修正」をブロックしている
= #38 を完了するまで #42 に着手できない
= #42 は #38 が終わるまで待機状態
```

ブロック依存数が多いタスク（多くのタスクから待たれているタスク）は、重要度シグナル（Importance）が高くなる（→ [重要度](#シグナルsignal)）。

コードでは `task_blocks` テーブルで管理する。
`blocked_by_id` のタスクが `done` になったとき、対応するレコードをアプリケーション層で削除する。

---

### スヌーズ（Snooze）

タスクを指定日まで一時的にスコア計算と一覧表示から除外する操作。
スヌーズを行うとステータスが `todo`／`in_progress` から `snoozed` に変わる。

`snoozed_until` の日付を過ぎると、`task list`／`task next` 実行時に自動的に `todo` に戻る（遅延評価）。デーモン等のバックグラウンドプロセスは使用しない。

`task pin` によるピン固定とは逆の操作。スヌーズは「しばらく見たくない」、ピン固定は「常に目に入れたい」という意図。

---

### ピン固定（Pin）

スコアに関わらずタスクを常にリスト最上位に表示する設定。
`pinned = 1` のタスクはスコア計算をスキップして先頭に表示される。

**解除条件：**
- `task unpin <id>` コマンドでのみ手動解除する
- `task done <id>` で完了しても自動的にピンは外れない（完了済みタスクは一覧に表示されなくなるため実害はない）

---

### 期限切れ（Overdue）

締め切り日を過ぎてもステータスが `done` になっていないタスクの状態。

スコア計算では緊急度（Urgency）が40点を超えて超過日数×2点が加算される（上限なし）。
UI では `due` の表示を赤色にして視覚的に通知する。

コード上の判定：`due_date < 現在日付 AND status != 'done'`

---

### TODO同期（TODO Sync）

ソースコード内の `TODO:` コメントをタスクとして取り込む操作。
`task sync` コマンドで実行する。

重複判定は `source_file`（ファイルパス）+ `source_line`（行番号）の組み合わせで行う。
同じファイルの同じ行から取り込まれたタスクは再登録しない。

---

### Gitコンテキスト（Git Context）

スコア計算時に参照するGit情報。現在は「現在のブランチ名」のみを含む。

コードでは `GitContext` 構造体として表現し、`Provider` インターフェース経由で取得する。
gitリポジトリ外では空の `GitContext` を返し、エラーにしない。

---

### 疲労ペナルティ（Fatigue Penalty）

`task next` で直近に表示されたタスクのスコアを下げるペナルティ。
同じタスクが連続して表示され続けることへの違和感を防ぎ、着手できない理由がある場合でも別のタスクを提案できるようにする。

直近2件の表示履歴を `display_history` テーブルで管理する。

---

### 経年加点（Aging Bonus）

作成してから長期間 `todo` または `in_progress` のまま放置されているタスクのスコアを引き上げる加点。
タスクが永遠に後回しにされることを防ぐ。最大14日で上限（15点）に達する。

「着手」の定義：ステータスが `done` または `snoozed` になること。
`in_progress` は着手とみなさず、経年加点は継続して加算される。

---

## 2. UI / UX 用語

### タスク一覧（Task List）

`task list` コマンドの出力。スコア順（ピン固定タスクが先頭）で全未完了タスクを表示する。

---

### `task next`

現在のコンテキストでスコアが最も高いタスクを1件表示するコマンド。
このツールの中心的な操作であり、「次に何をすべきか」の答えを返す。

---

### スコア内訳（Score Breakdown）

`task show` で表示される、スコアを構成する各シグナルの得点一覧。
ユーザーがスコアの根拠を確認し、推薦に納得するために使う。

---

### ヒント表示（Action Hint）

`task next` の出力下部に表示される「次に打てるコマンド」の提示。
コマンドを暗記しなくても次のアクションに進めるようにする。

```
$ task done 42     $ task snooze 42 1d
```

---

### ドライラン（Dry Run）

`task sync --dry-run` で実行する「実際には登録せず、登録予定の内容を表示するだけ」のモード。
破壊的操作を実行前に確認するために使う。

---

## 3. 英語・日本語対応表

| 英語（コード・ドキュメント） | 日本語（UI・ドキュメント） |
|--------------------------|------------------------|
| Task | タスク |
| Score | スコア |
| Signal | シグナル |
| Preset | プリセット |
| Status | ステータス |
| Urgency | 緊急度 |
| Importance | 重要度 |
| Context | コンテキスト加点 |
| Aging | 経年加点 |
| Fatigue | 疲労ペナルティ |
| Block / Dependency | ブロック依存 |
| Snooze | スヌーズ |
| Pin | ピン固定 |
| Due date | 締め切り |
| Overdue | 期限切れ |
| Branch | ブランチ |
| Tag | タグ |
| Dry run | ドライラン |
| Migration | マイグレーション |
| Score breakdown | スコア内訳 |
| Action hint | ヒント表示 |
| Git context | Gitコンテキスト |
| TODO sync | TODO同期 |
| Display history | 表示履歴 |
| Weight | 重み係数 |
| Task list | タスク一覧 |

---

## 4. コード上の命名規則まとめ

### 型・構造体

| 概念 | 型名 | 備考 |
|------|------|------|
| タスク | `Task` | |
| スコア計算結果 | `ScoreResult` | |
| Git情報 | `GitContext` | |
| アプリ設定 | `Config` | |
| スコア重み係数セット | `WeightMap` | `Config` 内のフィールドとして定義。`map[string]float64` の型エイリアス |
| Gitアクセス抽象 | `Provider`（インターフェース） | `internal/git/context.go` で定義 |
| Git実装 | `ExecProvider` | `internal/git/context.go` で定義 |

> テスト専用の `MockProvider` はプロダクションコードではないため、ここには掲載しない。
> 定義場所は `internal/git/context_test.go`。

### ステータス型

ステータスは型定義を使い、文字列の直書きを防ぐ。

```go
type Status string

const (
    StatusTodo       Status = "todo"
    StatusInProgress Status = "in_progress"
    StatusDone       Status = "done"
    StatusSnoozed    Status = "snoozed"
)
```

`Task.Status` フィールドの型は `Status`（`string` ではない）。

### プリセット定数

```go
type Preset string

const (
    PresetDeadline Preset = "deadline"
    PresetBalanced Preset = "balanced"
    PresetFlow     Preset = "flow"
)
```

### エラー変数

すべて `errors.New()` による sentinel error として定義する。
呼び出し側は `errors.Is(err, ErrTaskNotFound)` で比較する。

| 状況 | エラー変数名 |
|------|------------|
| 指定IDのタスクが存在しない | `ErrTaskNotFound` |
| 不正なプリセット値 | `ErrInvalidPreset` |
| gitリポジトリ外での実行 | `ErrNotGitRepository` |
| DBの初期化またはマイグレーション失敗 | `ErrDBSetup` |
