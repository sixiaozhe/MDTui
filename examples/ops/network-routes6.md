# IPv6 路由

IPv6 主路由表。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["ip", "-6", "route", "show"]
```

:::mdtui page
version: 1
id: ops-network-routes6
title: IPv6 路由
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["ip", "-6", "route", "show"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
