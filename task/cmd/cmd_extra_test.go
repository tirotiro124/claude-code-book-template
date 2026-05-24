package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/task-cli/task/internal/db"
)

// --- snooze ---

func TestSnooze_Basic(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "スヌーズするタスク")
	resetFlags()

	out, _ := runCmd(t, "snooze", "1", "2d")
	assert.Contains(t, out, "Snoozed #1")
}

func TestSnooze_NotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	_, errOut := runCmd(t, "snooze", "999", "1d")
	assert.Contains(t, errOut, "Error:")
}

func TestSnooze_InvalidID(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	_, _ = runCmd(t, "snooze", "abc", "1d")
}

func TestSnooze_WeekDuration(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "1週間スヌーズ")
	resetFlags()

	out, _ := runCmd(t, "snooze", "1", "1w")
	assert.Contains(t, out, "Snoozed #1")
}

func TestParseSnoozeDuration(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"1d", false},
		{"2h", false},
		{"30m", false},
		{"1w", false},
		{"1h30m", false}, // Go 標準形式
		{"x", true},
		{"0d", true},
		{"1x", true},
	}
	for _, tt := range tests {
		_, err := parseSnoozeDuration(tt.input)
		if tt.wantErr {
			assert.Error(t, err, "input: %s", tt.input)
		} else {
			assert.NoError(t, err, "input: %s", tt.input)
		}
	}
}

// --- pin / unpin ---

func TestPin_Basic(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "ピン固定タスク")
	resetFlags()

	out, _ := runCmd(t, "pin", "1")
	assert.Contains(t, out, "Pinned #1")
}

func TestUnpin_Basic(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "ピン解除タスク")
	runCmd(t, "pin", "1")
	resetFlags()

	out, _ := runCmd(t, "unpin", "1")
	assert.Contains(t, out, "Unpinned #1")
}

func TestPin_NotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	_, errOut := runCmd(t, "pin", "99")
	assert.Contains(t, errOut, "Error:")
}

// --- block / unblock ---

func TestBlock_Basic(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "タスクA")
	runCmd(t, "add", "タスクB")
	resetFlags()

	out, _ := runCmd(t, "block", "1", "2")
	assert.Contains(t, out, "Block added")
	assert.Contains(t, out, "#1")
	assert.Contains(t, out, "#2")
}

func TestUnblock_Basic(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "タスクA")
	runCmd(t, "add", "タスクB")
	runCmd(t, "block", "1", "2")
	resetFlags()

	out, _ := runCmd(t, "unblock", "1", "2")
	assert.Contains(t, out, "Block removed")
}

func TestBlock_SelfBlock(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "タスク")
	resetFlags()

	_, errOut := runCmd(t, "block", "1", "1")
	assert.Contains(t, errOut, "Error:")
}

func TestBlock_NotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "タスクのみ")
	resetFlags()

	_, errOut := runCmd(t, "block", "1", "999")
	assert.Contains(t, errOut, "Error:")
}

// --- export ---

func TestExport_JSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "エクスポートタスク")
	resetFlags()
	exportFormat = "json"

	out, _ := runCmd(t, "export", "--format", "json")
	assert.Contains(t, out, "エクスポートタスク")
	assert.Contains(t, out, `"Title"`)
}

func TestExport_Markdown(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "MDエクスポートタスク")
	resetFlags()

	out, _ := runCmd(t, "export", "--format", "md")
	assert.Contains(t, out, "MDエクスポートタスク")
	assert.Contains(t, out, "# Tasks")
}

func TestExport_InvalidFormat(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "タスク")
	resetFlags()

	_, errOut := runCmd(t, "export", "--format", "csv")
	assert.Contains(t, errOut, "Error:")
}

func TestExport_Done(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "完了タスク")
	runCmd(t, "done", "1")
	resetFlags()

	out, _ := runCmd(t, "export", "--format", "md")
	assert.Contains(t, out, "Done")
}

// --- applyEditable ---

func TestApplyEditable_EmptyTitle(t *testing.T) {
	task := mustCreateEditableTask()
	err := applyEditable(task, editableTask{Title: "", Status: "todo"})
	assert.Error(t, err)
}

func TestApplyEditable_InvalidDue(t *testing.T) {
	task := mustCreateEditableTask()
	err := applyEditable(task, editableTask{Title: "t", Due: "not-a-date", Status: "todo"})
	assert.Error(t, err)
}

func TestApplyEditable_InvalidStatus(t *testing.T) {
	task := mustCreateEditableTask()
	err := applyEditable(task, editableTask{Title: "t", Status: "unknown"})
	assert.Error(t, err)
}

func TestApplyEditable_Tags(t *testing.T) {
	task := mustCreateEditableTask()
	err := applyEditable(task, editableTask{Title: "t", Tags: "auth, ui", Status: "todo"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"auth", "ui"}, task.Tags)
}

func TestApplyEditable_ValidDue(t *testing.T) {
	task := mustCreateEditableTask()
	err := applyEditable(task, editableTask{Title: "t", Due: "2030-06-01", Status: "todo"})
	assert.NoError(t, err)
	assert.NotNil(t, task.DueDate)
}

