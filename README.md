# MDTUI

**把已有 Linux 命令，编排成可以通过 SSH 使用的交互式 Markdown 工作台。**

MDTUI turns command-line diagnostics into a navigable terminal workspace. 选择对象、查看关联结果、刷新状态和跟随日志，都在一份 Markdown 页面中完成。

![MDTUI 工作区演示](docs/assets/linux-workbench.gif)

> 演示画面由 MDTUI 的终端渲染器和固定示例数据生成，不包含真实设备信息。[查看静态截图](docs/assets/linux-workbench.png)。

## 为什么用 MDTUI

设备上已经有 `ip`、`ss`、`ps`、`df` 等命令，却常常需要记住参数、复制对象 ID，再拼凑多条输出。MDTUI 把这些步骤编排到一个页面里，让操作者按任务浏览，而不是背命令。它适合嵌入式设备调试、远程 Linux 运维，以及为现有命令制作轻量内部工具。

| 能力 | 使用时的体验 |
| --- | --- |
| 交互式 Markdown 页面 | 标题、列表、引用、代码块和 GFM 表格在终端中清晰呈现 |
| 列表与参数联动 | 选中对象后，详情面板自动获取它的稳定 ID 并更新 |
| 多种布局 | 工作区、主从分栏和堆叠布局；窄终端自动调整 |
| 实时状态 | 定时刷新、一次性动作、流式日志和暂停跟随 |
| 可控执行 | 页面先校验；命令以参数数组执行，不通过 shell 拼接；输出不会被当作页面代码执行 |
| 轻量部署 | Go 构建的 Linux 二进制；通过普通 SSH PTY 使用，目标机器无需 Go 或 Python |

## 立即体验

在 Linux 或 WSL 构建机上安装 **Go 1.24+**，然后在项目根目录运行：

```sh
bash scripts/build.sh
./bin/mdtui check examples/linux-workbench.md
./bin/mdtui examples/linux-workbench.md
```

这个示例是由 MDTUI 页面驱动的 Linux 运维工作台。左侧选择系统、网络、进程、存储、服务或日志任务，右侧查看结果；也能输入主机名或服务名，触发连通性诊断和服务状态查询。`linux-tools` 适配器在当前 Linux 环境调用固定命令，缺少工具或权限时会在结果中说明。

另有 [模拟网络调试页面](examples/network.md)，可在没有设备和网络权限的环境中练习对象选择、刷新与流式诊断；[输入参数示例](examples/input.md) 展示表单交互。

新增 [功能体验台](examples/showcase.md)，使用固定模拟数据集中演示导航路径、列表过滤和诊断导出：

```sh
./bin/mdtui examples/showcase.md
```

按 `F3` 输入 `wlan0` 并 Enter 过滤，再按 Enter 进入子页；按 `f` 产生模拟错误，按 `F4` 导出带参数、日志和错误信息的 Markdown。关闭导出弹窗后按 `q` 返回，父页保留过滤条件。导出文件保存在启动目录中。

[表单控件示例](examples/controls.md) 集中展示开关、下拉框、带范围校验的数字输入和操作按钮：运行 `./bin/mdtui examples/controls.md`，用 Tab / Shift+Tab 切换控件；Space 切换开关，Enter 编辑或选择，聚焦按钮后 Enter 执行。参数值会原样显示在模拟结果中，方便检查命令传参。

[三级导航示例](examples/multilevel.md) 展示“设备列表 → 接口列表 → 接口诊断”：运行 `./bin/mdtui examples/multilevel.md`，每层列表按 Enter 进入，Esc / q 逐层返回。父页保留状态并在返回后恢复调度；顶部显示路径和设备上下文。返回后可再次 Enter 进入同一项。含根页最多 8 层，`check` 会校验深度和循环引用。

[面板高度示例](examples/panel-height.md) 展示 `max_height`：运行 `./bin/mdtui examples/panel-height.md`，概览最多 5 行、接口列表最多 6 行，日志占用剩余空间。高度包含边框，超出内容仍可滚动查看。

[页面跳转示例](examples/navigation.md) 演示按运维主题打开不同的全屏子页，并在子页读取父页选中的对象；每个主题页可以拥有自己的布局、输入、动作和流式面板。[巡检台示例](examples/inspection.md) 演示 `on_focus` 与 `on_select`：光标移到巡检主题时预览采集计划，按 Enter 确认后才执行完整详情。

通过 SSH 使用时，需要分配 PTY，例如 `ssh -t user@device`，再在设备上运行已部署的 `mdtui PAGE.md`。构建与文件摆放方式见[部署指南](docs/build.md)。

## 复杂示例：Linux 运维控制台

[Linux 运维控制台](examples/ops/README.md) 用一系列 Markdown 页面集成现有 Linux 运维命令，展示如何通过 MDTUI 编排一个完整的工具集。命令和参数直接声明在页面的 `exec` 中，例如 `ps`、`ip`、`systemctl`、`journalctl`、`docker` 和 `kubectl`，不需要给 MDTUI 增加运维子命令。

在 Linux 或 WSL 环境中，从项目根目录运行：

```sh
bash scripts/build.sh
./bin/mdtui check examples/ops/index.md
./bin/mdtui examples/ops/index.md
```

示例覆盖 15 个主题、110 项诊断任务，并提供 6 项需要确认的变更操作：

| 范围 | 内容 |
| --- | --- |
| 系统与性能 | 系统与硬件、进程与性能、磁盘与文件系统 |
| 网络与服务 | 网络与连通性、服务与启动、日志与事件 |
| 日常管理 | 用户与会话、安全检查、软件包、定时任务、文件诊断 |
| 应用与平台 | Docker、Kubernetes、数据库检查 |
| 显式变更 | 服务启动、停止、重启、重载，进程终止，容器重启 |

