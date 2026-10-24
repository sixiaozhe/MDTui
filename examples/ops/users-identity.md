# 用户身份

UID、GID 和补充组。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["id", "--", {"ref": "target.value"}]
```

对象说明：用户名。

:::mdtui page
version: 1
id: ops-users-identity
title: 用户身份
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 用户名
default: "root"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["id", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
