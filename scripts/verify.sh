#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p docs/evidence
go version > docs/evidence/environment.txt
uname -a >> docs/evidence/environment.txt
go mod verify > docs/evidence/modules.txt
go vet ./... > docs/evidence/vet.txt 2>&1
go test -count=1 -v ./... > docs/evidence/unit.txt 2>&1
go test -count=1 -race ./... > docs/evidence/race.txt 2>&1
bash scripts/build.sh
./bin/mdtui check examples/network.md > docs/evidence/check.txt
./bin/mdtui check examples/input.md >> docs/evidence/check.txt
./bin/mdtui check examples/linux-workbench.md >> docs/evidence/check.txt
./bin/mdtui check examples/navigation.md >> docs/evidence/check.txt
python3 tests/integration/cli_test.py > docs/evidence/cli.json
python3 tests/integration/pty_test.py > docs/evidence/pty.json
python3 tests/integration/linux_workbench_test.py > docs/evidence/linux-workbench.json
python3 tests/integration/latency.py > docs/evidence/latency.json
printf 'Verification passed. Evidence: docs/evidence/\n'
