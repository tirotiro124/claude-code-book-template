"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.getDb = getDb;
exports.shortId = shortId;
exports.findTaskByShortId = findTaskByShortId;
exports.now = now;
const better_sqlite3_1 = __importDefault(require("better-sqlite3"));
const path_1 = __importDefault(require("path"));
const os_1 = __importDefault(require("os"));
const fs_1 = __importDefault(require("fs"));
const DB_DIR = path_1.default.join(os_1.default.homedir(), '.devtask');
const DB_PATH = path_1.default.join(DB_DIR, 'devtask.db');
let _db = null;
function getDb() {
    if (_db)
        return _db;
    fs_1.default.mkdirSync(DB_DIR, { recursive: true });
    _db = new better_sqlite3_1.default(DB_PATH);
    _db.pragma('journal_mode = WAL');
    migrate(_db);
    return _db;
}
function migrate(db) {
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
function shortId(id) {
    return id.slice(0, 8);
}
function findTaskByShortId(db, shortId) {
    const tasks = db.prepare('SELECT * FROM tasks WHERE id LIKE ?').all(`${shortId}%`);
    if (tasks.length === 0)
        return null;
    if (tasks.length > 1) {
        console.error(`複数のタスクが見つかりました。より長いIDを指定してください。`);
        process.exit(1);
    }
    return tasks[0];
}
function now() {
    return new Date().toISOString();
}
