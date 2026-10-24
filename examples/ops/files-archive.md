# tar 内容清单

仅列出 tar 归档，不解包或覆盖文件。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["tar", "-tf", {"ref": "target.value"}]
```

对象说明：tar 归档绝对路径。

:::mdtui page
version: 1
id: ops-files-archive
title: tar 内容清单
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: tar 归档绝对路径
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
exec: ["tar", "-tf", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
