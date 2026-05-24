# 初回実装 タスクリスト

## 進捗凡例

- `[ ]` 未着手
- `[x]` 完了

---

## フェーズ1：プロジェクトセットアップ

- [x] 1-1. `go mod init github.com/task-cli/task` でモジュールを初期化する
- [x] 1-2. `docs/repository-structure.md` に従いディレクトリ構造を作成する
- [x] 1-3. 依存ライブラリを `go.mod` に追加する
  - `github.com/spf13/cobra`
  - `modernc.org/sqlite`
  - `github.com/BurntSushi/toml`
  - `github.com/fatih/color`
  - `github.com/jedib0t/go-pretty`
  - `github.com/sabhiram/go-gitignore`
  - `github.com/stretchr/testify`
- [x] 1-4. `Makefile` を作成する（`build` / `test` / `lint` / `fmt` / `bench` / `release`）
- [x] 1-5. `.go-version` を作成する（`1.22.3`）
- [x] 1-6. `.gitignore` を作成する
- [x] 1-7. `.golangci.yml` を作成する（`errcheck` / `govet` / `staticcheck` / `gofmt` / `gosec`）
- [x] 1-8. `main.go` を作成する（`cmd.Execute()` を呼ぶだけ）
- [x] 1-9. `cmd/root.go` を作成する（`rootCmd` / `Execute()` / グローバルフラグ定義）

**完了条件：** `make build` でバイナリが生成される

---

## フェーズ2：DB・マイグレーション

- [x] 2-1. `internal/db/migrations.go` を作成する（v1 のSQL定義：tasks / task_blocks / display_history / インデックス）
- [x] 2-2. `internal/db/db.go` を実装する
  - `~/.task/` ディレクトリの自動作成（`os.MkdirAll` / パーミッション `0700`）
  - `tasks.db` の作成とパーミッション `600` の設定
  - `PRAGMA journal_mode=WAL; foreign_keys=ON;` の設定
  - `schema_versions` テーブルのブートストラップ作成
  - マイグレーション実行ロジック
- [x] 2-3. `internal/db/db_test.go` を作成する（`:memory:` SQLite でマイグレーションのテスト）

**完了条件：** `make test ./internal/db/...` が通る

---

## フェーズ3：スコアエンジン（テストファースト）

- [x] 3-1. `internal/scoring/engine_test.go` を作成する（テーブルドリブンテスト）
  - 緊急度の各閾値（8日以上・4-7日・1-3日・当日・超過）
  - 重要度の被依存数（0〜3件以上）
  - コンテキスト加点（ブランチ一致・不一致・空文字）
  - 経年加点（3日未満・3-6日・7-13日・14日以上）
  - 疲労ペナルティ（直前・2回前・それ以外）
  - プリセット重み係数の適用（deadline / balanced / flow）
- [x] 3-2. `internal/scoring/engine.go` を実装する（テストが通るまで）
  - `Input` / `ScoreResult` / `WeightMap` 型の定義
  - `Calculate(input Input) ScoreResult` の実装
  - 各シグナル計算関数の実装

**完了条件：** `make test ./internal/scoring/...` がカバレッジ 100% で通る

---

## フェーズ4：タスクリポジトリ

- [x] 4-1. `internal/db/task.go` に `Task` 構造体と `Status` 型を定義する
- [x] 4-2. `CreateTask` / `GetTask` / `ListTasks` / `UpdateTask` を実装する
- [x] 4-3. `ReactivateSnoozed` を実装する（スヌーズ期限切れタスクを `todo` に戻す）
- [x] 4-4. `AddBlock` / `RemoveBlock` / `RemoveAllBlocksBy` を実装する
- [x] 4-5. `RecordShown` / `GetRecentShown` を実装する（直近2件を維持）
- [x] 4-6. `internal/db/task_test.go` を作成する（`:memory:` SQLite で各操作をテスト）

**完了条件：** `make test ./internal/db/...` がカバレッジ 80% 以上で通る

---

## フェーズ5：Gitコンテキスト・コンフィグ・テストヘルパー

- [x] 5-1. `internal/testhelper/helper.go` を実装する（`NewTempGitRepo(t *testing.T) string`）
- [x] 5-2. `internal/git/context.go` を実装する（`Provider` インターフェース / `ExecProvider`）
- [x] 5-3. `internal/git/context_test.go` を作成する（`testhelper.NewTempGitRepo` を使った結合テスト）
- [x] 5-4. `internal/config/config.go` を実装する
  - `Config` 構造体 / `Preset` 型 / プリセット重み係数マップの定義
  - `.taskrc` の探索ロジック（カレントから上位へ / `~/.taskrc` へのフォールバック）
  - `NO_COLOR` 環境変数の読み込み
