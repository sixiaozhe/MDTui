# SQLite 结构

以只读方式查看 SQLite 数据库 schema。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["sqlite3", "-readonly", {"ref": "target.value"}, ".schema"]
```

对象说明：已有 SQLite 数据库绝对路径。

:::mdtui page
version: 1
id: ops-databases-sqlite
title: SQLite 结构
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 已有 SQLite 数据库绝对路径
default: ""
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["sqlite3", "-readonly", {"ref": "target.value"}, ".schema"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
