package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := OpenMemory()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCreateAndGetTask(t *testing.T) {
	db := newTestDB(t)
	due := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	created, err := db.CreateTask(Task{
		Title:   "ログイン実装",
		Branch:  "feat/login",
		Tags:    []string{"auth"},
		DueDate: &due,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), created.ID)
	assert.Equal(t, "ログイン実装", created.Title)
	assert.Equal(t, StatusTodo, created.Status)
	assert.Equal(t, "feat/login", created.Branch)
	assert.Equal(t, []string{"auth"}, created.Tags)
	assert.Equal(t, "2026-06-01", created.DueDate.Format("2006-01-02"))

	fetched, err := db.GetTask(1)
	require.NoError(t, err)
	assert.Equal(t, created.Title, fetched.Title)
}

func TestGetTask_NotFound(t *testing.T) {
	db := newTestDB(t)
	_, err := db.GetTask(999)
	assert.ErrorIs(t, err, ErrTaskNotFound)
}

func TestListTasks_DefaultFilter(t *testing.T) {
	db := newTestDB(t)

	_, err := db.CreateTask(Task{Title: "未着手"})
	require.NoError(t, err)
	done, err := db.CreateTask(Task{Title: "完了済み"})
	require.NoError(t, err)

	done.Status = StatusDone
	require.NoError(t, db.UpdateTask(done))

	tasks, err := db.ListTasks(TaskFilter{})
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, "未着手", tasks[0].Title)
}

func TestListTasks_BranchFilter(t *testing.T) {
	db := newTestDB(t)

	_, err := db.CreateTask(Task{Title: "ログイン", Branch: "feat/login"})
	require.NoError(t, err)
	_, err = db.CreateTask(Task{Title: "サインアップ", Branch: "feat/signup"})
	require.NoError(t, err)

	tasks, err := db.ListTasks(TaskFilter{Branch: "feat/login"})
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, "ログイン", tasks[0].Title)
}

func TestListTasks_TagFilter(t *testing.T) {
	db := newTestDB(t)

	_, err := db.CreateTask(Task{Title: "認証", Tags: []string{"auth"}})
	require.NoError(t, err)
	_, err = db.CreateTask(Task{Title: "UI修正", Tags: []string{"front"}})
	require.NoError(t, err)

	tasks, err := db.ListTasks(TaskFilter{Tag: "auth"})
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, "認証", tasks[0].Title)
}

func TestUpdateTask(t *testing.T) {
	db := newTestDB(t)

	task, err := db.CreateTask(Task{Title: "元のタイトル"})
	require.NoError(t, err)

	task.Title = "更新後のタイトル"
	task.Status = StatusInProgress
	require.NoError(t, db.UpdateTask(task))

	fetched, err := db.GetTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "更新後のタイトル", fetched.Title)
	assert.Equal(t, StatusInProgress, fetched.Status)
}

func TestReactivateSnoozed(t *testing.T) {
	db := newTestDB(t)

	yesterday := time.Now().AddDate(0, 0, -1)
	task, err := db.CreateTask(Task{Title: "スヌーズ中"})
	require.NoError(t, err)

	task.Status = StatusSnoozed
	task.SnoozedUntil = &yesterday
	require.NoError(t, db.UpdateTask(task))

	require.NoError(t, db.ReactivateSnoozed(time.Now()))

	fetched, err := db.GetTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusTodo, fetched.Status)
	assert.Nil(t, fetched.SnoozedUntil)
}

