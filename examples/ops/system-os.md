# 发行版与内核

查看发行版标识和内核版本。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["cat", "/etc/os-release"]
```

备用命令：

```text
["uname", "-a"]
```

:::mdtui page
version: 1
id: ops-system-os
title: 发行版与内核
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["cat", "/etc/os-release"]
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
exec: ["uname", "-a"]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::
