# 技術仕様書

## 1. テクノロジースタック

### 言語：Go

他の候補と比較した上でGoを選定する。

| 言語 | 単一バイナリ | 起動速度 | クロスコンパイル | 採用 |
|------|------------|---------|----------------|------|
| **Go** | ✓ | ✓ | ✓（CGO不使用時） | **採用** |
| Rust | ✓ | ✓ | △（ツールチェイン設定が複雑） | 不採用 |
| TypeScript | ✗（Node.js必要） | △ | ✓ | 不採用 |
| Python | ✗（インタープリタ必要） | △ | ✓ | 不採用 |

RustはGoと同等の性能要件を満たせるが、学習コストと開発速度の観点でGoを優先する。

**バージョン：** Go 1.22以上

**バージョン管理：** リポジトリルートに `.go-version` を配置し、`goenv` / `mise` で開発者間のバージョンを統一する。`go.mod` にも `toolchain go1.22.0` を明記する。

---

### 主要依存ライブラリ

| ライブラリ | 用途 | 選定理由 |
|-----------|------|---------|
| `github.com/spf13/cobra` | CLIフレームワーク | サブコマンド・フラグ管理が充実。`--help` 自動生成 |
| `modernc.org/sqlite` | SQLiteドライバ | CGO不要の純粋Go実装。クロスコンパイル時にCコンパイラが不要 |
| `github.com/BurntSushi/toml` | TOMLパーサ | `.taskrc` の読み込み |
| `github.com/fatih/color` | カラー出力 | `NO_COLOR` 環境変数を自動で尊重する |
| `github.com/jedib0t/go-pretty` | テーブル出力 | `task list` のカラム整形 |
| `github.com/sabhiram/go-gitignore` | `.gitignore` パース | `task sync` でのファイル除外判定。ワイルドカード・ネガティブパターンに対応 |

### 開発・テスト用

| ライブラリ | 用途 |
|-----------|------|
| `github.com/stretchr/testify` | アサーションヘルパー |
| `github.com/goreleaser/goreleaser` | バイナリビルド・配布自動化 |

---

## 2. ディレクトリ構成

```
task/
├── cmd/
│   ├── root.go          # CLIエントリーポイント・グローバルフラグ
│   ├── add.go
│   ├── list.go
│   ├── next.go
│   ├── done.go
│   ├── show.go
│   ├── edit.go
│   ├── snooze.go
│   ├── pin.go
│   ├── block.go
│   ├── sync.go
│   └── export.go
├── internal/
│   ├── config/
│   │   ├── config.go        # コンフィグローダー
│   │   └── config_test.go
│   ├── db/
│   │   ├── db.go            # DB接続・初期化・マイグレーション実行
│   │   ├── db_test.go
│   │   ├── migrations.go    # マイグレーションSQL定義（配列）
│   │   └── task.go          # タスクリポジトリ（CRUD）
│   ├── scoring/
│   │   ├── engine.go        # スコアエンジン（副作用なし）
│   │   └── engine_test.go
│   ├── git/
│   │   ├── context.go       # Gitコンテキストプロバイダー（インターフェース定義含む）
│   │   └── context_test.go
│   ├── scanner/
│   │   ├── todo.go          # TODOスキャナー
│   │   └── todo_test.go
│   ├── render/
│   │   ├── renderer.go      # 出力レンダラー
│   │   └── renderer_test.go
│   └── testhelper/
│       └── helper.go        # テスト共通ユーティリティ（一時gitリポジトリ作成など）
├── main.go
├── go.mod                   # toolchain go1.22.0 を明記
├── go.sum
├── .go-version              # 1.22.x
├── .goreleaser.yaml
├── .golangci.yml
└── Makefile
```

テストファイル（`_test.go`）はソースファイルと同じディレクトリに配置する（Goの標準慣習）。
`internal/testhelper/` は複数パッケージから参照する共通テストユーティリティを置く。

---

## 3. データベース設計方針

### マイグレーション

`schema_versions` テーブルはマイグレーション管理の基盤であるため、通常のマイグレーション実行前にブートストラップ処理として作成する。

