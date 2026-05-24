import { Command } from 'commander';
import { getDb, now, shortId, findTaskByShortId, Task } from '../db/database';
import { success, error, info } from '../display/format';

export function registerAdd(program: Command): void {
  program
    .command('add <title>')
    .description('タスクを追加する')
    .option('--due <date>', '期限 (YYYY-MM-DD)')
    .option('--assign <name>', '担当者名')
    .option('--estimate <hours>', '見積もり時間')
    .option('--blocks <id>', 'ブロックするタスクID')
    .action((title: string, opts) => {
      const db = getDb();
      const id = crypto.randomUUID();
      const ts = now();

      // 担当者解決
      let assigneeId: string | null = null;
      if (opts.assign) {
        const member = db
          .prepare('SELECT * FROM members WHERE name LIKE ?')
          .get(`%${opts.assign}%`) as { id: string; name: string } | undefined;
        if (!member) {
          console.log(error(`メンバー「${opts.assign}」が見つかりません。先に devtask member add で追加してください。`));
          process.exit(1);
        }
        assigneeId = member.id;
      }

      // アクティブスプリント取得
      const sprint = db.prepare('SELECT * FROM sprints WHERE active = 1').get() as
        | { id: string }
        | undefined;

      db.prepare(`
        INSERT INTO tasks (id, title, assignee_id, status, due_date, estimate_hours, sprint_id, created_at, updated_at)
        VALUES (?, ?, ?, 'open', ?, ?, ?, ?, ?)
      `).run(
        id,
        title.trim(),
        assigneeId,
        opts.due ?? null,
        opts.estimate ? parseFloat(opts.estimate) : null,
        sprint?.id ?? null,
        ts,
        ts
      );

      // ブロック関係設定
      if (opts.blocks) {
        const blocker = findTaskByShortId(db, opts.blocks);
        if (!blocker) {
          console.log(error(`タスク「${opts.blocks}」が見つかりません。`));
          process.exit(1);
        }
        db.prepare('INSERT INTO task_blocks (blocker_id, blocked_id) VALUES (?, ?)').run(
          id,
          blocker.id
        );
        console.log(info(`「${blocker.title}」をブロックするタスクとして設定しました。`));
      }

      console.log(success(`タスクを追加しました: ${title}`));
      console.log(`  ID: ${shortId(id)}`);
      if (opts.due) console.log(`  期限: ${opts.due}`);
      if (opts.assign) console.log(`  担当: ${opts.assign}`);
      if (opts.estimate) console.log(`  見積: ${opts.estimate}h`);
    });
}
