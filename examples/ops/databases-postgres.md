# PostgreSQL 会话

使用现有 libpq 环境/认证配置，-w 禁止交互询问密码。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["psql", "-X", "-w", "-P", "pager=off", "-c", "SELECT pid, usename, datname, state, wait_event_type FROM pg_stat_activity ORDER BY pid;"]
```

:::mdtui page
version: 1
id: ops-databases-postgres
title: PostgreSQL 会话
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["psql", "-X", "-w", "-P", "pager=off", "-c", "SELECT pid, usename, datname, state, wait_event_type FROM pg_stat_activity ORDER BY pid;"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
