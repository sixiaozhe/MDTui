# 最近登录

读取登录历史，不修改记录。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["last", "-n", {"ref": "limit.value"}]
```

:::mdtui page
version: 1
id: ops-users-last
title: 最近登录
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
exec: ["last", "-n", {"ref": "limit.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
