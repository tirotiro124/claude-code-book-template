# 初回実装 設計

## 1. 実装アプローチ

### 基本方針

「動くものを最短で作り、テストで固める」順番で進める。
依存関係のない内側のパッケージ（`scoring`）から実装し、外側（`cmd`）へと積み上げる。

```
実装順序:
1. DB・マイグレーション（データの土台）
2. スコアエンジン（副作用なし・最初にテストを書ける）
3. タスクリポジトリ（CRUDの土台）
4. Gitコンテキスト（外部依存の抽象化）
5. コンフィグローダー（設定の読み込み）
6. コアコマンド（add / done / list / next / show）
7. 追加コマンド（snooze / pin / block / edit / sync / export）
8. 出力レンダラーの整備（カラー・JSON・テーブル）
```

---

## 2. コンポーネントごとの実装設計

### 2-1. DB・マイグレーション（`internal/db/`）

**`db.go`：**

```go
type DB struct {
    *sql.DB
}

func Open() (*DB, error)
// 1. os.MkdirAll("~/.task/", 0700)
// 2. sql.Open("sqlite", "~/.task/tasks.db")
// 3. os.Chmod(dbPath, 0600)
// 4. PRAGMA journal_mode=WAL; foreign_keys=ON;
// 5. db.migrate()
```

**`migrations.go`：**

```go
var migrations = []string{
    // v1
    `CREATE TABLE tasks (
        id           INTEGER PRIMARY KEY AUTOINCREMENT,
        title        TEXT NOT NULL,
        status       TEXT NOT NULL DEFAULT 'todo'
                         CHECK(status IN ('todo','in_progress','done','snoozed')),
        due_date     TEXT,
        branch       TEXT,
        tags         TEXT NOT NULL DEFAULT '[]',
        pinned       INTEGER NOT NULL DEFAULT 0 CHECK(pinned IN (0,1)),
        snoozed_until TEXT,
        source_file  TEXT,
        source_line  INTEGER,
        created_at   TEXT NOT NULL,
        updated_at   TEXT NOT NULL
    )`,
    `CREATE TABLE task_blocks (
        task_id       INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
        blocked_by_id INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
        PRIMARY KEY (task_id, blocked_by_id)
    )`,
    `CREATE TABLE display_history (
        id           INTEGER PRIMARY KEY AUTOINCREMENT,
        task_id      INTEGER NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
        displayed_at TEXT NOT NULL
    )`,
    `CREATE INDEX idx_tasks_status   ON tasks(status)`,
    `CREATE INDEX idx_tasks_branch   ON tasks(branch)`,
    `CREATE INDEX idx_tasks_due_date ON tasks(due_date)`,
    `CREATE INDEX idx_tasks_pinned   ON tasks(pinned)`,
}
```

---

### 2-2. スコアエンジン（`internal/scoring/`）

外部依存を持たない純粋関数として実装する。最初にテストを書いてから実装する。

**インターフェース：**

```go
type Input struct {
    Task          db.Task
    CurrentBranch string
    Now           time.Time
    RecentShown   []int64  // display_history の直近2件のタスクID
    Weights       WeightMap
}

type ScoreResult struct {
    Total      int
    Urgency    int
    Importance int
    Context    int
    Aging      int
    Fatigue    int
}

func Calculate(input Input) ScoreResult
```

**各シグナルの計算：**

```go
// 緊急度（締め切りまでの残日数）
func calcUrgency(dueDate *time.Time, now time.Time) int {
    // due_date が nil → 0点
    // 残日数 >= 8 → 0点
    // 残日数 4-7  → 15点
    // 残日数 1-3  → 30点
    // 残日数 0    → 40点
    // 残日数 < 0  → 40 + abs(残日数) * 2
}

// 重要度（被依存タスク数）
func calcImportance(blockedByCount int) int {
    // 0件 → 0点 / 1件 → 10点 / 2件 → 20点 / 3件以上 → 30点
}

// コンテキスト加点（ブランチ一致）
func calcContext(taskBranch, currentBranch string) int {
    // 一致かつ空文字でない → 15点 / それ以外 → 0点
    // 上限 20点（将来の拡張余地）
}

// 経年加点（未着手日数）
func calcAging(createdAt time.Time, now time.Time) int {
    // 3日未満 → 0点 / 3-6日 → 5点 / 7-13日 → 10点 / 14日以上 → 15点
}

// 疲労ペナルティ
func calcFatigue(taskID int64, recentShown []int64) int {
    // recentShown[0] と一致 → 10点 / recentShown[1] と一致 → 5点 / それ以外 → 0点
}
```

**重み係数の適用：**

```go
func Apply(base ScoreResult, weights WeightMap) ScoreResult {
    return ScoreResult{
        Total:      int(float64(base.Urgency)*weights["urgency"]) +
                    int(float64(base.Importance)*weights["importance"]) +
                    int(float64(base.Context)*weights["context"]) +
                    int(float64(base.Aging)*weights["aging"]) -
                    int(float64(base.Fatigue)*weights["fatigue"]),
        ...
    }
}
```

---

### 2-3. タスクリポジトリ（`internal/db/task.go`）

```go
// 主要メソッド
func (db *DB) CreateTask(t Task) (Task, error)
func (db *DB) GetTask(id int64) (Task, error)
func (db *DB) ListTasks(filter TaskFilter) ([]Task, error)
func (db *DB) UpdateTask(t Task) error
func (db *DB) DeleteTask(id int64) error

// スヌーズ再アクティブ化（list/next 実行時に呼ぶ）
func (db *DB) ReactivateSnoozed(now time.Time) error

// 依存関係
func (db *DB) AddBlock(taskID, blockedByID int64) error
func (db *DB) RemoveBlock(taskID, blockedByID int64) error
func (db *DB) RemoveAllBlocksBy(blockedByID int64) error  // done 時に呼ぶ

// 表示履歴
func (db *DB) RecordShown(taskID int64) error   // 直近2件を超えたら古いものを削除
func (db *DB) GetRecentShown() ([]int64, error)
```

