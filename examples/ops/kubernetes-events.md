# 近期事件

查询命名空间事件，按时间排序。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["kubectl", "--request-timeout=5s", "get", "events", "-n", "--", {"ref": "target.value"}, "--sort-by=.metadata.creationTimestamp"]
```

对象说明：命名空间。

:::mdtui page
version: 1
id: ops-kubernetes-events
title: 近期事件
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 命名空间
default: "default"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["kubectl", "--request-timeout=5s", "get", "events", "-n", "--", {"ref": "target.value"}, "--sort-by=.metadata.creationTimestamp"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
