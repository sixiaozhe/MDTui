# 构建、部署和卸载

Linux 构建环境：Go ≥1.24。go.mod 固定直接/间接版本，go.sum 固定内容摘要；`go mod verify` 校验模块缓存。生产构建 CGO_ENABLED=0，无额外运行时。

```sh
go mod download
go mod verify
go test ./...
bash scripts/build.sh
./bin/mdtui --version
```

`build.sh` 同时生成 `bin/mdtui`、模拟示例所需的 `bin/net-debug` 和 Linux 运维示例所需的 `bin/linux-tools`。运行 `examples/linux-workbench.md` 时，应将页面与对应架构的两个二进制一起部署，并保持页面中 `../bin/linux-tools` 的相对目录关系。

构建脚本启用 trimpath、去除调试符号、版本 0.1.0，并通过固定临时文件替换目标，避免正在运行的旧二进制被静默保留，同时保持构建 ID 可重复。不要并发运行同一目录下的构建脚本。要自定义版本：

```sh
CGO_ENABLED=0 go build -trimpath -ldflags '-s -w -X main.version=0.1.0-custom' -o mdtui ./cmd/mdtui
```

## 交叉编译

先确认设备架构，不能从 Linux 名称推断 CPU/ABI。示例：

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o bin/mdtui-linux-arm64 ./cmd/mdtui
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -o bin/mdtui-linux-armv7 ./cmd/mdtui
CGO_ENABLED=0 GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -trimpath -o bin/mdtui-linux-mipsle ./cmd/mdtui
```

`bash scripts/cross-build.sh` 还会为相同架构生成 `linux-tools-linux-*`。部署时将对应文件命名为页面引用的 `linux-tools`，并放在页面上一级目录的 `bin/` 中。

示例不构成目标硬件验证。实际还需核对内核版本、端序、ARM 指令集/浮点约束、内存和 TERM。

上传匹配的 mdtui、页面及必要适配器，在用户有执行权限的目录 `chmod +x mdtui adapter`。先 check，再从 SSH PTY 启动。不需要提权或服务安装。卸载仅删除自己部署的程序、页面、适配器，以及显式创建的日志文件；不修改系统配置。

## 开发环境与部署边界

参考开发环境是 Ubuntu 22.04 / WSL2；标准 Linux 构建环境直接运行上述命令即可。`.tools/`、`bin/` 和 `docs/evidence/` 是本地生成目录，不进入源码仓库。

交叉编译成功只证明对应目标格式的程序能够构建。部署前请确认目标机器的 CPU、ABI、内核和终端支持；再在真实 SSH PTY 中进行功能与资源测试。当前尚无真实设备验收结果。
