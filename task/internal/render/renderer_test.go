package render

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/task-cli/task/internal/db"
	"github.com/task-cli/task/internal/scoring"
)

func newTestRenderer() (*Renderer, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	r := &Renderer{
		Stdout:  stdout,
		Stderr:  stderr,
		JSON:    false,
		NoColor: true,
	}
	return r, stdout, stderr
}

func TestNextTask_Output(t *testing.T) {
	r, stdout, _ := newTestRenderer()
	due := time.Now().AddDate(0, 0, 2)
	task := db.Task{ID: 42, Title: "ログイン実装", Branch: "feat/login", DueDate: &due}
	score := scoring.ScoreResult{Total: 65, Urgency: 30, Importance: 20, Context: 15, Aging: 10, Fatigue: 10}

	r.NextTask(task, score, "feat/login")

	out := stdout.String()
	assert.Contains(t, out, "#42")
	assert.Contains(t, out, "ログイン実装")
	assert.Contains(t, out, "65")
	assert.Contains(t, out, "branch: ✓")
}

func TestNextTask_JSON(t *testing.T) {
	r, stdout, _ := newTestRenderer()
	r.JSON = true
	task := db.Task{ID: 1, Title: "テスト"}
	score := scoring.ScoreResult{Total: 10}

	r.NextTask(task, score, "main")

	assert.Contains(t, stdout.String(), `"Total"`)
}

func TestTaskList_Output(t *testing.T) {
	r, stdout, _ := newTestRenderer()
	tasks := []db.Task{
		{ID: 1, Title: "タスク1", Branch: "feat/a"},
		{ID: 2, Title: "タスク2"},
	}
	scores := []scoring.ScoreResult{{Total: 50}, {Total: 30}}

	r.TaskList(tasks, scores)

	out := stdout.String()
	assert.Contains(t, out, "タスク1")
	assert.Contains(t, out, "タスク2")
	assert.Contains(t, out, "50")
}

func TestTaskList_Empty(t *testing.T) {
	r, stdout, _ := newTestRenderer()
	r.TaskList(nil, nil)
	assert.Contains(t, stdout.String(), "No tasks found")
}

func TestTaskDetail_Output(t *testing.T) {
	r, stdout, _ := newTestRenderer()
	task := db.Task{
		ID:        42,
		Title:     "ログイン実装",
		Status:    db.StatusTodo,
		Branch:    "feat/login",
		Tags:      []string{"auth"},
		CreatedAt: time.Now().AddDate(0, 0, -5),
	}
	score := scoring.ScoreResult{Total: 65, Urgency: 30, Importance: 20}
	blockedBy := []db.Task{{ID: 38, Title: "前提タスク"}}

	r.TaskDetail(task, score, blockedBy)

	out := stdout.String()
	assert.Contains(t, out, "#42")
	assert.Contains(t, out, "ログイン実装")
	assert.Contains(t, out, "feat/login")
	assert.Contains(t, out, "#38")
}

func TestAdded_Output(t *testing.T) {
	r, stdout, _ := newTestRenderer()
	task := db.Task{ID: 1, Title: "新しいタスク", Branch: "feat/new"}

	r.Added(task)

	out := stdout.String()
	assert.Contains(t, out, "Added #1")
	assert.Contains(t, out, "新しいタスク")
	assert.Contains(t, out, "feat/new")
}

func TestDone_Output(t *testing.T) {
	r, stdout, _ := newTestRenderer()
	task := db.Task{ID: 42, Title: "完了タスク"}

	r.Done(task)

	assert.Contains(t, stdout.String(), "Done #42")
}

func TestError_Output(t *testing.T) {
	r, _, stderr := newTestRenderer()
	r.Error("タスク #99 が見つかりません")
	assert.Contains(t, stderr.String(), "Error:")
	assert.Contains(t, stderr.String(), "#99")
}

