# 数据卷清单

Docker 数据卷清单，不删除卷。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["docker", "volume", "ls"]
```

:::mdtui page
version: 1
id: ops-containers-volumes
title: 数据卷清单
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["docker", "volume", "ls"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
