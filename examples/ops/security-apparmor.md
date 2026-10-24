# AppArmor 状态

查看已加载策略与模式。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["aa-status"]
```

:::mdtui page
version: 1
id: ops-security-apparmor
title: AppArmor 状态
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["aa-status"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
