import { Command } from 'commander';
import { getDb, shortId, Task, Sprint } from '../db/database';
import { calculateScore } from '../scorer/priority';
import { formatDueDate, formatScore, formatUrgency, formatNeglect, header, separator, dim } from '../display/format';
import chalk from 'chalk';
import dayjs from 'dayjs';

export function registerToday(program: Command): void {
  program
    .command('today')
    .description('優先度順のタスクを表示する（スコア上位5件）')
    .option('--assignee <name>', '担当者で絞り込み')
    .option('--limit <n>', '表示件数', '5')
    .action((opts) => {
      const db = getDb();

      const activeSprint = db.prepare('SELECT * FROM sprints WHERE active = 1').get() as
        | Sprint
        | undefined;

      let sql = `
        SELECT t.*, m.name as assignee_name
        FROM tasks t
        LEFT JOIN members m ON t.assignee_id = m.id
        WHERE t.status != 'done'
      `;
      const params: string[] = [];

      if (opts.assignee) {
        sql += ` AND m.name LIKE ?`;
        params.push(`%${opts.assignee}%`);
      }

      const tasks = db.prepare(sql).all(...params) as (Task & { assignee_name: string | null })[];

      // スコア計算
      const scored = tasks.map((task) => {
        const blockedCount = (
          db
            .prepare('SELECT COUNT(*) as cnt FROM task_blocks WHERE blocker_id = ?')
            .get(task.id) as { cnt: number }
        ).cnt;
        const breakdown = calculateScore(task, blockedCount, activeSprint ?? null);
        return { task, breakdown, blockedCount };
      });

      scored.sort((a, b) => b.breakdown.total - a.breakdown.total);

      const limit = parseInt(opts.limit, 10);
      const top = scored.slice(0, limit);

      // ヘッダー
      let sprintInfo = '';
      if (activeSprint) {
        const daysLeft = dayjs(activeSprint.end_date).diff(dayjs(), 'day');
        sprintInfo = chalk.gray(` | スプリント: ${activeSprint.name} 残り${daysLeft}日`);
      }
      console.log(header(`\n今日のタスク${sprintInfo}`));
      console.log(separator());

      if (top.length === 0) {
        console.log(dim('  未完了のタスクがありません。'));
        console.log();
        return;
      }

      top.forEach(({ task, breakdown, blockedCount }, i) => {
        const rank = chalk.dim(`${i + 1}.`);
        const urgency = formatUrgency(breakdown.total);
        const id = chalk.dim(shortId(task.id));
        const title = chalk.white(task.title.padEnd(24).slice(0, 24));
        const due = formatDueDate(task.due_date);
        const neglect = formatNeglect(task.last_viewed_at);
        const blocks =
          blockedCount > 0 ? chalk.red(`blocks:${blockedCount}`) : '';
        const assignee = task.assignee_name ? chalk.cyan(task.assignee_name) : chalk.gray('未割当');
        const score = formatScore(breakdown.total);

        const meta = [due, neglect, blocks].filter(Boolean).join('  ');
        console.log(
          `  ${rank.padEnd(3)} ${urgency} ${id}  ${title}  ${meta.padEnd(20)}  ${assignee.padEnd(6)}  ${score}`
        );
      });

      console.log();
    });
}
