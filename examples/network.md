# 网络调试工作台

## 快速开始

- 选择接口后按 **Enter** 查看详情，按 `r` 刷新。
- 按 `d` 打开诊断确认框；日志面板可用 **Space** 暂停跟随。

> 这是模拟页面，不会操作真实网络设备。

:::mdtui page
version: 1
id: network
title: 网络调试 / 模拟设备
layout: master-detail
sidebar: interfaces
max_parallel: 2
:::

:::mdtui panel
id: interfaces
title: 接口列表
format: items-json
exec: ["../bin/net-debug", "list", "--json"]
refresh: 5s
:::

:::mdtui panel
id: detail
title: 接口详情
exec: ["../bin/net-debug", "show", "--name", {ref: "interfaces.selected.id"}]
:::

:::mdtui panel
id: counters
title: 计数器
exec: ["../bin/net-debug", "stats", "--name", {ref: "interfaces.selected.id"}]
refresh: 1s
:::

:::mdtui panel
id: diagnostic
title: 诊断日志
mode: stream
format: text
autostart: false
exec: ["../bin/net-debug", "diagnose", "--name", {ref: "interfaces.selected.id"}]
:::

:::mdtui action
id: refresh_details
label: 刷新详情
key: r
refresh: [detail, counters]
:::

:::mdtui action
id: diagnose
label: 运行诊断
key: d
run: diagnostic
confirm: "启动当前接口的模拟诊断？"
:::
