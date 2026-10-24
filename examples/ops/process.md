# 进程与性能

CPU、线程、打开文件和进程下钻。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-process
title: 进程与性能
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  process.cpu: ./process-cpu.md
  process.memory: ./process-memory.md
  process.tree: ./process-tree.md
  process.detail: ./processes.md
  process.threads: ./process-threads.md
  process.files: ./process-files.md
  process.limits: ./process-limits.md
  process.status: ./process-status.md
  process.top: ./process-top.md
  process.pidstat: ./process-pidstat.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"process.cpu\", \"label\": \"CPU 排序\", \"description\": \"按 CPU 使用率显示进程\"}, {\"id\": \"process.memory\", \"label\": \"内存排序\", \"description\": \"按内存使用率显示进程\"}, {\"id\": \"process.tree\", \"label\": \"进程树\", \"description\": \"显示 PID 与父子关系\"}, {\"id\": \"process.detail\", \"label\": \"按 PID 查看\", \"description\": \"进入真实进程列表，再下钻查看进程\"}, {\"id\": \"process.threads\", \"label\": \"线程列表\", \"description\": \"指定进程的线程、CPU 与状态\"}, {\"id\": \"process.files\", \"label\": \"打开文件\", \"description\": \"指定 PID 的文件描述符；缺少 lsof 时查看 /proc\"}, {\"id\": \"process.limits\", \"label\": \"资源限制\", \"description\": \"指定进程的资源限制\"}, {\"id\": \"process.status\", \"label\": \"内核进程状态\", \"description\": \"查看 /proc/PID/status 的进程状态\"}, {\"id\": \"process.top\", \"label\": \"top 单次采样\", \"description\": \"非交互式 top 快照\"}, {\"id\": \"process.pidstat\", \"label\": \"进程短时采样\", \"description\": \"采样进程 CPU、内存与 IO，两次\"}]}"]
:::
