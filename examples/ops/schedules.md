# 定时任务

cron、systemd timers 和 at 队列。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-schedules
title: 定时任务
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  schedules.cron: ./schedules-cron.md
  schedules.cron-system: ./schedules-cron-system.md
  schedules.cron-directory: ./schedules-cron-directory.md
  schedules.timers: ./schedules-timers.md
  schedules.at: ./schedules-at.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"schedules.cron\", \"label\": \"当前用户 crontab\", \"description\": \"只读列出当前用户计划任务\"}, {\"id\": \"schedules.cron-system\", \"label\": \"系统 cron 配置\", \"description\": \"只读系统 crontab\"}, {\"id\": \"schedules.cron-directory\", \"label\": \"cron 目录\", \"description\": \"列出 /etc/cron.d 的配置文件\"}, {\"id\": \"schedules.timers\", \"label\": \"systemd 定时器\", \"description\": \"列出全部 timer 及下次触发时间\"}, {\"id\": \"schedules.at\", \"label\": \"at 队列\", \"description\": \"查看当前用户的延后执行队列\"}]}"]
:::
