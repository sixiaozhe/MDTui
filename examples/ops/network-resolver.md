# DNS 配置

系统名称解析器状态。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["resolvectl", "status"]
```

备用命令：

```text
["cat", "/etc/resolv.conf"]
```

:::mdtui page
version: 1
id: ops-network-resolver
title: DNS 配置
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["resolvectl", "status"]
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
exec: ["cat", "/etc/resolv.conf"]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::
