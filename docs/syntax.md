# MDTUI Markdown 语法参考

本文是 MDTUI 页面语言的完整语法说明。页面作者指南见 [author.md](author.md)，运行与按键说明见 [user.md](user.md)，需求基线见 [requirements.md](requirements.md)。

## 1. 文件结构

一个页面是一个 UTF-8 Markdown 文件，后缀通常为 `.md`。文件由两部分组成：

1. 普通 Markdown 说明文字
2. 一个或多个 `:::mdtui TYPE` 扩展块

扩展块只从第 1 列开始识别：

```text
:::mdtui page
version: 1
id: demo
:::
```

- 开始行必须是 `:::mdtui TYPE`，其中 `TYPE` 是 `page`、`panel`、`input` 或 `action`。
- 结束行必须是独占一行的 `:::`，不能有缩进或其他字符。
- 扩展块内部是一个 YAML 映射，不允许嵌套 `:::mdtui` 块。
- 未闭合的扩展块是页面错误。
- 围栏代码块（` ``` ` 或 `~~~`）中的 `:::mdtui` 只作为文字显示，不执行。
- 不是以 `:::mdtui` 开头的 `:::` 行按普通 Markdown 处理。

页面必须恰好有一个 `page` 块，并且它必须是第一个扩展块。普通 Markdown 可以出现在 `page` 块之前或之后，最终都会合并为页面说明区。

## 2. 全局限制

| 限制 | 规则 |
|---|---|
| 文件大小 | ≤ 1 MiB |
| 编码 | UTF-8 |
| page 块 | 恰好 1 个 |
| panel | ≤ 16 个 |
| input | ≤ 16 个 |
| action | ≤ 32 个 |
| YAML 嵌套深度 | ≤ 8 |
| 扩展块嵌套 | 禁止 |
| YAML 特性 | 禁止别名、合并键、自定义标签、重复键 |
| 扩展 ID | `[A-Za-z][A-Za-z0-9_-]{0,63}` |
| ID 命名空间 | page、panel、input、action 共享，不能重复 |
| exec 参数 | 不能包含 NUL |
| 子页数量 | `check` 最多跟随 256 个页面 |
| 子页深度 | 最多 8 层 |

YAML 标量类型是严格的：

- 整数必须写成整数，例如 `max_parallel: 2`
- 布尔值必须写成 `true` / `false`
- 时长通常写成字符串，例如 `5s`、`250ms`；`0` 可以写成整数 `0`
- 命令参数中的数字、布尔值必须加引号，例如 `"80"`

未知字段、未知块类型、类型错误都会让 `mdtui check` 失败。

## 3. 标识符

扩展 ID 必须匹配：

```text
[A-Za-z][A-Za-z0-9_-]{0,63}
```

也就是：字母开头，后续只能是字母、数字、下划线或连字符，最长 64 个字符。

所有扩展块共享同一命名空间：

```text
page.id、panel.id、input.id、action.id 不能互相重名
```

## 4. page 块

```text
:::mdtui page
version: 1
id: network
title: 网络调试
layout: master-detail
sidebar: interfaces
parent: parent-page
max_parallel: 2
:::
```

| 字段 | 必填 | 默认 | 规则 |
|---|---|---|---|
| `version` | 是 | — | 只接受整数 `1` |
| `id` | 是 | — | 全局唯一标识 |
| `title` | 否 | `id` | 页面标题 |
| `layout` | 否 | `stack` | `stack`、`master-detail` 或 `workspace` |
| `sidebar` | 条件 | 空 | `master-detail` / `workspace` 必填，必须是存在的 panel ID |
| `parent` | 否 | 空 | 子页声明父页面中触发它的 panel ID；声明后才可使用 `parent.selected.id` |
| `max_parallel` | 否 | `2` | 1～8 |

`stack` 布局下 `sidebar` 没有视觉效果；`master-detail` 和 `workspace` 会校验 `sidebar` 是否存在。

`parent` 只影响引用解析和父子校验，不会自动建立导航关系；导航关系由父页 panel 的 `open` 声明。

## 5. panel 块

```text
:::mdtui panel
id: interfaces
title: 接口列表
mode: snapshot
format: items-json
exec: ["../bin/net-debug", "list", "--json"]
refresh: 5s
timeout: 5s
autostart: true
open: ./interface-detail.md
on_focus: preview_interface
on_select: inspect_interface
:::
```

| 字段 | 必填 | 默认 | 规则 |
|---|---|---|---|
| `id` | 是 | — | 全局唯一标识 |
| `title` | 否 | `id` | 面板标题 |
| `mode` | 否 | `snapshot` | `snapshot` 或 `stream` |
| `format` | 否 | `markdown` | `markdown`、`text` 或 `items-json` |
| `exec` | 是 | — | 非空参数数组 |
| `refresh` | 否 | 关闭 | 仅 snapshot，最小 `250ms` |
| `timeout` | 否 | snapshot `5s` / stream `0` | snapshot 必须为正；stream 可为 `0`（不限时） |
| `autostart` | 否 | `true` | `false` 时等待动作或事件触发 |
| `open` | 否 | 空 | 仅 items-json snapshot；子页路径或选项映射 |
| `on_focus` | 否 | 空 | items-json 列表光标项变化时触发的 action ID |
| `on_select` | 否 | 空 | items-json 列表确认选择变化时触发的 action ID |

### 5.1 mode

- `snapshot`：执行一次，保存结果；可设置 `refresh`。
- `stream`：持续读取输出；不能使用 `items-json`，不能设置 `refresh`；可设置正超时或 `0`。

stream 至少需要两个并发槽位，因此页面通常应设置 `max_parallel: 2` 或更高。

### 5.2 format

- `markdown`：输出按 Markdown 渲染。
- `text`：输出按纯文本显示。
- `items-json`：输出必须是下面的列表协议；只支持 snapshot。

### 5.3 refresh 与 timeout

- `refresh` 从上次执行完成后开始计时，最小 `250ms`。
- 同一面板不会重叠执行；连续刷新会合并为最多一次后续查询。
- snapshot 的 `timeout` 必须大于 0。
- stream 的 `timeout` 为 `0` 表示不限时；正数表示总时长。
- 时长格式为 Go duration，例如 `5s`、`250ms`、`1m`。

## 6. exec 与引用

`exec` 必须是非空数组：

```text
exec: ["program", "literal", {ref: "list.selected.id"}]
```

规则：

- 第一个元素必须是字符串，表示程序名。
- 程序名不能是引用，不能为空。
- 后续元素只能是字符串，或只包含 `ref` 一个键的映射。
- 引用必须占据整个参数，不支持字符串内插、表达式、算术或 shell 求值。
- 数字和布尔值必须显式加引号。
- 参数不会经过 shell 拼接；需要管道时请编写包装脚本。

相对可执行文件路径（包含 `/` 但不是绝对路径）会相对页面文件所在目录解析。

### 6.1 引用类型

| 引用 | 含义 | 生效条件 |
|---|---|---|
| `<panel-id>.selected.id` | 列表当前确认对象的稳定 ID | 被引用 panel 必须是 items-json |
| `<panel-id>.cursor.id` | 列表当前光标所在项的 ID | 被引用 panel 必须是 items-json |
| `<input-id>.value` | 输入控件当前提交值 | 输入必须存在 |
| `parent.selected.id` | 打开当前子页时的父页选中对象 ID | 当前 page 必须声明 `parent` |

`cursor.id` 不要求按 Enter 确认，适合做“焦点预览”；`selected.id` 只反映已确认选择。

### 6.2 缺失依赖

引用缺失时，面板进入 `blocked`：

- 未选择对象：`请选择对象: <panel-id>`
- 未填写必填输入：`请填写参数: <label>`
- 光标列表为空：`列表为空: <panel-id>`
- 子页没有父页上下文：`请选择父页对象`

不会传入空字符串猜测参数，也不会启动进程。

### 6.3 依赖与失效

- `selected.id` 和 `cursor.id` 会建立面板依赖。
- 选择变化会失效依赖 `selected.id` 的下游面板。
- 光标变化会失效依赖 `cursor.id` 的下游面板。
- 输入提交会失效依赖 `.value` 的下游面板。
- 失效会传递到更下游的面板。
- 旧 generation 的结果不会覆盖新状态。
- 依赖图必须无环；自引用或循环引用会报错。

### 6.4 保留名 `parent`

在声明了 `parent` 的子页中，`parent.selected.id` 是保留引用，优先于同名本地 panel。若子页同时需要本地 panel `parent` 的列表引用，应避免使用该 ID。

在根页面（没有 `parent`）中，`parent.selected.id` 可以表示本地 items-json panel `parent` 的选中项。

## 7. input 块

```text
:::mdtui input
id: target
label: 待诊断主机 / IP
default: 127.0.0.1
required: true
max_length: 253
:::
```

| 字段 | 必填 | 默认 | 规则 |
|---|---|---|---|
| `id` | 是 | — | 全局唯一标识 |
| `label` | 否 | `id` | 显示名称 |
| `default` | 否 | 空串 | 不能包含换行；长度受 max_length 限制 |
| `required` | 否 | `false` | 为空时依赖面板 blocked |
| `max_length` | 否 | `256` | 1～4096 个 Unicode 码点 |

输入是单行文本，不用于密码。编辑草稿和已提交值分开：

- `Enter` 开始编辑；编辑中再次按 `Enter` 提交。
- `Esc` 放弃本次编辑。
- `←` / `→` 按字符簇移动光标。
- `Home` / `End` 跳到首尾。
- `Backspace` / `Delete` 删除。
- 输入中的 `q`、`/`、`n` 和动作键都作为普通文字。

## 8. action 块

```text
:::mdtui action
id: inspect
label: 查看详情
key: i
run: detail
:::

