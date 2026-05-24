package testhelper

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// NewTempGitRepo は t.TempDir() 上に一時的な git リポジトリを作成してパスを返す。
// テスト終了時に自動削除される。
func NewTempGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %s", args, out)
		}
	}

	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test User")

	// 最初のコミットを作成しておく（ブランチ名が確定するため）
	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# test\n"), 0600); err != nil {
		t.Fatalf("README.md の作成に失敗: %v", err)
	}
	run("add", ".")
	run("commit", "-m", "initial commit")

	return dir
}

// WriteTempFile は dir 内に name のファイルを content で作成してパスを返す。
func WriteTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatalf("ディレクトリの作成に失敗: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("ファイルの作成に失敗: %v", err)
	}
	return path
}
