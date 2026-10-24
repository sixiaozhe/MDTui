# 服务详情

引用父页对象 ID，执行原生命令查询状态。r 刷新，e 读取补充信息，F4 导出，Esc / q 逐层返回。

:::mdtui page
version: 1
id: ops-object-service
title: 服务详情
parent: objects
max_parallel: 3
:::

:::mdtui panel
id: detail
title: 对象状态
format: text
timeout: 15s
exec: ["systemctl", "status", "--no-pager", "--full", "--", {"ref": "parent.selected.id"}]
:::

:::mdtui action
id: refresh_object
label: 刷新对象状态
key: r
button: true
run: detail
:::

:::mdtui panel
id: extra
title: 补充诊断 · e 执行
format: text
autostart: false
timeout: 15s
exec: ["journalctl", "-u", {"ref": "parent.selected.id"}, "-n", "80", "--no-pager", "-o", "short-iso"]
:::

:::mdtui action
id: collect_extra
label: 采集补充信息
key: e
button: true
run: extra
:::

:::mdtui panel
id: follow
title: 对象实时日志
mode: stream
format: text
autostart: false
exec: ["journalctl", "-f", "-u", {"ref": "parent.selected.id"}, "-n", "50", "--no-pager", "-o", "short-iso"]
:::

:::mdtui action
id: start_follow
label: 开始日志跟随
key: l
button: true
run: follow
:::
