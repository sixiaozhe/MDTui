# 服务详情

当前服务来自父页面的选择。

:::mdtui page
version: 1
id: navigation-detail
title: 服务详情
parent: services
:::

:::mdtui panel
id: status
title: 状态
exec: ["printf", "selected=%s\n", {ref: "parent.selected.id"}]
timeout: 5s
:::
