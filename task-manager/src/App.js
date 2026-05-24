import { useState, useEffect, useCallback, useRef } from 'react';
import './App.css';

const API = 'http://localhost:4000';

const PRIORITY_LABELS = { high: '高', medium: '中', low: '低' };
const PRIORITY_ORDER  = ['high', 'medium', 'low'];

function nextPriority(p) {
  const i = PRIORITY_ORDER.indexOf(p);
  return PRIORITY_ORDER[(i + 1) % PRIORITY_ORDER.length];
}

function App() {
  const [tasks, setTasks] = useState([]);
  const [input, setInput] = useState('');
  const [priority, setPriority] = useState('medium');
  const [error, setError] = useState(null);
  const [editingId, setEditingId] = useState(null);
  const [editText, setEditText] = useState('');
  const [editPriority, setEditPriority] = useState('medium');
  const editRef = useRef(null);

  const fetchTasks = useCallback(async () => {
    try {
      const res = await fetch(`${API}/tasks`);
      if (!res.ok) throw new Error();
      setTasks(await res.json());
      setError(null);
    } catch {
      setError('サーバーに接続できません');
    }
  }, []);

  useEffect(() => { fetchTasks(); }, [fetchTasks]);
  useEffect(() => { if (editingId !== null) editRef.current?.focus(); }, [editingId]);

  async function addTask() {
    const text = input.trim();
    if (!text) return;
    const res = await fetch(`${API}/tasks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text, priority }),
    });
    if (res.ok) {
      const task = await res.json();
      setTasks(prev => [...prev, task]);
      setInput('');
      setPriority('medium');
    }
  }

  async function toggleTask(id) {
    const res = await fetch(`${API}/tasks/${id}/toggle`, { method: 'PATCH' });
    if (res.ok) setTasks(prev => prev.map(t => t.id === id ? { ...t, done: !t.done } : t));
  }

  async function deleteTask(id) {
    const res = await fetch(`${API}/tasks/${id}`, { method: 'DELETE' });
    if (res.ok) setTasks(prev => prev.filter(t => t.id !== id));
  }

  function startEdit(task) {
    setEditingId(task.id);
    setEditText(task.text);
    setEditPriority(task.priority);
  }

  async function commitEdit(id) {
    const text = editText.trim();
    if (!text) { cancelEdit(); return; }
    const res = await fetch(`${API}/tasks/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text, priority: editPriority }),
    });
    if (res.ok) setTasks(prev => prev.map(t => t.id === id ? { ...t, text, priority: editPriority } : t));
    cancelEdit();
  }

  function cancelEdit() { setEditingId(null); setEditText(''); }

  const remaining = tasks.filter(t => !t.done).length;

  return (
    <div className="app">
      <h1>タスク管理</h1>

      {error && <p className="error">{error}</p>}

      <div className="input-row">
        <input
          type="text"
          placeholder="新しいタスクを入力..."
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && addTask()}
          disabled={!!error}
        />
        <div className="priority-toggle-group">
          {PRIORITY_ORDER.map(p => (
            <button
              key={p}
              className={`priority-toggle priority-${p} ${priority === p ? 'active' : ''}`}
              onClick={() => setPriority(p)}
            >
              {PRIORITY_LABELS[p]}
            </button>
          ))}
        </div>
        <button className="add-btn" onClick={addTask} disabled={!!error}>追加</button>
      </div>

      {tasks.length > 0 && (
        <p className="summary">{tasks.length} 件中 {remaining} 件が未完了</p>
      )}

      <ul className="task-list">
        {tasks.map(task => (
          <li key={task.id} className={task.done ? 'done' : ''}>
            <button
              className="check-btn"
              onClick={() => toggleTask(task.id)}
              title={task.done ? '未完了に戻す' : '完了にする'}
            >
              {task.done ? '✓' : '○'}
            </button>

            {editingId === task.id ? (
              <>
                <input
                  ref={editRef}
                  className="edit-input"
                  value={editText}
                  onChange={e => setEditText(e.target.value)}
                  onKeyDown={e => { if (e.key === 'Enter') commitEdit(task.id); if (e.key === 'Escape') cancelEdit(); }}
                  onBlur={() => commitEdit(task.id)}
                />
                <div className="priority-toggle-group small">
                  {PRIORITY_ORDER.map(p => (
                    <button
                      key={p}
                      className={`priority-toggle priority-${p} ${editPriority === p ? 'active' : ''}`}
                      onMouseDown={e => { e.preventDefault(); setEditPriority(p); }}
                    >
                      {PRIORITY_LABELS[p]}
                    </button>
                  ))}
                </div>
                <button className="save-btn" onMouseDown={e => { e.preventDefault(); commitEdit(task.id); }} title="保存">✓</button>
              </>
            ) : (
              <>
                <span
                  className={`priority-badge priority-${task.priority}`}
                  onClick={() => !task.done && setTasks(prev => prev.map(t =>
                    t.id === task.id ? { ...t, priority: nextPriority(t.priority) } : t
                  ))}
                  title={task.done ? '' : 'クリックで優先度を変更'}
                >
                  {PRIORITY_LABELS[task.priority]}
                </span>
                <span
                  className="task-text"
                  onDoubleClick={() => !task.done && startEdit(task)}
                  title={task.done ? '' : 'ダブルクリックで編集'}
                >
                  {task.text}
                </span>
                <button className="edit-btn" onClick={() => startEdit(task)} disabled={task.done} title="編集">✎</button>
              </>
            )}

            <button className="delete-btn" onClick={() => deleteTask(task.id)} title="削除">✕</button>
          </li>
        ))}
      </ul>

      {!error && tasks.length === 0 && (
        <p className="empty">タスクがありません。上から追加してください。</p>
      )}
    </div>
  );
}

export default App;
