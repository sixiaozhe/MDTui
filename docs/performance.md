# 性能与低资源模式

MDTUI 的资源使用包含终端渲染、输出缓存和调度。页面启动的 `ps`、Python、Docker、数据库客户端等子进程另有开销，设备总占用需要一起观察。

## 缓存和更新

| 设置 | 普通模式 | `--low-resource` |
| --- | --- | --- |
| 日志窗口 | 2 MiB / 10000 行 | 256 KiB / 2000 行 |
| 文本日志显示间隔 | 100ms | 500ms |
| Markdown 日志显示间隔 | 最快 500ms | 最快 1s |
| 每个暂停父页的文本缓存预算 | 不额外裁减 | 约 1 MiB |

窗口字节和行数限制同时生效，最旧完整行先淘汰；单行仍有截断限制。窗口大小不等于进程总内存，Go 运行时、终端单元格、Markdown 解析与折行结果、列表和不可变快照都需要额外空间。

日志使用追加式缓冲区，读取共享不可变字符串，淘汰时推进起点，周期性压缩以释放旧前缀。命令输出仍经过控制字符过滤和 UTF-8 处理。Markdown 合并更新后按完整窗口解析，保留跨行表格、列表、代码块的语义。命令结束时立即发布最后的输出，不等待显示间隔。

流式更新只重绘变化的可见面板；布局、输入、状态变化使用完整绘制。列表过滤和行文本按当前结果及过滤条件缓存，数据刷新、选择、光标和页面变化时相应失效。

低资源模式进入子页后会裁减超出预算的父页面板文本，保留列表 ID、选择、输入和执行状态。被释放的结果显示手动刷新提示；缓存裁减不会增加自动执行或重试。此预算不包含列表、元数据及运行时开销。

## 自定义

```sh
./bin/mdtui --low-resource PAGE.md
./bin/mdtui --stream-buffer-kib 512 --stream-interval 500ms PAGE.md
./bin/mdtui --low-resource --stream-buffer-kib 128 --stream-interval 1s PAGE.md
```

`--stream-buffer-kib` 范围为 64..2048，`--stream-interval` 范围为 100ms..5s。显式设置覆盖对应预设，Markdown 仍采用不小于其预设下限的间隔。显示更新与页面的命令 `refresh` 周期是两个设置：降低显示频率不会减少周期命令的执行次数。普通文本日志可继续使用 100ms 更新，输入事件始终即时处理。缓冲区、显示间隔随导航传给子页。

设备整体占用偏高时，也要检查页面中过短的 `refresh`、频繁启动的外部解释器或大型查询，以及 `max_parallel`。这些命令成本不计入下述 MDTUI 进程 CPU 数据。

## 复测

2026-10-24 在同一 WSL2 / Linux AMD64 环境下，将修改前的 `1e032eb` 与本次实现作 30 秒流式负载对比（8 个面板，约 2 MB/s）：

| 场景 | CPU：修改前 → 修改后 | 采样 RSS 峰值：修改前 → 修改后 |
| --- | --- | --- |
| 文本日志，普通模式 | 5.79% → 2.03% | 31.9 → 32.9 MiB |
| Markdown 日志，普通模式 | 6.80% → 2.31% | 57.6 → 52.5 MiB |
| Markdown 日志，低资源模式 | 修改后 1.66% | 修改后 14.0 MiB |

普通模式保留原有日志窗口容量；文本场景内存峰值未改善。低资源模式使用更小窗口和更长显示间隔，不能将其内存降低全部归因于缓冲算法。空闲 RSS 约 9.3 MiB。日志窗口读取微基准为 0 B/op、0 allocs/op；正确性回归中的本地 PTY 按键 P95 约 1.3ms，文本日志显示 P95 约 57ms。不同运行的 RSS 和短时 CPU 会波动，不能把这些值作为目标设备的硬上限。

```sh
bash scripts/build.sh
python3 tests/integration/performance.py --idle-seconds 8 --seconds 30 --output docs/evidence/performance-text.json
python3 tests/integration/performance.py --idle-seconds 8 --seconds 30 --format markdown --output docs/evidence/performance-markdown.json
python3 tests/integration/performance.py --idle-seconds 8 --seconds 30 --format markdown --low-resource --output docs/evidence/performance-low-resource.json
go test ./internal/content -run '^$' -bench Stream -benchmem
python3 tests/integration/latency.py
```

测试创建 8 个面板，通过真实的 120×40 PTY 接收约 2 MB/s 的日志，每秒从 `/proc` 采样 MDTUI 进程 CPU 和 RSS。CPU 百分比按单核计算，不包含日志源子进程；RSS 峰值为采样峰值。短测适合对比同机同负载，不代表嵌入式设备的长期上限。持续压力测试可将 `--seconds` 调为 1800 或更长；延迟测试衡量本地 PTY 输出，未包含 SSH 网络和物理显示。

原始报告输出至不进入仓库的 `docs/evidence/`。完整正确性回归运行 `bash scripts/verify.sh`，包含竞态、真实 PTY、导航、鼠标和日志延迟检查。
