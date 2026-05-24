import chalk from 'chalk';
import dayjs from 'dayjs';
import 'dayjs/locale/ja';

dayjs.locale('ja');

export function formatDueDate(dueDate: string | null): string {
  if (!dueDate) return '';
  const today = dayjs().startOf('day');
  const due = dayjs(dueDate).startOf('day');
  const diff = due.diff(today, 'day');

  if (diff < 0) return chalk.red(`期限超過${Math.abs(diff)}日`);
  if (diff === 0) return chalk.red('今日');
  if (diff === 1) return chalk.yellow('明日');
  if (diff <= 3) return chalk.yellow(`${diff}日後`);
  if (diff <= 7) return chalk.cyan(`${diff}日後`);
  return chalk.gray(due.format('MM/DD'));
}

export function formatStatus(status: string): string {
  switch (status) {
    case 'open':
      return chalk.gray('未着手');
    case 'in_progress':
      return chalk.blue('進行中');
    case 'done':
      return chalk.green('完了');
    default:
      return status;
  }
}

export function formatScore(score: number): string {
  if (score >= 80) return chalk.red(String(score).padStart(3));
  if (score >= 50) return chalk.yellow(String(score).padStart(3));
  return chalk.gray(String(score).padStart(3));
}

export function formatUrgency(score: number): string {
  if (score >= 80) return chalk.red('[!]');
  if (score >= 50) return chalk.yellow('[~]');
  return '   ';
}

export function loadBar(current: number, max: number, width = 10): string {
  const filled = Math.min(Math.round((current / max) * width), width);
  const empty = width - filled;
  const bar = chalk.cyan('█').repeat(filled) + chalk.gray('░').repeat(empty);
  return bar;
}

export function separator(width = 60): string {
  return chalk.gray('─'.repeat(width));
}

export function header(text: string): string {
  return chalk.bold.white(text);
}

export function dim(text: string): string {
  return chalk.dim(text);
}

export function success(text: string): string {
  return chalk.green(`✓ ${text}`);
}

export function warn(text: string): string {
  return chalk.yellow(`⚠ ${text}`);
}

export function info(text: string): string {
  return chalk.cyan(`→ ${text}`);
}

export function error(text: string): string {
  return chalk.red(`✗ ${text}`);
}

export function formatNeglect(lastViewedAt: string | null): string {
  if (!lastViewedAt) return '';
  const days = Math.floor(
    (Date.now() - new Date(lastViewedAt).getTime()) / (1000 * 60 * 60 * 24)
  );
  if (days < 1) return '';
  return chalk.gray(`放置:${days}日`);
}
