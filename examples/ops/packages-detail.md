# 软件包详情

指定软件包的版本、描述和依赖。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["dpkg-query", "-s", {"ref": "target.value"}]
```

备用命令：

```text
["rpm", "-qi", {"ref": "target.value"}]
```

对象说明：软件包名称，例如 bash。

:::mdtui page
version: 1
id: ops-packages-detail
title: 软件包详情
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 软件包名称，例如 bash
default: "bash"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["dpkg-query", "-s", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::

:::mdtui panel
id: fallback
title: 备用命令输出
format: text
autostart: false
timeout: 10s
exec: ["rpm", "-qi", {"ref": "target.value"}]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::
