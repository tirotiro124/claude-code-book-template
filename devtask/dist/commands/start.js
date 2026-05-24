"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerStart = registerStart;
const database_1 = require("../db/database");
const format_1 = require("../display/format");
const chalk_1 = __importDefault(require("chalk"));
function registerStart(program) {
    program
        .command('start <id>')
        .description('タスクを進行中にする')
        .action((id) => {
        const db = (0, database_1.getDb)();
        const task = (0, database_1.findTaskByShortId)(db, id);
        if (!task) {
            console.log((0, format_1.error)(`タスク「${id}」が見つかりません。`));
            process.exit(1);
        }
        if (task.status === 'in_progress') {
            console.log(chalk_1.default.gray(`タスク「${task.title}」はすでに進行中です。`));
            return;
        }
        db.prepare(`
        UPDATE tasks SET status = 'in_progress', last_viewed_at = ?, updated_at = ? WHERE id = ?
      `).run((0, database_1.now)(), (0, database_1.now)(), task.id);
        console.log((0, format_1.success)(`着手: ${task.title} (${(0, database_1.shortId)(task.id)})`));
        console.log();
    });
}
