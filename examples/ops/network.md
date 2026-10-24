# 网络与连通性

网卡、路由、端口、DNS、HTTP 和 TLS。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-network
title: 网络与连通性
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  network.interfaces: ./interfaces.md
  network.interface: ./network-interface.md
  network.routes: ./network-routes.md
  network.routes6: ./network-routes6.md
  network.listeners: ./network-listeners.md
  network.connections: ./network-connections.md
  network.neighbors: ./network-neighbors.md
  network.resolver: ./network-resolver.md
  network.dns: ./network-dns.md
  network.ping: ./network-ping.md
  network.trace: ./network-trace.md
  network.http: ./network-http.md
  network.tls: ./network-tls.md
  network.statistics: ./network-statistics.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"network.interfaces\", \"label\": \"网卡列表\", \"description\": \"进入真实网卡列表并查看地址与链路\"}, {\"id\": \"network.interface\", \"label\": \"指定网卡统计\", \"description\": \"指定网卡的地址、链路和收发计数\"}, {\"id\": \"network.routes\", \"label\": \"IPv4 路由\", \"description\": \"主路由表\"}, {\"id\": \"network.routes6\", \"label\": \"IPv6 路由\", \"description\": \"IPv6 主路由表\"}, {\"id\": \"network.listeners\", \"label\": \"监听端口\", \"description\": \"TCP/UDP 监听与进程信息，权限不足时部分信息不可见\"}, {\"id\": \"network.connections\", \"label\": \"活动连接\", \"description\": \"TCP 活动连接与计时器\"}, {\"id\": \"network.neighbors\", \"label\": \"邻居与 ARP\", \"description\": \"IPv4/IPv6 邻居缓存\"}, {\"id\": \"network.resolver\", \"label\": \"DNS 配置\", \"description\": \"系统名称解析器状态\"}, {\"id\": \"network.dns\", \"label\": \"名称解析\", \"description\": \"使用本机 NSS/DNS 配置解析目标\"}, {\"id\": \"network.ping\", \"label\": \"ICMP 诊断\", \"description\": \"向明确指定的目标发送三次探测\"}, {\"id\": \"network.trace\", \"label\": \"路由追踪\", \"description\": \"最多八跳的路由诊断；超时也显示已有输出\"}, {\"id\": \"network.http\", \"label\": \"HTTP 响应头\", \"description\": \"仅请求 HTTP/HTTPS 响应头，最多八秒\"}, {\"id\": \"network.tls\", \"label\": \"TLS 握手\", \"description\": \"读取指定端点的 TLS 握手摘要；stdin 关闭，不发送业务数据\"}, {\"id\": \"network.statistics\", \"label\": \"协议栈统计\", \"description\": \"汇总协议栈计数器\"}]}"]
:::
