# 资源详情

查询当前命名空间的资源，例如 pod/example。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["kubectl", "--request-timeout=5s", "describe", "--", {"ref": "target.value"}]
```

对象说明：资源类型/名称，例如 pod/example。

:::mdtui page
version: 1
id: ops-kubernetes-describe
title: 资源详情
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 资源类型/名称，例如 pod/example
default: ""
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["kubectl", "--request-timeout=5s", "describe", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
