"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.calculateScore = calculateScore;
exports.urgencyLabel = urgencyLabel;
function calculateScore(task, blockedCount, activeSprint) {
    const components = [];
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    // 期限プレッシャー
    if (task.due_date) {
        const due = new Date(task.due_date);
        due.setHours(0, 0, 0, 0);
        const daysLeft = Math.ceil((due.getTime() - today.getTime()) / (1000 * 60 * 60 * 24));
        if (daysLeft < 0) {
            components.push({ label: '期限超過', points: 50, reason: `${Math.abs(daysLeft)}日超過` });
        }
        else if (daysLeft === 0) {
            components.push({ label: '期限', points: 50, reason: '今日が期限' });
        }
        else if (daysLeft === 1) {
            components.push({ label: '期限', points: 40, reason: '明日が期限' });
        }
        else if (daysLeft <= 3) {
            components.push({ label: '期限', points: 25, reason: `${daysLeft}日後が期限` });
        }
        else if (daysLeft <= 7) {
            components.push({ label: '期限', points: 15, reason: `${daysLeft}日後が期限` });
        }
        else {
            components.push({ label: '期限', points: 5, reason: `${daysLeft}日後が期限` });
        }
    }
    // ブロック中（最大30）
    if (blockedCount > 0) {
        const pts = Math.min(blockedCount * 10, 30);
        components.push({
            label: 'ブロック中',
            points: pts,
            reason: `${blockedCount}件のタスクをブロック中`,
        });
    }
    // 先送りペナルティ（最大15）
    if (task.snoozed_count > 0) {
        const pts = Math.min(task.snoozed_count * 5, 15);
        components.push({
            label: '先送りペナルティ',
            points: pts,
            reason: `${task.snoozed_count}回先送り済み`,
        });
    }
    // 放置ペナルティ
    if (task.last_viewed_at) {
        const daysSince = (Date.now() - new Date(task.last_viewed_at).getTime()) / (1000 * 60 * 60 * 24);
        if (daysSince > 3) {
            components.push({
                label: '放置ペナルティ',
                points: 9,
                reason: `${Math.floor(daysSince)}日間未確認`,
            });
        }
        else if (daysSince > 1) {
            components.push({
                label: '放置ペナルティ',
                points: 5,
                reason: `${Math.floor(daysSince)}日間未確認`,
            });
        }
    }
    // スプリントプレッシャー
    if (activeSprint) {
        const sprintStart = new Date(activeSprint.start_date).getTime();
        const sprintEnd = new Date(activeSprint.end_date).getTime();
        const now = Date.now();
        const total = sprintEnd - sprintStart;
        const remaining = sprintEnd - now;
        const ratio = remaining / total;
        if (ratio < 0.2) {
            components.push({
                label: 'スプリント',
                points: 15,
                reason: 'スプリント終盤（残り20%未満）',
            });
        }
        else if (ratio < 0.5) {
            components.push({
                label: 'スプリント',
                points: 8,
                reason: 'スプリント後半',
            });
        }
    }
    const total = components.reduce((sum, c) => sum + c.points, 0);
    return { total, components };
}
function urgencyLabel(score) {
    if (score >= 80)
        return '[!]';
    if (score >= 50)
        return '[~]';
    return '   ';
}