func TestBlockOperations(t *testing.T) {
	db := newTestDB(t)

	t1, err := db.CreateTask(Task{Title: "ブロックされるタスク"})
	require.NoError(t, err)
	t2, err := db.CreateTask(Task{Title: "ブロックするタスク"})
	require.NoError(t, err)

	// t1 は t2 の完了を待つ
	require.NoError(t, db.AddBlock(t1.ID, t2.ID))

	count, err := db.BlockedByCount(t2.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// t2 が完了 → t1 のブロックを解除
	require.NoError(t, db.RemoveAllBlocksBy(t2.ID))

	count, err = db.BlockedByCount(t2.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestRemoveBlock(t *testing.T) {
	db := newTestDB(t)

	t1, _ := db.CreateTask(Task{Title: "タスク1"})
	t2, _ := db.CreateTask(Task{Title: "タスク2"})

	require.NoError(t, db.AddBlock(t1.ID, t2.ID))
	count, _ := db.BlockedByCount(t2.ID)
	assert.Equal(t, 1, count)

	require.NoError(t, db.RemoveBlock(t1.ID, t2.ID))
	count, _ = db.BlockedByCount(t2.ID)
	assert.Equal(t, 0, count)
}

func TestCreateTask_WithoutDueDate(t *testing.T) {
	db := newTestDB(t)

	task, err := db.CreateTask(Task{Title: "締め切りなし"})
	require.NoError(t, err)
	assert.Nil(t, task.DueDate)
	assert.Empty(t, task.Tags)
	assert.False(t, task.Pinned)
}

func TestUpdateTask_Pin(t *testing.T) {
	db := newTestDB(t)

	task, err := db.CreateTask(Task{Title: "ピン固定テスト"})
	require.NoError(t, err)
	assert.False(t, task.Pinned)

	task.Pinned = true
	require.NoError(t, db.UpdateTask(task))
	updated, err := db.GetTask(task.ID)
	require.NoError(t, err)
	assert.True(t, updated.Pinned)
}

func TestListTasks_SnoozedStatus(t *testing.T) {
	db := newTestDB(t)

	_, err := db.CreateTask(Task{Title: "通常タスク"})
	require.NoError(t, err)
	snoozed, err := db.CreateTask(Task{Title: "スヌーズタスク"})
	require.NoError(t, err)

	tomorrow := time.Now().AddDate(0, 0, 1)
	snoozed.Status = StatusSnoozed
	snoozed.SnoozedUntil = &tomorrow
	require.NoError(t, db.UpdateTask(snoozed))

	// snoozed も含めて取得
	tasks, err := db.ListTasks(TaskFilter{
		Statuses: []Status{StatusTodo, StatusSnoozed},
	})
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestAddBlock_Duplicate(t *testing.T) {
	db := newTestDB(t)
	t1, _ := db.CreateTask(Task{Title: "タスク1"})
	t2, _ := db.CreateTask(Task{Title: "タスク2"})

	require.NoError(t, db.AddBlock(t1.ID, t2.ID))
	// 重複追加は OR IGNORE でエラーにならない
	require.NoError(t, db.AddBlock(t1.ID, t2.ID))

	count, _ := db.BlockedByCount(t2.ID)
	assert.Equal(t, 1, count)
}

func TestCreateTask_WithSourceInfo(t *testing.T) {
	db := newTestDB(t)

	task, err := db.CreateTask(Task{
		Title:      "TODO from code",
		SourceFile: "src/auth/login.go",
		SourceLine: 42,
	})
	require.NoError(t, err)
	assert.Equal(t, "src/auth/login.go", task.SourceFile)
	assert.Equal(t, 42, task.SourceLine)
}

func TestListTasks_Done(t *testing.T) {
	db := newTestDB(t)

	task, err := db.CreateTask(Task{Title: "完了タスク"})
	require.NoError(t, err)
	task.Status = StatusDone
	require.NoError(t, db.UpdateTask(task))

	// done を明示的に指定して取得
	tasks, err := db.ListTasks(TaskFilter{Statuses: []Status{StatusDone}})
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, StatusDone, tasks[0].Status)
}

func TestGetRecentShown_Empty(t *testing.T) {
	db := newTestDB(t)
	ids, err := db.GetRecentShown()
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestReactivateSnoozed_FutureSnooze(t *testing.T) {
	db := newTestDB(t)

	tomorrow := time.Now().AddDate(0, 0, 1)
	task, err := db.CreateTask(Task{Title: "未来スヌーズ"})
	require.NoError(t, err)

	task.Status = StatusSnoozed
	task.SnoozedUntil = &tomorrow
	require.NoError(t, db.UpdateTask(task))

	// 未来のスヌーズは解除されない
	require.NoError(t, db.ReactivateSnoozed(time.Now()))
	fetched, err := db.GetTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, StatusSnoozed, fetched.Status)
}

func TestGetPrerequisites(t *testing.T) {
	database := newTestDB(t)

	t1, _ := database.CreateTask(Task{Title: "前提タスク"})
	t2, _ := database.CreateTask(Task{Title: "依存タスク"})

	// t2 は t1 の完了を待つ
	require.NoError(t, database.AddBlock(t2.ID, t1.ID))

	// t2 の前提タスクは t1
	prereqs, err := database.GetPrerequisites(t2.ID)
	require.NoError(t, err)
	require.Len(t, prereqs, 1)
	assert.Equal(t, t1.ID, prereqs[0].ID)

	// t1 には前提タスクなし
	prereqs, err = database.GetPrerequisites(t1.ID)
	require.NoError(t, err)
	assert.Empty(t, prereqs)
}

func TestDisplayHistory(t *testing.T) {
	db := newTestDB(t)

	t1, _ := db.CreateTask(Task{Title: "タスク1"})
	t2, _ := db.CreateTask(Task{Title: "タスク2"})
	t3, _ := db.CreateTask(Task{Title: "タスク3"})

	require.NoError(t, db.RecordShown(t1.ID))
	require.NoError(t, db.RecordShown(t2.ID))
	require.NoError(t, db.RecordShown(t3.ID)) // 直近2件を超えた場合、t1が消える

	ids, err := db.GetRecentShown()
	require.NoError(t, err)
	assert.Equal(t, []int64{t3.ID, t2.ID}, ids) // 最新順
}
