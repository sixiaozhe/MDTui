# 目录占用

显式执行目录容量统计；大目录可能超时。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["du", "-h", "--max-depth=1", "--", {"ref": "target.value"}]
```

对象说明：绝对目录路径。

:::mdtui page
version: 1
id: ops-storage-usage
title: 目录占用
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 绝对目录路径
default: "/var/log"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 14s
exec: ["du", "-h", "--max-depth=1", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
