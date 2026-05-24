package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanTODOs_Basic(t *testing.T) {
	dir := t.TempDir()

	write(t, dir, "main.go", `package main

// TODO: ログイン機能を実装する
func main() {
	// TODO: エラーハンドリングを追加する
}
`)

	items, err := ScanTODOs(dir)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Contains(t, items[0].Text, "ログイン機能")
	assert.Equal(t, 3, items[0].Line)
	assert.Contains(t, items[1].Text, "エラーハンドリング")
}

func TestScanTODOs_MultipleFiles(t *testing.T) {
	dir := t.TempDir()

	write(t, dir, "a.go", "// TODO: タスクA\n")
	write(t, dir, "b.go", "// TODO: タスクB\n")

	items, err := ScanTODOs(dir)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestScanTODOs_NoTODOs(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "clean.go", "package main\n\nfunc main() {}\n")

	items, err := ScanTODOs(dir)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestScanTODOs_GitignoreRespected(t *testing.T) {
	dir := t.TempDir()

	write(t, dir, ".gitignore", "vendor/\n")
	vendorDir := filepath.Join(dir, "vendor")
	require.NoError(t, os.Mkdir(vendorDir, 0700))
	write(t, vendorDir, "lib.go", "// TODO: vendorのTODOは無視される\n")
	write(t, dir, "main.go", "// TODO: メインのTODO\n")

	items, err := ScanTODOs(dir)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Contains(t, items[0].Text, "メインのTODO")
}

func TestScanTODOs_SkipGitDir(t *testing.T) {
	dir := t.TempDir()

	gitDir := filepath.Join(dir, ".git")
	require.NoError(t, os.Mkdir(gitDir, 0700))
	write(t, gitDir, "COMMIT_EDITMSG", "TODO: gitディレクトリは無視\n")
	write(t, dir, "code.go", "// TODO: 有効なTODO\n")

	items, err := ScanTODOs(dir)
	require.NoError(t, err)
	assert.Len(t, items, 1)
}

func TestScanTODOs_SubDir(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "internal", "auth")
	require.NoError(t, os.MkdirAll(subDir, 0700))
	write(t, subDir, "login.go", "// TODO: OAuth実装\n")

	items, err := ScanTODOs(dir)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Contains(t, items[0].Text, "OAuth")
}

func TestScanTODOs_HashComment(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "script.py", "# TODO: Pythonのコメント\n")

	items, err := ScanTODOs(dir)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Contains(t, items[0].Text, "Python")
}

func TestIsTextFile_Binary(t *testing.T) {
	dir := t.TempDir()
	binaryPath := filepath.Join(dir, "data.bin")
	data := []byte{0x00, 0x01, 0x02, 0x03, 0xFF}
	require.NoError(t, os.WriteFile(binaryPath, data, 0600))

	assert.False(t, isTextFile(binaryPath))
}

func TestIsTextFile_KnownBinaryExt(t *testing.T) {
	dir := t.TempDir()
	pngPath := filepath.Join(dir, "image.png")
	require.NoError(t, os.WriteFile(pngPath, []byte("fake"), 0600))
	assert.False(t, isTextFile(pngPath))
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
}