func mustCreateEditableTask() *db.Task {
	return &db.Task{}
}

// --- taskToEditable / applyEditable ユニットテスト ---

func TestTaskToEditable_WithDue(t *testing.T) {
	due := time.Date(2030, 6, 1, 0, 0, 0, 0, time.Local)
	task := db.Task{
		Title:   "タイトル",
		Branch:  "feat/test",
		Tags:    []string{"auth", "ui"},
		Status:  db.StatusTodo,
		DueDate: &due,
	}
	et := taskToEditable(task)
	assert.Equal(t, "タイトル", et.Title)
	assert.Equal(t, "2030-06-01", et.Due)
	assert.Equal(t, "feat/test", et.Branch)
	assert.Contains(t, et.Tags, "auth")
	assert.Equal(t, "todo", et.Status)
}

func TestTaskToEditable_NoDue(t *testing.T) {
	task := db.Task{Title: "締め切りなし", Status: db.StatusInProgress}
	et := taskToEditable(task)
	assert.Equal(t, "", et.Due)
	assert.Equal(t, "in_progress", et.Status)
}

// --- isDuplicate ユニットテスト ---

func TestIsDuplicate_Match(t *testing.T) {
	tasks := []db.Task{
		{SourceFile: "src/auth.go", SourceLine: 10},
		{SourceFile: "src/main.go", SourceLine: 5},
	}
	assert.True(t, isDuplicate(tasks, "src/auth.go", 10))
}

func TestIsDuplicate_NoMatch(t *testing.T) {
	tasks := []db.Task{
		{SourceFile: "src/auth.go", SourceLine: 10},
	}
	assert.False(t, isDuplicate(tasks, "src/auth.go", 99))
	assert.False(t, isDuplicate(tasks, "other.go", 10))
}

func TestIsDuplicate_Empty(t *testing.T) {
	assert.False(t, isDuplicate(nil, "file.go", 1))
}

// --- edit (EDITOR=true でファイルを変更せずに保存) ---

func TestEdit_WithTrueEditor(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("EDITOR", "true") // 何もせず終了
	resetFlags()

	runCmd(t, "add", "編集するタスク")
	resetFlags()

	out, _ := runCmd(t, "edit", "1")
	assert.Contains(t, out, "Updated #1")
}

func TestEdit_NoEditor(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("EDITOR", "")
	t.Setenv("VISUAL", "")
	resetFlags()

	runCmd(t, "add", "タスク")
	resetFlags()

	_, _ = runCmd(t, "edit", "1") // $EDITOR 未設定でエラーが返る
}

func TestEdit_NotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("EDITOR", "true")
	resetFlags()

	_, errOut := runCmd(t, "edit", "999")
	assert.Contains(t, errOut, "Error:")
}

// --- sync ---

func TestSync_NonGitRepo(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Chdir(tmp)
	resetFlags()

	_, errOut := runCmd(t, "sync")
	assert.Contains(t, errOut, "Error:")
}

func TestSync_NoTODOs(t *testing.T) {
	gitDir := t.TempDir()
	initGitRepoForTest(t, gitDir)
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "main.go"), []byte("package main\n"), 0600))

	t.Setenv("HOME", t.TempDir())
	t.Chdir(gitDir)
	resetFlags()

	out, _ := runCmd(t, "sync")
	assert.Contains(t, out, "見つかりません")
}

func TestSync_WithTODOs(t *testing.T) {
	gitDir := t.TempDir()
	initGitRepoForTest(t, gitDir)
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "main.go"),
		[]byte("// TODO: テスト用タスクを追加する\n"), 0600))

	t.Setenv("HOME", t.TempDir())
	t.Chdir(gitDir)
	resetFlags()

	out, _ := runCmd(t, "sync")
	assert.Contains(t, out, "Added:")
	assert.Contains(t, out, "テスト用タスク")
}

func TestSync_DryRun_WithTODOs(t *testing.T) {
	gitDir := t.TempDir()
	initGitRepoForTest(t, gitDir)
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "code.go"),
		[]byte("// TODO: ドライランテスト\n"), 0600))

	t.Setenv("HOME", t.TempDir())
	t.Chdir(gitDir)
	resetFlags()
	syncDryRun = true

	out, _ := runCmd(t, "sync", "--dry-run")
	assert.Contains(t, out, "dry-run")
	assert.Contains(t, out, "ドライラン")
}

func TestSync_SkipDuplicate(t *testing.T) {
	gitDir := t.TempDir()
	initGitRepoForTest(t, gitDir)
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "a.go"),
		[]byte("// TODO: 重複チェックタスク\n"), 0600))

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Chdir(gitDir)
	resetFlags()

	// 1回目
	runCmd(t, "sync")
	resetFlags()
	// 2回目（重複スキップ）
	out, _ := runCmd(t, "sync")
	assert.Contains(t, out, "skipped")
}

func initGitRepoForTest(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	} {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = dir
		_ = c.Run()
	}
}
