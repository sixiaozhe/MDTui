# 打开文件

指定 PID 的文件描述符；缺少 lsof 时查看 /proc。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["lsof", "-nP", "-p", {"ref": "target.value"}]
```

对象说明：进程 PID。

:::mdtui page
version: 1
id: ops-process-files
title: 打开文件
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 进程 PID
default: "1"
required: true
max_length: 1024
type: number
min: 1
max: 2147483647
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["lsof", "-nP", "-p", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
