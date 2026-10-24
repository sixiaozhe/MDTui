# 资源限制

指定进程的资源限制。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["cat", "--", {"ref": "target.value"}]
```

对象说明：完整绝对路径，例如 /proc/1/limits。

:::mdtui page
version: 1
id: ops-process-limits
title: 资源限制
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 完整绝对路径，例如 /proc/1/limits
default: "/proc/1/limits"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["cat", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
