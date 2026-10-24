# Linux 运维 Markdown 示例集

每个任务都在普通 MDTUI Markdown 的 exec 中直接声明 Linux 原生命令；MDTUI 核心没有增加运维子命令或专用分发器。

```sh
bash scripts/build.sh
./bin/mdtui check examples/ops/index.md
./bin/mdtui examples/ops/index.md
```

## 结构与操作

- 首页 → 主题 → 独立任务页。进程、网卡、服务和容器支持主题 → 对象列表 → 对象详情，共四层。
- 静态导航通过 GNU coreutils 的 printf 输出 items-json，不依赖 linux-tools 或 net-debug。
- 四种动态对象列表只在 Markdown 中使用 Python3 小片段，把 ps/ip/systemctl/docker 输出转换成既有 items-json；不是 MDTUI 内置功能。
- 普通任务页先显示原生命令及参数，r/按钮执行，b 使用备用命令；备用命令不自动运行。对象详情页自动查询所选对象状态。
- 日志页 l 启动跟随，Ctrl+C 取消，Space 暂停显示跟随。F3 过滤列表，F4 导出。

## 环境与边界

需要 GNU/Linux 及各任务对应的原生工具。systemd、iproute2、procps、sysstat、lsof、smartmontools、Docker、kubectl、数据库客户端按主题使用，不会自动安装或提权。仅四种动态对象列表需要 Python3。缺少程序、权限不足、daemon 不可用或对象消失时展示原生错误。

Kubernetes 使用当前 kubeconfig；数据库使用现有客户端环境或认证配置，页面不收集密码。进程参数、Docker inspect 和配置可能包含敏感内容，分享导出前请检查。快照和日志沿用 MDTUI 的有界输出缓存；日志行数由数字输入传给原生命令。

变更页面直接执行原生命令，但所有面板禁止自动启动和刷新，操作必须通过确认框。PID 输入至少为 2；服务和容器对象前用 -- 结束选项。自动化验证不会实际执行变更。

## 主题目录

| 主题 | 任务数 | 内容 |
|---|---|---|
| [系统与硬件](system.md) | 10 | 发行版、CPU、内存、内核和虚拟化 |
| [进程与性能](process.md) | 10 | CPU、线程、打开文件和进程下钻 |
| [网络与连通性](network.md) | 14 | 网卡、路由、端口、DNS、HTTP 和 TLS |
| [磁盘与文件系统](storage.md) | 8 | 容量、inode、挂载、IO 与磁盘健康 |
| [服务与启动](services.md) | 7 | systemd、失败服务、依赖与服务下钻 |
| [日志与事件](logs.md) | 7 | journal、内核、服务日志与文件跟随 |
| [用户与会话](users.md) | 6 | 在线会话、账户、组和登录历史 |
| [安全检查](security.md) | 8 | 防火墙、SELinux、AppArmor 和审计 |
| [软件包](packages.md) | 6 | Debian/RPM 包、文件归属和本地更新清单 |
| [Docker 容器](containers.md) | 9 | 容器下钻、镜像、网络、卷和资源 |
| [Kubernetes](kubernetes.md) | 8 | 节点、Pod、工作负载、事件和日志 |
| [定时任务](schedules.md) | 5 | cron、systemd timers 和 at 队列 |
| [文件诊断](files.md) | 8 | 属性、目录、权限、占用、查找和哈希 |
| [数据库检查](databases.md) | 4 | Redis、PostgreSQL、MySQL 和 SQLite |
| [变更操作](changes.md) | 6 | 单独确认的服务操作、进程终止和容器重启 |

## 任务和原生命令

