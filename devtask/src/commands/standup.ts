import { Command } from 'commander';
import { getDb, shortId, Task, Sprint } from '../db/database';
import { calculateScore } from '../scorer/priority';
import { header, separator } from '../display/format';
import chalk from 'chalk';
import dayjs from 'dayjs';

export function registerStandup(program: Command): void {
  program
    .command('standup')
    .description('スタンドアップレポートを生成する')
    .option('--assignee <name>', '担当者名')
    .action((opts) => {
      const db = getDb();

      // 担当者解決
      let memberName = opts.assignee ?? null;
      let memberId: string | null = null;

      if (opts.assignee) {
        const member = db
          .prepare('SELECT * FROM members WHERE name LIKE ?')
          .get(`%${opts.assignee}%`) as { id: string; name: string } | undefined;
        if (!member) {
          console.log(chalk.red(`メンバー「${opts.assignee}」が見つかりません。`));
          process.exit(1);
        }
        memberId = member.id;
        memberName = member.name;
      }

      // 昨日完了したタスク（updated_atが昨日以降かつdone）
      const yesterday = dayjs().subtract(1, 'day').startOf('day').toISOString();
      let completedSql = `SELECT * FROM tasks WHERE status = 'done' AND updated_at >= ?`;
      const completedParams: string[] = [yesterday];
      if (memberId) {
        completedSql += ` AND assignee_id = ?`;
        completedParams.push(memberId);
      }
      const completed = db.prepare(completedSql).all(...completedParams) as Task[];

      // 今日着手予定（未完了・スコア上位3件）
      const activeSprint = db.prepare('SELECT * FROM sprints WHERE active = 1').get() as
        | Sprint
        | undefined;

      let todaySql = `SELECT * FROM tasks WHERE status != 'done'`;
      const todayParams: string[] = [];
      if (memberId) {
        todaySql += ` AND assignee_id = ?`;
        todayParams.push(memberId);
      }
      const openTasks = db.prepare(todaySql).all(...todayParams) as Task[];

      const scored = openTasks
        .map((task) => {
          const blockedCount = (
            db
              .prepare('SELECT COUNT(*) as cnt FROM task_blocks WHERE blocker_id = ?')
              .get(task.id) as { cnt: number }
          ).cnt;
          const breakdown = calculateScore(task, blockedCount, activeSprint ?? null);
          return { task, score: breakdown.total };
        })
        .sort((a, b) => b.score - a.score)
        .slice(0, 3);

      // ブロッカー（このメンバーの未完了タスクがブロックされているか）
      let blockersSql = `
        SELECT DISTINCT bt.title, bt.id
        FROM task_blocks tb
        JOIN tasks bt ON tb.blocker_id = bt.id
        JOIN tasks t ON tb.blocked_id = t.id
        WHERE bt.status != 'done' AND t.status != 'done'
      `;
      const blockersParams: string[] = [];
      if (memberId) {
        blockersSql += ` AND t.assignee_id = ?`;
        blockersParams.push(memberId);
      }
      const blockers = db.prepare(blockersSql).all(...blockersParams) as {
        title: string;
        id: string;
      }[];

      // 出力
      const title = memberName ? `=== スタンドアップ (${memberName}) ===` : '=== スタンドアップ ===';
      console.log(header(`\n${title}`));
      console.log(separator(44));

      console.log(chalk.bold('\n昨日完了:'));
      if (completed.length === 0) {
        console.log(chalk.gray('  なし'));
      } else {
        for (const t of completed) {
          console.log(`  ${chalk.green('✓')} ${t.title} (${shortId(t.id)})`);
        }
      }

      console.log(chalk.bold('\n今日の予定:'));
      if (scored.length === 0) {
        console.log(chalk.gray('  なし'));
      } else {
        for (const { task, score } of scored) {
          console.log(
            `  ${chalk.cyan('→')} ${task.title} (${shortId(task.id)}) ${chalk.dim(`[score: ${score}]`)}`
          );
        }
      }

      console.log(chalk.bold('\nブロッカー:'));
      if (blockers.length === 0) {
        console.log(chalk.gray('  なし'));
      } else {
        for (const b of blockers) {
          console.log(`  ${chalk.red('!')} ${b.title} (${shortId(b.id)})`);
        }
      }

      console.log();
    });
}
