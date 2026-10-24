# 本次启动错误

本次启动中 warning 及以上日志。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["journalctl", "-b", "-p", "warning", "-n", {"ref": "limit.value"}, "--no-pager"]
```

:::mdtui page
version: 1
id: ops-logs-errors
title: 本次启动错误
parent: tasks
max_parallel: 3
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
exec: ["journalctl", "-b", "-p", "warning", "-n", {"ref": "limit.value"}, "--no-pager"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
