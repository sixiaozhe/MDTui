# 网卡详情

引用父页对象 ID，执行原生命令查询状态。r 刷新，e 读取补充信息，F4 导出，Esc / q 逐层返回。

:::mdtui page
version: 1
id: ops-object-interface
title: 网卡详情
parent: objects
max_parallel: 3
:::

:::mdtui panel
id: detail
title: 对象状态
format: text
timeout: 15s
exec: ["ip", "-s", "addr", "show", "dev", {"ref": "parent.selected.id"}]
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
exec: ["nstat", "-az"]
:::

:::mdtui action
id: collect_extra
label: 采集补充信息
key: e
button: true
run: extra
:::