从首页选择主题，再进入任务页，填写参数后按 `r` 或点击按钮执行。进程、网卡、服务和容器支持“首页 → 主题 → 对象列表 → 对象详情”的四层导航，选中的对象 ID 自动传给详情命令。列表可用 `F3` 过滤，结果可用 `F4` 导出；日志任务还支持启动和取消实时跟随。变更操作单独放在确认页面，浏览页面不会自动执行变更。

页面组织方式和可复用的写法：

| 文件 | 展示的能力 |
| --- | --- |
| [index.md](examples/ops/index.md) | 首页主题导航与系统概览 |
| [network.md](examples/ops/network.md) | 通过 `open` 映射组织主题下的任务 |
| [network-ping.md](examples/ops/network-ping.md) | 输入参数、引用参数执行原生命令、按钮触发 |
| [processes.md](examples/ops/processes.md) / [process-detail.md](examples/ops/process-detail.md) | 动态对象列表、下钻与父页对象引用 |
| [logs-file.md](examples/ops/logs-file.md) | 日志行数输入、快照和流式跟随 |
| [changes.md](examples/ops/changes.md) | 禁止自动启动的变更面板与执行确认 |

目标机器需要安装所选任务使用的原生工具；仅四种动态对象列表额外使用 Python3，将命令输出转换成 `items-json`。缺少程序或权限时显示原生错误，不会自动安装工具或提权。所有页面已包含在示例目录中，可直接编辑运行；`scripts/generate-ops-examples.py` 和目录内的 `catalog.json` 仅用于开发时批量生成页面，运行时不需要生成器。完整命令目录、依赖和操作说明见[示例说明](examples/ops/README.md)。

## 几个按键就能操作

| 按键 | 作用 |
| --- | --- |
| `↑` / `↓`、`Enter` | 移动列表光标、确认选择 |
| `Tab` / `Shift+Tab` | 切换说明、参数和面板焦点 |
| `[` / `]` | 切换工作区详情视图 |
| 大写 `Z`、`Esc` | 放大当前面板、退出放大 |
| `F2` | 打开页面动作菜单 |
| `F3` / `F4` | 过滤当前列表 / 导出当前页诊断快照 |
| 鼠标 | 单击定位、双击列表确认，滚轮滚动；点击控件、按钮和子页返回 |
| `/`、`n` / `N` | 搜索当前区域、跳转匹配 |
| `q` | 退出并恢复终端 |

完整操作、输入编辑、日志跟随和错误状态见[使用手册](docs/user.md)。

支持鼠标上报的终端默认启用鼠标交互。需要终端原生选择、复制时，可运行 `./bin/mdtui --no-mouse PAGE.md` 关闭鼠标接管；全部功能仍可用键盘操作。

资源有限的设备可以使用 `./bin/mdtui --low-resource PAGE.md`：缩小日志窗口、合并日志显示更新，并释放超出预算的父页文本缓存。也可通过 `--stream-buffer-kib 512 --stream-interval 500ms` 单独调整缓存和显示间隔。按键即时处理；这些设置不改变页面的命令刷新周期。具体默认值、取舍和性能复测方法见[性能与低资源模式](docs/performance.md)。

## 页面就是 Markdown

普通 Markdown 是说明文字，`:::mdtui` 块声明交互元素。下面的详情面板引用列表中选中的对象；MDTUI 会在选择改变时重新查询：

```md
# 网络调试

选择一个接口，按 Enter 查看详情。

:::mdtui page
version: 1
id: network
layout: master-detail
sidebar: interfaces
:::

:::mdtui panel
id: interfaces
title: 接口列表
format: items-json
exec: ["./adapter", "list"]
:::

:::mdtui panel
id: detail
title: 接口详情
exec: ["./adapter", "show", {ref: "interfaces.selected.id"}]
:::
```

列表命令返回带稳定 `id` 的 `items-json`，详情命令可返回 Markdown。页面还支持输入、动作、确认框、定时刷新和流式面板。完整语法与约束见[语法参考](docs/syntax.md)、[页面作者指南](docs/author.md)及[命令适配指南](docs/adapter.md)。

## 项目结构与验证

- `cmd/mdtui`：工作台入口；`internal/document` 编译页面，`internal/scheduler` 管理依赖和刷新，`internal/tui` 渲染交互界面。
- `cmd/linux-tools`：真实 Linux 只读诊断适配器；`cmd/net-debug`：确定性的模拟适配器。
- `examples`：可直接运行的页面；`tests/integration`：真实 PTY、CLI 与性能测试。

```sh
bash scripts/verify.sh
bash scripts/cross-build.sh
```

验证脚本执行模块校验、`go vet`、单元与竞态测试、页面校验以及真实 PTY 交互测试；结果生成在本地 `docs/evidence/`，该目录不会进入仓库。[验证范围与已知边界](docs/acceptance.md)说明了当前覆盖范围。

**状态说明：**项目已在 Ubuntu 22.04 / WSL2 参考环境完成自动化验证，并构建 ARM64、ARM、MIPSLE Linux 产物；这些产物尚未在真实嵌入式设备上验收。页面具有执行本机程序的能力，请只运行可信来源的页面。`check` 仅解析和校验，不执行页面命令。

欢迎通过 Issue 反馈终端兼容问题或设备适配需求；贡献流程见[CONTRIBUTING.md](CONTRIBUTING.md)。

## 许可证

MDTUI 使用 [MIT License](LICENSE)。第三方依赖的许可证及声明见 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)。
