"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerTeam = registerTeam;
const database_1 = require("../db/database");
const format_1 = require("../display/format");
const chalk_1 = __importDefault(require("chalk"));
const OVERLOAD_TASKS = 6;
const OVERLOAD_HOURS = 12;
function registerTeam(program) {
    program
        .command('team')
        .description('チームの負荷状況を表示する')
        .action(() => {
        const db = (0, database_1.getDb)();
        const members = db.prepare('SELECT * FROM members ORDER BY name').all();
        if (members.length === 0) {
            console.log((0, format_1.dim)('  メンバーが登録されていません。devtask member add で追加してください。'));
            return;
        }
        const stats = members.map((m) => {
            const tasks = db
                .prepare(`SELECT * FROM tasks WHERE assignee_id = ? AND status != 'done'`)
                .all(m.id);
            const totalHours = tasks.reduce((sum, t) => sum + (t.estimate_hours ?? 0), 0);
            return { member: m, taskCount: tasks.length, totalHours };
        });
        // 未割当タスク
        const unassigned = db
            .prepare(`SELECT COUNT(*) as cnt FROM tasks WHERE assignee_id IS NULL AND status != 'done'`)
            .get().cnt;
        const maxTasks = Math.max(...stats.map((s) => s.taskCount), 1);
        console.log((0, format_1.header)('\nチームの負荷状況'));
        console.log((0, format_1.separator)(52));
        for (const s of stats) {
            const bar = (0, format_1.loadBar)(s.taskCount, Math.max(maxTasks, OVERLOAD_TASKS));
            const nameStr = chalk_1.default.white(s.member.name.padEnd(6).slice(0, 6));
            const taskStr = chalk_1.default.gray(`${s.taskCount}タスク`);
            const hourStr = s.totalHours > 0 ? chalk_1.default.gray(`推定${s.totalHours}h`) : chalk_1.default.gray('見積なし');
            let annotation = '';
            if (s.taskCount >= OVERLOAD_TASKS || s.totalHours >= OVERLOAD_HOURS) {
                annotation = chalk_1.default.red(' ⚠ オーバーロード');
            }
            else if (s.taskCount === 0) {
                annotation = chalk_1.default.green(' ← 余力あり');
            }
            console.log(`  ${nameStr}  ${bar}  ${taskStr.padEnd(8)}  ${hourStr.padEnd(8)}${annotation}`);
        }
        if (unassigned > 0) {
            console.log((0, format_1.separator)(52));
            console.log(`  ${chalk_1.default.yellow(`未割当タスク: ${unassigned}件`)}`);
        }
        console.log();
    });
}
