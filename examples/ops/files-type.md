# 文件类型

识别文件格式，不执行文件。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["file", "--", {"ref": "target.value"}]
```

对象说明：文件绝对路径。

:::mdtui page
version: 1
id: ops-files-type
title: 文件类型
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 文件绝对路径
default: "/etc/hosts"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["file", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
