# 当前用户 crontab

只读列出当前用户计划任务。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["crontab", "-l"]
```

:::mdtui page
version: 1
id: ops-schedules-cron
title: 当前用户 crontab
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["crontab", "-l"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
