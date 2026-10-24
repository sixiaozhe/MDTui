# 页面作者指南

完整语法参考见 [syntax.md](syntax.md)，完整字段契约见 [需求基线第 5 节](requirements.md)。可复制 [network.md](../examples/network.md) 或 [input.md](../examples/input.md) 开始编写。

需要表单时可复制 [controls.md](../examples/controls.md)：input 的 type 可选 text（默认）、toggle、select、number。select 使用静态 options 区分显示 label 和命令 value；number 可声明 min/max 整数边界。所有控件继续使用 `{ref: "参数ID.value"}` 传字符串。给现有 action 增加 `button: true` 即可显示可聚焦按钮，并保留快捷键、依赖校验和 confirm 行为。

所有扩展从第 1 列的 `:::mdtui TYPE` 开始，单独一行 `:::` 结束。围栏代码块中的声明只显示，不执行。普通 Markdown 合并成页面上方说明区。

必须有且只有一个 page，并放在其他声明前。所有 ID 共享命名空间，格式 `[A-Za-z][A-Za-z0-9_-]{0,63}`。未知类型/字段、重复键/ID、别名/合并键、自定义标签和循环依赖均报错。页面 ≤1 MiB、≤16 panel、≤16 input、≤32 action，YAML 嵌套 ≤8。

```text
:::mdtui page
version: 1
id: diagnostics
layout: stack
max_parallel: 2
:::
```

主从布局使用 `layout: master-detail` 与 `sidebar: 面板ID`。max_parallel 为 1～8。

任务较多时可用 `layout: workspace` 和 `sidebar: 列表面板ID`。宽终端左侧保留列表、右侧一次显示一个详情面板；窄终端显示当前焦点面板。列表选中对象后，工作区会自动显示首个引用该列表的详情面板。`[` / `]` 切换详情，Tab 切换焦点，Z 放大当前面板并在放大时用 Tab 切换面板，Esc 退出放大。见 [Linux 运维工作台](../examples/linux-workbench.md)。

panel 声明必填 id、exec；可选 title、mode（snapshot 默认/stream）、format（markdown 默认/text/items-json）、open（仅 items-json snapshot 可用；单个相对路径，或选项 ID 到子页路径的映射）、on_focus、on_select、refresh、timeout、autostart（默认 true）。snapshot 默认超时 5s，refresh 最小 250ms；stream 默认超时 0（不限时），不能设置 refresh 或 items-json。

panel 可设置 `max_height: 6` 限制总高度（包含上下边框，最小 3 行），省略时自动分配。高度上限也适用于放大视图；超出的内容保留并可滚动。堆叠及主从布局右列会将省出的空间分配给其他未达到上限的面板。示例见 [panel-height.md](../examples/panel-height.md)。

列表可以通过 `open: ./child.md` 进入同一个子页，也可以用映射让不同选项打开不同子页：

```text
open:
  system: ./system.md
  network: ./network.md
```

未出现在映射中的选项不会跳转。子页在 page 块声明 `parent: 列表ID` 后，可用 `{ref: "parent.selected.id"}` 读取父页选中对象。子页全屏显示并拥有独立面板；`Esc` 或 `q` 返回父页，父页的状态和刷新会恢复。`mdtui check` 会递归校验子页、循环引用和 8 层深度限制。示例见 [navigation.md](../examples/navigation.md)。

`on_focus` 和 `on_select` 绑定已有 action，用于区分“列表项获得焦点”和“选项被确认”：

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

`on_focus` 在列表光标移动到新的选项时触发；页面初始加载时，聚焦列表的首个光标项也会触发一次。`on_select` 仅在列表按 Enter 改变选中值时触发。刷新本身不触发，只有光标项 ID 变化才会再次触发 `on_focus`。自动触发动作不能设置 `confirm`，两个事件都只能用于 items-json 面板，`on_select` 不能与 `open` 同时使用。可运行示例见 [inspection.md](../examples/inspection.md)。

exec 是非空数组，程序名必须为字面字符串，后续参数只接受字符串或 `{ref: "列表.selected.id"}`、`{ref: "列表.cursor.id"}`、`{ref: "输入.value"}`、`{ref: "parent.selected.id"}`。数字参数必须加引号。引用占整个参数，不支持插值、表达式和 shell 求值。缺失选择/必填输入会阻塞，不会传空值猜测参数。

input 的字段为 id、label（默认 ID）、default（默认空串）、required（默认 false）、max_length（默认 256，1～4096 Unicode 码点）。单行输入支持末尾追加和 Backspace；编辑草稿与提交值独立。不用于密码。

action 的字段为 id、label（默认 ID）、key、confirm，以及二选一的 `run: 面板ID` 或 `refresh: [面板ID,...]`。refresh 只能针对 snapshot，run 可针对任意面板；常配合 `autostart: false`。run 为一次性执行，不启动目标的周期重试；需要持续查询时用 refresh。key 为除 q 外单个小写字母，可不设置并从 F2 触发。不要复用动作键。对改变设备状态的命令显式配置非空 confirm。

先运行 `mdtui check page.md`；校验输出包含文件、行号、字段与修正提示，不探测也不执行程序。配置检查通过不保证目标程序已安装；程序缺失在运行时显示面板错误。

Markdown 在终端中保留标题层级、段落、列表、任务项、引用、强调、代码块、链接地址和 GFM 表格的视觉样式。普通正文自动换行；代码块和表格保持原行，左右键可横向查看。HTML/图片不会执行或加载；HTML 显示为文字，图片显示替代文字。命令输出只作为内容，永不递归执行扩展。
