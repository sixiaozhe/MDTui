# 系统近期日志

读取最近日志，可进一步启用跟随。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["journalctl", "-n", {"ref": "limit.value"}, "--no-pager", "-o", "short-iso"]
```

备用命令：

```text
["tail", "-n", {"ref": "limit.value"}, "--", "/var/log/syslog"]
```

:::mdtui page
version: 1
id: ops-logs-recent
title: 系统近期日志
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
exec: ["journalctl", "-n", {"ref": "limit.value"}, "--no-pager", "-o", "short-iso"]
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
exec: ["tail", "-n", {"ref": "limit.value"}, "--", "/var/log/syslog"]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::

:::mdtui panel
id: follow
title: 实时日志
mode: stream
format: text
autostart: false
exec: ["journalctl", "-f", "-n", {"ref": "limit.value"}, "--no-pager", "-o", "short-iso"]
:::

:::mdtui action
id: start_follow
label: 开始日志跟随
key: l
button: true
run: follow
:::

:::mdtui panel
id: follow_fallback
title: 备用日志跟随
mode: stream
format: text
autostart: false
exec: ["tail", "-n", {"ref": "limit.value"}, "-F", "--", "/var/log/syslog"]
:::

:::mdtui action
id: start_follow_fallback
label: 备用方式跟随
key: f
button: true
run: follow_fallback
:::
