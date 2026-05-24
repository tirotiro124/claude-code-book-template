package db

// schema_versions テーブル自体はブートストラップ処理で作成する。ここには含めない。
var migrations = []string{
	// v1: 初期スキーマ
	`CREATE TABLE tasks (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		title         TEXT NOT NULL,
		status        TEXT NOT NULL DEFAULT 'todo'
		                   CHECK(status IN ('todo','in_progress','done','snoozed')),
		due_date      TEXT,
		branch        TEXT,
		tags          TEXT NOT NULL DEFAULT '[]',
		pinned        INTEGER NOT NULL DEFAULT 0 CHECK(pinned IN (0,1)),
		snoozed_until TEXT,
		source_file   TEXT,
		source_line   INTEGER,
		created_at    TEXT NOT NULL,
		updated_at    TEXT NOT NULL
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
