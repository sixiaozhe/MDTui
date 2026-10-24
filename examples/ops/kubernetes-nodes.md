# 节点清单

查询当前上下文的节点状态。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["kubectl", "--request-timeout=5s", "get", "nodes", "-o", "wide"]
```

:::mdtui page
version: 1
id: ops-kubernetes-nodes
title: 节点清单
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["kubectl", "--request-timeout=5s", "get", "nodes", "-o", "wide"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
