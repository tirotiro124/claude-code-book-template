# Contributing

## 開発環境セットアップ

```bash
# Go 1.22+ をインストール
go version  # go1.22.x 以上であることを確認

# リポジトリをクローン
git clone https://github.com/task-cli/task.git
cd task

# 依存関係をインストール
CGO_ENABLED=0 go mod download

# テストを実行して開発環境を確認
CGO_ENABLED=0 go test ./...
```

## ビルド

```bash
CGO_ENABLED=0 go build -o task .
./task --help
```

## テスト

```bash
# 全テスト実行
CGO_ENABLED=0 go test ./...

# カバレッジ付き
CGO_ENABLED=0 go test ./... -cover

# 特定パッケージのみ
CGO_ENABLED=0 go test ./internal/scoring/... -v

# ベンチマーク
CGO_ENABLED=0 go test ./internal/scoring/... -bench=. -benchmem -run=^$
```

カバレッジ目標: `scoring` 100% / その他 80%+

## リント

```bash
golangci-lint run
```

設定は `.golangci.yml` を参照してください。

## ディレクトリ構成

```
task/
├── main.go                     # エントリポイント
├── cmd/                        # CLI コマンド定義（cobra）
├── internal/
│   ├── db/                     # SQLite データ層
│   ├── scoring/                # スコアエンジン（純粋関数）
│   ├── config/                 # 設定ローダー
│   ├── git/                    # Git コンテキスト取得
│   ├── render/                 # ターミナル出力
│   ├── scanner/                # TODO コメントスキャナー
│   └── testhelper/             # テストユーティリティ
└── .github/workflows/          # CI/CD
```

## コーディング規約

- エラーは `fmt.Errorf("context: %w", err)` でラップする
- コメントは非自明な場合のみ記述する
- パッケージ外から使うインターフェースは consumer 側に定義する
- `internal/` パッケージ間の依存は `cmd/ → internal/*` の方向のみ

## PR 手順

1. `main` ブランチから feature ブランチを切る
2. 変更を実装してテストを追加する
3. `CGO_ENABLED=0 go test ./...` が通ることを確認する
4. `golangci-lint run` が通ることを確認する
5. Squash Merge でマージする

コミットメッセージは [Conventional Commits](https://www.conventionalcommits.org/) に従います:

```
feat: task snooze コマンドを追加する
fix: スヌーズ期限切れの判定ロジックを修正する
test: scoring パッケージのカバレッジを100%にする
docs: README にインストール手順を追加する
```
