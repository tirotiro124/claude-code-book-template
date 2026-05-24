"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerMember = registerMember;
const database_1 = require("../db/database");
const format_1 = require("../display/format");
const chalk_1 = __importDefault(require("chalk"));
function registerMember(program) {
    const member = program.command('member').description('チームメンバーを管理する');
    member
        .command('add <name>')
        .description('メンバーを追加する')
        .option('--email <email>', 'メールアドレス')
        .action((name, opts) => {
        const db = (0, database_1.getDb)();
        const existing = db.prepare('SELECT * FROM members WHERE name = ?').get(name.trim());
        if (existing) {
            console.log(chalk_1.default.gray(`メンバー「${name}」はすでに登録されています。`));
            return;
        }
        const id = crypto.randomUUID();
        db.prepare('INSERT INTO members (id, name, email, created_at) VALUES (?, ?, ?, ?)').run(id, name.trim(), opts.email ?? null, (0, database_1.now)());
        console.log((0, format_1.success)(`メンバーを追加しました: ${name}`));
        console.log(`  ID: ${(0, database_1.shortId)(id)}`);
        if (opts.email)
            console.log(`  Email: ${opts.email}`);
        console.log();
    });
    member
        .command('list')
        .description('メンバー一覧を表示する')
        .action(() => {
        const db = (0, database_1.getDb)();
        const members = db.prepare('SELECT * FROM members ORDER BY name').all();
        if (members.length === 0) {
            console.log((0, format_1.dim)('  メンバーが登録されていません。'));
            return;
        }
        console.log((0, format_1.header)('\nチームメンバー'));
        console.log((0, format_1.separator)(40));
        for (const m of members) {
            const taskCount = db
                .prepare(`SELECT COUNT(*) as cnt FROM tasks WHERE assignee_id = ? AND status != 'done'`)
                .get(m.id).cnt;
            const id = chalk_1.default.dim((0, database_1.shortId)(m.id));
            const name = chalk_1.default.white(m.name);
            const email = m.email ? chalk_1.default.gray(m.email) : '';
            const tasks = chalk_1.default.cyan(`${taskCount}件担当`);
            console.log(`  ${id}  ${name.padEnd(10)}  ${tasks.padEnd(8)}  ${email}`);
        }
        console.log();
    });
    member
        .command('remove <name>')
        .description('メンバーを削除する')
        .action((name) => {
        const db = (0, database_1.getDb)();
        const m = db
            .prepare('SELECT * FROM members WHERE name LIKE ?')
            .get(`%${name}%`);
        if (!m) {
            console.log((0, format_1.error)(`メンバー「${name}」が見つかりません。`));
            process.exit(1);
        }
        db.prepare('DELETE FROM members WHERE id = ?').run(m.id);
        console.log((0, format_1.success)(`メンバーを削除しました: ${m.name}`));
        console.log();
    });
}
