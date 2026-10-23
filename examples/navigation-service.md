# 服务状态

该页面由父页面的 `service` 选项打开。它同时演示参数输入和一次性动作：填写服务名后按 `s` 查看状态。

:::mdtui page
version: 1
id: navigation-service
title: 服务状态
parent: topics
:::

:::mdtui input
id: service_name
label: 服务名（例如 ssh）
required: true
max_length: 128
:::

:::mdtui panel
id: detail
title: 运行中的服务
exec: ["../bin/linux-tools", "topic", {ref: "parent.selected.id"}]
timeout: 20s
:::

:::mdtui panel
id: service_detail
title: 服务详情
autostart: false
exec: ["../bin/linux-tools", "service", {ref: "service_name.value"}]
timeout: 12s
:::

:::mdtui action
id: inspect_service
label: 查看服务状态
key: s
run: service_detail
:::

:::mdtui action
id: refresh_service
label: 重新采集服务列表
key: r
refresh: [detail]
:::
