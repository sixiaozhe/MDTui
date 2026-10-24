# IPv6 防火墙

只读 ip6tables 规则。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["ip6tables", "-S"]
```

:::mdtui page
version: 1
id: ops-security-iptables6
title: IPv6 防火墙
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["ip6tables", "-S"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
