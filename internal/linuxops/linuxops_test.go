package linuxops

import (
	"context"
	"errors"
	"strings"
	"testing"

	"mdtui/internal/content"
)

type fakeExecutor struct {
	calls []invocation
	text  string
	err   error
}

func (f *fakeExecutor) Run(_ context.Context, program string, args ...string) (string, error) {
	f.calls = append(f.calls, invocation{program, append([]string(nil), args...)})
	return f.text, f.err
}

func TestTopicCatalogIsSelectable(t *testing.T) {
	items, err := content.Items([]byte(Topics()))
	if err != nil || len(items) != 6 {
		t.Fatalf("invalid MDTUI catalog: %d items, %v", len(items), err)
	}
	for _, item := range items {
		if !validTopic(item.ID) || item.Label == "" {
			t.Fatalf("unusable topic: %+v", item)
		}
	}
}

func TestDiagnosticValidationAndArgv(t *testing.T) {
	f := &fakeExecutor{text: "ok"}
	for _, bad := range []string{"-h", "host;reboot", "a b", "a/../b"} {
		if _, err := Render(context.Background(), f, "diagnose", bad); err == nil {
			t.Fatalf("accepted unsafe target %q", bad)
		}
	}
	if len(f.calls) != 0 {
		t.Fatal("ran command before validation")
	}
	_, err := Render(context.Background(), f, "diagnose", "host.example")
	if err != nil || len(f.calls) != 2 {
		t.Fatalf("diagnostic failed: %v, calls=%d", err, len(f.calls))
	}
	for _, call := range f.calls {
		if call.program == "sh" || call.program == "bash" || call.args[len(call.args)-1] != "host.example" {
			t.Fatalf("unexpected command: %+v", call)
		}
	}
}

func TestNonzeroStatusShowsUsefulOutput(t *testing.T) {
	f := &fakeExecutor{text: "inactive (dead)", err: errors.New("exit status 3")}
	result, err := Render(context.Background(), f, "service", "ssh")
	if err != nil || !strings.Contains(result, "inactive (dead)") || !strings.Contains(result, "返回非零状态") {
		t.Fatalf("lost command output: %q, %v", result, err)
	}
}

func TestExplainListsCommandsWithoutExecuting(t *testing.T) {
	out, err := Explain("network")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"网络与端口 · 采集计划", "ip -br addr", "ss -tuln"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %s", want, out)
		}
	}
	if _, err := Explain("missing"); err == nil {
		t.Fatal("accepted unknown topic")
	}
}