| 页面 | 首选原生命令 |
|---|---|
| [发行版与内核](system-os.md) | `cat /etc/os-release` |
| [内核版本](system-kernel.md) | `uname -a` |
| [运行时间与负载](system-uptime.md) | `uptime` |
| [CPU 拓扑](system-cpu.md) | `lscpu` |
| [内存与交换区](system-memory.md) | `free -h` |
| [短时性能采样](system-vmstat.md) | `vmstat 1 3` |
| [时钟与时区](system-time.md) | `timedatectl status` |
| [虚拟化环境](system-virtualization.md) | `systemd-detect-virt` |
| [PCI 设备](system-pci.md) | `lspci -nn` |
| [USB 设备](system-usb.md) | `lsusb` |
| [CPU 排序](process-cpu.md) | `ps -eo pid,ppid,user,%cpu,%mem,comm --sort=-%cpu` |
| [内存排序](process-memory.md) | `ps -eo pid,ppid,user,%cpu,%mem,rss,comm --sort=-%mem` |
| [进程树](process-tree.md) | `pstree -p` |
| [按 PID 查看](process-detail.md) | `ps -p {target} -o pid,ppid,user,lstart,etime,state,%cpu,%mem,args` |
| [线程列表](process-threads.md) | `ps -T -p {target} -o pid,tid,psr,pcpu,state,comm` |
| [打开文件](process-files.md) | `lsof -nP -p {target}` |
| [资源限制](process-limits.md) | `cat -- {target}` |
| [内核进程状态](process-status.md) | `cat -- {target}` |
| [top 单次采样](process-top.md) | `top -b -n 1` |
| [进程短时采样](process-pidstat.md) | `pidstat -dur 1 2` |
| [网卡列表](network-interfaces.md) | `ip -br addr` |
| [指定网卡统计](network-interface.md) | `ip -s addr show dev {target}` |
| [IPv4 路由](network-routes.md) | `ip route show` |
| [IPv6 路由](network-routes6.md) | `ip -6 route show` |
| [监听端口](network-listeners.md) | `ss -tulnp` |
| [活动连接](network-connections.md) | `ss -tanpo` |
| [邻居与 ARP](network-neighbors.md) | `ip neigh show` |
| [DNS 配置](network-resolver.md) | `resolvectl status` |
| [名称解析](network-dns.md) | `getent ahosts -- {target}` |
| [ICMP 诊断](network-ping.md) | `ping -c 3 -W 2 -- {target}` |
| [路由追踪](network-trace.md) | `traceroute -m 8 -w 1 -- {target}` |
| [HTTP 响应头](network-http.md) | `curl --proto =http,https --proto-redir =http,https --head --location --max-time 8 --connect-timeout 3 --silent --show-error -- {target}` |
| [TLS 握手](network-tls.md) | `openssl s_client -connect {target} -brief` |
| [协议栈统计](network-statistics.md) | `nstat -az` |
| [块设备与文件系统](storage-blocks.md) | `lsblk -o NAME,TYPE,SIZE,FSTYPE,UUID,MOUNTPOINT` |
| [容量使用率](storage-space.md) | `df -hT` |
| [inode 使用率](storage-inodes.md) | `df -hi` |
| [挂载关系](storage-mounts.md) | `findmnt -r` |
| [目录占用](storage-usage.md) | `du -h --max-depth=1 -- {target}` |
| [磁盘 IO 采样](storage-io.md) | `iostat -xz 1 2` |
| [磁盘 SMART 报告](storage-smart.md) | `smartctl -a -- {target}` |
| [交换区](storage-swap.md) | `swapon --show` |
| [运行中的服务](services-running.md) | `systemctl list-units --type=service --state=running --no-pager --plain --no-legend` |
| [失败的单元](services-failed.md) | `systemctl --failed --no-pager --plain` |
| [已安装的服务](services-installed.md) | `systemctl list-unit-files --type=service --no-pager` |
| [服务对象下钻](services-detail.md) | `systemctl status --no-pager --full -- {target}` |
| [服务定义](services-unit.md) | `systemctl cat --no-pager -- {target}` |
| [服务依赖](services-dependencies.md) | `systemctl list-dependencies --no-pager -- {target}` |
| [启动耗时](services-boot.md) | `systemd-analyze blame --no-pager` |
| [系统近期日志](logs-recent.md) | `journalctl -n {limit} --no-pager -o short-iso` |
| [内核日志](logs-kernel.md) | `journalctl -k -n {limit} --no-pager` |
| [指定服务日志](logs-unit.md) | `journalctl -u {target} -n {limit} --no-pager -o short-iso` |
| [日志文件尾部](logs-file.md) | `tail -n {limit} -- {target}` |
| [本次启动错误](logs-errors.md) | `journalctl -b -p warning -n {limit} --no-pager` |
| [启动记录](logs-boots.md) | `journalctl --list-boots --no-pager` |
| [日志磁盘占用](logs-usage.md) | `journalctl --disk-usage` |
| [在线会话](users-sessions.md) | `w` |
| [账户清单](users-accounts.md) | `getent passwd` |
| [组清单](users-groups.md) | `getent group` |
| [用户身份](users-identity.md) | `id -- {target}` |
| [最近登录](users-last.md) | `last -n {limit}` |
| [失败登录](users-failed.md) | `lastb -n {limit}` |
| [nftables 规则](security-nft.md) | `nft list ruleset` |
| [IPv4 防火墙](security-iptables.md) | `iptables -S` |
| [IPv6 防火墙](security-iptables6.md) | `ip6tables -S` |
| [UFW 状态](security-ufw.md) | `ufw status verbose` |
| [SELinux 状态](security-selinux.md) | `sestatus` |
| [AppArmor 状态](security-apparmor.md) | `aa-status` |
| [审计状态](security-audit.md) | `auditctl -s` |
| [SSH 服务配置](security-ssh.md) | `sshd -T` |
| [已安装软件包](packages-installed.md) | `dpkg-query -W` |
| [本地可升级清单](packages-upgradable.md) | `apt list --upgradable` |
| [软件包详情](packages-detail.md) | `dpkg-query -s {target}` |
| [包内文件](packages-files.md) | `dpkg-query -L {target}` |
| [文件所属软件包](packages-owner.md) | `dpkg-query -S {target}` |
| [包管理日志](packages-history.md) | `tail -n {limit} -- /var/log/dpkg.log` |
| [容器对象下钻](containers-list.md) | `docker ps -a --no-trunc` |
| [镜像清单](containers-images.md) | `docker image ls --digests` |
| [容器资源快照](containers-stats.md) | `docker stats --no-stream` |
| [容器详情](containers-inspect.md) | `docker inspect -- {target}` |
| [容器日志](containers-logs.md) | `docker logs --tail {limit} -- {target}` |
| [容器网络](containers-networks.md) | `docker network ls` |
| [数据卷清单](containers-volumes.md) | `docker volume ls` |
| [Docker 磁盘占用](containers-disk.md) | `docker system df` |
| [Compose 项目](containers-compose.md) | `docker compose ls` |
| [集群上下文](kubernetes-contexts.md) | `kubectl config get-contexts` |
| [节点清单](kubernetes-nodes.md) | `kubectl --request-timeout=5s get nodes -o wide` |
| [命名空间 Pod](kubernetes-pods.md) | `kubectl --request-timeout=5s get pods -n -- {target} -o wide` |
| [工作负载](kubernetes-deployments.md) | `kubectl --request-timeout=5s get deployments -n -- {target} -o wide` |
| [Service 清单](kubernetes-services.md) | `kubectl --request-timeout=5s get services -n -- {target} -o wide` |
| [近期事件](kubernetes-events.md) | `kubectl --request-timeout=5s get events -n -- {target} --sort-by=.metadata.creationTimestamp` |
| [资源详情](kubernetes-describe.md) | `kubectl --request-timeout=5s describe -- {target}` |
| [Pod 日志](kubernetes-logs.md) | `kubectl --request-timeout=5s logs --tail {limit} -- {target}` |
| [当前用户 crontab](schedules-cron.md) | `crontab -l` |
| [系统 cron 配置](schedules-cron-system.md) | `cat /etc/crontab` |
| [cron 目录](schedules-cron-directory.md) | `ls -la -- /etc/cron.d` |
| [systemd 定时器](schedules-timers.md) | `systemctl list-timers --all --no-pager` |
| [at 队列](schedules-at.md) | `atq` |
| [文件属性](files-stat.md) | `stat -- {target}` |
| [目录清单](files-directory.md) | `ls -lah -- {target}` |
| [路径权限链](files-permissions.md) | `namei -l -- {target}` |
| [浅层文件查找](files-find.md) | `find {target} -maxdepth 2 -type f -print` |
| [文件类型](files-type.md) | `file -- {target}` |
| [SHA-256 校验](files-sha256.md) | `sha256sum -- {target}` |
| [文件占用](files-open.md) | `lsof -nP -- {target}` |
| [tar 内容清单](files-archive.md) | `tar -tf {target}` |
| [Redis 概况](databases-redis.md) | `redis-cli --no-auth-warning INFO server` |
| [PostgreSQL 会话](databases-postgres.md) | `psql -X -w -P pager=off -c SELECT pid, usename, datname, state, wait_event_type FROM pg_stat_activity ORDER BY pid;` |
| [MySQL 状态](databases-mysql.md) | `mysql --connect-timeout=3 --batch -e SHOW GLOBAL STATUS;` |
| [SQLite 结构](databases-sqlite.md) | `sqlite3 -readonly {target} .schema` |

## 维护

这些 Markdown 可以直接编辑和部署。catalog.json 只是生成器的开发期数据，不会被 MDTUI 加载；批量重建运行 `python3 scripts/generate-ops-examples.py`，会覆盖生成页面。部署运行不需要生成器、catalog.json 或项目适配器。
