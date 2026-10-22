# 架构决策 / ADR 001

日期：2026-10-21。目标：Linux SSH TUI，不包含 Web、桌面壳、内置 SSH 或表达式执行。

采用 Go 1.24 最低版本，CGO_DISABLED 单文件部署。Linux 实现 Setpgid、SIGTERM/SIGKILL 进程组清理；非 Linux 构建仅有直接进程取消降级，不作为 V1 部署验收目标。

| 模块 | 职责 |
|---|---|
| internal/document | 完整页面编译、YAML 节点限制、字段/引用/依赖图校验 |
| internal/scheduler | 单事件循环状态、可注入时钟/runner、队列、代际隔离、选择失效 |
| internal/runner | 真进程、空 stdin、cwd、限额、超时、进程组取消 |
| internal/content | 严格 items-json、增量 UTF-8/控制序列过滤、有界日志、Markdown AST |
| internal/tui | tcell 生命周期、焦点、布局、菜单、确认、输入草稿、单元格裁切 |
| cmd | CLI、信号、版本与确定性模拟适配器 |

schema/state 并入 document/scheduler，避免只有转发接口的空模块。工作 goroutine 执行进程、快照过滤和列表协议解析，只写有锁有界流缓冲与 16 容量结果通道。所有 UI/调度状态在主事件循环修改，查询以 request ID + generation 双重检查；旧进程退出前不释放其并发槽位，同面板不重叠。

流缓冲与 UI 每 100ms 汇合一次；空闲只做轻量 tick，不重绘。snapshot Markdown 在工作协程预渲染，stream Markdown 只解析有界窗口；终端按当前宽度重排普通正文，渲染缓存以内容和宽度为键。搜索在当前区域的可见文本上查找，匹配位置独立于滚动视口。stderr 与 stdout 分开限额。所有文字经控制序列过滤后进入 tcell 单元格绘制。

## 依赖选择

| 库 | 固定版本 | 用途 / 许可证 |
|---|---|---|
| github.com/gdamore/tcell/v2 | 2.8.1 | 终端输入、alternate screen、resize、模拟屏 / Apache-2.0 |
| github.com/rivo/uniseg | 0.4.7 | 字素簇和终端单元格宽度 / MIT |
| github.com/yuin/goldmark | 1.7.8 | Markdown AST、表格扩展 / MIT |
| go.yaml.in/yaml/v3 | 3.0.4 | YAML 节点解析，额外实现严格字段约束 / MIT/Apache-2.0 |
| golang.org/x/term | 0.30.0 | TTY 判断 / BSD-3-Clause |

使用成熟库而非重写 Markdown/YAML/终端协议。锁定已使用的 v2 tcell API；不跟随主版本自动升级。官方接口核对：[tcell](https://pkg.go.dev/github.com/gdamore/tcell/v2)、[YAML v3](https://pkg.go.dev/go.yaml.in/yaml/v3)。升级需重新跑行为测试和 PTY 生命周期测试。

## 假设与风险

设备 CPU、内核、ABI、资源、SSH 客户端和真实命令未提供。参考环境验证不等于真实设备验证。系统不自动推断副作用，页面作者负责确认文案与信任边界。命令自行脱离进程组不保证清理。性能预算需独立记录，不能由编译成功推导。
