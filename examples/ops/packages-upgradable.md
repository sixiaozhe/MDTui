# 本地可升级清单

仅使用已有元数据，不执行 apt update 或安装。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["apt", "list", "--upgradable"]
```

备用命令：

```text
["dnf", "--cacheonly", "check-update"]
```

:::mdtui page
version: 1
id: ops-packages-upgradable
title: 本地可升级清单
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 14s
exec: ["apt", "list", "--upgradable"]
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
timeout: 14s
exec: ["dnf", "--cacheonly", "check-update"]
:::

:::mdtui action
id: collect_fallback
label: 执行备用命令
key: b
button: true
run: fallback
:::
