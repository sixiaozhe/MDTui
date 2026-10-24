# MySQL 状态

使用现有客户端认证配置，仅执行 SHOW GLOBAL STATUS。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["mysql", "--connect-timeout=3", "--batch", "-e", "SHOW GLOBAL STATUS;"]
```

:::mdtui page
version: 1
id: ops-databases-mysql
title: MySQL 状态
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["mysql", "--connect-timeout=3", "--batch", "-e", "SHOW GLOBAL STATUS;"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
