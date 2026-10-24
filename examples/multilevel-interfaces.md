# 第二层：设备接口

列表由父页设备 ID 驱动，描述中显示所属设备。选择接口后 Enter 进入诊断；Esc / q 返回设备列表。

:::mdtui page
version: 1
id: multilevel-interfaces
title: 接口列表
parent: devices
layout: master-detail
sidebar: interfaces
:::

:::mdtui panel
id: interfaces
title: 设备接口 · Enter 进入诊断
format: items-json
exec: ["../bin/net-debug", "device-interfaces", "--device", {ref: "parent.selected.id"}]
open: ./multilevel-diagnostic.md
refresh: 5s
:::

:::mdtui panel
id: device
title: 所属设备
format: text
exec: ["../bin/net-debug", "echo", "设备", {ref: "parent.selected.id"}]
:::
