# 短时性能采样

每秒采样一次，共三次，不启动交互式监控。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["vmstat", "1", "3"]
```

:::mdtui page
version: 1
id: ops-system-vmstat
title: 短时性能采样
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["vmstat", "1", "3"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
