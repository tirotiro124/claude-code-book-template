import { Command } from 'commander';
import { getDb, shortId, findTaskByShortId } from '../db/database';
import { success, error, info } from '../display/format';
import chalk from 'chalk';

export function registerBlock(program: Command): void {
  program
    .command('block <id>')
    .description('タスク間のブロック関係を設定する')
    .requiredOption('--blocks <id>', 'ブロックされるタスクID')
    .action((id: string, opts) => {
      const db = getDb();

      const blocker = findTaskByShortId(db, id);
      if (!blocker) {
        console.log(error(`タスク「${id}」が見つかりません。`));
        process.exit(1);
      }

      const blocked = findTaskByShortId(db, opts.blocks);
      if (!blocked) {
        console.log(error(`タスク「${opts.blocks}」が見つかりません。`));
        process.exit(1);
      }

      if (blocker.id === blocked.id) {
        console.log(error('同じタスクをブロック対象に指定できません。'));
        process.exit(1);
      }

      // 既存チェック
      const existing = db
        .prepare('SELECT 1 FROM task_blocks WHERE blocker_id = ? AND blocked_id = ?')
        .get(blocker.id, blocked.id);
      if (existing) {
        console.log(chalk.gray('すでにブロック関係が設定されています。'));
        return;
      }

      // 循環チェック（blockedがblockerをブロックしていないか）
      const circular = db
        .prepare('SELECT 1 FROM task_blocks WHERE blocker_id = ? AND blocked_id = ?')
        .get(blocked.id, blocker.id);
      if (circular) {
        console.log(error('循環するブロック関係は設定できません。'));
        process.exit(1);
      }

      db.prepare('INSERT INTO task_blocks (blocker_id, blocked_id) VALUES (?, ?)').run(
        blocker.id,
        blocked.id
      );

      console.log(success('ブロック関係を設定しました'));
      console.log(
        `  ${chalk.white(blocker.title)} (${shortId(blocker.id)}) → ブロック → ${chalk.white(blocked.title)} (${shortId(blocked.id)})`
      );
      console.log(info(`「${blocker.title}」が完了するまで「${blocked.title}」はブロックされます。`));
      console.log();
    });
}
