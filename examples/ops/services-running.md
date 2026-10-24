# 运行中的服务

当前运行的 systemd 服务。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["systemctl", "list-units", "--type=service", "--state=running", "--no-pager", "--plain", "--no-legend"]
```

备用命令：

```text
["service", "--status-all"]
```

:::mdtui page
version: 1
id: ops-services-running
title: 运行中的服务
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["systemctl", "list-units", "--type=service", "--state=running", "--no-pager", "--plain", "--no-legend"]
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
exec: ["service", "--status-all"]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::
