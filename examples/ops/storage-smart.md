# 磁盘 SMART 报告

只读取磁盘健康报告，不执行自检或修改参数。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["smartctl", "-a", "--", {"ref": "target.value"}]
```

对象说明：块设备绝对路径，例如 /dev/sda。

:::mdtui page
version: 1
id: ops-storage-smart
title: 磁盘 SMART 报告
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 块设备绝对路径，例如 /dev/sda
default: "/dev/sda"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["smartctl", "-a", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
