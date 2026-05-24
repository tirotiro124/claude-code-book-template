# リポジトリ構造定義書

## 1. ディレクトリ・ファイル構成

```
task/
├── .github/
│   └── workflows/
│       ├── ci.yml              # PR・pushごとのlint + test + build
│       └── release.yml         # タグプッシュ時のリリース
├── cmd/
│   ├── root.go                 # rootCmd 定義・Execute() 関数・グローバルフラグ
│   ├── add.go                  # task add
│   ├── list.go                 # task list
│   ├── next.go                 # task next
│   ├── done.go                 # task done
│   ├── show.go                 # task show
│   ├── edit.go                 # task edit
│   ├── snooze.go               # task snooze
│   ├── pin.go                  # task pin / task unpin
│   ├── block.go                # task block / task unblock
│   ├── sync.go                 # task sync
│   └── export.go               # task export
├── internal/
│   ├── config/
│   │   ├── config.go           # コンフィグローダー・デフォルト値
│   │   └── config_test.go
│   ├── db/
│   │   ├── db.go               # DB接続・ライフサイクル・マイグレーション実行
│   │   ├── db_test.go
│   │   ├── migrations.go       # マイグレーションSQL定義（配列）
│   │   ├── task.go             # Task ドメインに関するすべてのクエリ
│   │   └── task_test.go
│   ├── scoring/
│   │   ├── engine.go           # スコアエンジン（副作用なし）
│   │   ├── engine_test.go
│   │   └── testdata/           # スコア計算のテスト入力データ（JSONなど）
│   ├── git/
│   │   ├── context.go          # Provider インターフェース + ExecProvider 実装
│   │   └── context_test.go
│   ├── scanner/
│   │   ├── todo.go             # TODOスキャナー（.gitignore 尊重）
│   │   ├── todo_test.go
│   │   └── testdata/           # TODOスキャンのテスト用ソースファイル群
│   ├── render/
│   │   ├── renderer.go         # Renderer 構造体・共通出力ロジック
│   │   ├── task.go             # タスク関連の出力メソッド（next, list, show）
│   │   ├── message.go          # 完了・エラーメッセージの出力メソッド
│   │   └── renderer_test.go
│   └── testhelper/
│       └── helper.go           # テスト共通ユーティリティ（Go関数のみ）
├── docs/                       # 永続的ドキュメント
│   ├── product-requirements.md
│   ├── functional-design.md
│   ├── architecture.md
│   ├── repository-structure.md （本ファイル）
│   ├── development-guidelines.md
│   └── glossary.md
├── .steering/                  # 作業単位のドキュメント（履歴として保持）
│   └── YYYYMMDD-[タイトル]/
│       ├── requirements.md
│       ├── design.md
│       └── tasklist.md
├── main.go                     # エントリーポイント（cmd.Execute() を呼ぶだけ）
├── go.mod                      # toolchain go1.22.0 を明記
├── go.sum                      # 自動生成。手動編集禁止
├── .go-version                 # 使用するGoバージョン（例: 1.22.3）
├── .gitignore
├── .goreleaser.yaml            # リリースビルド設定
├── .golangci.yml               # リントルール設定
├── Makefile                    # 開発用タスクランナー
├── CONTRIBUTING.md             # コントリビューションガイド（README には概要のみ）
├── LICENSE                     # MIT ライセンス
└── README.md                   # インストール方法・基本的な使い方
```

---

## 2. ディレクトリの役割

### `cmd/`

CLIコマンドの定義層。cobra の `Command` 構造体を定義し、フラグの解析と `internal/` パッケージへの処理の委譲を行う。

**ルール：**
- ビジネスロジックは持たない。`internal/` パッケージを呼び出すだけにする
- 1ファイル1コマンド（または `pin/unpin`・`block/unblock` のように対になるコマンド）を原則とする
- 各コマンドファイルの `init()` 内で `rootCmd.AddCommand()` を呼ぶ（後述）
- `root.go` には `rootCmd` の定義・`Execute()` 関数・`--verbose`・`--json`・`--no-color` のグローバルフラグ・`version` 変数のみを置く

