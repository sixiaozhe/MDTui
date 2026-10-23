# Linux 运维导航示例

从左侧选择运维主题，按 Enter 会打开该主题自己的子页：网络页带主机诊断，服务页带服务名输入，日志页带流式跟随。返回时父页的选择、滚动和刷新状态会保留。

示例使用随 MDTUI 一起部署的 `linux-tools` 适配器，不依赖 `printf` 等额外命令。底部参数栏可填写主机或服务名，再用 `d`、`o` 触发诊断和概览刷新。

:::mdtui page
version: 1
id: navigation
title: Linux 运维导航
layout: workspace
sidebar: topics
max_parallel: 4
:::

:::mdtui panel
id: topics
title: 运维主题
format: items-json
exec: ["../bin/linux-tools", "topics"]
open:
  system: ./navigation-system.md
  network: ./navigation-network.md
  process: ./navigation-process.md
  storage: ./navigation-storage.md
  service: ./navigation-service.md
  logs: ./navigation-logs.md
:::

:::mdtui input
id: target
label: 待诊断主机 / IP
default: 127.0.0.1
required: true
max_length: 253
:::

:::mdtui panel
id: overview
title: 系统概览
exec: ["../bin/linux-tools", "overview", "-"]
refresh: 30s
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
id: refresh_overview
label: 更新系统概览
key: o
refresh: [overview]
:::

:::mdtui action
id: diagnose_target
label: 诊断目标连通性
key: d
run: connectivity
:::