func TestNoTasks_Output(t *testing.T) {
	r, stdout, _ := newTestRenderer()
	r.NoTasks()
	out := stdout.String()
	assert.Contains(t, out, "No tasks found")
	assert.Contains(t, out, "task add")
}

func TestFormatDue_Colors(t *testing.T) {
	now := time.Now()
	red := func(s string, a ...interface{}) string { return "[RED]" + s }
	orange := func(s string, a ...interface{}) string { return "[ORANGE]" + s }

	tests := []struct {
		name     string
		due      *time.Time
		contains string
	}{
		{"締め切りなし", nil, ""},
		{"当日", ptr(now), "[RED]"},
		{"2日後", ptr(now.AddDate(0, 0, 2)), "[ORANGE]"},
		{"10日後", ptr(now.AddDate(0, 0, 10)), "10d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDue(tt.due, now, red, orange)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestNew(t *testing.T) {
	r := New(false, true)
	assert.NotNil(t, r)
	assert.True(t, r.NoColor)
}

func TestNew_NoColor_Env(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	r := New(false, false)
	assert.NotNil(t, r)
}

func TestDone_JSON(t *testing.T) {
	r, stdout, _ := newTestRenderer()
	r.JSON = true
	task := db.Task{ID: 1, Title: "テスト"}
	r.Done(task)
	assert.Contains(t, stdout.String(), `"ID"`)
}

func TestTaskList_JSON(t *testing.T) {
	r, stdout, _ := newTestRenderer()
	r.JSON = true
	tasks := []db.Task{{ID: 1, Title: "テスト"}}
	scores := []scoring.ScoreResult{{Total: 10}}
	r.TaskList(tasks, scores)
	assert.Contains(t, stdout.String(), `"Total"`)
}

func TestTaskDetail_NoBranch_NoBlocks(t *testing.T) {
	r, stdout, _ := newTestRenderer()
	task := db.Task{
		ID:        1,
		Title:     "シンプルタスク",
		Status:    db.StatusTodo,
		CreatedAt: time.Now(),
	}
	r.TaskDetail(task, scoring.ScoreResult{}, nil)
	out := stdout.String()
	assert.Contains(t, out, "#1")
	assert.NotContains(t, out, "branch")
}

func TestFormatDuePlain(t *testing.T) {
	now := time.Now()
	red := func(s string, a ...interface{}) string { return "[RED]" }
	orange := func(s string, a ...interface{}) string { return "[ORANGE]" }

	assert.Equal(t, "-", formatDuePlain(nil, now, red, orange))
	assert.Contains(t, formatDuePlain(ptr(now), now, red, orange), "[RED]")
	assert.Contains(t, formatDuePlain(ptr(now.AddDate(0, 0, 2)), now, red, orange), "[ORANGE]")
	assert.Contains(t, formatDuePlain(ptr(now.AddDate(0, 0, 10)), now, red, orange), "10d")
	assert.Contains(t, formatDuePlain(ptr(now.AddDate(0, 0, -1)), now, red, orange), "[RED]")
}

func TestDueDesc(t *testing.T) {
	now := time.Now()
	assert.Equal(t, "締め切りなし", dueDesc(nil, now))
	assert.Contains(t, dueDesc(ptr(now.AddDate(0, 0, 5)), now), "5日")
	assert.Contains(t, dueDesc(ptr(now.AddDate(0, 0, -2)), now), "超過")
}

func TestColorize_WithColor(t *testing.T) {
	r, _, _ := newTestRenderer()
	r.NoColor = false
	fn := r.colorize()
	assert.NotNil(t, fn)
}

func ptr(t time.Time) *time.Time { return &t }

func TestAdded_NoJSON_NoBranch(t *testing.T) {
	r, stdout, _ := newTestRenderer()
	task := db.Task{ID: 5, Title: "ブランチなし"}
	r.Added(task)
	out := stdout.String()
	assert.Contains(t, out, "Added #5")
	assert.False(t, strings.Contains(out, "branch:"))
}
