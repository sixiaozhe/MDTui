# 容器资源快照

一次性 CPU、内存、网络和 IO 统计。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["docker", "stats", "--no-stream"]
```

:::mdtui page
version: 1
id: ops-containers-stats
title: 容器资源快照
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 12s
exec: ["docker", "stats", "--no-stream"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
