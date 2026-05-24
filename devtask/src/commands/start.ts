import { Command } from 'commander';
import { getDb, shortId, findTaskByShortId, now } from '../db/database';
import { success, error } from '../display/format';
import chalk from 'chalk';

export function registerStart(program: Command): void {
  program
    .command('start <id>')
    .description('タスクを進行中にする')
    .action((id: string) => {
      const db = getDb();

      const task = findTaskByShortId(db, id);
      if (!task) {
        console.log(error(`タスク「${id}」が見つかりません。`));
        process.exit(1);
      }

      if (task.status === 'in_progress') {
        console.log(chalk.gray(`タスク「${task.title}」はすでに進行中です。`));
        return;
      }

      db.prepare(`
        UPDATE tasks SET status = 'in_progress', last_viewed_at = ?, updated_at = ? WHERE id = ?
      `).run(now(), now(), task.id);

      console.log(success(`着手: ${task.title} (${shortId(task.id)})`));
      console.log();
    });
}
