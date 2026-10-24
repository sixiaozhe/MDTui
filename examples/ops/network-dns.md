# 名称解析

使用本机 NSS/DNS 配置解析目标。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["getent", "ahosts", "--", {"ref": "target.value"}]
```

对象说明：主机名或 IP。

:::mdtui page
version: 1
id: ops-network-dns
title: 名称解析
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 主机名或 IP
default: "localhost"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["getent", "ahosts", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
