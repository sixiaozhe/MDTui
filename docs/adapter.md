# 命令适配指南

命令从页面目录执行，继承用户、环境与 PATH；路径含分隔符时相对页面目录解析。stdin 为空设备，不分配子进程 PTY。stdout 是内容，stderr 是诊断；两者分别采集，不保证跨流顺序。退出码非 0 为失败。

## 稳定选择协议

```json
{"version":1,"items":[{"id":"eth0","label":"以太网 0","description":"UP"}]}
```

stdout 必须是完整单个 JSON 对象。version=1，items 为数组，≤10000 项。id 为非空字符串，≤256 Unicode 码点，不含控制字符，不重复；label 必填字符串，description 可选。未知字段被拒绝。

ID 应来自业务主键，不能用显示行号或易变文案。首次加载不确认选择，Enter 才确认；刷新按 ID 保留选择。列表丢失已选 ID 会取消下游、清除旧上下文。协议错误保留旧列表，不部分覆盖。

## 查询与日志

已有 Markdown 命令可直接作为 snapshot。旧命令可由独立包装脚本转出 items-json；包装脚本须能单独运行并维护错误码。需要管道时在脚本内明确实现，工作台不会代为拼接 shell 命令。

日志用 stream，及时 flush stdout，避免 isatty 才输出。支持多次 write 拆分 UTF-8 字符。默认保留最近 2 MiB/10000 行，单行 64 KiB；超限截断有提示。stdout 快照超 4 MiB 会终止命令；stderr 保留最近 64 KiB。

不要启动交互式终端程序、读取 stdin 或 daemonize。取消时对进程组 SIGTERM，最多等待 1 秒，然后 SIGKILL；退出的 leader 遗留组内进程也清理。主动脱离进程组不保证清理。

演示适配器源代码在 `cmd/net-debug`。替换示例页面中的 exec 即可接入实际设备命令，无须修改工作台源码。
