# systemd 定时器

列出全部 timer 及下次触发时间。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["systemctl", "list-timers", "--all", "--no-pager"]
```

:::mdtui page
version: 1
id: ops-schedules-timers
title: systemd 定时器
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["systemctl", "list-timers", "--all", "--no-pager"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