:::mdtui action
id: refresh_all
label: 刷新全部
key: r
refresh: [detail, counters]
:::

:::mdtui action
id: restart
label: 重启服务
key: s
run: service_action
confirm: "确定重启当前服务？"
:::
```

| 字段 | 必填 | 默认 | 规则 |
|---|---|---|---|
| `id` | 是 | — | 全局唯一标识 |
| `label` | 否 | `id` | 菜单显示名称 |
| `key` | 否 | 空 | 除 `q` 外的一个小写 ASCII 字母 |
| `run` | 二选一 | — | 目标 panel ID；可针对任意 panel |
| `refresh` | 二选一 | — | 非空 panel ID 数组；只能针对 snapshot |
| `confirm` | 否 | 空 | 非空字符串；存在时弹出确认框 |

规则：

- `run` 和 `refresh` 必须且只能填写一个。
- `run` 是一次性执行，不会启动目标的周期重试。
- `refresh` 可以一次刷新多个 snapshot 面板。
- 快捷键不能重复，`q` 保留。
- 输入编辑时动作键不生效。
- 被 `on_focus` / `on_select` 自动触发的动作不能设置 `confirm`。
- 动作键、F2 菜单和事件绑定可以同时使用。

确认框默认选中“取消”，`Tab` 或左右键切换，`Enter` 执行，`Esc` 关闭。

## 9. open 与子页导航

`open` 只适用于 `format: items-json` 的 snapshot panel。

### 9.1 固定子页

```text
open: ./interface-detail.md
```

### 9.2 按选项映射

```text
open:
  system: ./navigation-system.md
  network: ./navigation-network.md
  process: ./navigation-process.md
