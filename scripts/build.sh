#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
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
