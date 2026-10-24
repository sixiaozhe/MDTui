# 面板最大高度

顶部概览最多 5 行，接口列表最多 6 行，日志使用剩余空间。高度包含上下边框，内容超出时可滚动。

选择接口后 Enter 自动启动模拟日志；调整终端尺寸或按 Z 放大，面板仍遵守 max_height。

:::mdtui page
version: 1
id: panel-height
title: 面板高度体验
layout: stack
max_parallel: 3
:::

:::mdtui panel
id: overview
title: 概览 · 最多 5 行
format: text
max_height: 5
exec: ["../bin/net-debug", "echo", "这是一个固定高度上限的概览面板"]
:::

:::mdtui panel
id: interfaces
title: 接口列表 · 最多 6 行
format: items-json
max_height: 6
exec: ["../bin/net-debug", "list"]
:::

:::mdtui panel
id: log
title: 日志 · 使用剩余高度
mode: stream
format: text
exec: ["../bin/net-debug", "diagnose", "--name", {ref: "interfaces.selected.id"}]
:::
