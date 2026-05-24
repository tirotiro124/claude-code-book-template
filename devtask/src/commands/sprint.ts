import { Command } from 'commander';
import { getDb, shortId, now, Sprint } from '../db/database';
import { success, error, header, separator, dim, formatStatus } from '../display/format';
import chalk from 'chalk';
import dayjs from 'dayjs';

export function registerSprint(program: Command): void {
  const sprint = program.command('sprint').description('スプリントを管理する');

  sprint
    .command('create <name>')
    .description('スプリントを作成する')
    .requiredOption('--start <date>', '開始日 (YYYY-MM-DD)')
    .requiredOption('--end <date>', '終了日 (YYYY-MM-DD)')
    .action((name: string, opts) => {
      const db = getDb();

      const id = crypto.randomUUID();
      db.prepare(
        'INSERT INTO sprints (id, name, start_date, end_date, active) VALUES (?, ?, ?, ?, 0)'
      ).run(id, name.trim(), opts.start, opts.end);

      console.log(success(`スプリントを作成しました: ${name}`));
      console.log(`  ID: ${shortId(id)}`);
      console.log(`  期間: ${opts.start} 〜 ${opts.end}`);
      console.log(chalk.gray('  ヒント: devtask sprint activate で有効化できます。'));
      console.log();
    });

  sprint
    .command('activate <id>')
    .description('スプリントをアクティブにする')
    .action((id: string) => {
      const db = getDb();

      const target = db
        .prepare('SELECT * FROM sprints WHERE id LIKE ?')
        .get(`${id}%`) as Sprint | undefined;

      if (!target) {
        console.log(error(`スプリント「${id}」が見つかりません。`));
        process.exit(1);
      }

      // 他を非アクティブ化
      db.prepare('UPDATE sprints SET active = 0').run();
      db.prepare('UPDATE sprints SET active = 1 WHERE id = ?').run(target.id);

      console.log(success(`スプリントをアクティブにしました: ${target.name}`));
      console.log(`  期間: ${target.start_date} 〜 ${target.end_date}`);
      console.log();
    });

  sprint
    .command('list')
    .description('スプリント一覧を表示する')
    .action(() => {
      const db = getDb();
      const sprints = db
        .prepare('SELECT * FROM sprints ORDER BY start_date DESC')
        .all() as Sprint[];

      if (sprints.length === 0) {
        console.log(dim('  スプリントが登録されていません。'));
        return;
      }

      console.log(header('\nスプリント一覧'));
      console.log(separator(52));

      for (const s of sprints) {
        const id = chalk.dim(shortId(s.id));
        const name = chalk.white(s.name);
        const period = chalk.gray(`${s.start_date} 〜 ${s.end_date}`);
        const active = s.active ? chalk.green('[アクティブ]') : chalk.gray('');
        const daysLeft = dayjs(s.end_date).diff(dayjs(), 'day');
        const remaining =
          s.active && daysLeft >= 0
            ? chalk.yellow(`残り${daysLeft}日`)
            : daysLeft < 0
              ? chalk.red('終了')
              : '';

        console.log(`  ${id}  ${name.padEnd(12)}  ${period}  ${active}  ${remaining}`);
      }
      console.log();
    });
}