- [x] 5-5. `internal/config/config_test.go` を作成する（一時ファイルを使った探索・パーステスト）

**完了条件：** `make test ./internal/git/... ./internal/config/...` が通る

---

## フェーズ6：出力レンダラー

- [x] 6-1. `internal/render/renderer.go` を実装する（`Renderer` 構造体 / JSON切り替え / カラー切り替え）
- [x] 6-2. `internal/render/task.go` を実装する（`NextTask` / `TaskList` / `TaskDetail`）
- [x] 6-3. `internal/render/message.go` を実装する（`Added` / `Done` / `Error` / `NoTasks`）
- [x] 6-4. `internal/render/renderer_test.go` を作成する（stdout キャプチャによる出力検証）

**完了条件：** `make test ./internal/render/...` が通る

---

## フェーズ7：コアコマンド実装

- [x] 7-1. `cmd/add.go` を実装する（`task add`）
- [x] 7-2. `cmd/list.go` を実装する（`task list`）
- [x] 7-3. `cmd/next.go` を実装する（`task next`）
- [x] 7-4. `cmd/show.go` を実装する（`task show`）
- [x] 7-5. `cmd/done.go` を実装する（`task done`）
- [x] 7-6. コアコマンドの E2E テストを作成する（`cobra.ExecuteC()` を使用）

**完了条件：** `task add` `task list` `task next` `task show` `task done` が正常動作する

---

## フェーズ8：追加コマンド実装

- [x] 8-1. `cmd/snooze.go` を実装する（`task snooze <id> <duration>`）
- [x] 8-2. `cmd/pin.go` を実装する（`task pin` / `task unpin`）
- [x] 8-3. `cmd/block.go` を実装する（`task block` / `task unblock`）
- [x] 8-4. `cmd/edit.go` を実装する（`$EDITOR` でTOML編集）
- [x] 8-5. `internal/scanner/todo.go` を実装する（`.gitignore` 尊重・`TODO:` 抽出・重複チェック）
- [x] 8-6. `internal/scanner/todo_test.go` を作成する
- [x] 8-7. `cmd/sync.go` を実装する（`task sync` / `task sync --dry-run`）
- [x] 8-8. `cmd/export.go` を実装する（`task export --format json|md`）
- [x] 8-9. 追加コマンドの E2E テストを作成する

**完了条件：** 全コマンドが正常動作する

---

## フェーズ9：品質チェック・仕上げ

- [x] 9-1. `make test` でカバレッジ目標を確認する（scoring: 100% / その他: 80%以上）
- [ ] 9-2. `make lint` を通す（CI環境で実施）
- [x] 9-3. `make bench` を実行してベンチマーク基準値を確認する（scoring: 106 ns/op / 0 allocs）
- [x] 9-4. 全コマンドの応答時間が 200ms 以内であることを確認する
- [x] 9-5. `NO_COLOR=1` / `--json` / `--verbose` フラグの動作を確認する
- [x] 9-6. gitリポジトリ外での `task add` 動作を確認する（ブランチなしで正常動作）
- [x] 9-7. `~/.task/tasks.db` のパーミッションが `600` であることを確認する
- [x] 9-8. `.github/workflows/ci.yml` を作成する（lint + test + build）
- [x] 9-9. `README.md` を作成する（インストール方法・基本的な使い方）
- [x] 9-10. `CONTRIBUTING.md` を作成する（開発環境セットアップ・PR手順）

**完了条件：** `requirements.md` の受け入れ条件がすべて満たされる

---

## 完了条件サマリー

| フェーズ | 完了条件 |
|---------|---------|
| 1 | `make build` が通る |
| 2 | DB マイグレーションテストが通る |
| 3 | スコアエンジンがカバレッジ 100% で通る |
| 4 | タスクリポジトリテストがカバレッジ 80% 以上で通る |
| 5 | Git / Config テストが通る |
| 6 | レンダラーテストが通る |
| 7 | コアコマンドが正常動作する |
| 8 | 全コマンドが正常動作する |
| 9 | `requirements.md` の受け入れ条件がすべて満たされる |