---

### `internal/`

アプリケーションの内部実装。Goの `internal` パッケージ制約により、このリポジトリ外からインポートできない。

#### `internal/config/`

`.taskrc` の探索・読み込み・バリデーションを担う。

- `config.go` — `Config` 構造体の定義、`.taskrc` の探索ロジック（カレントから上位へ）、デフォルト値の適用

#### `internal/db/`

SQLiteへのアクセスを担う。2ファイルで役割を明確に分離する。

| ファイル | 責務 |
|--------|------|
| `db.go` | DB接続・`~/.task/` の自動作成・PRAGMA設定・マイグレーション実行のみ |
| `task.go` | `Task` ドメインに関するすべてのSQL（CRUD・フィルタリング・スヌーズ再アクティブ化）|

`db.go` にタスクのクエリを書いてはならない。`task.go` にDB接続・初期化ロジックを書いてはならない。

#### `internal/scoring/`

スコア計算ロジック。外部状態への依存を持たない純粋な計算関数として実装する。

- `engine.go` — `Calculate(task, context, config) ScoreResult` 関数の実装
- `testdata/` — テーブルドリブンテスト用の入力データファイル（必要に応じて配置）

#### `internal/git/`

gitコマンドへのアクセスを抽象化する。

- `context.go` — `Provider` インターフェース定義、`ExecProvider`（実プロセス実行）の実装

#### `internal/scanner/`

ソースコード内の `TODO:` コメントを抽出する。

- `todo.go` — `.gitignore` パース、ファイル走査、`TODO:` 抽出、重複チェック
- `testdata/` — スキャン対象の仮想ソースファイル群（`.gitignore` のテストケース含む）

#### `internal/render/`

ターミナル出力のフォーマットを担う。コマンドが増えるにつれて肥大化しないよう、最初から責務ごとにファイルを分割する。

| ファイル | 責務 |
|--------|------|
| `renderer.go` | `Renderer` 構造体・`--json`/`NO_COLOR` の切り替え・stderr 出力の共通ロジック |
| `task.go` | `next`・`list`・`show` などタスク表示メソッド |
| `message.go` | 完了メッセージ・エラーメッセージ出力メソッド |

新たな出力パターンが増えた場合は `renderer.go` に追記せず、責務に応じた新しいファイルに追加する。

#### `internal/testhelper/`

複数のテストパッケージから参照する共通のGo関数を置く。データファイルは含めない（データファイルは各パッケージの `testdata/` に置く）。

**ここに置くもの：** 複数パッケージで使うテスト用Go関数（例：`NewTempGitRepo(t)`）
**ここに置かないもの：** テスト用のデータファイル、特定パッケージ専用のモック

---

### `docs/`

アプリケーション全体の永続的ドキュメント。基本設計・方針が変わらない限り更新しない。

---

### `.steering/`

特定の開発作業に対応する作業単位のドキュメント。作業完了後も履歴として保持する。新しい作業は必ず新しいディレクトリを作成する。

---

## 3. ルートファイル

| ファイル | 役割 |
|--------|------|
| `main.go` | エントリーポイント。`cmd.Execute()` を呼ぶだけ。ロジックは持たない |
| `go.mod` | モジュール定義。`toolchain go1.22.0` を明記する |
| `go.sum` | 依存ライブラリのチェックサム。自動生成のため手動編集しない |
| `.go-version` | `goenv` / `mise` 向けのGoバージョン固定ファイル |
| `.gitignore` | ビルド成果物・機密ファイルの除外設定（詳細は後述） |
| `.goreleaser.yaml` | リリースビルド・配布設定 |
| `.golangci.yml` | 有効化するリントルールの設定 |
| `Makefile` | `build`, `test`, `lint`, `fmt`, `bench`, `release` ターゲットを定義 |
| `CONTRIBUTING.md` | コントリビューションガイド。開発環境セットアップ・PR手順を記載 |
| `LICENSE` | MIT ライセンス |
| `README.md` | インストール方法・基本的な使い方。詳細は `CONTRIBUTING.md` へ誘導 |

