import { Command } from 'commander';
import { getDb, shortId, findTaskByShortId, now } from '../db/database';
import { success, info, error } from '../display/format';
import chalk from 'chalk';

export function registerDone(program: Command): void {
  program
    .command('done <id>')
    .description('タスクを完了にする')
    .action((id: string) => {
      const db = getDb();

      const task = findTaskByShortId(db, id);
      if (!task) {
        console.log(error(`タスク「${id}」が見つかりません。`));
        process.exit(1);
      }

      if (task.status === 'done') {
        console.log(chalk.gray(`タスク「${task.title}」はすでに完了しています。`));
        return;
      }

      db.prepare(`UPDATE tasks SET status = 'done', updated_at = ? WHERE id = ?`).run(
        now(),
        task.id
      );

      console.log(success(`完了: ${task.title} (${shortId(task.id)})`));

      // このタスクをブロックしていた→このタスクが完了したことで解除されるタスク
      const unblocked = db
        .prepare(`
          SELECT t.id, t.title
          FROM task_blocks tb
          JOIN tasks t ON tb.blocked_id = t.id
          WHERE tb.blocker_id = ?
        `)
        .all(task.id) as { id: string; title: string }[];

      if (unblocked.length > 0) {
        console.log();
        console.log(info('ブロック解除されたタスク:'));
        for (const u of unblocked) {
          console.log(`  ${chalk.green('✓')} ${u.title} (${shortId(u.id)})`);
        }
      }

      console.log();
    });
}
