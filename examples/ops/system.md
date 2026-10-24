# 系统与硬件

发行版、CPU、内存、内核和虚拟化。选择任务 Enter 查看原生命令和参数；进入普通任务页不会自动采集。

:::mdtui page
version: 1
id: ops-system
title: 系统与硬件
parent: areas
max_parallel: 2
:::

:::mdtui panel
id: tasks
title: 任务清单 · F3 过滤
format: items-json
open:
  system.os: ./system-os.md
  system.kernel: ./system-kernel.md
  system.uptime: ./system-uptime.md
  system.cpu: ./system-cpu.md
  system.memory: ./system-memory.md
  system.vmstat: ./system-vmstat.md
  system.time: ./system-time.md
  system.virtualization: ./system-virtualization.md
  system.pci: ./system-pci.md
  system.usb: ./system-usb.md
exec: ["printf", "%s\n", "{\"version\": 1, \"items\": [{\"id\": \"system.os\", \"label\": \"发行版与内核\", \"description\": \"查看发行版标识和内核版本\"}, {\"id\": \"system.kernel\", \"label\": \"内核版本\", \"description\": \"内核、主机名和体系架构\"}, {\"id\": \"system.uptime\", \"label\": \"运行时间与负载\", \"description\": \"运行时间和 1/5/15 分钟平均负载\"}, {\"id\": \"system.cpu\", \"label\": \"CPU 拓扑\", \"description\": \"CPU 型号、核心和 NUMA 信息\"}, {\"id\": \"system.memory\", \"label\": \"内存与交换区\", \"description\": \"内存、缓存和 swap 使用情况\"}, {\"id\": \"system.vmstat\", \"label\": \"短时性能采样\", \"description\": \"每秒采样一次，共三次，不启动交互式监控\"}, {\"id\": \"system.time\", \"label\": \"时钟与时区\", \"description\": \"时间、时区及同步状态\"}, {\"id\": \"system.virtualization\", \"label\": \"虚拟化环境\", \"description\": \"判断容器或虚拟机环境\"}, {\"id\": \"system.pci\", \"label\": \"PCI 设备\", \"description\": \"PCI 设备及厂商标识\"}, {\"id\": \"system.usb\", \"label\": \"USB 设备\", \"description\": \"USB 设备清单\"}]}"]
:::
