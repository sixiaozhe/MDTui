# 接口诊断子页

顶部显示“功能体验台 → 接口诊断”和父页对象 ID。所有输出均为模拟数据。

- **F4** 导出当前参数、详情、计数器和已有日志；弹窗显示保存路径。
- **f** 运行模拟失败，可将错误状态及 stderr 一起导出。
- **Tab** 可聚焦备注参数，**Enter** 编辑；**Esc / q** 返回父页。

:::mdtui page
version: 1
id: showcase-interface
title: 接口诊断
parent: interfaces
layout: stack
max_parallel: 4
:::

:::mdtui input
id: note
label: 巡检备注（会写入导出文件）
default: 演示采集
max_length: 128
:::

:::mdtui panel
id: detail
title: 接口详情
exec: ["../bin/net-debug", "show", "--name", {ref: "parent.selected.id"}]
:::

:::mdtui panel
id: counters
title: 接口计数器
exec: ["../bin/net-debug", "stats", "--name", {ref: "parent.selected.id"}]
refresh: 2s
:::

:::mdtui panel
id: diagnostic
title: 模拟实时日志
mode: stream
format: text
exec: ["../bin/net-debug", "diagnose", "--name", {ref: "parent.selected.id"}]
:::

:::mdtui panel
id: failure
title: 模拟失败 · 按 f 触发
autostart: false
exec: ["../bin/net-debug", "fail"]
:::

:::mdtui action
id: simulate_failure
label: 采集模拟故障
key: f
run: failure
:::