---

## 4. `.gitignore` の管理対象

以下のファイル・ディレクトリをコミット対象から除外する。

```gitignore
# ビルド成果物
task
dist/

# テスト・カバレッジ
coverage.out
*.test

# エディタ・OS
.DS_Store
.idea/
.vscode/
*.swp

# 機密情報（誤コミット防止）
.env
*.pem
*.key
```

**リポジトリに置いてはいけないもの：**
- APIキー・パスワード・トークンなどの機密情報
- `~/.task/tasks.db`（ユーザーデータ）
- `task` バイナリ（`go build` で再生成できるため）
- `dist/`（`goreleaser` の出力ディレクトリ）

---

## 5. ファイル配置ルール

### ビジネスロジックの配置

| 種別 | 配置先 |
|------|--------|
| CLIのフラグ定義・ルーティング | `cmd/` |
| ドメインロジック（スコア計算・タスク操作） | `internal/` |
| 外部リソースへのアクセス（DB・Git・ファイル） | `internal/` の各サブパッケージ |
| 出力フォーマット | `internal/render/` |

`cmd/` にビジネスロジックを書いてはならない。`cmd/` のコードは `internal/` への薄いブリッジにとどめる。

### テストファイルの配置

| 種別 | 配置先 |
|------|--------|
| 単体テスト・結合テスト | テスト対象ファイルと同ディレクトリの `_test.go` |
| テスト用データファイル | 各パッケージの `testdata/` サブディレクトリ |
| 複数パッケージで使う共通Go関数 | `internal/testhelper/helper.go` |
| 特定パッケージ専用のモック | そのパッケージ内の `_test.go` に定義 |

### 新規 `internal/` パッケージを作る基準

既存パッケージに追加するのではなく新しいパッケージを作成する場合：

- 既存パッケージとは独立した外部リソース（新しい外部APIや形式）を扱う場合
- 既存パッケージが単一責任の原則を超えて肥大化する場合（目安：500行超）
- 他のパッケージとの循環インポートを避けるために分離が必要な場合

### 新規コマンドの追加手順

1. `cmd/<command>.go` を作成し cobra の `Command` を定義する
2. **そのファイルの `init()` 内** で `rootCmd.AddCommand(<command>Cmd)` を呼ぶ

   ```go
   // cmd/add.go
   func init() {
       rootCmd.AddCommand(addCmd)   // root.go ではなくここで登録する
       addCmd.Flags().StringP("due", "d", "", "締め切り日")
   }
   ```

3. 必要なビジネスロジックは `internal/` の適切なパッケージに追加する
4. `cmd/<command>.go` からは `internal/` を呼ぶだけにする

> `root.go` ではなく各コマンドファイルで `AddCommand` を呼ぶことで、
> コマンドが増えても `root.go` が肥大化しない。

### マイグレーションの追加手順

1. `internal/db/migrations.go` の `migrations` 配列の末尾にSQL文を追加する
2. **既存のSQL文は絶対に編集しない**（順序・内容を変えると本番DBが壊れる）
3. 適用済みのマイグレーションにバグがあった場合も、既存エントリを修正するのではなく、新しいマイグレーションとして修正SQLを追加する

   ```go
   var migrations = []string{
       // v1: 初期スキーマ
       `CREATE TABLE tasks ( ... );`,
       // v2: v1 で不足していたインデックスを追加（v1 を編集せずここに追記）
       `CREATE INDEX idx_tasks_status ON tasks(status);`,
   }
   ```
