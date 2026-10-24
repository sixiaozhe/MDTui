# SELinux 状态

查看强制访问控制状态。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["sestatus"]
```

备用命令：

```text
["getenforce"]
```

:::mdtui page
version: 1
id: ops-security-selinux
title: SELinux 状态
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["sestatus"]
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
exec: ["getenforce"]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::
