"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerStandup = registerStandup;
const database_1 = require("../db/database");
const priority_1 = require("../scorer/priority");
const format_1 = require("../display/format");
const chalk_1 = __importDefault(require("chalk"));
const dayjs_1 = __importDefault(require("dayjs"));
function registerStandup(program) {
    program
        .command('standup')
        .description('スタンドアップレポートを生成する')
        .option('--assignee <name>', '担当者名')
        .action((opts) => {
        const db = (0, database_1.getDb)();
        // 担当者解決
        let memberName = opts.assignee ?? null;
        let memberId = null;
        if (opts.assignee) {
            const member = db
                .prepare('SELECT * FROM members WHERE name LIKE ?')
                .get(`%${opts.assignee}%`);
            if (!member) {
                console.log(chalk_1.default.red(`メンバー「${opts.assignee}」が見つかりません。`));
                process.exit(1);
            }
            memberId = member.id;
            memberName = member.name;
        }
        // 昨日完了したタスク（updated_atが昨日以降かつdone）
        const yesterday = (0, dayjs_1.default)().subtract(1, 'day').startOf('day').toISOString();
        let completedSql = `SELECT * FROM tasks WHERE status = 'done' AND updated_at >= ?`;
        const completedParams = [yesterday];
        if (memberId) {
            completedSql += ` AND assignee_id = ?`;
            completedParams.push(memberId);
        }
        const completed = db.prepare(completedSql).all(...completedParams);
        // 今日着手予定（未完了・スコア上位3件）
        const activeSprint = db.prepare('SELECT * FROM sprints WHERE active = 1').get();
        let todaySql = `SELECT * FROM tasks WHERE status != 'done'`;
        const todayParams = [];
        if (memberId) {
            todaySql += ` AND assignee_id = ?`;
            todayParams.push(memberId);
        }
        const openTasks = db.prepare(todaySql).all(...todayParams);
        const scored = openTasks
            .map((task) => {
            const blockedCount = db
                .prepare('SELECT COUNT(*) as cnt FROM task_blocks WHERE blocker_id = ?')
                .get(task.id).cnt;
            const breakdown = (0, priority_1.calculateScore)(task, blockedCount, activeSprint ?? null);
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
        const blockersParams = [];
        if (memberId) {
            blockersSql += ` AND t.assignee_id = ?`;
            blockersParams.push(memberId);
        }
        const blockers = db.prepare(blockersSql).all(...blockersParams);
        // 出力
        const title = memberName ? `=== スタンドアップ (${memberName}) ===` : '=== スタンドアップ ===';
        console.log((0, format_1.header)(`\n${title}`));
        console.log((0, format_1.separator)(44));
        console.log(chalk_1.default.bold('\n昨日完了:'));
        if (completed.length === 0) {
            console.log(chalk_1.default.gray('  なし'));
        }
        else {
            for (const t of completed) {
                console.log(`  ${chalk_1.default.green('✓')} ${t.title} (${(0, database_1.shortId)(t.id)})`);
            }
        }
        console.log(chalk_1.default.bold('\n今日の予定:'));
        if (scored.length === 0) {
            console.log(chalk_1.default.gray('  なし'));
        }
        else {
            for (const { task, score } of scored) {
                console.log(`  ${chalk_1.default.cyan('→')} ${task.title} (${(0, database_1.shortId)(task.id)}) ${chalk_1.default.dim(`[score: ${score}]`)}`);
            }
        }
        console.log(chalk_1.default.bold('\nブロッカー:'));
        if (blockers.length === 0) {
            console.log(chalk_1.default.gray('  なし'));
        }
        else {
            for (const b of blockers) {
                console.log(`  ${chalk_1.default.red('!')} ${b.title} (${(0, database_1.shortId)(b.id)})`);
            }
        }
        console.log();
    });
}
