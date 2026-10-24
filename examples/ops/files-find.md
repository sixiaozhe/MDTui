# 浅层文件查找

仅在指定路径下查找两层以内文件。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["find", {"ref": "target.value"}, "-maxdepth", "2", "-type", "f", "-print"]
```

对象说明：查找起点绝对路径。

:::mdtui page
version: 1
id: ops-files-find
title: 浅层文件查找
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 查找起点绝对路径
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
exec: ["find", {"ref": "target.value"}, "-maxdepth", "2", "-type", "f", "-print"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
