# 监听端口

TCP/UDP 监听与进程信息，权限不足时部分信息不可见。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["ss", "-tulnp"]
```

备用命令：

```text
["netstat", "-tulnp"]
```

:::mdtui page
version: 1
id: ops-network-listeners
title: 监听端口
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["ss", "-tulnp"]
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
exec: ["netstat", "-tulnp"]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::
