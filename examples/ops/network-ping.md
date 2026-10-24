# ICMP 诊断

向明确指定的目标发送三次探测。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["ping", "-c", "3", "-W", "2", "--", {"ref": "target.value"}]
```

对象说明：主机名或 IP。

:::mdtui page
version: 1
id: ops-network-ping
title: ICMP 诊断
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 主机名或 IP
default: "127.0.0.1"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 14s
exec: ["ping", "-c", "3", "-W", "2", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
