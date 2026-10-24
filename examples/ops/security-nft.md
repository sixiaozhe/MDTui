# nftables 规则

只读规则清单，不修改或清空规则。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["nft", "list", "ruleset"]
```

:::mdtui page
version: 1
id: ops-security-nft
title: nftables 规则
parent: tasks
max_parallel: 3
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 10s
exec: ["nft", "list", "ruleset"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
