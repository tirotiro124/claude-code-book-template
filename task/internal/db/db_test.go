package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	db, err := Open()
	require.NoError(t, err)
	defer db.Close()

	info, err := os.Stat(filepath.Join(dir, ".task", "tasks.db"))
	require.NoError(t, err)
	assert.Equal(t, "-rw-------", info.Mode().String())
}

func TestOpenAt_Reopen(t *testing.T) {
	dir := t.TempDir()

	// 1回目：DB作成
	db1, err := OpenAt(dir)
	require.NoError(t, err)
	_, err = db1.Exec(`INSERT INTO tasks (title, status, tags, pinned, created_at, updated_at) VALUES ('test','todo','[]',0,datetime('now'),datetime('now'))`)
	require.NoError(t, err)
	db1.Close()

	// 2回目：既存DBを再オープン（マイグレーションはスキップされる）
	db2, err := OpenAt(dir)
	require.NoError(t, err)
	defer db2.Close()

	var count int
	require.NoError(t, db2.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&count))
	assert.Equal(t, 1, count)
}

func TestOpenAt(t *testing.T) {
	dir := t.TempDir()
	db, err := OpenAt(dir)
	require.NoError(t, err)
	defer db.Close()

	// tasks.db が作成されていること
	info, err := os.Stat(filepath.Join(dir, "tasks.db"))
	require.NoError(t, err)
	assert.Equal(t, "-rw-------", info.Mode().String())

	// マイグレーションが適用されていること
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM schema_versions`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, len(migrations), count)
}

func TestOpenMemory(t *testing.T) {
	db, err := OpenMemory()
	require.NoError(t, err)
	defer db.Close()

	// schema_versions テーブルが存在し、マイグレーションが適用されていること
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM schema_versions`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, len(migrations), count)
}

func TestMigrate_Idempotent(t *testing.T) {
	db, err := OpenMemory()
	require.NoError(t, err)
	defer db.Close()

	// 2回目のマイグレーションはエラーにならない
	err = db.migrate()
	assert.NoError(t, err)

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM schema_versions`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, len(migrations), count)
}

func TestMigrate_TablesExist(t *testing.T) {
	db, err := OpenMemory()
	require.NoError(t, err)
	defer db.Close()

	tables := []string{"tasks", "task_blocks", "display_history"}
	for _, table := range tables {
		var name string
		err := db.QueryRow(
			`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&name)
		require.NoError(t, err, "テーブル %s が存在しない", table)
		assert.Equal(t, table, name)
	}
}
