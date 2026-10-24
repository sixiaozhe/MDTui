# 日志文件尾部

读取绝对路径文件末尾，可跟随；不自动创建文件。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["tail", "-n", {"ref": "limit.value"}, "--", {"ref": "target.value"}]
```

对象说明：日志文件绝对路径。

:::mdtui page
version: 1
id: ops-logs-file
title: 日志文件尾部
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 日志文件绝对路径
default: "/var/log/syslog"
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
exec: ["tail", "-n", {"ref": "limit.value"}, "--", {"ref": "target.value"}]
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
exec: ["tail", "-n", {"ref": "limit.value"}, "-F", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: start_follow
label: 开始日志跟随
key: l
button: true
run: follow
:::
