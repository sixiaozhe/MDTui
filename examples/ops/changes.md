# 显式变更操作

exec 直接调用 systemctl、kill 或 docker。变更面板全部为 autostart: false，仅通过带确认框的动作执行；确认框显示原生 argv。浏览不会修改系统，不会自动 sudo。

:::mdtui page
version: 1
id: ops-changes
title: 变更操作
parent: areas
max_parallel: 2
:::

:::mdtui input
id: service
label: 明确的服务名
required: true
max_length: 128
:::

:::mdtui input
id: pid
label: 终止的进程 PID（至少为 2）
type: number
min: 2
max: 2147483647
required: true
:::

:::mdtui input
id: container
label: 明确的 Docker 容器名 / ID
required: true
max_length: 128
:::

:::mdtui panel
id: service_start
title: 启动服务
format: text
autostart: false
timeout: 20s
exec: ["systemctl", "--no-ask-password", "start", "--", {"ref": "service.value"}]
:::

:::mdtui action
id: do_start
label: 启动服务
key: a
button: true
run: service_start
confirm: "确认启动服务？请核对下方原生命令中的对象。"
:::

:::mdtui panel
id: service_stop
title: 停止服务
format: text
autostart: false
timeout: 20s
exec: ["systemctl", "--no-ask-password", "stop", "--", {"ref": "service.value"}]
:::

:::mdtui action
id: do_stop
label: 停止服务
key: s
button: true
run: service_stop
confirm: "确认停止服务？请核对下方原生命令中的对象。"
:::

:::mdtui panel
id: service_restart
title: 重启服务
format: text
autostart: false
timeout: 20s
exec: ["systemctl", "--no-ask-password", "restart", "--", {"ref": "service.value"}]
:::

:::mdtui action
id: do_restart
label: 重启服务
key: r
button: true
run: service_restart
confirm: "确认重启服务？请核对下方原生命令中的对象。"
:::

:::mdtui panel
id: service_reload
title: 重载服务
format: text
autostart: false
timeout: 20s
exec: ["systemctl", "--no-ask-password", "reload", "--", {"ref": "service.value"}]
:::

:::mdtui action
id: do_reload
label: 重载服务
key: l
button: true
run: service_reload
confirm: "确认重载服务？请核对下方原生命令中的对象。"
:::

:::mdtui panel
id: terminate
title: 发送 TERM
format: text
autostart: false
timeout: 10s
exec: ["kill", "--signal", "TERM", "--", {"ref": "pid.value"}]
:::

:::mdtui action
id: do_term
label: 终止指定进程
key: t
button: true
run: terminate
confirm: "确认向下方 PID 发送 TERM？该进程可能承担关键业务。"
:::

:::mdtui panel
id: restart_container
title: 重启容器
format: text
autostart: false
timeout: 25s
exec: ["docker", "restart", "--time", "10", "--", {"ref": "container.value"}]
:::

:::mdtui action
id: do_container
label: 重启指定容器
key: c
button: true
run: restart_container
confirm: "确认重启下方 Docker 容器？"
:::
