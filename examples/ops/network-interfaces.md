# 网卡列表

进入真实网卡列表并查看地址与链路。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["ip", "-br", "addr"]
```

备用命令：

```text
["ifconfig", "-a"]
```

:::mdtui page
version: 1
id: ops-network-interfaces
title: 网卡列表
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["ip", "-br", "addr"]
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
exec: ["ifconfig", "-a"]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::
