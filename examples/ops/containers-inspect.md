# 容器详情

指定容器的 inspect 报告。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["docker", "inspect", "--", {"ref": "target.value"}]
```

对象说明：容器 ID 或名称，可通过容器列表选择。

:::mdtui page
version: 1
id: ops-containers-inspect
title: 容器详情
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 容器 ID 或名称，可通过容器列表选择
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
exec: ["docker", "inspect", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
