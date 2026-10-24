# 文件属性

所有者、权限、大小和时间戳。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["stat", "--", {"ref": "target.value"}]
```

对象说明：文件或目录绝对路径。

:::mdtui page
version: 1
id: ops-files-stat
title: 文件属性
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 文件或目录绝对路径
default: "/etc/hosts"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["stat", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
