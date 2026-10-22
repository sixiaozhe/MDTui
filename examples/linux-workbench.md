# Linux 运维工作台

从左侧选择**系统、网络、进程、存储、服务或日志**，按 Enter 查看该主题。`[` / `]` 切换右侧视图，Z 放大长输出。无需输入或记忆 Linux 命令。

需要排查具体问题时，切到参数栏填写主机或服务名，再按 `d` 诊断连通性、`s` 查看服务。`r` 更新当前主题，`o` 更新概览，`l` 开始跟随日志。所有操作只读取状态；日志跟随可按 Ctrl+C 停止。

:::mdtui page
version: 1
id: linux-workbench
title: Linux 运维 / 任务导航
layout: workspace
sidebar: topics
max_parallel: 4
:::

:::mdtui panel
id: topics
title: 运维任务
format: items-json
exec: ["../bin/linux-tools", "topics"]
:::

:::mdtui input
id: target
label: 待诊断主机 / IP
default: 127.0.0.1
required: true
max_length: 253
:::

:::mdtui input
id: service_name
label: 服务名（例如 ssh）
required: true
max_length: 128
:::

:::mdtui panel
id: overview
title: 系统概览
exec: ["../bin/linux-tools", "overview", "-"]
refresh: 30s
timeout: 20s
:::

:::mdtui panel
id: topic_detail
title: 主题详情
exec: ["../bin/linux-tools", "topic", {ref: "topics.selected.id"}]
timeout: 20s
:::

:::mdtui panel
id: connectivity
title: 连通性诊断
autostart: false
exec: ["../bin/linux-tools", "diagnose", {ref: "target.value"}]
timeout: 20s
:::

:::mdtui panel
id: service_detail
title: 服务详情
autostart: false
exec: ["../bin/linux-tools", "service", {ref: "service_name.value"}]
timeout: 12s
:::

:::mdtui panel
id: live_logs
title: 实时日志
mode: stream
format: text
autostart: false
exec: ["../bin/linux-tools", "follow"]
:::

:::mdtui action
id: refresh_topic
label: 更新所选主题
key: r
refresh: [topic_detail]
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

:::mdtui action
id: inspect_service
label: 查看服务状态
key: s
run: service_detail
:::

:::mdtui action
id: follow_logs
label: 跟随系统日志
key: l
run: live_logs
:::
