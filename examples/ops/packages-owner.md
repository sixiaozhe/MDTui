# 文件所属软件包

查找绝对路径文件属于哪个软件包。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["dpkg-query", "-S", {"ref": "target.value"}]
```

备用命令：

```text
["rpm", "-qf", {"ref": "target.value"}]
```

对象说明：文件绝对路径。

:::mdtui page
version: 1
id: ops-packages-owner
title: 文件所属软件包
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 文件绝对路径
default: "/bin/bash"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["dpkg-query", "-S", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::

:::mdtui panel
id: fallback
title: 备用命令输出
format: text
autostart: false
timeout: 10s
exec: ["rpm", "-qf", {"ref": "target.value"}]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::
