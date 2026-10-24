package scheduler

import (
	"strings"
	"testing"
	"time"

	"mdtui/internal/document"
)

func TestTypedInputValidationAndArguments(t *testing.T) {
	minValue, maxValue := int64(1), int64(60)
	p := &document.Page{MaxParallel: 2, Inputs: []document.Input{
		{ID: "verbose", Type: "toggle", Default: "false"},
		{ID: "protocol", Type: "select", Default: "tcp", Options: []document.Option{{Value: "tcp"}, {Value: "udp"}}},
		{ID: "timeout", Type: "number", Default: "5", Min: &minValue, Max: &maxValue, Required: true},
	}, Panels: []document.Panel{{ID: "result", Mode: "snapshot", Autostart: true, Exec: []document.Arg{{Literal: "echo"}, {Ref: "verbose.value"}, {Ref: "protocol.value"}, {Ref: "timeout.value"}}}}}
	f := &fake{make(chan call, 10)}
	e := New(p, f, time.Now)
	e.Start()
	first := receive(t, f)
	if strings.Join(first.spec.Args[1:], ",") != "false,tcp,5" {
		t.Fatal(first.spec.Args)
	}
	finish(t, e, first, "initial")
	before := e.Panels["result"].Generation
	for _, pair := range [][2]string{{"verbose", "1"}, {"protocol", "sctp"}, {"timeout", "61"}, {"timeout", "1.5"}, {"missing", "x"}} {
		if reason := e.Submit(pair[0], pair[1]); reason == "" {
			t.Fatalf("accepted %v", pair)
		}
	}
	if e.Panels["result"].Generation != before || len(f.calls) != 0 {
		t.Fatal("invalid submission changed dependencies")
	}
	if reason := e.Submit("protocol", "udp"); reason != "" {
		t.Fatal(reason)
	}
	next := receive(t, f)
	if next.spec.Args[2] != "udp" {
		t.Fatal(next.spec.Args)
	}
	finish(t, e, next, "updated")
	if reason := e.Submit("timeout", ""); reason != "" {
		t.Fatal(reason)
	}
	if e.Panels["result"].Status != "blocked" {
		t.Fatal("required number did not block execution")
	}
	e.Inputs["protocol"] = "invalid"
	if _, reason := e.Args("result"); reason == "" {
		t.Fatal("invalid stored select value reached command")
	}
}
