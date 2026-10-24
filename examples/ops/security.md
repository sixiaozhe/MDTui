# 安全检查

防火墙、SELinux、AppArmor 和审计。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-security
title: 安全检查
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  security.nft: ./security-nft.md
  security.iptables: ./security-iptables.md
  security.iptables6: ./security-iptables6.md
  security.ufw: ./security-ufw.md
  security.selinux: ./security-selinux.md
  security.apparmor: ./security-apparmor.md
  security.audit: ./security-audit.md
  security.ssh: ./security-ssh.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"security.nft\", \"label\": \"nftables 规则\", \"description\": \"只读规则清单，不修改或清空规则\"}, {\"id\": \"security.iptables\", \"label\": \"IPv4 防火墙\", \"description\": \"只读 iptables 规则\"}, {\"id\": \"security.iptables6\", \"label\": \"IPv6 防火墙\", \"description\": \"只读 ip6tables 规则\"}, {\"id\": \"security.ufw\", \"label\": \"UFW 状态\", \"description\": \"查看 UFW 配置与规则\"}, {\"id\": \"security.selinux\", \"label\": \"SELinux 状态\", \"description\": \"查看强制访问控制状态\"}, {\"id\": \"security.apparmor\", \"label\": \"AppArmor 状态\", \"description\": \"查看已加载策略与模式\"}, {\"id\": \"security.audit\", \"label\": \"审计状态\", \"description\": \"只读内核审计状态\"}, {\"id\": \"security.ssh\", \"label\": \"SSH 服务配置\", \"description\": \"只读解析 sshd 配置，不启动服务\"}]}"]
:::