```go
// internal/db/db.go

func (db *DB) migrate() error {
    // Step 1: schema_versions テーブルのみブートストラップとして作成
    // （マイグレーション番号管理の前提となるテーブルのため別扱い）
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS schema_versions (
            version    INTEGER PRIMARY KEY,
            applied_at TEXT NOT NULL
        )
    `)
    if err != nil {
        return err
    }

    // Step 2: 現在の適用済みバージョンを取得
    var current int
    row := db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_versions`)
    if err := row.Scan(&current); err != nil {
        return err
    }

    // Step 3: 未適用のマイグレーションを順に実行
    for i, sql := range migrations {
        version := i + 1
        if version <= current {
            continue
        }
        if _, err := db.Exec(sql); err != nil {
            return fmt.Errorf("migration v%d failed: %w", version, err)
        }
        db.Exec(
            `INSERT INTO schema_versions (version, applied_at) VALUES (?, ?)`,
            version, time.Now().UTC().Format(time.RFC3339),
        )
    }
    return nil
}
```

```go
// internal/db/migrations.go
// schema_versions テーブル自体はここに含めない（ブートストラップ処理で作成する）

var migrations = []string{
    // v1: 初期スキーマ
    `CREATE TABLE tasks ( ... );`,
    `CREATE TABLE task_blocks ( ... );`,
    `CREATE TABLE display_history ( ... );`,
    // v2以降はここに追加（インデックスなどの変更も含む）
}
```

### 初回起動時の初期化

```go
// internal/db/db.go

func Open() (*DB, error) {
    dir := filepath.Join(os.UserHomeDir(), ".task")

    // ~/.task/ ディレクトリが存在しない場合は自動作成
    if err := os.MkdirAll(dir, 0700); err != nil {
        return nil, fmt.Errorf("~/.task ディレクトリの作成に失敗しました: %w", err)
    }

    dbPath := filepath.Join(dir, "tasks.db")
    db, err := sql.Open("sqlite", dbPath)
    if err != nil {
        return nil, err
    }

    // 新規作成の場合はパーミッションを 600 に設定
    os.Chmod(dbPath, 0600)

    db.Exec(`PRAGMA journal_mode=WAL;`)
    db.Exec(`PRAGMA foreign_keys=ON;`)

    return &DB{db}, db.migrate()
}
```

### ファイルロック

SQLiteのWALモードを有効化し、並行読み取りのパフォーマンスを確保する。書き込み操作はSQLiteの排他ロック機構に委ねる。

### バックアップ・リカバリ

- データファイルは `~/.task/tasks.db` の単一ファイルのため、ユーザーが任意のタイミングでコピーによるバックアップが可能
- アプリ側でのバックアップ自動化はv1スコープ外とし、`task export --format json > backup.json` によるデータ退避を代替手段として案内する
- DBファイルが破損した場合は `~/.task/tasks.db` を削除して再起動することでDBを再作成できる（データは失われる）ことをドキュメントに明記する

### データ保存先

```
~/.task/
└── tasks.db    # 全データ（タスク・履歴・スキーマバージョン）
```

---

## 4. 開発ツールと手法

### Makefile

```makefile
.PHONY: build test lint fmt release

build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o task ./main.go

test:
	go test -race -cover ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .
	goimports -w .

bench:
	go test -bench=. -benchmem ./...

release:
	goreleaser release --snapshot --clean
```

### ビルドとバージョン埋め込み

```bash
go build -ldflags="-s -w -X main.version=v1.0.0" -o task ./main.go
```

`main.version` 変数をビルド時に注入することで `task --version` が正確なバージョンを返せるようにする。リリースビルドでは `goreleaser` がタグ名を自動で注入する。

### テスト

```bash
make test    # go test -race -cover ./...
make bench   # go test -bench=. -benchmem ./...
```

**カバレッジ目標：** `internal/scoring/` は100%、その他パッケージは80%以上を目標とする。

**テスト方針：**

| 対象 | 方針 |
|------|------|
| スコアエンジン | 副作用なしのため純粋なユニットテスト。入力値と期待スコアのテーブルドリブンテスト |
| タスクリポジトリ | インメモリSQLite（`:memory:`）を使用した結合テスト |
| Gitコンテキスト | `testhelper.NewTempGitRepo()` で `t.TempDir()` 上に一時gitリポジトリを作成して実行 |
| CLIコマンド | `cobra` の `ExecuteC()` を使ったコマンドレベルの結合テスト |

