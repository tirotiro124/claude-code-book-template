# 開発ガイドライン

## 1. コーディング規約

### 基本方針

- 標準的なGoの慣習（[Effective Go](https://go.dev/doc/effective_go)・[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)）に従う
- `gofmt` + `goimports` で自動整形する。スタイルの議論はしない
- `golangci-lint` が通らないコードはマージしない

### エラーハンドリング

エラーは握りつぶさず、必ずコンテキストを付けてラップして返す。

```go
// Bad
db, _ := sql.Open("sqlite", path)

// Bad（コンテキストなし）
return err

// Good
db, err := sql.Open("sqlite", path)
if err != nil {
    return fmt.Errorf("DB接続に失敗しました: %w", err)
}
```

`cmd/` 層でのみエラーを最終的に処理する。`internal/` パッケージはエラーを返すだけにする。

```go
// cmd/add.go
if err := taskRepo.Create(task); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

### パッケージ設計

- パッケージ名は短く・小文字・単数形にする（`configs` ではなく `config`）
- パッケージのコメントはファイルの先頭に `// Package config ...` の形式で書く
- 循環インポートは禁止。`internal/` パッケージ間の依存方向は一方向に保つ

```
許可する依存方向:
cmd → internal/config
cmd → internal/db
cmd → internal/scoring
cmd → internal/render
internal/db → internal/config（設定値を参照する場合）
internal/scoring → （依存なし。純粋関数）
internal/scanner → internal/git（gitルート取得のため）

禁止:
internal/db → internal/render
internal/scoring → internal/db
```

### インターフェース

インターフェースは**使う側**のパッケージで定義する（Goの慣習）。

```go
// internal/git/context.go に Provider インターフェースを定義し、
// cmd/ 層や internal/scanner/ がそれを使う

type Provider interface {
    CurrentBranch() (string, error)
    RootDir() (string, error)
}
```

小さく保つ。1〜3メソッドが理想。

### コメント

公開シンボル（大文字始まりの関数・型）には GoDoc コメントを書く。

```go
// Calculate はタスクのスコアを計算して ScoreResult を返す。
// 副作用を持たない純粋関数として実装する。
func Calculate(task Task, ctx Context, cfg config.Config) ScoreResult {
```

内部実装のコメントは「なぜそうするか」が自明でない場合のみ書く。コードを読めばわかることは書かない。

---

## 2. 命名規則

### Go全般

| 種別 | 規則 | 例 |
|------|------|-----|
| パッケージ名 | 小文字・単数形 | `config`, `scoring`, `render` |
| 型名・インターフェース名 | UpperCamelCase | `Task`, `ScoreResult`, `Provider` |
| 関数・メソッド名 | UpperCamelCase（公開）/ lowerCamelCase（非公開） | `Calculate`, `applyWeights` |
| 変数名 | lowerCamelCase | `taskRepo`, `currentBranch` |
| 定数名 | UpperCamelCase | `StatusTodo`, `StatusDone` |
| エラー変数 | `Err` プレフィックス | `ErrTaskNotFound`, `ErrInvalidPreset` |

### ドメイン固有

| 概念 | Go上の名前 |
|------|----------|
| タスク | `Task` |
| スコア計算結果 | `ScoreResult` |
| Git情報 | `GitContext` |
| 設定 | `Config` |
| タスクステータス | `StatusTodo`, `StatusInProgress`, `StatusDone`, `StatusSnoozed` |
| プリセット | `PresetDeadline`, `PresetBalanced`, `PresetFlow` |

### ファイル名

- 小文字・スネークケース（Goの標準）
- テストファイルは `<対象ファイル名>_test.go`
- 1ファイル1責務を原則とし、ファイル名はその責務を表す

---

## 3. テスト規約

### テストの種類と方針

| 種類 | 対象 | 方針 |
|------|------|------|
| ユニットテスト | スコアエンジン・設定パース | テーブルドリブン。外部依存なし |
| 結合テスト | タスクリポジトリ | インメモリSQLite（`:memory:`）を使用 |
| 結合テスト | Gitコンテキスト | `testhelper.NewTempGitRepo(t)` で一時リポジトリ作成 |
| E2Eテスト | CLIコマンド | `cobra.ExecuteC()` でコマンドを実行し出力を検証 |

### テーブルドリブンテストの書き方

```go
func TestCalculate(t *testing.T) {
    tests := []struct {
        name     string
        task     Task
        ctx      Context
        expected int
    }{
        {
            name:     "締め切り2日前は緊急度30点",
            task:     Task{DueDate: today.AddDate(0, 0, 2)},
            ctx:      Context{Branch: ""},
            expected: 30,
        },
        {
            name:     "ブランチ一致でコンテキスト加点15点",
            task:     Task{Branch: "feat/login"},
            ctx:      Context{Branch: "feat/login"},
            expected: 15,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Calculate(tt.task, tt.ctx, defaultConfig())
            assert.Equal(t, tt.expected, result.Total)
        })
    }
}
```

### カバレッジ目標

| パッケージ | 目標カバレッジ |
|-----------|-------------|
| `internal/scoring/` | 100% |
| `internal/db/` | 80%以上 |
| `internal/config/` | 80%以上 |
| `internal/scanner/` | 80%以上 |
| その他 | 80%以上 |

### テスト実行

```bash
make test           # go test -race -cover ./...
make bench          # go test -bench=. -benchmem ./...
```

`-race` フラグを必ずつけてデータ競合を検出する。

---

## 4. Git規約

### ブランチ命名

```
feat/<内容>       # 新機能
fix/<内容>        # バグ修正
refactor/<内容>   # リファクタリング
docs/<内容>       # ドキュメントのみの変更
test/<内容>       # テストのみの変更
chore/<内容>      # ビルド・依存関係など
```

例：`feat/scoring-engine`、`fix/snooze-reactivation`、`docs/update-architecture`

### コミットメッセージ

[Conventional Commits](https://www.conventionalcommits.org/) に従う。

```
<type>(<scope>): <要約>

<本文（任意）>
```

```
feat(scoring): スコアエンジンの初期実装

緊急度・重要度・コンテキスト・経年・疲労の5要素を計算する。
プリセット（deadline/balanced/flow）の重み係数を適用する。

fix(db): スヌーズ期限切れタスクが復活しない問題を修正

task list / task next 実行時に snoozed_until の遅延評価を行うよう変更。
```

**1行目のルール：**
- 50文字以内
- 動詞から始める（「追加」「修正」「削除」など）
- 末尾にピリオドをつけない

### プルリクエスト

- 1PRに1つの目的。スコープを小さく保つ
- PRのタイトルはコミットメッセージの1行目と同じ形式にする
- マージ前に `golangci-lint` と `go test -race` がCIで通っていること
- セルフレビューでコードを読み直してからレビュー依頼する

### マージ戦略

- `main` ブランチへのマージはSquash Mergeを使用する
- 作業中のコミット履歴は整理しなくてよい（Squashでまとまるため）
- `main` への直接 push は禁止

---

## 5. 依存関係の管理

### 追加時のルール

- 新しいライブラリを追加する前に標準ライブラリで代替できないか確認する
- ライセンスが MIT / BSD / Apache 2.0 であることを確認する
- メンテナンスが継続されているか（直近1年以内のコミットがあるか）を確認する

### 更新

```bash
go get -u ./...       # 全依存関係を最新に更新
go mod tidy           # 不要な依存関係を削除
```

依存関係の更新は機能開発とは別のPRで行う。

---

## 6. 開発環境セットアップ

```bash
# 1. リポジトリのクローン
git clone https://github.com/task-cli/task.git
cd task

# 2. Go バージョンの確認（.go-version に合わせること）
go version

# 3. 依存関係のインストール
go mod download

# 4. ビルド確認
make build

# 5. テスト実行
make test

# 6. リント確認
make lint
```

### 推奨エディタ設定（VSCode）

`.vscode/settings.json` に以下を設定することを推奨する。

```json
{
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "editor.formatOnSave": true
}
```
