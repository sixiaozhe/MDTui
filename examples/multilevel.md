# 三级设备导航

固定模拟设备，不访问真实网络。选择设备后 Enter，再选择接口 Enter，进入第三级诊断页。

鼠标可单击定位、双击进入；点击子页顶部“‹ 返回”逐层返回。

- 导航路径显示当前层级及上层设备；子页用 Esc / q 逐层返回。
- F3 可过滤列表，返回后保留条件；同一项可再次 Enter 进入。
- F4 可导出当前页。根页按 q 退出。

:::mdtui page
version: 1
id: multilevel
title: 设备导航
layout: master-detail
sidebar: devices
:::

:::mdtui panel
id: devices
title: 模拟设备 · Enter 查看接口
format: items-json
exec: ["../bin/net-debug", "devices"]
open: ./multilevel-interfaces.md
:::

:::mdtui panel
id: preview
title: 光标设备 ID
format: text
exec: ["../bin/net-debug", "echo", "设备", {ref: "devices.cursor.id"}]
:::
