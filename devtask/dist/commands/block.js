"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerBlock = registerBlock;
const database_1 = require("../db/database");
const format_1 = require("../display/format");
const chalk_1 = __importDefault(require("chalk"));
function registerBlock(program) {
    program
        .command('block <id>')
        .description('タスク間のブロック関係を設定する')
        .requiredOption('--blocks <id>', 'ブロックされるタスクID')
        .action((id, opts) => {
        const db = (0, database_1.getDb)();
        const blocker = (0, database_1.findTaskByShortId)(db, id);
        if (!blocker) {
            console.log((0, format_1.error)(`タスク「${id}」が見つかりません。`));
            process.exit(1);
        }
        const blocked = (0, database_1.findTaskByShortId)(db, opts.blocks);
        if (!blocked) {
            console.log((0, format_1.error)(`タスク「${opts.blocks}」が見つかりません。`));
            process.exit(1);
        }
        if (blocker.id === blocked.id) {
            console.log((0, format_1.error)('同じタスクをブロック対象に指定できません。'));
            process.exit(1);
        }
        // 既存チェック
        const existing = db
            .prepare('SELECT 1 FROM task_blocks WHERE blocker_id = ? AND blocked_id = ?')
            .get(blocker.id, blocked.id);
        if (existing) {
            console.log(chalk_1.default.gray('すでにブロック関係が設定されています。'));
            return;
        }
        // 循環チェック（blockedがblockerをブロックしていないか）
        const circular = db
            .prepare('SELECT 1 FROM task_blocks WHERE blocker_id = ? AND blocked_id = ?')
            .get(blocked.id, blocker.id);
        if (circular) {
            console.log((0, format_1.error)('循環するブロック関係は設定できません。'));
            process.exit(1);
        }
        db.prepare('INSERT INTO task_blocks (blocker_id, blocked_id) VALUES (?, ?)').run(blocker.id, blocked.id);
        console.log((0, format_1.success)('ブロック関係を設定しました'));
        console.log(`  ${chalk_1.default.white(blocker.title)} (${(0, database_1.shortId)(blocker.id)}) → ブロック → ${chalk_1.default.white(blocked.title)} (${(0, database_1.shortId)(blocked.id)})`);
        console.log((0, format_1.info)(`「${blocker.title}」が完了するまで「${blocked.title}」はブロックされます。`));
        console.log();
    });
}
