# 磁盘与文件系统

容量、inode、挂载、IO 与磁盘健康。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-storage
title: 磁盘与文件系统
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  storage.blocks: ./storage-blocks.md
  storage.space: ./storage-space.md
  storage.inodes: ./storage-inodes.md
  storage.mounts: ./storage-mounts.md
  storage.usage: ./storage-usage.md
  storage.io: ./storage-io.md
  storage.smart: ./storage-smart.md
  storage.swap: ./storage-swap.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"storage.blocks\", \"label\": \"块设备与文件系统\", \"description\": \"磁盘、分区、文件系统和挂载点\"}, {\"id\": \"storage.space\", \"label\": \"容量使用率\", \"description\": \"各文件系统的容量和可用空间\"}, {\"id\": \"storage.inodes\", \"label\": \"inode 使用率\", \"description\": \"定位 inode 耗尽\"}, {\"id\": \"storage.mounts\", \"label\": \"挂载关系\", \"description\": \"挂载源、目标及选项\"}, {\"id\": \"storage.usage\", \"label\": \"目录占用\", \"description\": \"显式执行目录容量统计；大目录可能超时\"}, {\"id\": \"storage.io\", \"label\": \"磁盘 IO 采样\", \"description\": \"两次扩展 IO 采样，需要 sysstat\"}, {\"id\": \"storage.smart\", \"label\": \"磁盘 SMART 报告\", \"description\": \"只读取磁盘健康报告，不执行自检或修改参数\"}, {\"id\": \"storage.swap\", \"label\": \"交换区\", \"description\": \"当前 swap 设备及使用情况\"}]}"]
:::
