# 表单控件体验

固定模拟数据，不连接设备。按 **Shift+Tab** 从结果面板回到数字参数，再继续切换下拉框、开关和备注。

也可用鼠标点击参数、下拉选项和操作按钮；点击参数栏两端 ‹ / › 或滚轮切换参数。

- 开关：**Space / Enter** 切换；下拉框：**Enter** 打开，↑↓ 选择，Enter 确认，Esc 取消。
- 数字：**Enter** 编辑，↑↓ 加减，Enter 校验并提交，Esc 放弃。
- **Tab** 聚焦操作按钮，Enter 执行；也可按 **a**。**F4** 导出当前参数与结果。

:::mdtui page
version: 1
id: controls
title: 表单控件体验
:::

:::mdtui input
id: note
label: 巡检备注
default: 示例任务
max_length: 128
:::

:::mdtui input
id: verbose
label: 详细输出
type: toggle
default: false
:::

:::mdtui input
id: protocol
label: 协议
type: select
default: tcp
options:
  - value: tcp
    label: TCP 连接
  - value: udp
    label: UDP 数据报
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
id: result
title: 模拟适配器收到的原始参数
format: text
autostart: false
exec: ["../bin/net-debug", "echo", "--note", {ref: "note.value"}, "--verbose", {ref: "verbose.value"}, "--protocol", {ref: "protocol.value"}, "--timeout", {ref: "timeout.value"}]
:::

:::mdtui action
id: apply
label: 应用诊断参数
key: a
button: true
run: result
:::

:::mdtui action
id: confirmed_apply
label: 确认后重新采集
button: true
run: result
confirm: 使用当前参数重新生成模拟诊断结果？
:::
