"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.formatDueDate = formatDueDate;
exports.formatStatus = formatStatus;
exports.formatScore = formatScore;
exports.formatUrgency = formatUrgency;
exports.loadBar = loadBar;
exports.separator = separator;
exports.header = header;
exports.dim = dim;
exports.success = success;
exports.warn = warn;
exports.info = info;
exports.error = error;
exports.formatNeglect = formatNeglect;
const chalk_1 = __importDefault(require("chalk"));
const dayjs_1 = __importDefault(require("dayjs"));
require("dayjs/locale/ja");
dayjs_1.default.locale('ja');
function formatDueDate(dueDate) {
    if (!dueDate)
        return '';
    const today = (0, dayjs_1.default)().startOf('day');
    const due = (0, dayjs_1.default)(dueDate).startOf('day');
    const diff = due.diff(today, 'day');
    if (diff < 0)
        return chalk_1.default.red(`期限超過${Math.abs(diff)}日`);
    if (diff === 0)
        return chalk_1.default.red('今日');
    if (diff === 1)
        return chalk_1.default.yellow('明日');
    if (diff <= 3)
        return chalk_1.default.yellow(`${diff}日後`);
    if (diff <= 7)
        return chalk_1.default.cyan(`${diff}日後`);
    return chalk_1.default.gray(due.format('MM/DD'));
}
function formatStatus(status) {
    switch (status) {
        case 'open':
            return chalk_1.default.gray('未着手');
        case 'in_progress':
            return chalk_1.default.blue('進行中');
        case 'done':
            return chalk_1.default.green('完了');
        default:
            return status;
    }
}
function formatScore(score) {
    if (score >= 80)
        return chalk_1.default.red(String(score).padStart(3));
    if (score >= 50)
        return chalk_1.default.yellow(String(score).padStart(3));
    return chalk_1.default.gray(String(score).padStart(3));
}
function formatUrgency(score) {
    if (score >= 80)
        return chalk_1.default.red('[!]');
    if (score >= 50)
        return chalk_1.default.yellow('[~]');
    return '   ';
}
function loadBar(current, max, width = 10) {
    const filled = Math.min(Math.round((current / max) * width), width);
    const empty = width - filled;
    const bar = chalk_1.default.cyan('█').repeat(filled) + chalk_1.default.gray('░').repeat(empty);
    return bar;
}
function separator(width = 60) {
    return chalk_1.default.gray('─'.repeat(width));
}
function header(text) {
    return chalk_1.default.bold.white(text);
}
function dim(text) {
    return chalk_1.default.dim(text);
}
function success(text) {
    return chalk_1.default.green(`✓ ${text}`);
}
function warn(text) {
    return chalk_1.default.yellow(`⚠ ${text}`);
}
function info(text) {
    return chalk_1.default.cyan(`→ ${text}`);
}
function error(text) {
    return chalk_1.default.red(`✗ ${text}`);
}
function formatNeglect(lastViewedAt) {
    if (!lastViewedAt)
        return '';
    const days = Math.floor((Date.now() - new Date(lastViewedAt).getTime()) / (1000 * 60 * 60 * 24));
    if (days < 1)
        return '';
    return chalk_1.default.gray(`放置:${days}日`);
}
