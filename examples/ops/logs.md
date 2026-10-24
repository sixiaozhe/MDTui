# 日志与事件

journal、内核、服务日志与文件跟随。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-logs
title: 日志与事件
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  logs.recent: ./logs-recent.md
  logs.kernel: ./logs-kernel.md
  logs.unit: ./logs-unit.md
  logs.file: ./logs-file.md
  logs.errors: ./logs-errors.md
  logs.boots: ./logs-boots.md
  logs.usage: ./logs-usage.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"logs.recent\", \"label\": \"系统近期日志\", \"description\": \"读取最近日志，可进一步启用跟随\"}, {\"id\": \"logs.kernel\", \"label\": \"内核日志\", \"description\": \"读取当前内核日志\"}, {\"id\": \"logs.unit\", \"label\": \"指定服务日志\", \"description\": \"指定单元的近期日志，可跟随\"}, {\"id\": \"logs.file\", \"label\": \"日志文件尾部\", \"description\": \"读取绝对路径文件末尾，可跟随；不自动创建文件\"}, {\"id\": \"logs.errors\", \"label\": \"本次启动错误\", \"description\": \"本次启动中 warning 及以上日志\"}, {\"id\": \"logs.boots\", \"label\": \"启动记录\", \"description\": \"journal 中保存的系统启动列表\"}, {\"id\": \"logs.usage\", \"label\": \"日志磁盘占用\", \"description\": \"journal 占用空间，仅查询不清理\"}]}"]
:::
