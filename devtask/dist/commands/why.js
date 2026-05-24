"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerWhy = registerWhy;
const database_1 = require("../db/database");
const priority_1 = require("../scorer/priority");
const format_1 = require("../display/format");
const chalk_1 = __importDefault(require("chalk"));
function registerWhy(program) {
    program
        .command('why <id>')
        .description('タスクの優先度スコアの根拠を表示する')
        .action((id) => {
        const db = (0, database_1.getDb)();
        const task = (0, database_1.findTaskByShortId)(db, id);
        if (!task) {
            console.log((0, format_1.error)(`タスク「${id}」が見つかりません。`));
            process.exit(1);
        }
        // view_count・last_viewed_at を更新
        db.prepare(`
        UPDATE tasks SET view_count = view_count + 1, last_viewed_at = ? WHERE id = ?
      `).run(new Date().toISOString(), task.id);
        const activeSprint = db.prepare('SELECT * FROM sprints WHERE active = 1').get();
        const blockedCount = db
            .prepare('SELECT COUNT(*) as cnt FROM task_blocks WHERE blocker_id = ?')
            .get(task.id).cnt;
        const breakdown = (0, priority_1.calculateScore)(task, blockedCount, activeSprint ?? null);
        const taskTitle = chalk_1.default.bold.white(task.title);
        const sid = chalk_1.default.dim((0, database_1.shortId)(task.id));
        console.log((0, format_1.header)(`\nタスク ${sid}: ${taskTitle}`));
        console.log((0, format_1.separator)(44));
        if (breakdown.components.length === 0) {
            console.log(chalk_1.default.gray('  スコアに寄与する要素がありません。'));
        }
        else {
            const maxPoints = Math.max(...breakdown.components.map((c) => c.points));
            const pointsWidth = String(maxPoints).length + 1;
            for (const c of breakdown.components) {
                const pts = chalk_1.default.yellow(`+${String(c.points).padStart(pointsWidth)}`);
                const label = chalk_1.default.white(c.label.padEnd(16));
                const reason = chalk_1.default.gray(c.reason);
                console.log(`  ${pts}  ${label}  ${reason}`);
            }
        }
        console.log(chalk_1.default.gray('  ' + '─'.repeat(40)));
        let verdict = '';
        if (breakdown.total >= 80)
            verdict = chalk_1.default.red('→ 今日の最優先');
        else if (breakdown.total >= 50)
            verdict = chalk_1.default.yellow('→ 早めに着手推奨');
        else
            verdict = chalk_1.default.gray('→ 通常優先度');
        console.log(`  合計: ${(0, format_1.formatScore)(breakdown.total)}  ${verdict}`);
        console.log();
        // ブロック先のタスク表示
        if (blockedCount > 0) {
            const blocked = db
                .prepare(`
            SELECT t.id, t.title, m.name as assignee_name
            FROM task_blocks tb
            JOIN tasks t ON tb.blocked_id = t.id
            LEFT JOIN members m ON t.assignee_id = m.id
            WHERE tb.blocker_id = ?
          `)
                .all(task.id);
            console.log(chalk_1.default.red(`  ブロックしているタスク (${blockedCount}件):`));
            for (const b of blocked) {
                const assignee = b.assignee_name ? chalk_1.default.cyan(` (${b.assignee_name})`) : '';
                console.log(`    ${chalk_1.default.dim((0, database_1.shortId)(b.id))}  ${b.title}${assignee}`);
            }
            console.log();
        }
    });
}
