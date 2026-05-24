import { Command } from 'commander';
import { getDb, Member, Task } from '../db/database';
import { loadBar, header, separator, warn, dim } from '../display/format';
import chalk from 'chalk';

const OVERLOAD_TASKS = 6;
const OVERLOAD_HOURS = 12;

export function registerTeam(program: Command): void {
  program
    .command('team')
    .description('チームの負荷状況を表示する')
    .action(() => {
      const db = getDb();

      const members = db.prepare('SELECT * FROM members ORDER BY name').all() as Member[];

      if (members.length === 0) {
        console.log(dim('  メンバーが登録されていません。devtask member add で追加してください。'));
        return;
      }

      const stats = members.map((m) => {
        const tasks = db
          .prepare(`SELECT * FROM tasks WHERE assignee_id = ? AND status != 'done'`)
          .all(m.id) as Task[];
        const totalHours = tasks.reduce((sum, t) => sum + (t.estimate_hours ?? 0), 0);
        return { member: m, taskCount: tasks.length, totalHours };
      });

      // 未割当タスク
      const unassigned = (
        db
          .prepare(`SELECT COUNT(*) as cnt FROM tasks WHERE assignee_id IS NULL AND status != 'done'`)
          .get() as { cnt: number }
      ).cnt;

      const maxTasks = Math.max(...stats.map((s) => s.taskCount), 1);

      console.log(header('\nチームの負荷状況'));
      console.log(separator(52));

      for (const s of stats) {
        const bar = loadBar(s.taskCount, Math.max(maxTasks, OVERLOAD_TASKS));
        const nameStr = chalk.white(s.member.name.padEnd(6).slice(0, 6));
        const taskStr = chalk.gray(`${s.taskCount}タスク`);
        const hourStr =
          s.totalHours > 0 ? chalk.gray(`推定${s.totalHours}h`) : chalk.gray('見積なし');

        let annotation = '';
        if (s.taskCount >= OVERLOAD_TASKS || s.totalHours >= OVERLOAD_HOURS) {
          annotation = chalk.red(' ⚠ オーバーロード');
        } else if (s.taskCount === 0) {
          annotation = chalk.green(' ← 余力あり');
        }

        console.log(`  ${nameStr}  ${bar}  ${taskStr.padEnd(8)}  ${hourStr.padEnd(8)}${annotation}`);
      }

      if (unassigned > 0) {
        console.log(separator(52));
        console.log(`  ${chalk.yellow(`未割当タスク: ${unassigned}件`)}`);
      }

      console.log();
    });
}
