# 系统日志

该页面由父页面的 `logs` 选项打开。除了近期日志，还可以按 `l` 启动流式跟随，按 Space 暂停滚动。

:::mdtui page
version: 1
id: navigation-logs
title: 系统日志
parent: topics
max_parallel: 3
:::

:::mdtui panel
id: detail
title: 近期日志
exec: ["../bin/linux-tools", "topic", {ref: "parent.selected.id"}]
timeout: 20s
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
id: follow_logs
label: 开始跟随日志
key: l
run: live_logs
:::

:::mdtui action
id: refresh_logs
label: 重新采集近期日志
key: r
refresh: [detail]
:::
