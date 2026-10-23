# 磁盘与挂载

该页面由父页面的 `storage` 选项打开，展示块设备、文件系统容量和挂载点。

:::mdtui page
version: 1
id: navigation-storage
title: 磁盘与挂载
parent: topics
:::

:::mdtui panel
id: detail
title: 存储状态
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
id: refresh_storage
label: 重新采集存储信息
key: r
refresh: [detail, overview]
:::
