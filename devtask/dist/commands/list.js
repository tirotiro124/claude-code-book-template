"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerList = registerList;
const database_1 = require("../db/database");
const format_1 = require("../display/format");
const chalk_1 = __importDefault(require("chalk"));
function registerList(program) {
    program
        .command('list')
        .description('タスク一覧を表示する')
        .option('--assignee <name>', '担当者で絞り込み')
        .option('--status <status>', 'ステータスで絞り込み (open|in_progress|done)')
        .option('--all', '完了済みも含めて表示')
        .action((opts) => {
        const db = (0, database_1.getDb)();
        let sql = `
        SELECT t.*, m.name as assignee_name
        FROM tasks t
        LEFT JOIN members m ON t.assignee_id = m.id
        WHERE 1=1
      `;
        const params = [];
        if (!opts.all && !opts.status) {
            sql += ` AND t.status != 'done'`;
        }
        if (opts.status) {
            sql += ` AND t.status = ?`;
            params.push(opts.status);
        }
        if (opts.assignee) {
            sql += ` AND m.name LIKE ?`;
            params.push(`%${opts.assignee}%`);
        }
        sql += ` ORDER BY t.created_at DESC`;
        const tasks = db.prepare(sql).all(...params);
        if (tasks.length === 0) {
            console.log((0, format_1.dim)('タスクがありません。'));
            return;
        }
        console.log((0, format_1.header)(`\nタスク一覧 (${tasks.length}件)`));
        console.log((0, format_1.separator)());
        for (const task of tasks) {
            const id = chalk_1.default.dim((0, database_1.shortId)(task.id));
            const title = task.status === 'done' ? chalk_1.default.strikethrough.gray(task.title) : chalk_1.default.white(task.title);
            const status = (0, format_1.formatStatus)(task.status);
            const due = (0, format_1.formatDueDate)(task.due_date);
            const assignee = task.assignee_name ? chalk_1.default.cyan(task.assignee_name) : '';
            const estimate = task.estimate_hours ? chalk_1.default.gray(`${task.estimate_hours}h`) : '';
            const parts = [id, title, status, due, assignee, estimate].filter(Boolean);
            console.log('  ' + parts.join('  '));
        }
        console.log();
    });
}
