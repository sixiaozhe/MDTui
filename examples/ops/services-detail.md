# 服务对象下钻

从服务列表进入状态页。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["systemctl", "status", "--no-pager", "--full", "--", {"ref": "target.value"}]
```

对象说明：服务名，例如 ssh 或 ssh.service。

:::mdtui page
version: 1
id: ops-services-detail
title: 服务对象下钻
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 服务名，例如 ssh 或 ssh.service
default: "ssh"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["systemctl", "status", "--no-pager", "--full", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
