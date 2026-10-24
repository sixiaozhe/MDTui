# 第三层：接口诊断

顶部显示完整导航路径和设备上下文。parent.selected.id 读取直接父页选中的接口。

- d 或聚焦按钮后 Enter 启动模拟日志；Space 暂停跟随。
- F4 导出；Esc / q 返回接口列表，再按一次返回设备列表。

:::mdtui page
version: 1
id: multilevel-diagnostic
title: 接口诊断
parent: interfaces
max_parallel: 3
:::

:::mdtui input
id: note
label: 本页巡检备注
default: 第三级采集
max_length: 128
:::

:::mdtui panel
id: detail
title: 接口详情
exec: ["../bin/net-debug", "show", "--name", {ref: "parent.selected.id"}]
:::

:::mdtui panel
id: counters
title: 实时计数器
exec: ["../bin/net-debug", "stats", "--name", {ref: "parent.selected.id"}]
refresh: 2s
:::

:::mdtui panel
id: log
title: 模拟诊断日志
mode: stream
format: text
autostart: false
exec: ["../bin/net-debug", "diagnose", "--name", {ref: "parent.selected.id"}]
:::

:::mdtui action
id: diagnose
label: 启动接口诊断
key: d
button: true
run: log
:::
