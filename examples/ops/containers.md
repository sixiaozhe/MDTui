# Docker 容器

容器下钻、镜像、网络、卷和资源。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-containers
title: Docker 容器
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  containers.list: ./containers-list.md
  containers.images: ./containers-images.md
  containers.stats: ./containers-stats.md
  containers.inspect: ./containers-inspect.md
  containers.logs: ./containers-logs.md
  containers.networks: ./containers-networks.md
  containers.volumes: ./containers-volumes.md
  containers.disk: ./containers-disk.md
  containers.compose: ./containers-compose.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"containers.list\", \"label\": \"容器对象下钻\", \"description\": \"进入 Docker 容器清单，再查看容器状态\"}, {\"id\": \"containers.images\", \"label\": \"镜像清单\", \"description\": \"只读镜像列表，不拉取或删除\"}, {\"id\": \"containers.stats\", \"label\": \"容器资源快照\", \"description\": \"一次性 CPU、内存、网络和 IO 统计\"}, {\"id\": \"containers.inspect\", \"label\": \"容器详情\", \"description\": \"指定容器的 inspect 报告\"}, {\"id\": \"containers.logs\", \"label\": \"容器日志\", \"description\": \"容器近期日志，可跟随\"}, {\"id\": \"containers.networks\", \"label\": \"容器网络\", \"description\": \"Docker 网络清单\"}, {\"id\": \"containers.volumes\", \"label\": \"数据卷清单\", \"description\": \"Docker 数据卷清单，不删除卷\"}, {\"id\": \"containers.disk\", \"label\": \"Docker 磁盘占用\", \"description\": \"镜像、容器、卷和构建缓存占用，仅查询\"}, {\"id\": \"containers.compose\", \"label\": \"Compose 项目\", \"description\": \"Docker Compose 项目状态\"}]}"]
:::
