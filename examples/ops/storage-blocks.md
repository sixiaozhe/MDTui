# 块设备与文件系统

磁盘、分区、文件系统和挂载点。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["lsblk", "-o", "NAME,TYPE,SIZE,FSTYPE,UUID,MOUNTPOINT"]
```

备用命令：

```text
["lsblk"]
```

:::mdtui page
version: 1
id: ops-storage-blocks
title: 块设备与文件系统
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["lsblk", "-o", "NAME,TYPE,SIZE,FSTYPE,UUID,MOUNTPOINT"]
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
exec: ["lsblk"]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::
