#!/usr/bin/env node
"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const commander_1 = require("commander");
const add_1 = require("./commands/add");
const list_1 = require("./commands/list");
const today_1 = require("./commands/today");
const why_1 = require("./commands/why");
const done_1 = require("./commands/done");
const start_1 = require("./commands/start");
const block_1 = require("./commands/block");
const team_1 = require("./commands/team");
const standup_1 = require("./commands/standup");
const member_1 = require("./commands/member");
const sprint_1 = require("./commands/sprint");
const program = new commander_1.Command();
program
    .name('devtask')
    .description('開発者チーム向けCLIタスク管理ツール')
    .version('0.1.0');
(0, add_1.registerAdd)(program);
(0, list_1.registerList)(program);
(0, today_1.registerToday)(program);
(0, why_1.registerWhy)(program);
(0, done_1.registerDone)(program);
(0, start_1.registerStart)(program);
(0, block_1.registerBlock)(program);
(0, team_1.registerTeam)(program);
(0, standup_1.registerStandup)(program);
(0, member_1.registerMember)(program);
(0, sprint_1.registerSprint)(program);
program.parse(process.argv);
