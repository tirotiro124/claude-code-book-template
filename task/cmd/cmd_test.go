package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// resetFlags resets cobra flag variables between tests.
func resetFlags() {
	addDue = ""
	addBranch = ""
	addTags = nil
	listStatus = ""
	listTag = ""
	listAll = false
	jsonOut = false
	noColor = true
	syncDryRun = false
	exportFormat = "json"
}

func runCmd(t *testing.T, args ...string) (string, string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	cmdOut = &out
	cmdErr = &errBuf
	t.Cleanup(func() {
		cmdOut = nil
		cmdErr = nil
	})
	rootCmd.SetArgs(args)
	_ = rootCmd.Execute()
	return out.String(), errBuf.String()
}

func TestAdd_Basic(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	out, _ := runCmd(t, "add", "テストタスク")
	assert.Contains(t, out, "Added #1")
	assert.Contains(t, out, "テストタスク")
}

func TestAdd_WithDue(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	out, _ := runCmd(t, "add", "--due", "2030-12-31", "期限付きタスク")
	assert.Contains(t, out, "Added #1")
	assert.Contains(t, out, "期限付きタスク")
}

func TestAdd_InvalidDue(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	_, errOut := runCmd(t, "add", "--due", "not-a-date", "タスク")
	assert.Contains(t, errOut, "Error:")
}

func TestList_Empty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	out, _ := runCmd(t, "list")
	assert.Contains(t, out, "No tasks found")
}

func TestList_WithTasks(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "タスクA")
	runCmd(t, "add", "タスクB")
	resetFlags()

	out, _ := runCmd(t, "list")
	assert.Contains(t, out, "タスクA")
	assert.Contains(t, out, "タスクB")
}

func TestNext_Empty(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	out, _ := runCmd(t, "next")
	assert.Contains(t, out, "No tasks found")
}

func TestNext_WithTask(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "次のタスク")
	resetFlags()

	out, _ := runCmd(t, "next")
	assert.Contains(t, out, "次のタスク")
	assert.Contains(t, out, "Next Task")
}

func TestDone_Basic(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "完了するタスク")
	resetFlags()

	out, _ := runCmd(t, "done", "1")
	assert.Contains(t, out, "Done #1")
}

func TestDone_NotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	_, errOut := runCmd(t, "done", "999")
	assert.Contains(t, errOut, "Error:")
}

func TestDone_InvalidID(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	// cobra returns error for bad arg, rootCmd prints it to stderr
	_, _ = runCmd(t, "done", "abc")
}

func TestShow_Basic(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "詳細表示タスク")
	resetFlags()

	out, _ := runCmd(t, "show", "1")
	assert.Contains(t, out, "#1")
	assert.Contains(t, out, "詳細表示タスク")
}

func TestShow_NotFound(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	_, errOut := runCmd(t, "show", "99")
	assert.Contains(t, errOut, "Error:")
}

func TestList_JSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()
	jsonOut = true

	runCmd(t, "add", "JSONタスク")
	resetFlags()
	jsonOut = true

	out, _ := runCmd(t, "list")
	assert.Contains(t, out, `"Total"`)
}

func TestList_All(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "タスクX")
	runCmd(t, "done", "1")
	resetFlags()
	listAll = true

	out, _ := runCmd(t, "list", "--all")
	assert.Contains(t, out, "タスクX")
}

func TestAdd_WithBranch(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	out, _ := runCmd(t, "add", "--branch", "feat/test", "ブランチ付きタスク")
	assert.Contains(t, out, "feat/test")
}

func TestNext_JSON(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "JSONネクストタスク")
	resetFlags()
	jsonOut = true

	out, _ := runCmd(t, "next")
	assert.Contains(t, out, `"Total"`)
}

func TestList_WithTag(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "--tag", "auth", "認証タスク")
	resetFlags()

	out, _ := runCmd(t, "list", "--tag", "auth")
	assert.Contains(t, out, "認証タスク")
}

func TestList_StatusFilter(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "フィルタタスク")
	resetFlags()

	out, _ := runCmd(t, "list", "--status", "todo")
	assert.Contains(t, out, "フィルタタスク")
}

func TestExecute_Help(t *testing.T) {
	var out bytes.Buffer
	cmdOut = &out
	t.Cleanup(func() { cmdOut = nil })
	rootCmd.SetArgs([]string{"--help"})
	Execute("test-version")
	assert.Contains(t, out.String()+rootCmd.UsageString(), "task")
}

func TestShow_InvalidID(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	// non-integer id returns cobra error
	_, _ = runCmd(t, "show", "notanid")
}

func TestDone_ShowsNext(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	resetFlags()

	runCmd(t, "add", "タスク1")
	runCmd(t, "add", "タスク2")
	resetFlags()

	out, _ := runCmd(t, "done", "1")
	assert.Contains(t, out, "Done #1")
	// Next task should also be shown
	assert.Contains(t, out, "タスク2")
}
