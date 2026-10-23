# 进程与负载

该页面由父页面的 `process` 选项打开，聚焦 CPU、内存和进程列表。

:::mdtui page
version: 1
id: navigation-process
title: 进程与负载
parent: topics
:::

:::mdtui panel
id: detail
title: 进程概览
exec: ["../bin/linux-tools", "topic", {ref: "parent.selected.id"}]
refresh: 10s
timeout: 20s
:::

:::mdtui panel
id: overview
title: 主机负载
exec: ["../bin/linux-tools", "overview", "-"]
refresh: 30s
timeout: 20s
:::

:::mdtui action
id: refresh_process
label: 重新采集进程信息
key: r
refresh: [detail, overview]
:::
