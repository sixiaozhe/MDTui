# 服务与启动

systemd、失败服务、依赖与服务下钻。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-services
title: 服务与启动
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  services.running: ./services-running.md
  services.failed: ./services-failed.md
  services.installed: ./services-installed.md
  services.detail: ./services-list.md
  services.unit: ./services-unit.md
  services.dependencies: ./services-dependencies.md
  services.boot: ./services-boot.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"services.running\", \"label\": \"运行中的服务\", \"description\": \"当前运行的 systemd 服务\"}, {\"id\": \"services.failed\", \"label\": \"失败的单元\", \"description\": \"systemd 失败单元及状态\"}, {\"id\": \"services.installed\", \"label\": \"已安装的服务\", \"description\": \"服务文件和启用状态\"}, {\"id\": \"services.detail\", \"label\": \"服务对象下钻\", \"description\": \"从服务列表进入状态页\"}, {\"id\": \"services.unit\", \"label\": \"服务定义\", \"description\": \"只读展示单元配置及 drop-in\"}, {\"id\": \"services.dependencies\", \"label\": \"服务依赖\", \"description\": \"查看依赖关系，不改变服务状态\"}, {\"id\": \"services.boot\", \"label\": \"启动耗时\", \"description\": \"systemd 启动耗时排序\"}]}"]
:::
