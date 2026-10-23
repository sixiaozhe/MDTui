# 网络与端口

该页面由父页面的 `network` 选项打开。除了接口、路由和监听端口，还能对参数栏中的主机或 IP 发起连通性诊断。

:::mdtui page
version: 1
id: navigation-network
title: 网络与端口
parent: topics
max_parallel: 3
:::

:::mdtui input
id: target
label: 待诊断主机 / IP
default: 127.0.0.1
required: true
max_length: 253
:::

:::mdtui panel
id: detail
title: 网络状态
exec: ["../bin/linux-tools", "topic", {ref: "parent.selected.id"}]
timeout: 20s
:::

:::mdtui panel
id: connectivity
title: 连通性诊断
autostart: false
exec: ["../bin/linux-tools", "diagnose", {ref: "target.value"}]
timeout: 20s
:::

:::mdtui action
id: diagnose_network
label: 诊断目标连通性
key: d
run: connectivity
:::

:::mdtui action
id: refresh_network
label: 重新采集网络状态
key: r
refresh: [detail]
:::
