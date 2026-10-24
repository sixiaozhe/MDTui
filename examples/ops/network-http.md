# HTTP 响应头

仅请求 HTTP/HTTPS 响应头，最多八秒。原生命令在本页 exec 中直接声明；r 执行主命令，b 执行备用命令（如有），F4 导出。

主命令：

```text
["curl", "--proto", "=http,https", "--proto-redir", "=http,https", "--head", "--location", "--max-time", "8", "--connect-timeout", "3", "--silent", "--show-error", "--", {"ref": "target.value"}]
```

对象说明：完整 http:// 或 https:// URL；不接受内嵌账户密码。

:::mdtui page
version: 1
id: ops-network-http
title: HTTP 响应头
parent: tasks
max_parallel: 3
:::

:::mdtui input
id: target
label: 完整 http:// 或 https:// URL；不接受内嵌账户密码
default: "http://127.0.0.1"
required: true
max_length: 1024
:::

:::mdtui panel
id: result
title: 原生命令输出
format: text
autostart: false
timeout: 12s
exec: ["curl", "--proto", "=http,https", "--proto-redir", "=http,https", "--head", "--location", "--max-time", "8", "--connect-timeout", "3", "--silent", "--show-error", "--", {"ref": "target.value"}]
:::

:::mdtui action
id: collect
label: 执行主命令
key: r
button: true
run: result
:::
