package git

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/task-cli/task/internal/testhelper"
)

func TestCurrentBranch(t *testing.T) {
	dir := testhelper.NewTempGitRepo(t)
	p := ExecProvider{Dir: dir}

	branch, err := p.CurrentBranch()
	require.NoError(t, err)
	// git init のデフォルトブランチは "main" または "master"
	assert.NotEmpty(t, branch)
}

func TestCurrentBranch_CustomBranch(t *testing.T) {
	dir := testhelper.NewTempGitRepo(t)

	cmd := exec.Command("git", "checkout", "-b", "feat/login")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	p := ExecProvider{Dir: dir}
	branch, err := p.CurrentBranch()
	require.NoError(t, err)
	assert.Equal(t, "feat/login", branch)
}

func TestCurrentBranch_NotGitRepo(t *testing.T) {
	dir := t.TempDir() // git init していないディレクトリ
	p := ExecProvider{Dir: dir}

	_, err := p.CurrentBranch()
	assert.ErrorIs(t, err, ErrNotGitRepository)
}

func TestRootDir(t *testing.T) {
	dir := testhelper.NewTempGitRepo(t)
	p := ExecProvider{Dir: dir}

	root, err := p.RootDir()
	require.NoError(t, err)
	assert.NotEmpty(t, root)
}

func TestRootDir_NotGitRepo(t *testing.T) {
	dir := t.TempDir()
	p := ExecProvider{Dir: dir}

	_, err := p.RootDir()
	assert.ErrorIs(t, err, ErrNotGitRepository)
}
