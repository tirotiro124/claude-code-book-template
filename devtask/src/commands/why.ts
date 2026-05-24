import { Command } from 'commander';
import { getDb, shortId, findTaskByShortId, Sprint } from '../db/database';
import { calculateScore } from '../scorer/priority';
import { header, separator, formatScore, error } from '../display/format';
import chalk from 'chalk';

export function registerWhy(program: Command): void {
  program
    .command('why <id>')
    .description('タスクの優先度スコアの根拠を表示する')
    .action((id: string) => {
      const db = getDb();

      const task = findTaskByShortId(db, id);
      if (!task) {
        console.log(error(`タスク「${id}」が見つかりません。`));
        process.exit(1);
      }

      // view_count・last_viewed_at を更新
      db.prepare(`
        UPDATE tasks SET view_count = view_count + 1, last_viewed_at = ? WHERE id = ?
      `).run(new Date().toISOString(), task.id);

      const activeSprint = db.prepare('SELECT * FROM sprints WHERE active = 1').get() as
        | Sprint
        | undefined;

      const blockedCount = (
        db
          .prepare('SELECT COUNT(*) as cnt FROM task_blocks WHERE blocker_id = ?')
          .get(task.id) as { cnt: number }
      ).cnt;

      const breakdown = calculateScore(task, blockedCount, activeSprint ?? null);

      const taskTitle = chalk.bold.white(task.title);
      const sid = chalk.dim(shortId(task.id));
      console.log(header(`\nタスク ${sid}: ${taskTitle}`));
      console.log(separator(44));

      if (breakdown.components.length === 0) {
        console.log(chalk.gray('  スコアに寄与する要素がありません。'));
      } else {
        const maxPoints = Math.max(...breakdown.components.map((c) => c.points));
        const pointsWidth = String(maxPoints).length + 1;

        for (const c of breakdown.components) {
          const pts = chalk.yellow(`+${String(c.points).padStart(pointsWidth)}`);
          const label = chalk.white(c.label.padEnd(16));
          const reason = chalk.gray(c.reason);
          console.log(`  ${pts}  ${label}  ${reason}`);
        }
      }

      console.log(chalk.gray('  ' + '─'.repeat(40)));

      let verdict = '';
      if (breakdown.total >= 80) verdict = chalk.red('→ 今日の最優先');
      else if (breakdown.total >= 50) verdict = chalk.yellow('→ 早めに着手推奨');
      else verdict = chalk.gray('→ 通常優先度');

      console.log(`  合計: ${formatScore(breakdown.total)}  ${verdict}`);
      console.log();

      // ブロック先のタスク表示
      if (blockedCount > 0) {
        const blocked = db
          .prepare(`
            SELECT t.id, t.title, m.name as assignee_name
            FROM task_blocks tb
            JOIN tasks t ON tb.blocked_id = t.id
            LEFT JOIN members m ON t.assignee_id = m.id
            WHERE tb.blocker_id = ?
          `)
          .all(task.id) as { id: string; title: string; assignee_name: string | null }[];

        console.log(chalk.red(`  ブロックしているタスク (${blockedCount}件):`));
        for (const b of blocked) {
          const assignee = b.assignee_name ? chalk.cyan(` (${b.assignee_name})`) : '';
          console.log(`    ${chalk.dim(shortId(b.id))}  ${b.title}${assignee}`);
        }
        console.log();
      }
    });
}
