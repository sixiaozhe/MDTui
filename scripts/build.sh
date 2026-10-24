#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
if ! command -v go >/dev/null 2>&1; then
  printf '错误：找不到 go 命令。请在当前 Linux / WSL 环境安装 Go 1.24+，并将 Go 的 bin 目录加入 PATH。\n' >&2
  printf '运行 go version 确认可用后重试。详情见 docs/build.md。\n' >&2
  exit 1
fi
mkdir -p bin
export CGO_ENABLED=0
mdtui_tmp="bin/.mdtui.tmp"
fixture_tmp="bin/.net-debug.tmp"
linux_tmp="bin/.linux-tools.tmp"
cleanup() { rm -f "$mdtui_tmp" "$fixture_tmp" "$linux_tmp"; }
trap cleanup EXIT
go build -trimpath -ldflags '-s -w -X main.version=0.1.0' -o "$mdtui_tmp" ./cmd/mdtui
mv -f "$mdtui_tmp" bin/mdtui
go build -trimpath -ldflags '-s -w' -o "$fixture_tmp" ./cmd/net-debug
mv -f "$fixture_tmp" bin/net-debug
go build -trimpath -ldflags '-s -w' -o "$linux_tmp" ./cmd/linux-tools
mv -f "$linux_tmp" bin/linux-tools
