# TLS 握手

读取指定端点的 TLS 握手摘要；stdin 关闭，不发送业务数据。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["openssl", "s_client", "-connect", {"ref": "target.value"}, "-brief"]
```

对象说明：主机:端口，例如 localhost:443。

:::mdtui page
version: 1
id: ops-network-tls
title: TLS 握手
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 主机:端口，例如 localhost:443
default: "localhost:443"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 12s
exec: ["openssl", "s_client", "-connect", {"ref": "target.value"}, "-brief"]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
