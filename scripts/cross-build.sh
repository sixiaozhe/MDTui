#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p bin docs/evidence
export CGO_ENABLED=0 GOOS=linux
for arch in arm64 arm mipsle; do
 GOARCH=$arch GOARM=7 GOMIPS=softfloat go build -trimpath -ldflags '-s -w -X main.version=0.1.0' -o "bin/mdtui-linux-$arch" ./cmd/mdtui
 GOARCH=$arch GOARM=7 GOMIPS=softfloat go build -trimpath -ldflags '-s -w' -o "bin/linux-tools-linux-$arch" ./cmd/linux-tools
done
file bin/mdtui bin/linux-tools bin/mdtui-linux-* bin/linux-tools-linux-* > docs/evidence/cross-build.txt
sha256sum bin/mdtui bin/net-debug bin/linux-tools bin/mdtui-linux-* bin/linux-tools-linux-* > docs/evidence/binaries.sha256