```

规则：

- 路径必须是相对路径，不能是绝对路径。
- 路径首尾空白会被去掉；空路径报错。
- 映射的键是列表项 ID，值必须是字符串路径。
- 映射不能为空。
- 没有配置映射的选项不会跳转，界面会显示提示。
- 子页可以继续声明 `open`，最多 8 层。
- 循环引用会按规范化绝对路径检测。
- `check` 会跟随字符串路径和映射中的所有目标。
- 子页若声明 `parent`，它必须与引用它的 panel ID 一致；同一子页被不同 panel 引用时，应省略 `parent` 或拆分文件。

### 9.3 进入与返回

- 在带 `open` 的列表上按 `Enter` 进入子页。
- 子页全屏显示，拥有独立的 panel、input、action 和刷新状态。
- 进入时父页暂停调度并保留选择、滚动、焦点和已有结果。
- 子页内 `Esc` 或 `q` 返回父页；根页 `q` 退出。
- 返回后父页恢复刷新。
- 子页可以用 `{ref: "parent.selected.id"}` 读取打开它时的父页选中项。

### 9.4 与 on_select 的关系

`on_select` 和 `open` 不能同时配置：同一个 Enter 事件不能既执行动作又进入子页。若需要两种行为，请拆成不同列表或不同页面。

`on_focus` 可以与 `open` 共存：光标移动触发预览，Enter 进入子页。

## 10. on_focus 与 on_select

这两个字段只适用于 `format: items-json` 的 panel，值是 action ID。

### 10.1 on_focus

- 列表获得焦点后，光标移动到新的选项时触发。
- 页面初始加载、列表获得焦点且首项可用时触发一次。
- 刷新本身不触发；只有光标所在项的 ID 发生变化时才再次触发。
- 通常配合 `{ref: "<list>.cursor.id"}` 做预览面板。

### 10.2 on_select

- 列表按 `Enter` 真正改变选中值时触发。
- 上下移动光标不触发。
- 同一项重复按 Enter 不触发。
- 通常配合 `{ref: "<list>.selected.id"}` 做完整查询。

### 10.3 自动触发限制

- 被绑定的 action 必须存在。
- action 不能设置 `confirm`。
- 目标动作仍然遵循 `run` / `refresh` 规则。
- 建议把预览/详情 panel 设置为 `autostart: false`，由事件动作显式启动，避免重复执行。

示例：

```text
:::mdtui action
id: preview
run: preview
:::

:::mdtui action
id: inspect
run: detail
:::

:::mdtui panel
id: list
format: items-json
on_focus: preview
on_select: inspect
exec: ["adapter", "list"]
:::

:::mdtui panel
id: preview
autostart: false
exec: ["adapter", "show", {ref: "list.cursor.id"}]
:::