**gitコマンドのモック方針：**

`git/context.go` にインターフェースを定義し、本番実装とテスト用モックを差し替えられる設計にする。

```go
// internal/git/context.go
type Provider interface {
    CurrentBranch() (string, error)
    RootDir() (string, error)
}

// 本番実装：git バイナリを外部プロセスとして呼ぶ
type ExecProvider struct{}

// テスト用モック
type MockProvider struct {
    Branch string
    Root   string
}
```

### リント

```bash
make lint    # golangci-lint run
```

`.golangci.yml` で有効にするルール：`errcheck`, `govet`, `staticcheck`, `gofmt`, `gosec`

### フォーマット

```bash
make fmt    # gofmt + goimports
```

### CI/CD（GitHub Actions）

```
.github/workflows/
├── ci.yml       # PR・pushごとにlint + test + build を実行
└── release.yml  # タグプッシュ時に goreleaser でリリース
```

**ci.yml の主要ステップ：**

```
1. golangci-lint run
2. go test -race -cover ./...
3. カバレッジが目標値を下回った場合にワーニング
4. go build（ビルドが通ることの確認）
```

**release.yml の主要ステップ：**

```
1. goreleaser release（macOS/Linux向けバイナリ生成）
2. チェックサム（SHA256）ファイルの自動生成
3. GitHub Releases への公開
4. Homebrew Tap リポジトリの自動更新
```

---

## 5. 配布方法

**リポジトリ名：** `github.com/task-cli/task`（仮。リリース前に確定する）

### macOS（Homebrew）

```bash
brew install task-cli/tap/task
```

`goreleaser` が生成するバイナリを `task-cli/homebrew-tap` リポジトリに自動登録する。

**コード署名：** macOS Gatekeeperによるブロックを防ぐため、`goreleaser` の `notarize` 設定でApple Developer IDによる署名・公証（Notarization）を行う。

### Linux（直接ダウンロード）

GitHub Releases からバイナリとチェックサムをダウンロードして検証する。

```bash
# バイナリのダウンロード
curl -L https://github.com/task-cli/task/releases/latest/download/task-linux-amd64 \
  -o /usr/local/bin/task

# チェックサム検証
curl -L https://github.com/task-cli/task/releases/latest/download/checksums.txt \
  | sha256sum --check --ignore-missing

chmod +x /usr/local/bin/task
```

### インストールスクリプト（v2以降で対応）

プラットフォーム自動判定・チェックサム検証・パス設定をまとめた `install.sh` を提供する。v1ではHomebrew / 直接ダウンロードのみとし、ユーザーが一定数集まった段階で優先度を判断する。

---

## 6. 技術的制約と要件

| 制約 | 内容 |
|------|------|
| 外部サービス依存なし | 全機能がオフラインで動作すること |
| CGO不使用 | `modernc.org/sqlite` を採用することでクロスコンパイル時にCコンパイラ不要 |
| gitバイナリへの依存 | gitコマンドが存在しない環境ではGit連携機能を無効化し、他機能は動作すること |
| ファイルパーミッション | `~/.task/tasks.db` 作成時にパーミッションを `600` に設定すること |
| macOSコード署名 | 配布バイナリはApple Developer IDで署名・公証すること |

---

## 7. パフォーマンス要件と対策

| 要件 | 対策 |
|------|------|
| 全コマンド200ms以内 | SQLiteのWALモード有効化、インデックス設定、不要なファイルIOの排除 |
| `task list` のスコア計算 | タスク件数が増えてもO(n)で完了する設計。500件を超えたらページネーションを導入する（`--limit` オプション） |
| バイナリサイズ | `go build -ldflags="-s -w"` でデバッグシンボルを削除し10MB以下を目標とする |
| 起動オーバーヘッド | グローバル初期化処理を最小限に保ち、DB接続は実行コマンドが必要とするときのみ行う |

**ベンチマーク：**

`go test -bench` によるベンチマークを `internal/scoring/` と `internal/db/` に実装し、CI上でパフォーマンス劣化を早期検知する。基準値（ns/op）はv1リリース時に計測した結果を `BENCHMARKS.md` に記録し、以降のPRで比較する。
