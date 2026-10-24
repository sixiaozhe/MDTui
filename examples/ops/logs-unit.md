# 指定服务日志

指定单元的近期日志，可跟随。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["journalctl", "-u", {"ref": "target.value"}, "-n", {"ref": "limit.value"}, "--no-pager", "-o", "short-iso"]
```

对象说明：服务名。

:::mdtui page
version: 1
id: ops-logs-unit
title: 指定服务日志
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

:::mdtui input
id: limit
label: 日志行数
type: number
default: 50
min: 1
max: 200
required: true
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["journalctl", "-u", {"ref": "target.value"}, "-n", {"ref": "limit.value"}, "--no-pager", "-o", "short-iso"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::

:::mdtui panel
id: follow
title: 实时日志
mode: stream
format: text
autostart: false
exec: ["journalctl", "-f", "-u", {"ref": "target.value"}, "-n", {"ref": "limit.value"}, "--no-pager", "-o", "short-iso"]
:::

:::mdtui action
id: start_follow
label: 开始日志跟随
key: l
button: true
run: follow
:::
