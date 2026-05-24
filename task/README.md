# task

Gitコンテキストに応じてタスクの優先順位を自動調整するCLIタスク管理ツール。

## 特徴

- **スコアリング**: 締め切り・依存関係・Gitブランチ・経過日数・表示疲労の5シグナルで自動スコア計算
- **Git連携**: カレントブランチに関連するタスクを自動的に優先表示
- **TODO同期**: ソースコードの `TODO:` コメントをタスクとして取り込み
- **カラー出力**: 期限に応じたカラー表示（`NO_COLOR=1` で無効化）
- **JSON出力**: `--json` フラグでスクリプト連携が可能

## インストール

```bash
# Go 1.22+ が必要
go install github.com/task-cli/task@latest
```

またはリリースページからバイナリをダウンロードしてください。

## クイックスタート

```bash
# タスクを追加
task add "ログイン機能を実装する" --due 2026-06-01 --tag auth

# 次にやるべきタスクを確認
task next

# タスク一覧を表示
task list

# タスクを完了にする
task done 1

# TODO コメントを同期
task sync
```

## コマンド一覧

| コマンド | 説明 |
|---------|------|
| `task add <title>` | タスクを追加する |
| `task next` | 次に取り組むべきタスクを表示する |
| `task list` | タスク一覧を表示する |
| `task show <id>` | タスクの詳細を表示する |
| `task done <id>` | タスクを完了にする |
| `task snooze <id> <duration>` | タスクをスヌーズする (`1d` / `2h` / `1w`) |
| `task pin <id>` | タスクをピン固定する |
| `task unpin <id>` | ピン固定を解除する |
| `task block <id> <dep-id>` | 依存関係を追加する |
| `task unblock <id> <dep-id>` | 依存関係を解除する |
| `task edit <id>` | `$EDITOR` でタスクを編集する |
| `task sync` | TODO コメントをタスクに同期する |
| `task export` | タスクをエクスポートする (`--format json\|md`) |

## スコア計算

```
score = urgency(0-40) + importance(0-30) + context(0-20) + aging(0-15) - fatigue(0-10)
```

| シグナル | 説明 |
|---------|------|
| urgency | 締め切りまでの残日数が少ないほど高スコア |
| importance | このタスクを待っているタスク数が多いほど高スコア |
| context | カレントブランチと一致するタスクに加点 |
| aging | 長期間未着手のタスクに加点 |
| fatigue | 直前に表示されたタスクを減点（同じタスクを繰り返し推薦しない） |

### プリセット

`.taskrc` で重み係数を設定できます:

```toml
preset = "balanced"  # balanced | deadline | flow
```

- `balanced`: すべてのシグナルを均等に重視（デフォルト）
- `deadline`: 締め切りを2倍重視
- `flow`: コンテキスト（ブランチ一致）を2倍重視

## 設定

`~/.taskrc` またはプロジェクトルートの `.taskrc` に設定を記述します:

```toml
preset = "flow"
```

## データ

タスクは `~/.task/tasks.db`（SQLite）に保存されます。パーミッションは `0600` に設定されます。

## 開発

```bash
# テスト実行
CGO_ENABLED=0 go test ./...

# ビルド
CGO_ENABLED=0 go build -o task .

# リント
golangci-lint run
```
