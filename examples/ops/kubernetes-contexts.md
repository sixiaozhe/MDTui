# 集群上下文

读取 kubeconfig 上下文，不切换上下文。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["kubectl", "config", "get-contexts"]
```

:::mdtui page
version: 1
id: ops-kubernetes-contexts
title: 集群上下文
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["kubectl", "config", "get-contexts"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
