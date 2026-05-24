import { Command } from 'commander';
import { getDb, shortId, now, Member } from '../db/database';
import { success, error, header, separator, dim } from '../display/format';
import chalk from 'chalk';

export function registerMember(program: Command): void {
  const member = program.command('member').description('チームメンバーを管理する');

  member
    .command('add <name>')
    .description('メンバーを追加する')
    .option('--email <email>', 'メールアドレス')
    .action((name: string, opts) => {
      const db = getDb();

      const existing = db.prepare('SELECT * FROM members WHERE name = ?').get(name.trim());
      if (existing) {
        console.log(chalk.gray(`メンバー「${name}」はすでに登録されています。`));
        return;
      }

      const id = crypto.randomUUID();
      db.prepare('INSERT INTO members (id, name, email, created_at) VALUES (?, ?, ?, ?)').run(
        id,
        name.trim(),
        opts.email ?? null,
        now()
      );

      console.log(success(`メンバーを追加しました: ${name}`));
      console.log(`  ID: ${shortId(id)}`);
      if (opts.email) console.log(`  Email: ${opts.email}`);
      console.log();
    });

  member
    .command('list')
    .description('メンバー一覧を表示する')
    .action(() => {
      const db = getDb();
      const members = db.prepare('SELECT * FROM members ORDER BY name').all() as Member[];

      if (members.length === 0) {
        console.log(dim('  メンバーが登録されていません。'));
        return;
      }

      console.log(header('\nチームメンバー'));
      console.log(separator(40));

      for (const m of members) {
        const taskCount = (
          db
            .prepare(`SELECT COUNT(*) as cnt FROM tasks WHERE assignee_id = ? AND status != 'done'`)
            .get(m.id) as { cnt: number }
        ).cnt;

        const id = chalk.dim(shortId(m.id));
        const name = chalk.white(m.name);
        const email = m.email ? chalk.gray(m.email) : '';
        const tasks = chalk.cyan(`${taskCount}件担当`);

        console.log(`  ${id}  ${name.padEnd(10)}  ${tasks.padEnd(8)}  ${email}`);
      }
      console.log();
    });

  member
    .command('remove <name>')
    .description('メンバーを削除する')
    .action((name: string) => {
      const db = getDb();
      const m = db
        .prepare('SELECT * FROM members WHERE name LIKE ?')
        .get(`%${name}%`) as Member | undefined;

      if (!m) {
        console.log(error(`メンバー「${name}」が見つかりません。`));
        process.exit(1);
      }

      db.prepare('DELETE FROM members WHERE id = ?').run(m.id);
      console.log(success(`メンバーを削除しました: ${m.name}`));
      console.log();
    });
}
