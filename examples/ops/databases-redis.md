# Redis 概况

使用本机 redis-cli 默认连接，仅执行 INFO server。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["redis-cli", "--no-auth-warning", "INFO", "server"]
```

:::mdtui page
version: 1
id: ops-databases-redis
title: Redis 概况
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["redis-cli", "--no-auth-warning", "INFO", "server"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
