# 文件诊断

属性、目录、权限、占用、查找和哈希。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-files
title: 文件诊断
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  files.stat: ./files-stat.md
  files.directory: ./files-directory.md
  files.permissions: ./files-permissions.md
  files.find: ./files-find.md
  files.type: ./files-type.md
  files.sha256: ./files-sha256.md
  files.open: ./files-open.md
  files.archive: ./files-archive.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"files.stat\", \"label\": \"文件属性\", \"description\": \"所有者、权限、大小和时间戳\"}, {\"id\": \"files.directory\", \"label\": \"目录清单\", \"description\": \"显示隐藏文件、权限和容量\"}, {\"id\": \"files.permissions\", \"label\": \"路径权限链\", \"description\": \"逐级查看路径权限和符号链接\"}, {\"id\": \"files.find\", \"label\": \"浅层文件查找\", \"description\": \"仅在指定路径下查找两层以内文件\"}, {\"id\": \"files.type\", \"label\": \"文件类型\", \"description\": \"识别文件格式，不执行文件\"}, {\"id\": \"files.sha256\", \"label\": \"SHA-256 校验\", \"description\": \"读取文件计算哈希，大文件可能超时\"}, {\"id\": \"files.open\", \"label\": \"文件占用\", \"description\": \"查找使用指定文件的进程，不发送信号\"}, {\"id\": \"files.archive\", \"label\": \"tar 内容清单\", \"description\": \"仅列出 tar 归档，不解包或覆盖文件\"}]}"]
:::
