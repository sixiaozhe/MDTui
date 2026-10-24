# 时钟与时区

时间、时区及同步状态。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["timedatectl", "status"]
```

备用命令：

```text
["date", "--iso-8601=seconds"]
```

:::mdtui page
version: 1
id: ops-system-time
title: 时钟与时区
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["timedatectl", "status"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::

:::mdtui panel
id: fallback
title: 备用命令输出
format: text
autostart: false
timeout: 10s
exec: ["date", "--iso-8601=seconds"]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::
