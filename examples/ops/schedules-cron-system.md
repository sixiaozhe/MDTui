# 系统 cron 配置

只读系统 crontab。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["cat", "/etc/crontab"]
```

:::mdtui page
version: 1
id: ops-schedules-cron-system
title: 系统 cron 配置
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["cat", "/etc/crontab"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
