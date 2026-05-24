"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.registerAdd = registerAdd;
const database_1 = require("../db/database");
const format_1 = require("../display/format");
function registerAdd(program) {
    program
        .command('add <title>')
        .description('タスクを追加する')
        .option('--due <date>', '期限 (YYYY-MM-DD)')
        .option('--assign <name>', '担当者名')
        .option('--estimate <hours>', '見積もり時間')
        .option('--blocks <id>', 'ブロックするタスクID')
        .action((title, opts) => {
        const db = (0, database_1.getDb)();
        const id = crypto.randomUUID();
        const ts = (0, database_1.now)();
        // 担当者解決
        let assigneeId = null;
        if (opts.assign) {
            const member = db
                .prepare('SELECT * FROM members WHERE name LIKE ?')
                .get(`%${opts.assign}%`);
            if (!member) {
                console.log((0, format_1.error)(`メンバー「${opts.assign}」が見つかりません。先に devtask member add で追加してください。`));
                process.exit(1);
            }
            assigneeId = member.id;
        }
        // アクティブスプリント取得
        const sprint = db.prepare('SELECT * FROM sprints WHERE active = 1').get();
        db.prepare(`
        INSERT INTO tasks (id, title, assignee_id, status, due_date, estimate_hours, sprint_id, created_at, updated_at)
        VALUES (?, ?, ?, 'open', ?, ?, ?, ?, ?)
      `).run(id, title.trim(), assigneeId, opts.due ?? null, opts.estimate ? parseFloat(opts.estimate) : null, sprint?.id ?? null, ts, ts);
        // ブロック関係設定
        if (opts.blocks) {
            const blocker = (0, database_1.findTaskByShortId)(db, opts.blocks);
            if (!blocker) {
                console.log((0, format_1.error)(`タスク「${opts.blocks}」が見つかりません。`));
                process.exit(1);
            }
            db.prepare('INSERT INTO task_blocks (blocker_id, blocked_id) VALUES (?, ?)').run(id, blocker.id);
            console.log((0, format_1.info)(`「${blocker.title}」をブロックするタスクとして設定しました。`));
        }
        console.log((0, format_1.success)(`タスクを追加しました: ${title}`));
        console.log(`  ID: ${(0, database_1.shortId)(id)}`);
        if (opts.due)
            console.log(`  期限: ${opts.due}`);
        if (opts.assign)
            console.log(`  担当: ${opts.assign}`);
        if (opts.estimate)
            console.log(`  見積: ${opts.estimate}h`);
    });
}