**`TaskFilter`：**

```go
type TaskFilter struct {
    Status  []Status  // 空の場合は todo / in_progress のみ
    Branch  string
    Tag     string
    Pinned  *bool
}
```

---

### 2-4. Gitコンテキスト（`internal/git/context.go`）

```go
type Provider interface {
    CurrentBranch() (string, error)
    RootDir() (string, error)
}

type ExecProvider struct{}

func (p ExecProvider) CurrentBranch() (string, error) {
    // exec.Command("git", "branch", "--show-current")
    // エラー時（gitリポジトリ外など）は ("", ErrNotGitRepository)
}

func (p ExecProvider) RootDir() (string, error) {
    // exec.Command("git", "rev-parse", "--show-toplevel")
}
```

---

### 2-5. コンフィグローダー（`internal/config/config.go`）

```go
type Config struct {
    Preset   Preset
    Weights  WeightMap
    Timezone *time.Location
    NoColor  bool
}

var presetWeights = map[Preset]WeightMap{
    PresetDeadline: {"urgency": 2.0, "importance": 1.0, "context": 0.5, "aging": 1.0, "fatigue": 1.0},
    PresetBalanced: {"urgency": 1.0, "importance": 1.0, "context": 1.0, "aging": 1.0, "fatigue": 1.0},
    PresetFlow:     {"urgency": 0.5, "importance": 1.0, "context": 2.0, "aging": 1.0, "fatigue": 1.0},
}

func Load() (Config, error)
// 1. カレントから上位へ .taskrc を探索
// 2. 見つからなければ ~/.taskrc を確認
// 3. どちらもなければデフォルト値（PresetBalanced）を返す
// 4. NO_COLOR 環境変数を確認して NoColor フラグに反映
```

---

### 2-6. コアコマンドの実装設計

#### `task add`

```
1. cobra のフラグを解析（--due, --branch, --tag）
2. --branch 省略時: ExecProvider.CurrentBranch() で取得。エラーは無視して空文字
3. Task を構築して db.CreateTask()
4. 完了メッセージを表示: "Added #43: タイトル  [branch: feat/login]"
```

#### `task next`

```
1. config.Load()
2. ExecProvider.CurrentBranch()
3. db.ReactivateSnoozed(now)
4. db.ListTasks(filter{Status: [todo, in_progress]})
5. db.GetRecentShown()
6. 各タスクの blockedByCount を取得
7. scoring.Calculate() で全タスクのスコアを計算
8. ピン固定タスクを先頭、残りをスコア降順でソート
9. 1件目を render.NextTask() で表示
10. db.RecordShown(taskID)
```

#### `task done`

```
1. db.GetTask(id) で存在確認
2. db.UpdateTask(status=done)
3. db.RemoveAllBlocksBy(id)  // このタスクを待っていた依存を解除
4. 完了メッセージを表示
5. task next と同じフローで次のタスクを表示
```

#### `task list`

```
task next の 1〜8 と同じ
9. 全タスクを render.TaskList() で表示（テーブル形式）
```

#### `task show`

```
1. db.GetTask(id)
2. scoring.Calculate() でスコアと内訳を計算
3. render.TaskDetail() でスコア内訳つき詳細を表示
```

---

### 2-7. 出力レンダラー（`internal/render/`）

```go
type Renderer struct {
    JSON    bool
    NoColor bool
    Stderr  io.Writer
    Stdout  io.Writer
}

// internal/render/task.go
func (r Renderer) NextTask(task db.Task, score scoring.ScoreResult)
func (r Renderer) TaskList(tasks []db.Task, scores []scoring.ScoreResult)
func (r Renderer) TaskDetail(task db.Task, score scoring.ScoreResult, blocks []db.Task)

// internal/render/message.go
func (r Renderer) Added(task db.Task)
func (r Renderer) Done(task db.Task)
func (r Renderer) Error(err error)
func (r Renderer) NoTasks()
```

---

## 3. データ構造

### `db.Task` 構造体

```go
type Task struct {
    ID          int64
    Title       string
    Status      Status
    DueDate     *time.Time   // nil = 締め切りなし
    Branch      string
    Tags        []string     // JSON文字列から変換済み
    Pinned      bool
    SnoozedUntil *time.Time  // nil = スヌーズなし
    SourceFile  string
    SourceLine  int
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

---

## 4. 影響範囲の分析

### パッケージ依存関係

```
cmd/* → internal/config
cmd/* → internal/db
cmd/* → internal/scoring
cmd/* → internal/render
cmd/* → internal/git

internal/scanner → internal/git  （gitルート取得）
internal/db      → （依存なし）
internal/scoring → （依存なし）
internal/config  → （依存なし）
internal/render  → internal/scoring, internal/db
internal/git     → （依存なし）
```

循環インポートが発生しないことを確認済み。

### テスト戦略と影響範囲

| パッケージ | テスト手法 | 外部依存 |
|-----------|-----------|---------|
| `scoring` | ユニットテスト（純粋関数） | なし |
| `config` | ユニットテスト（一時ファイル） | なし |
| `db` | 結合テスト（`:memory:` SQLite） | なし |
| `git` | 結合テスト（`testhelper.NewTempGitRepo`） | git バイナリ |
| `scanner` | 結合テスト（`t.TempDir()` + 仮ファイル） | git バイナリ |
| `render` | ユニットテスト（stdout キャプチャ） | なし |
| `cmd` | E2Eテスト（`cobra.ExecuteC()`） | git バイナリ・SQLite |
