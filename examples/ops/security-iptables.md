# IPv4 防火墙

只读 iptables 规则。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["iptables", "-S"]
```

:::mdtui page
version: 1
id: ops-security-iptables
title: IPv4 防火墙
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["iptables", "-S"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
