# 命名空间 Pod

查询指定命名空间 Pod。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["kubectl", "--request-timeout=5s", "get", "pods", "-n", "--", {"ref": "target.value"}, "-o", "wide"]
```

对象说明：命名空间。

:::mdtui page
version: 1
id: ops-kubernetes-pods
title: 命名空间 Pod
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
exec: ["kubectl", "--request-timeout=5s", "get", "pods", "-n", "--", {"ref": "target.value"}, "-o", "wide"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
