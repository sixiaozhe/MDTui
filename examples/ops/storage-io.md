# 磁盘 IO 采样

两次扩展 IO 采样，需要 sysstat。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["iostat", "-xz", "1", "2"]
```

:::mdtui page
version: 1
id: ops-storage-io
title: 磁盘 IO 采样
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["iostat", "-xz", "1", "2"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
