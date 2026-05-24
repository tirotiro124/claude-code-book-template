const express = require('express');
const cors = require('cors');
const Database = require('better-sqlite3');
const path = require('path');

const PRIORITIES = ['high', 'medium', 'low'];

const db = new Database(path.join(__dirname, 'tasks.db'));

db.exec(`
  CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    text TEXT NOT NULL,
    done INTEGER NOT NULL DEFAULT 0,
    priority TEXT NOT NULL DEFAULT 'medium',
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
  )
`);

// migrate existing tables that lack the priority column
const cols = db.prepare("PRAGMA table_info(tasks)").all().map(c => c.name);
if (!cols.includes('priority')) {
  db.exec("ALTER TABLE tasks ADD COLUMN priority TEXT NOT NULL DEFAULT 'medium'");
}

const getAll = db.prepare('SELECT id, text, done, priority FROM tasks ORDER BY created_at ASC');
const insert = db.prepare("INSERT INTO tasks (text, priority) VALUES (?, ?)");
const toggle = db.prepare('UPDATE tasks SET done = CASE WHEN done = 0 THEN 1 ELSE 0 END WHERE id = ?');
const update = db.prepare('UPDATE tasks SET text = ?, priority = ? WHERE id = ?');
const remove = db.prepare('DELETE FROM tasks WHERE id = ?');

const app = express();
app.use(cors());
app.use(express.json());

function row(r) { return { ...r, done: r.done === 1 }; }

app.get('/tasks', (_req, res) => res.json(getAll.all().map(row)));

app.post('/tasks', (req, res) => {
  const text = (req.body.text || '').trim();
  if (!text) return res.status(400).json({ error: 'text is required' });
  const priority = PRIORITIES.includes(req.body.priority) ? req.body.priority : 'medium';
  const info = insert.run(text, priority);
  res.status(201).json({ id: info.lastInsertRowid, text, done: false, priority });
});

app.patch('/tasks/:id', (req, res) => {
  const text = (req.body.text || '').trim();
  if (!text) return res.status(400).json({ error: 'text is required' });
  const priority = PRIORITIES.includes(req.body.priority) ? req.body.priority : 'medium';
  const info = update.run(text, priority, Number(req.params.id));
  if (info.changes === 0) return res.status(404).json({ error: 'not found' });
  res.json({ ok: true });
});

app.patch('/tasks/:id/toggle', (req, res) => {
  const info = toggle.run(Number(req.params.id));
  if (info.changes === 0) return res.status(404).json({ error: 'not found' });
  res.json({ ok: true });
});

app.delete('/tasks/:id', (req, res) => {
  const info = remove.run(Number(req.params.id));
  if (info.changes === 0) return res.status(404).json({ error: 'not found' });
  res.json({ ok: true });
});

const PORT = process.env.PORT || 4000;
app.listen(PORT, () => console.log(`API listening on :${PORT}`));
