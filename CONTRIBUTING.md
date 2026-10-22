# 参与 MDTUI

欢迎提交错误报告、设备兼容反馈、页面示例和代码改进。提 Issue 时请说明 MDTUI 版本、Linux 发行版与架构、终端类型和窗口尺寸、复现步骤，以及预期与实际结果。请删除日志中的主机名、IP、令牌和设备专有信息后再公开。

## 本地开发

需要 Linux 或 WSL、Go 1.24+ 和 Python 3。运行：

```sh
go test ./...
bash scripts/build.sh
./bin/mdtui check examples/linux-workbench.md
python3 tests/integration/linux_workbench_test.py
```

提交前运行 `bash scripts/verify.sh`。它会生成本地 `docs/evidence/`，该目录已被 Git 忽略。修改页面语法时，请更新 `docs/author.md` 和编译器测试；修改交互时，请更新 `docs/user.md`、帮助文本及 TUI 测试。新增适配器应明确输入校验、超时、输出上限和所需系统权限。

页面可以执行程序，只接受可信来源的页面；不要在示例中放入真实凭据或设备地址。请为新增代码提供与现有项目兼容的 MIT 授权，并保留第三方许可证声明。
