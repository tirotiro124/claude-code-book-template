import { Command } from 'commander';
import { getDb, shortId, Task, Member } from '../db/database';
import { formatDueDate, formatStatus, header, separator, dim } from '../display/format';
import chalk from 'chalk';

export function registerList(program: Command): void {
  program
    .command('list')
    .description('タスク一覧を表示する')
    .option('--assignee <name>', '担当者で絞り込み')
    .option('--status <status>', 'ステータスで絞り込み (open|in_progress|done)')
    .option('--all', '完了済みも含めて表示')
    .action((opts) => {
      const db = getDb();

      let sql = `
        SELECT t.*, m.name as assignee_name
        FROM tasks t
        LEFT JOIN members m ON t.assignee_id = m.id
        WHERE 1=1
      `;
      const params: (string | null)[] = [];

      if (!opts.all && !opts.status) {
        sql += ` AND t.status != 'done'`;
      }
      if (opts.status) {
        sql += ` AND t.status = ?`;
        params.push(opts.status);
      }
      if (opts.assignee) {
        sql += ` AND m.name LIKE ?`;
        params.push(`%${opts.assignee}%`);
      }

      sql += ` ORDER BY t.created_at DESC`;

      const tasks = db.prepare(sql).all(...params) as (Task & { assignee_name: string | null })[];

      if (tasks.length === 0) {
        console.log(dim('タスクがありません。'));
        return;
      }

      console.log(header(`\nタスク一覧 (${tasks.length}件)`));
      console.log(separator());

      for (const task of tasks) {
        const id = chalk.dim(shortId(task.id));
        const title =
          task.status === 'done' ? chalk.strikethrough.gray(task.title) : chalk.white(task.title);
        const status = formatStatus(task.status);
        const due = formatDueDate(task.due_date);
        const assignee = task.assignee_name ? chalk.cyan(task.assignee_name) : '';
        const estimate = task.estimate_hours ? chalk.gray(`${task.estimate_hours}h`) : '';

        const parts = [id, title, status, due, assignee, estimate].filter(Boolean);
        console.log('  ' + parts.join('  '));
      }
      console.log();
    });
}
