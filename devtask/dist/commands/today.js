"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerToday = registerToday;
const database_1 = require("../db/database");
const priority_1 = require("../scorer/priority");
const format_1 = require("../display/format");
const chalk_1 = __importDefault(require("chalk"));
const dayjs_1 = __importDefault(require("dayjs"));
function registerToday(program) {
    program
        .command('today')
        .description('優先度順のタスクを表示する（スコア上位5件）')
        .option('--assignee <name>', '担当者で絞り込み')
        .option('--limit <n>', '表示件数', '5')
        .action((opts) => {
        const db = (0, database_1.getDb)();
        const activeSprint = db.prepare('SELECT * FROM sprints WHERE active = 1').get();
        let sql = `
        SELECT t.*, m.name as assignee_name
        FROM tasks t
        LEFT JOIN members m ON t.assignee_id = m.id
        WHERE t.status != 'done'
      `;
        const params = [];
        if (opts.assignee) {
            sql += ` AND m.name LIKE ?`;
            params.push(`%${opts.assignee}%`);
        }
        const tasks = db.prepare(sql).all(...params);
        // スコア計算
        const scored = tasks.map((task) => {
            const blockedCount = db
                .prepare('SELECT COUNT(*) as cnt FROM task_blocks WHERE blocker_id = ?')
                .get(task.id).cnt;
            const breakdown = (0, priority_1.calculateScore)(task, blockedCount, activeSprint ?? null);
            return { task, breakdown, blockedCount };
        });
        scored.sort((a, b) => b.breakdown.total - a.breakdown.total);
        const limit = parseInt(opts.limit, 10);
        const top = scored.slice(0, limit);
        // ヘッダー
        let sprintInfo = '';
        if (activeSprint) {
            const daysLeft = (0, dayjs_1.default)(activeSprint.end_date).diff((0, dayjs_1.default)(), 'day');
            sprintInfo = chalk_1.default.gray(` | スプリント: ${activeSprint.name} 残り${daysLeft}日`);
        }
        console.log((0, format_1.header)(`\n今日のタスク${sprintInfo}`));
        console.log((0, format_1.separator)());
        if (top.length === 0) {
            console.log((0, format_1.dim)('  未完了のタスクがありません。'));
            console.log();
            return;
        }
        top.forEach(({ task, breakdown, blockedCount }, i) => {
            const rank = chalk_1.default.dim(`${i + 1}.`);
            const urgency = (0, format_1.formatUrgency)(breakdown.total);
            const id = chalk_1.default.dim((0, database_1.shortId)(task.id));
            const title = chalk_1.default.white(task.title.padEnd(24).slice(0, 24));
            const due = (0, format_1.formatDueDate)(task.due_date);
            const neglect = (0, format_1.formatNeglect)(task.last_viewed_at);
            const blocks = blockedCount > 0 ? chalk_1.default.red(`blocks:${blockedCount}`) : '';
            const assignee = task.assignee_name ? chalk_1.default.cyan(task.assignee_name) : chalk_1.default.gray('未割当');
            const score = (0, format_1.formatScore)(breakdown.total);
            const meta = [due, neglect, blocks].filter(Boolean).join('  ');
            console.log(`  ${rank.padEnd(3)} ${urgency} ${id}  ${title}  ${meta.padEnd(20)}  ${assignee.padEnd(6)}  ${score}`);
        });
        console.log();
    });
}
