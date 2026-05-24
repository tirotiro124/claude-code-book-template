package db

import (
	"encoding/json"
	"fmt"
	"time"
)

// Status はタスクのステータス型。
type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusSnoozed    Status = "snoozed"
)

// Task はタスクのドメインオブジェクト。
type Task struct {
	ID           int64
	Title        string
	Status       Status
	DueDate      *time.Time
	Branch       string
	Tags         []string
	Pinned       bool
	SnoozedUntil *time.Time
	SourceFile   string
	SourceLine   int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TaskFilter はタスク一覧取得時のフィルタ条件。
type TaskFilter struct {
	Statuses []Status // 空の場合は todo / in_progress のみ
	Branch   string
	Tag      string
}

// CreateTask はタスクを新規作成して返す。
func (db *DB) CreateTask(t Task) (Task, error) {
	now := time.Now().UTC()
	tags, err := json.Marshal(t.Tags)
	if err != nil {
		return Task{}, fmt.Errorf("タグのシリアライズに失敗しました: %w", err)
	}

	res, err := db.Exec(
		`INSERT INTO tasks
			(title, status, due_date, branch, tags, pinned, snoozed_until,
			 source_file, source_line, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.Title,
		string(StatusTodo),
		formatDate(t.DueDate),
		t.Branch,
		string(tags),
		boolToInt(t.Pinned),
		formatDate(t.SnoozedUntil),
		t.SourceFile,
		t.SourceLine,
		now.Format(time.RFC3339),
		now.Format(time.RFC3339),
	)
	if err != nil {
		return Task{}, fmt.Errorf("タスクの作成に失敗しました: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Task{}, fmt.Errorf("タスクIDの取得に失敗しました: %w", err)
	}

	return db.GetTask(id)
}

// GetTask は指定IDのタスクを返す。
func (db *DB) GetTask(id int64) (Task, error) {
	row := db.QueryRow(`SELECT * FROM tasks WHERE id = ?`, id)
	t, err := scanTask(row)
	if err != nil {
		return Task{}, fmt.Errorf("タスク #%d が見つかりません: %w", id, ErrTaskNotFound)
	}
	return t, nil
}

// ListTasks はフィルタ条件に合うタスクを返す。
func (db *DB) ListTasks(filter TaskFilter) ([]Task, error) {
	statuses := filter.Statuses
	if len(statuses) == 0 {
		statuses = []Status{StatusTodo, StatusInProgress}
	}

	// IN 句のプレースホルダを動的に構築
	placeholders := make([]any, len(statuses))
	query := `SELECT * FROM tasks WHERE status IN (`
	for i, s := range statuses {
		if i > 0 {
			query += ","
		}
		query += "?"
		placeholders[i] = string(s)
	}
	query += ")"

	if filter.Branch != "" {
		query += " AND branch = ?"
		placeholders = append(placeholders, filter.Branch)
	}

	rows, err := db.Query(query, placeholders...)
	if err != nil {
		return nil, fmt.Errorf("タスク一覧の取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		if filter.Tag != "" && !containsTag(t.Tags, filter.Tag) {
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// UpdateTask はタスクを更新する。
func (db *DB) UpdateTask(t Task) error {
	tags, err := json.Marshal(t.Tags)
	if err != nil {
		return fmt.Errorf("タグのシリアライズに失敗しました: %w", err)
	}

	_, err = db.Exec(
		`UPDATE tasks SET
			title=?, status=?, due_date=?, branch=?, tags=?, pinned=?,
			snoozed_until=?, source_file=?, source_line=?, updated_at=?
		 WHERE id=?`,
		t.Title,
		string(t.Status),
		formatDate(t.DueDate),
		t.Branch,
		string(tags),
		boolToInt(t.Pinned),
		formatDate(t.SnoozedUntil),
		t.SourceFile,
		t.SourceLine,
		time.Now().UTC().Format(time.RFC3339),
		t.ID,
	)
	if err != nil {
		return fmt.Errorf("タスクの更新に失敗しました: %w", err)
	}
	return nil
}

// ReactivateSnoozed はスヌーズ期限切れのタスクを todo に戻す（遅延評価）。
func (db *DB) ReactivateSnoozed(now time.Time) error {
	_, err := db.Exec(
		`UPDATE tasks SET status='todo', snoozed_until=NULL, updated_at=?
		 WHERE status='snoozed' AND snoozed_until <= ?`,
		now.UTC().Format(time.RFC3339),
		now.UTC().Format("2006-01-02"),
	)
	if err != nil {
		return fmt.Errorf("スヌーズ解除に失敗しました: %w", err)
	}
	return nil
}

// BlockedByCount はタスクを待っているタスク数を返す（重要度計算用）。
func (db *DB) BlockedByCount(taskID int64) (int, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM task_blocks WHERE blocked_by_id = ?`, taskID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ブロック依存数の取得に失敗しました: %w", err)
	}
	return count, nil
}

// AddBlock はブロック依存を追加する（taskID は blockedByID の完了を待つ）。
func (db *DB) AddBlock(taskID, blockedByID int64) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO task_blocks (task_id, blocked_by_id) VALUES (?, ?)`,
		taskID, blockedByID,
	)
	if err != nil {
		return fmt.Errorf("ブロック依存の追加に失敗しました: %w", err)
	}
	return nil
}

// RemoveBlock はブロック依存を解除する。
func (db *DB) RemoveBlock(taskID, blockedByID int64) error {
	_, err := db.Exec(
		`DELETE FROM task_blocks WHERE task_id=? AND blocked_by_id=?`,
		taskID, blockedByID,
	)
	if err != nil {
		return fmt.Errorf("ブロック依存の解除に失敗しました: %w", err)
	}
	return nil
}

// RemoveAllBlocksBy はタスクが完了したとき、そのタスクを待っていた依存をすべて解除する。
func (db *DB) RemoveAllBlocksBy(blockedByID int64) error {
	_, err := db.Exec(
		`DELETE FROM task_blocks WHERE blocked_by_id=?`, blockedByID,
	)
	if err != nil {
		return fmt.Errorf("ブロック依存の一括解除に失敗しました: %w", err)
	}
	return nil
}

// RecordShown は task next の表示履歴を記録する。直近2件を超えたら古いものを削除する。
func (db *DB) RecordShown(taskID int64) error {
	_, err := db.Exec(
		`INSERT INTO display_history (task_id, displayed_at) VALUES (?, ?)`,
		taskID, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("表示履歴の記録に失敗しました: %w", err)
	}

	// 古いレコードを削除して直近2件のみ保持
	_, err = db.Exec(
		`DELETE FROM display_history WHERE id NOT IN
		 (SELECT id FROM display_history ORDER BY id DESC LIMIT 2)`,
	)
	if err != nil {
		return fmt.Errorf("表示履歴の整理に失敗しました: %w", err)
	}
	return nil
}

// GetPrerequisites はタスクが待っている依存タスク（前提タスク）を返す。
func (db *DB) GetPrerequisites(taskID int64) ([]Task, error) {
	rows, err := db.Query(
		`SELECT t.* FROM tasks t
		 INNER JOIN task_blocks tb ON t.id = tb.blocked_by_id
		 WHERE tb.task_id = ?`, taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("前提タスクの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// GetRecentShown は直近の表示履歴（最新順）のタスクIDを返す。
func (db *DB) GetRecentShown() ([]int64, error) {
	rows, err := db.Query(
		`SELECT task_id FROM display_history ORDER BY id DESC LIMIT 2`,
	)
	if err != nil {
		return nil, fmt.Errorf("表示履歴の取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// --- スキャン・ユーティリティ ---

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(s scanner) (Task, error) {
	var (
		t            Task
		dueDateStr   *string
		snoozeStr    *string
		tagsStr      string
		pinnedInt    int
		createdAtStr string
		updatedAtStr string
	)
	err := s.Scan(
		&t.ID, &t.Title, &t.Status,
		&dueDateStr, &t.Branch, &tagsStr, &pinnedInt,
		&snoozeStr, &t.SourceFile, &t.SourceLine,
		&createdAtStr, &updatedAtStr,
	)
	if err != nil {
		return Task{}, err
	}

	t.Pinned = pinnedInt == 1
	t.DueDate = parseDate(dueDateStr)
	t.SnoozedUntil = parseDate(snoozeStr)

	if err := json.Unmarshal([]byte(tagsStr), &t.Tags); err != nil {
		t.Tags = []string{}
	}
	if t.Tags == nil {
		t.Tags = []string{}
	}

	if ca, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
		t.CreatedAt = ca
	}
	if ua, err := time.Parse(time.RFC3339, updatedAtStr); err == nil {
		t.UpdatedAt = ua
	}

	return t, nil
}

func formatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

func parseDate(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil
	}
	return &t
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
