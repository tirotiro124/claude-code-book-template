"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerDone = registerDone;
const database_1 = require("../db/database");
const format_1 = require("../display/format");
const chalk_1 = __importDefault(require("chalk"));
function registerDone(program) {
    program
        .command('done <id>')
        .description('タスクを完了にする')
        .action((id) => {
        const db = (0, database_1.getDb)();
        const task = (0, database_1.findTaskByShortId)(db, id);
        if (!task) {
            console.log((0, format_1.error)(`タスク「${id}」が見つかりません。`));
            process.exit(1);
        }
        if (task.status === 'done') {
            console.log(chalk_1.default.gray(`タスク「${task.title}」はすでに完了しています。`));
            return;
        }
        db.prepare(`UPDATE tasks SET status = 'done', updated_at = ? WHERE id = ?`).run((0, database_1.now)(), task.id);
        console.log((0, format_1.success)(`完了: ${task.title} (${(0, database_1.shortId)(task.id)})`));
        // このタスクをブロックしていた→このタスクが完了したことで解除されるタスク
        const unblocked = db
            .prepare(`
          SELECT t.id, t.title
          FROM task_blocks tb
          JOIN tasks t ON tb.blocked_id = t.id
          WHERE tb.blocker_id = ?
        `)
            .all(task.id);
        if (unblocked.length > 0) {
            console.log();
            console.log((0, format_1.info)('ブロック解除されたタスク:'));
            for (const u of unblocked) {
                console.log(`  ${chalk_1.default.green('✓')} ${u.title} (${(0, database_1.shortId)(u.id)})`);
            }
        }
        console.log();
    });
}