:::mdtui panel
id: detail
autostart: false
exec: ["adapter", "inspect", {ref: "list.selected.id"}]
:::
```

## 11. items-json 协议

`format: items-json` 的面板必须输出：

```json
{
  "version": 1,
  "items": [
    {"id": "eth0", "label": "以太网接口 0", "description": "UP"},
    {"id": "wlan0", "label": "无线接口"}
  ]
}
```

规则：

- 顶层必须是对象，只允许 `version` 和 `items`。
- `version` 必须是整数 `1`。
- `items` 必须是数组，最多 10000 项；空数组是合法的，列表显示为空。
- 每项只允许 `id`、`label`、`description`。
- `id` 和 `label` 必填，`description` 可选且必须是字符串。
- `id` 非空、最长 256 个 Unicode 码点、不能重复、不能含控制字符。
- 未知字段、错误类型、重复 ID 都会让该次查询失败。
- `items-json` 只支持 snapshot，不能用于 stream。
- 列表选中值变化后，引用 `.selected.id` 的 autostart 面板会重新查询；`autostart: false` 的面板需要动作触发。
- 光标移动后，引用 `.cursor.id` 的 autostart 面板会重新查询；通常用 `on_focus` 显式启动预览面板。

## 12. 布局

### stack

面板按声明顺序纵向排列。

### master-detail

```text
layout: master-detail
sidebar: interfaces
```

- 终端宽度至少 100 列时，左侧显示 sidebar，右侧纵向排列其他面板。
- sidebar 宽度约为可用宽度的 30%，最少 24 列。
- 宽度不足时退化为 stack。
- 高度不足时只显示当前面板，用 Tab 切换并显示索引。

### workspace

```text
layout: workspace
sidebar: topics
```

- 终端宽度至少 90 列时，左侧显示 sidebar，右侧一次显示一个详情面板。
- sidebar 宽度约为可用宽度的 28%，最少 26 列。
- `[` / `]` 切换详情视图。
- 列表选中对象后，工作区会优先显示引用该列表的详情面板。
- 窄终端只显示当前焦点面板。

### 焦点与放大

- `Tab` / `Shift+Tab` 切换说明区、输入和面板。
- `Z` 放大当前面板；放大时 `Tab` 切换面板，`Esc` 退出放大。
- 小于 60×16 时显示尺寸不足提示，仍可 `q` 退出。

## 13. 普通 Markdown

扩展块之外的文字会合并为页面说明区，位于面板区上方。

支持：

- 标题层级
- 段落和换行
- 有序/无序列表、任务项
- 引用
- 粗体、斜体、删除线
- 代码块和行内代码
- 链接地址
- GFM 表格
- 图片替代文字

规则：

- 普通正文按区域宽度自动换行。
- 代码块和表格保持原行，可用左右键横向查看。
- HTML 不会执行或加载，按文字显示。
- 图片不会加载，显示替代文字。
- 命令输出只作为内容，永远不会递归执行其中的 `:::mdtui` 声明。
- 搜索只作用于当前焦点区域。

## 14. 校验与错误

运行：

```sh
mdtui check PAGE.md
```

`check` 会：

- 解析并校验根页面
- 递归跟随所有 `open` 目标
- 检查父子声明一致性
- 检查循环引用和 8 层深度
- 检查页面数量上限
- 检查字段、类型、引用、依赖图和容量限制
- 不执行任何命令

错误包含文件、行号、字段路径和修复建议。校验通过不代表目标程序已安装或具有权限。

## 15. 安全边界

- 页面可以执行本机程序，请只运行可信来源的页面。
- `exec` 以参数数组启动进程，不经过 shell。
- 引用只取稳定 ID 或输入值，不做字符串插值。
- 命令输出不会创建动作、修改页面或递归执行扩展声明。
- `check` 不执行命令。
- 需要管道、重定向或复杂参数处理时，由页面作者维护包装脚本。

## 16. 最小页面

```markdown
# 示例

说明文字。

:::mdtui page
version: 1
id: demo
:::

:::mdtui panel
id: hello
title: 问候
exec: ["echo", "hello"]
:::
```

## 17. 字段速查

### page

```text
version, id, title, layout, sidebar, parent, max_parallel
```

### panel

```text
id, title, mode, format, exec, refresh, timeout, autostart, open, on_focus, on_select
```

### input

```text
id, label, default, required, max_length
```

### action

```text
id, label, key, run, refresh, confirm
```

## 18. 可运行示例

- [Linux 运维工作台](../examples/linux-workbench.md)
- [网络调试](../examples/network.md)
- [输入参数](../examples/input.md)
- [页面跳转](../examples/navigation.md)
- [巡检台](../examples/inspection.md)
