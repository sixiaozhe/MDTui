package scheduler

import (
	"strings"
	"testing"
	"time"

	"mdtui/internal/document"
	"mdtui/internal/runner"
)

func TestStreamCoalescesButCompletionFlushes(t *testing.T) {
	for _, format := range []string{"text", "markdown"} {
		t.Run(format, func(t *testing.T) {
			now := time.Unix(100, 0)
			page := &document.Page{MaxParallel: 2, Panels: []document.Panel{{ID: "log", Mode: "stream", Format: format, Exec: []document.Arg{{Literal: "log"}}}}}
			f := &fake{make(chan call, 2)}
			e := New(page, f, func() time.Time { return now })
			e.StreamBytes, e.StreamLines = 64<<10, 3
			if format == "text" {
				e.StreamInterval = 500 * time.Millisecond
			}
			e.Run("log")
			c := receive(t, f)
			p := e.Panels["log"]
			p.active.buffer.Write([]byte("first\n"))
			if !e.Tick() {
				t.Fatal("first update missing")
			}
			p.active.buffer.Write([]byte("second\n"))
			now = now.Add(100 * time.Millisecond)
			if e.Tick() || strings.Contains(p.Text, "second") {
				t.Fatal("update was not coalesced")
			}
			now = now.Add(400 * time.Millisecond)
			if !e.Tick() || !strings.Contains(p.Text, "second") {
				t.Fatal("coalesced update lost")
			}
			p.active.buffer.Write([]byte("third\nfinal\n"))
			c.reply <- runner.Result{}
			e.Consume(<-e.Events)
			if !strings.HasSuffix(p.Text, "final\n") || p.Dropped != 1 {
				t.Fatalf("final bounded output lost: %+v", p)
			}
		})
	}
}

func TestLowResourceParentBudgetPreservesSelectionAndOneShot(t *testing.T) {
	e, f, _ := loaded(t)
	e.Panels["list"].Selected = "A"
	p := e.Panels["detail"]
	p.Text, p.Status = strings.Repeat("x", 2048), "ready"
	p.pausedOnce = true
	e.TrimPausedText(1024) // Active pages must not be trimmed.
	if len(p.Text) != 2048 {
		t.Fatal("trimmed active page")
	}
	e.Pause()
	p.pausedOnce = true
	e.TrimPausedText(1024)
	if len(p.Text) >= 1024 || !p.Stale || e.Panels["list"].Selected != "A" || len(e.Panels["list"].Items) == 0 {
		t.Fatal("lost navigation state")
	}
	e.Resume()
	if len(f.calls) != 0 {
		t.Fatal("cache eviction reran a one-shot command")
	}
}
