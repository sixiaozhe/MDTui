# 参数传递演示
Tab 聚焦输入，Enter 编辑并提交。空格、引号、分号和 $() 都只作为参数。

:::mdtui page
version: 1
id: parameters
:::
:::mdtui input
id: filter
label: 过滤关键字
default: ""
max_length: 128
:::
:::mdtui panel
id: echo
format: text
exec: ["../bin/net-debug", "echo", {ref: "filter.value"}]
:::
