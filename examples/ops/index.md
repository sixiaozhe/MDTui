# Linux 运维命令工作台

所有任务在 Markdown exec 中直接调用 Linux 原生命令，没有 MDTUI 运维子命令。Enter / 双击进入，F3 过滤，F4 导出。普通入口只读，变更操作在独立页面确认。

:::mdtui page
version: 1
id: ops-home
title: Linux 运维命令工作台
layout: master-detail
sidebar: areas
max_parallel: 4
:::

:::mdtui panel
id: areas
title: 运维主题 · Enter / 双击进入
format: items-json
open:
  system: ./system.md
  process: ./process.md
  network: ./network.md
  storage: ./storage.md
  services: ./services.md
  logs: ./logs.md
  users: ./users.md
  security: ./security.md
  packages: ./packages.md
  containers: ./containers.md
  kubernetes: ./kubernetes.md
  schedules: ./schedules.md
  files: ./files.md
  databases: ./databases.md
  changes: ./changes.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"system\", \"label\": \"系统与硬件\", \"description\": \"发行版、CPU、内存、内核和虚拟化\"}, {\"id\": \"process\", \"label\": \"进程与性能\", \"description\": \"CPU、线程、打开文件和进程下钻\"}, {\"id\": \"network\", \"label\": \"网络与连通性\", \"description\": \"网卡、路由、端口、DNS、HTTP 和 TLS\"}, {\"id\": \"storage\", \"label\": \"磁盘与文件系统\", \"description\": \"容量、inode、挂载、IO 与磁盘健康\"}, {\"id\": \"services\", \"label\": \"服务与启动\", \"description\": \"systemd、失败服务、依赖与服务下钻\"}, {\"id\": \"logs\", \"label\": \"日志与事件\", \"description\": \"journal、内核、服务日志与文件跟随\"}, {\"id\": \"users\", \"label\": \"用户与会话\", \"description\": \"在线会话、账户、组和登录历史\"}, {\"id\": \"security\", \"label\": \"安全检查\", \"description\": \"防火墙、SELinux、AppArmor 和审计\"}, {\"id\": \"packages\", \"label\": \"软件包\", \"description\": \"Debian/RPM 包、文件归属和本地更新清单\"}, {\"id\": \"containers\", \"label\": \"Docker 容器\", \"description\": \"容器下钻、镜像、网络、卷和资源\"}, {\"id\": \"kubernetes\", \"label\": \"Kubernetes\", \"description\": \"节点、Pod、工作负载、事件和日志\"}, {\"id\": \"schedules\", \"label\": \"定时任务\", \"description\": \"cron、systemd timers 和 at 队列\"}, {\"id\": \"files\", \"label\": \"文件诊断\", \"description\": \"属性、目录、权限、占用、查找和哈希\"}, {\"id\": \"databases\", \"label\": \"数据库检查\", \"description\": \"Redis、PostgreSQL、MySQL 和 SQLite\"}, {\"id\": \"changes\", \"label\": \"变更操作\", \"description\": \"单独确认的服务操作、进程终止和容器重启\"}]}"]
:::

:::mdtui panel
id: kernel
title: 主机内核
format: text
max_height: 5
exec: ["uname", "-a"]
:::

:::mdtui panel
id: load
title: 运行时间与负载
format: text
max_height: 5
exec: ["uptime"]
:::

:::mdtui panel
id: memory
title: 内存概览
format: text
max_height: 7
exec: ["free", "-h"]
:::

:::mdtui panel
id: filesystems
title: 文件系统
format: text
exec: ["df", "-hT"]
:::
