# CPU 拓扑

CPU 型号、核心和 NUMA 信息。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["lscpu"]
```

备用命令：

```text
["cat", "/proc/cpuinfo"]
```

:::mdtui page
version: 1
id: ops-system-cpu
title: CPU 拓扑
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["lscpu"]
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
exec: ["cat", "/proc/cpuinfo"]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::
