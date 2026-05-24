"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerSprint = registerSprint;
const database_1 = require("../db/database");
const format_1 = require("../display/format");
const chalk_1 = __importDefault(require("chalk"));
const dayjs_1 = __importDefault(require("dayjs"));
function registerSprint(program) {
    const sprint = program.command('sprint').description('スプリントを管理する');
    sprint
        .command('create <name>')
        .description('スプリントを作成する')
        .requiredOption('--start <date>', '開始日 (YYYY-MM-DD)')
        .requiredOption('--end <date>', '終了日 (YYYY-MM-DD)')
        .action((name, opts) => {
        const db = (0, database_1.getDb)();
        const id = crypto.randomUUID();
        db.prepare('INSERT INTO sprints (id, name, start_date, end_date, active) VALUES (?, ?, ?, ?, 0)').run(id, name.trim(), opts.start, opts.end);
        console.log((0, format_1.success)(`スプリントを作成しました: ${name}`));
        console.log(`  ID: ${(0, database_1.shortId)(id)}`);
        console.log(`  期間: ${opts.start} 〜 ${opts.end}`);
        console.log(chalk_1.default.gray('  ヒント: devtask sprint activate で有効化できます。'));
        console.log();
    });
    sprint
        .command('activate <id>')
        .description('スプリントをアクティブにする')
        .action((id) => {
        const db = (0, database_1.getDb)();
        const target = db
            .prepare('SELECT * FROM sprints WHERE id LIKE ?')
            .get(`${id}%`);
        if (!target) {
            console.log((0, format_1.error)(`スプリント「${id}」が見つかりません。`));
            process.exit(1);
        }
        // 他を非アクティブ化
        db.prepare('UPDATE sprints SET active = 0').run();
        db.prepare('UPDATE sprints SET active = 1 WHERE id = ?').run(target.id);
        console.log((0, format_1.success)(`スプリントをアクティブにしました: ${target.name}`));
        console.log(`  期間: ${target.start_date} 〜 ${target.end_date}`);
        console.log();
    });
    sprint
        .command('list')
        .description('スプリント一覧を表示する')
        .action(() => {
        const db = (0, database_1.getDb)();
        const sprints = db
            .prepare('SELECT * FROM sprints ORDER BY start_date DESC')
            .all();
        if (sprints.length === 0) {
            console.log((0, format_1.dim)('  スプリントが登録されていません。'));
            return;
        }
        console.log((0, format_1.header)('\nスプリント一覧'));
        console.log((0, format_1.separator)(52));
        for (const s of sprints) {
            const id = chalk_1.default.dim((0, database_1.shortId)(s.id));
            const name = chalk_1.default.white(s.name);
            const period = chalk_1.default.gray(`${s.start_date} 〜 ${s.end_date}`);
            const active = s.active ? chalk_1.default.green('[アクティブ]') : chalk_1.default.gray('');
            const daysLeft = (0, dayjs_1.default)(s.end_date).diff((0, dayjs_1.default)(), 'day');
            const remaining = s.active && daysLeft >= 0
                ? chalk_1.default.yellow(`残り${daysLeft}日`)
                : daysLeft < 0
                    ? chalk_1.default.red('終了')
                    : '';
            console.log(`  ${id}  ${name.padEnd(12)}  ${period}  ${active}  ${remaining}`);
        }
        console.log();
    });
}
