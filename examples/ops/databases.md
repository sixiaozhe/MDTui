# 数据库检查

Redis、PostgreSQL、MySQL 和 SQLite。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-databases
title: 数据库检查
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  databases.redis: ./databases-redis.md
  databases.postgres: ./databases-postgres.md
  databases.mysql: ./databases-mysql.md
  databases.sqlite: ./databases-sqlite.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"databases.redis\", \"label\": \"Redis 概况\", \"description\": \"使用本机 redis-cli 默认连接，仅执行 INFO server\"}, {\"id\": \"databases.postgres\", \"label\": \"PostgreSQL 会话\", \"description\": \"使用现有 libpq 环境/认证配置，-w 禁止交互询问密码\"}, {\"id\": \"databases.mysql\", \"label\": \"MySQL 状态\", \"description\": \"使用现有客户端认证配置，仅执行 SHOW GLOBAL STATUS\"}, {\"id\": \"databases.sqlite\", \"label\": \"SQLite 结构\", \"description\": \"以只读方式查看 SQLite 数据库 schema\"}]}"]
:::
