# Linux 巡检台

这个页面把“列表项获得焦点”和“选项被确认”分开：

- 左侧巡检主题的光标移动到某一项时，右侧「采集计划」显示该主题将执行哪些命令。
- 按 Enter 确认选项后，右侧「主题详情」才真正执行完整采集。
- 上下移动只预览，不会运行耗时的诊断命令；`i` 可重新采集当前详情。

所有命令来自随 MDTUI 部署的 `linux-tools` 适配器，不依赖额外脚本。

:::mdtui page
version: 1
id: inspection
title: Linux 巡检台
layout: master-detail
sidebar: topics
max_parallel: 3
:::

:::mdtui panel
id: topics
title: 巡检主题
format: items-json
on_focus: preview_topic
on_select: inspect_topic
exec: ["../bin/linux-tools", "topics"]
:::

:::mdtui panel
id: preview
title: 采集计划
autostart: false
exec: ["../bin/linux-tools", "explain", {ref: "topics.cursor.id"}]
timeout: 5s
:::

:::mdtui panel
id: detail
title: 主题详情
autostart: false
exec: ["../bin/linux-tools", "topic", {ref: "topics.selected.id"}]
timeout: 20s
:::

:::mdtui action
id: preview_topic
label: 预览采集计划
run: preview
:::

:::mdtui action
id: inspect_topic
label: 采集所选主题
key: i
run: detail
:::
