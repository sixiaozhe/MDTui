# 日志磁盘占用

journal 占用空间，仅查询不清理。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["journalctl", "--disk-usage"]
```

:::mdtui page
version: 1
id: ops-logs-usage
title: 日志磁盘占用
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["journalctl", "--disk-usage"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
