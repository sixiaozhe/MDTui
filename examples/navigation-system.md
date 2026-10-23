# 系统与资源

该页面由父页面的 `system` 选项打开，展示内核、运行时间、内存和文件系统概览。

:::mdtui page
version: 1
id: navigation-system
title: 系统与资源
parent: topics
:::

:::mdtui panel
id: detail
title: 当前主题
exec: ["../bin/linux-tools", "topic", {ref: "parent.selected.id"}]
timeout: 20s
:::

:::mdtui panel
id: overview
title: 系统概览
exec: ["../bin/linux-tools", "overview", "-"]
refresh: 30s
timeout: 20s
:::

:::mdtui action
id: refresh_system
label: 重新采集系统信息
key: r
refresh: [detail, overview]
:::
