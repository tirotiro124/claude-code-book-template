package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// DB はSQLite接続のラッパー。
type DB struct {
	*sql.DB
}

// Open はDBファイルを開き、マイグレーションを実行する。
// ~/.task/ ディレクトリが存在しない場合は自動作成する。
func Open() (*DB, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("ホームディレクトリの取得に失敗しました: %w", err)
	}
	dir := filepath.Join(home, ".task")
	return OpenAt(dir)
}

// OpenAt は指定ディレクトリにDBを開く。テストからも呼び出せる。
func OpenAt(dir string) (*DB, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("ディレクトリの作成に失敗しました: %w", err)
	}

	dbPath := filepath.Join(dir, "tasks.db")
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("DB接続に失敗しました: %w", err)
	}

	// PRAGMA を実行することでファイルが物理的に作成される
	if _, err := sqlDB.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return nil, fmt.Errorf("PRAGMA journal_mode の設定に失敗しました: %w", err)
	}

	if err := os.Chmod(dbPath, 0600); err != nil {
		return nil, fmt.Errorf("DBファイルのパーミッション設定に失敗しました: %w", err)
	}

	if _, err := sqlDB.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return nil, fmt.Errorf("PRAGMA journal_mode の設定に失敗しました: %w", err)
	}
	if _, err := sqlDB.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		return nil, fmt.Errorf("PRAGMA foreign_keys の設定に失敗しました: %w", err)
	}

	db := &DB{sqlDB}
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("マイグレーションに失敗しました: %w", err)
	}
	return db, nil
}

// OpenMemory はテスト用のインメモリDBを開く。
func OpenMemory() (*DB, error) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	if _, err := sqlDB.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		return nil, err
	}
	db := &DB{sqlDB}
	return db, db.migrate()
}

func (db *DB) migrate() error {
	// schema_versions テーブルをブートストラップとして作成する
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_versions (
			version    INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("schema_versions テーブルの作成に失敗しました: %w", err)
	}

	var current int
	row := db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_versions`)
	if err := row.Scan(&current); err != nil {
		return fmt.Errorf("スキーマバージョンの取得に失敗しました: %w", err)
	}

	for i, sqlStmt := range migrations {
		version := i + 1
		if version <= current {
			continue
		}
		if _, err := db.Exec(sqlStmt); err != nil {
			return fmt.Errorf("マイグレーション v%d の実行に失敗しました: %w", version, err)
		}
		if _, err := db.Exec(
			`INSERT INTO schema_versions (version, applied_at) VALUES (?, ?)`,
			version, time.Now().UTC().Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("スキーマバージョンの記録に失敗しました: %w", err)
		}
	}
	return nil
}
