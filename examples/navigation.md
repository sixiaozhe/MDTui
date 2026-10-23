# 页面跳转示例

选择服务后按 Enter 进入服务详情子页。

:::mdtui page
version: 1
id: navigation
title: 页面跳转示例
:::

:::mdtui panel
id: services
title: 服务列表
format: items-json
exec: ["printf", "{\"version\":1,\"items\":[{\"id\":\"sshd\",\"label\":\"sshd\"},{\"id\":\"cron\",\"label\":\"cron\"}]}"]
open: ./navigation-detail.md
:::
