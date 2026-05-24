import Database from 'better-sqlite3';
import path from 'path';
import os from 'os';
import fs from 'fs';

export interface Task {
  id: string;
  title: string;
  description: string | null;
  assignee_id: string | null;
  status: 'open' | 'in_progress' | 'done';
  due_date: string | null;
  estimate_hours: number | null;
  snoozed_count: number;
  view_count: number;
  last_viewed_at: string | null;
  sprint_id: string | null;
  created_at: string;
  updated_at: string;
}

export interface TaskBlock {
  blocker_id: string;
  blocked_id: string;
}

export interface Member {
  id: string;
  name: string;
  email: string | null;
  created_at: string;
}

export interface Sprint {
  id: string;
  name: string;
  start_date: string;
  end_date: string;
  active: number;
}

const DB_DIR = path.join(os.homedir(), '.devtask');
const DB_PATH = path.join(DB_DIR, 'devtask.db');

let _db: Database.Database | null = null;

export function getDb(): Database.Database {
  if (_db) return _db;

  fs.mkdirSync(DB_DIR, { recursive: true });
  _db = new Database(DB_PATH);
  _db.pragma('journal_mode = WAL');
  migrate(_db);
  return _db;
}

function migrate(db: Database.Database): void {
  db.exec(`
    CREATE TABLE IF NOT EXISTS tasks (
      id TEXT PRIMARY KEY,
      title TEXT NOT NULL,
      description TEXT,
      assignee_id TEXT,
      status TEXT DEFAULT 'open',
      due_date TEXT,
      estimate_hours REAL,
      snoozed_count INTEGER DEFAULT 0,
      view_count INTEGER DEFAULT 0,
      last_viewed_at TEXT,
      sprint_id TEXT,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS task_blocks (
      blocker_id TEXT NOT NULL,
      blocked_id TEXT NOT NULL,
      PRIMARY KEY (blocker_id, blocked_id)
    );

    CREATE TABLE IF NOT EXISTS members (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      email TEXT,
      created_at TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS sprints (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      start_date TEXT NOT NULL,
      end_date TEXT NOT NULL,
      active INTEGER DEFAULT 0
    );
  `);
}

export function shortId(id: string): string {
  return id.slice(0, 8);
}

export function findTaskByShortId(db: Database.Database, shortId: string): Task | null {
  const tasks = db.prepare('SELECT * FROM tasks WHERE id LIKE ?').all(`${shortId}%`) as Task[];
  if (tasks.length === 0) return null;
  if (tasks.length > 1) {
    console.error(`複数のタスクが見つかりました。より長いIDを指定してください。`);
    process.exit(1);
  }
  return tasks[0];
}

export function now(): string {
  return new Date().toISOString();
}
