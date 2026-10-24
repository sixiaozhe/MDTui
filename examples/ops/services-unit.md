# 服务定义

只读展示单元配置及 drop-in。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["systemctl", "cat", "--no-pager", "--", {"ref": "target.value"}]
```

对象说明：服务名。

:::mdtui page
version: 1
id: ops-services-unit
title: 服务定义
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 服务名
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
exec: ["systemctl", "cat", "--no-pager", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
