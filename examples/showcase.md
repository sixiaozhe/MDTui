# 导航、过滤与导出体验台

这是固定模拟数据，不需要设备或网络权限。顶部显示导航路径。

- **F3** 输入 `wlan0` 后 Enter 过滤；再按 **Enter** 进入接口子页。
- 子页按 **F4** 导出诊断快照，关闭弹窗后按 **q** 返回。
- 返回后过滤条件保留；**F3** 删除过滤词并 Enter 恢复全部接口。
- **Shift+Tab** 可切换下拉框、开关和数字参数；按 **a** 或聚焦按钮应用参数。

:::mdtui page
version: 1
id: showcase
title: 功能体验台
layout: master-detail
sidebar: interfaces
max_parallel: 3
:::

:::mdtui input
id: verbose
label: 详细输出
type: toggle
default: false
:::

:::mdtui input
id: protocol
label: 诊断协议
type: select
options:
  - value: tcp
    label: TCP
  - value: udp
    label: UDP
:::

:::mdtui input
id: timeout
label: 超时时间（秒，1～60）
type: number
default: 5
min: 1
max: 60
required: true
:::

:::mdtui panel
id: interfaces
title: 接口列表 · F3 过滤 / Enter 进入
format: items-json
exec: ["../bin/net-debug", "list"]
open: ./showcase-interface.md
refresh: 5s
:::

:::mdtui panel
id: preview
title: 光标预览 · 无需 Enter
exec: ["../bin/net-debug", "show", "--name", {ref: "interfaces.cursor.id"}]
:::

:::mdtui panel
id: counters
title: 光标对象计数器
exec: ["../bin/net-debug", "stats", "--name", {ref: "interfaces.cursor.id"}]
:::

:::mdtui panel
id: parameters
title: 模拟诊断参数
format: text
autostart: false
exec: ["../bin/net-debug", "echo", "--verbose", {ref: "verbose.value"}, "--protocol", {ref: "protocol.value"}, "--timeout", {ref: "timeout.value"}]
:::

:::mdtui action
id: apply
label: 应用诊断参数
key: a
button: true
run: parameters
:::
