#!/usr/bin/env node
import { Command } from 'commander';
import { registerAdd } from './commands/add';
import { registerList } from './commands/list';
import { registerToday } from './commands/today';
import { registerWhy } from './commands/why';
import { registerDone } from './commands/done';
import { registerStart } from './commands/start';
import { registerBlock } from './commands/block';
import { registerTeam } from './commands/team';
import { registerStandup } from './commands/standup';
import { registerMember } from './commands/member';
import { registerSprint } from './commands/sprint';

const program = new Command();

program
  .name('devtask')
  .description('開発者チーム向けCLIタスク管理ツール')
  .version('0.1.0');

registerAdd(program);
registerList(program);
registerToday(program);
registerWhy(program);
registerDone(program);
registerStart(program);
registerBlock(program);
registerTeam(program);
registerStandup(program);
registerMember(program);
registerSprint(program);

program.parse(process.argv);
