# 指定网卡统计

指定网卡的地址、链路和收发计数。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["ip", "-s", "addr", "show", "dev", {"ref": "target.value"}]
```

备用命令：

```text
["ifconfig", {"ref": "target.value"}]
```

对象说明：网卡名称，例如 lo 或 eth0。

:::mdtui page
version: 1
id: ops-network-interface
title: 指定网卡统计
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 网卡名称，例如 lo 或 eth0
default: "lo"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["ip", "-s", "addr", "show", "dev", {"ref": "target.value"}]
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
exec: ["ifconfig", {"ref": "target.value"}]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::
