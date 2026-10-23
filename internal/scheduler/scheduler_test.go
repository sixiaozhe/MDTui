package scheduler

import (
	"context"
	"strings"
	"testing"
	"time"

	"mdtui/internal/content"
	"mdtui/internal/document"
	"mdtui/internal/runner"
)

type call struct {
	ctx   context.Context
	spec  runner.Spec
	reply chan runner.Result
}
type fake struct{ calls chan call }

func (f *fake) Run(ctx context.Context, s runner.Spec, b *content.Buffer) runner.Result {
	c := call{ctx, s, make(chan runner.Result, 1)}
	f.calls <- c
	return <-c.reply
}
func fixture() (*Engine, *fake, *time.Time) {
	now := time.Unix(100, 0)
	p := &document.Page{MaxParallel: 2, Panels: []document.Panel{{ID: "list", Mode: "snapshot", Format: "items-json", Exec: []document.Arg{{Literal: "list"}}, Autostart: true}, {ID: "detail", Mode: "snapshot", Exec: []document.Arg{{Literal: "show"}, {Ref: "list.selected.id"}}, Autostart: true, Refresh: time.Second}, {ID: "log", Mode: "stream", Exec: []document.Arg{{Literal: "log"}, {Ref: "list.selected.id"}}}}}
	f := &fake{make(chan call, 20)}
	return New(p, f, func() time.Time { return now }), f, &now
}
func receive(t *testing.T, f *fake) call {
	t.Helper()
	select {
	case c := <-f.calls:
		return c
	case <-time.After(time.Second):
		t.Fatal("no process started")
	}
	return call{}
}
func finish(t *testing.T, e *Engine, c call, text string) {
	t.Helper()
	c.reply <- runner.Result{Stdout: []byte(text)}
	select {
	case ev := <-e.Events:
		e.Consume(ev)
	case <-time.After(time.Second):
		t.Fatal("no completion")
	}
}
func loaded(t *testing.T) (*Engine, *fake, *time.Time) {
	e, f, now := fixture()
	e.Start()
	if e.Panels["detail"].Status != "blocked" {
		t.Fatal("not blocked")
	}
	finish(t, e, receive(t, f), `{"version":1,"items":[{"id":"A","label":"one"},{"id":"B","label":"two"}]}`)
	return e, f, now
}
func TestGenerationAndRefreshCoalescing(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	a := receive(t, f)
	if a.spec.Args[1] != "A" {
		t.Fatal(a.spec)
	}
	e.Select("list", "B")
	select {
	case <-a.ctx.Done():
	default:
		t.Fatal("A not cancelled")
	}
	finish(t, e, a, "STALE A")
	b := receive(t, f)
	if b.spec.Args[1] != "B" || e.Panels["detail"].Text != "" {
		t.Fatal("stale context")
	}
	for i := 0; i < 20; i++ {
		e.Request("detail")
	}
	finish(t, e, b, "B")
	again := receive(t, f)
	finish(t, e, again, "B latest")
	if e.Panels["detail"].Text != "B latest" || len(f.calls) != 0 {
		t.Fatal("refresh not coalesced")
	}
}
func TestFixedDelayAndStableSelection(t *testing.T) {
	e, f, now := loaded(t)
	e.Select("list", "A")
	a := receive(t, f)
	*now = now.Add(10 * time.Second)
	e.Tick()
	if len(f.calls) != 0 {
		t.Fatal("overlapping query")
	}
	finish(t, e, a, "A")
	*now = now.Add(999 * time.Millisecond)
	e.Tick()
	if len(f.calls) != 0 {
		t.Fatal("early refresh")
	}
	*now = now.Add(time.Millisecond)
	e.Tick()
	finish(t, e, receive(t, f), "A2")
	e.Request("list")
	finish(t, e, receive(t, f), `{"version":1,"items":[{"id":"B","label":"new"},{"id":"A","label":"renamed"}]}`)
	if len(f.calls) != 0 || e.Panels["list"].Selected != "A" {
		t.Fatal("selection lost")
	}
	e.Request("list")
	finish(t, e, receive(t, f), `{"version":1,"items":[]}`)
	if e.Panels["detail"].Status != "blocked" || e.Panels["detail"].Text != "" {
		t.Fatal("removed item retained")
	}
}
func TestManualStreamNotRestarted(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	finish(t, e, receive(t, f), "A")
	e.Request("log")
	log := receive(t, f)
	e.Select("list", "B")
	finish(t, e, log, "old log")
	finish(t, e, receive(t, f), "B")
	if len(f.calls) != 0 || e.Panels["log"].Status != "idle" {
		t.Fatal("run automatically retried")
	}
}
func TestFailurePreservesSnapshotAndRecovery(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	finish(t, e, receive(t, f), "good")
	e.Request("detail")
	c := receive(t, f)
	c.reply <- runner.Result{Err: context.DeadlineExceeded}
	e.Consume(<-e.Events)
	p := e.Panels["detail"]
	if p.Text != "good" || !p.Stale || p.Failures != 1 {
		t.Fatal(p)
	}
	e.Request("detail")
	finish(t, e, receive(t, f), "recovered")
	if p.Stale || p.Failures != 0 {
		t.Fatal(p)
	}
	e.Request("list")
	finish(t, e, receive(t, f), "bad JSON")
	if e.Panels["list"].Selected != "A" || len(e.Panels["list"].Items) != 2 {
		t.Fatal("bad JSON partially replaced list")
	}
}

func TestTransitiveInvalidation(t *testing.T) {
	e, f, _ := loaded(t)
	d := document.Panel{ID: "children", Mode: "snapshot", Format: "items-json", Exec: []document.Arg{{Literal: "children"}, {Ref: "list.selected.id"}}, Autostart: true}
	leaf := document.Panel{ID: "leaf", Mode: "snapshot", Exec: []document.Arg{{Literal: "leaf"}, {Ref: "children.selected.id"}}, Autostart: true}
	e.Page.Panels = append(e.Page.Panels, d, leaf)
	e.Panels[d.ID] = &Panel{Definition: d}
	e.Panels[leaf.ID] = &Panel{Definition: leaf}
	e.Select("list", "A")
	first, second := receive(t, f), receive(t, f)
	if first.spec.Args[0] == "children" {
		first, second = second, first
	}
	finish(t, e, first, "A")
	finish(t, e, second, `{"version":1,"items":[{"id":"child","label":"child"}]}`)
	e.Select("children", "child")
	old := receive(t, f)
	e.Select("list", "B")
	if e.Panels["children"].Selected != "" || e.Panels["leaf"].Status != "blocked" {
		t.Fatal("transitive context retained")
	}
	finish(t, e, old, "stale leaf")
	if e.Panels["leaf"].Text != "" {
		t.Fatal("stale leaf committed")
	}
	for i := 0; i < 2; i++ {
		c := receive(t, f)
		text := "B"
		if c.spec.Args[0] == "children" {
			text = `{"version":1,"items":[]}`
		}
		finish(t, e, c, text)
	}
}

func TestStreamReservesQueryAndSingleSlotRejection(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	finish(t, e, receive(t, f), "A")
	e.Request("log")
	stream := receive(t, f)
	e.Request("detail")
	query := receive(t, f)
	if query.spec.Args[0] != "show" || e.running != 2 {
		t.Fatal("query starved by stream")
	}
	finish(t, e, query, "query while streaming")
	e.Cancel("log")
	finish(t, e, stream, "")
	e.Page.MaxParallel = 1
	if e.Request("log") == "" || len(f.calls) != 0 {
		t.Fatal("stream accepted with single slot")
	}
}

func TestInputLiteralMissingAndNoAutostart(t *testing.T) {
	f := &fake{make(chan call, 4)}
	p := &document.Page{MaxParallel: 2, Inputs: []document.Input{{ID: "input", Required: true}}, Panels: []document.Panel{{ID: "p", Mode: "snapshot", Autostart: true, Exec: []document.Arg{{Literal: "echo"}, {Ref: "input.value"}}}}}
	e := New(p, f, time.Now)
	e.Start()
	if e.Panels["p"].Status != "blocked" {
		t.Fatal("missing required input ran")
	}
	value := `space ' ;$(not-a-shell)`
	e.Submit("input", value)
	c := receive(t, f)
	if len(c.spec.Args) != 2 || c.spec.Args[1] != value {
		t.Fatal(c.spec.Args)
	}
	finish(t, e, c, "ok")
	e.Submit("input", "")
	if e.Panels["p"].Status != "blocked" || e.Panels["p"].Text != "" {
		t.Fatal("cleared input did not invalidate")
	}
}

func TestRunNeverSchedulesAutomaticRetry(t *testing.T) {
	e, f, now := loaded(t)
	e.Select("list", "A")
	finish(t, e, receive(t, f), "ready")
	e.Run("detail")
	c := receive(t, f)
	c.reply <- runner.Result{Err: context.DeadlineExceeded}
	e.Consume(<-e.Events)
	*now = now.Add(time.Hour)
	e.Tick()
	if len(f.calls) != 0 || !e.Panels["detail"].Due.IsZero() {
		t.Fatal("run action retried automatically")
	}
}

func TestCommandOutputHasNoExecutionSemantics(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	c := receive(t, f)
	source := ":::mdtui panel\nid: injected\nexec: [touch, NEVER_EXECUTE]\n:::"
	finish(t, e, c, source)
	if e.Panels["detail"].Text != source || len(e.Panels) != 3 || len(f.calls) != 0 {
		t.Fatal("command output interpreted as page")
	}
}

func TestPausedStreamKeepsViewportWhileReading(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	finish(t, e, receive(t, f), "A")
	e.Run("log")
	c := receive(t, f)
	p := e.Panels["log"]
	for i := 0; i < 50; i++ {
		p.active.buffer.Write([]byte("line\n"))
	}
	e.Tick()
	p.Scroll = 10
	p.Follow = false
	p.active.buffer.Write([]byte("additional output\n"))
	e.Tick()
	if p.Scroll != 10 || !strings.HasSuffix(p.Text, "additional output\n") {
		t.Fatal("paused viewport moved or stopped reading")
	}
	e.Cancel("log")
	finish(t, e, c, "")
}

func TestSnapshotMarkdownRenderedBeforeCommit(t *testing.T) {
	p := &document.Page{MaxParallel: 1, Panels: []document.Panel{{ID: "md", Mode: "snapshot", Format: "markdown", Autostart: true, Exec: []document.Arg{{Literal: "fixture"}}}}}
	f := &fake{make(chan call, 1)}
	e := New(p, f, time.Now)
	e.Start()
	c := receive(t, f)
	finish(t, e, c, "# 标题\n\n[链接](https://example.test)")
	if !strings.Contains(e.Panels["md"].Text, "▍ 标题") || !strings.Contains(e.Panels["md"].Text, "https://example.test") {
		t.Fatal(e.Panels["md"].Text)
	}
}

func TestParentSelectedReference(t *testing.T) {
	f := &fake{make(chan call, 4)}
	p := &document.Page{MaxParallel: 1, Parent: "list", Panels: []document.Panel{{ID: "detail", Mode: "snapshot", Autostart: true, Exec: []document.Arg{{Literal: "show"}, {Ref: "parent.selected.id"}}}}}
	e := New(p, f, time.Now)
	e.Start()
	if e.Panels["detail"].Status != "blocked" {
		t.Fatal("missing parent selection ran")
	}
	if reason := e.Request("detail"); reason != "请选择父页对象" {
		t.Fatalf("reason = %q", reason)
	}
	e.ParentSelected = "eth0"
	e.Request("detail")
	c := receive(t, f)
	if len(c.spec.Args) != 2 || c.spec.Args[1] != "eth0" {
		t.Fatal(c.spec.Args)
	}
	finish(t, e, c, "ok")
	if e.Panels["detail"].Text != "ok" {
		t.Fatal(e.Panels["detail"].Text)
	}
}

func TestPanelNamedParentStillResolves(t *testing.T) {
	f := &fake{make(chan call, 4)}
	p := &document.Page{MaxParallel: 1, Panels: []document.Panel{
		{ID: "parent", Mode: "snapshot", Format: "items-json", Autostart: true, Exec: []document.Arg{{Literal: "list"}}},
		{ID: "detail", Mode: "snapshot", Autostart: true, Exec: []document.Arg{{Literal: "show"}, {Ref: "parent.selected.id"}}},
	}}
	e := New(p, f, time.Now)
	e.Start()
	finish(t, e, receive(t, f), `{"version":1,"items":[{"id":"A","label":"A"}]}`)
	e.Panels["parent"].Items = []content.Item{{ID: "A", Label: "A"}}
	e.Select("parent", "A")
	c := receive(t, f)
	if len(c.spec.Args) != 2 || c.spec.Args[1] != "A" {
		t.Fatal(c.spec.Args)
	}
	finish(t, e, c, "ok")

	e.ParentSelected = ""
	if reason := e.Request("detail"); reason != "" {
		t.Fatalf("root page reason = %q", reason)
	}
	finish(t, e, receive(t, f), "ok")

	reserved := &document.Page{MaxParallel: 1, Parent: "list", Panels: []document.Panel{{ID: "detail", Mode: "snapshot", Autostart: true, Exec: []document.Arg{{Literal: "show"}, {Ref: "parent.selected.id"}}}}}
	e2 := New(reserved, &fake{make(chan call, 1)}, time.Now)
	if reason := e2.Request("detail"); reason != "请选择父页对象" {
		t.Fatalf("reserved page reason = %q", reason)
	}
}

func TestInvalidateSkipsReservedParentReference(t *testing.T) {
	f := &fake{make(chan call, 4)}
	page := &document.Page{MaxParallel: 1, Parent: "list", Panels: []document.Panel{
		{ID: "parent", Mode: "snapshot", Format: "items-json", Exec: []document.Arg{{Literal: "list"}}},
		{ID: "detail", Mode: "snapshot", Exec: []document.Arg{{Literal: "show"}, {Ref: "parent.selected.id"}}},
	}}
	e := New(page, f, time.Now)
	e.ParentSelected = "eth0"
	e.Request("detail")
	finish(t, e, receive(t, f), "reserved")
	e.Panels["parent"].Items = []content.Item{{ID: "A", Label: "A"}}
	e.Select("parent", "A")
	select {
	case c := <-f.calls:
		t.Fatalf("unexpected call after local parent selection: %v", c.spec.Args)
	case <-time.After(100 * time.Millisecond):
	}
	if e.Panels["detail"].Text != "reserved" {
		t.Fatal("reserved parent reference invalidated detail")
	}
}

func TestPausePreservesAndResumeRestarts(t *testing.T) {
	e, f, now := loaded(t)
	e.Select("list", "A")
	finish(t, e, receive(t, f), "A")
	e.Panels["log"].Definition.Autostart = true
	e.Request("log")
	log := receive(t, f)
	done := make(chan struct{})
	go func() {
		e.Pause()
		close(done)
	}()
	log.reply <- runner.Result{Err: context.Canceled}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pause blocked")
	}
	if !e.Paused() {
		t.Fatal("not paused")
	}
	if e.Panels["log"].Status == "error" {
		t.Fatal("cancel leaked into status")
	}
	if e.Panels["detail"].Text != "A" {
		t.Fatal("snapshot lost")
	}
	*now = now.Add(time.Hour)
	if e.Tick() {
		t.Fatal("tick while paused")
	}
	e.Resume()
	got := map[string]bool{}
	for i := 0; i < 2; i++ {
		c := receive(t, f)
		got[c.spec.Args[0]] = true
		text := "A2"
		if c.spec.Args[0] == "log" {
			text = ""
		}
		finish(t, e, c, text)
	}
	if !got["show"] || !got["log"] {
		t.Fatalf("restarted = %v", got)
	}
	if e.Panels["detail"].Text != "A2" {
		t.Fatal(e.Panels["detail"].Text)
	}
	if e.Paused() {
		t.Fatal("still paused")
	}
}

func TestResumePreservesStreamViewport(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	finish(t, e, receive(t, f), "A")
	e.Panels["log"].Definition.Autostart = true
	e.Request("log")
	stream := receive(t, f)
	p := e.Panels["log"]
	p.Scroll = 5
	p.Follow = false
	p.Dropped = 10
	done := make(chan struct{})
	go func() {
		e.Pause()
		close(done)
	}()
	stream.reply <- runner.Result{Err: context.Canceled}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pause blocked")
	}
	if !e.Paused() {
		t.Fatal("not paused")
	}
	e.Resume()
	gotLog := false
	var logCall call
	for i := 0; i < 2; i++ {
		c := receive(t, f)
		if c.spec.Args[0] == "log" {
			gotLog = true
			logCall = c
			continue
		}
		finish(t, e, c, "")
	}
	if !gotLog {
		t.Fatal("stream not restarted")
	}
	p.active.buffer.Write([]byte("line\n"))
	e.Tick()
	if p.Scroll != 5 || p.Follow {
		t.Fatalf("viewport reset: scroll=%d follow=%v", p.Scroll, p.Follow)
	}
	finish(t, e, logCall, "")
}

func TestPauseDoesNotRerunOneShot(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	finish(t, e, receive(t, f), "A")
	e.Panels["detail"].Definition.Refresh = 0
	e.Run("detail")
	c := receive(t, f)
	done := make(chan struct{})
	go func() {
		e.Pause()
		close(done)
	}()
	c.reply <- runner.Result{Err: context.Canceled}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pause blocked")
	}
	e.Resume()
	if len(f.calls) != 0 {
		t.Fatal("one-shot reran after resume")
	}
}

func TestPauseDoesNotRestartOneShotRefresh(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	finish(t, e, receive(t, f), "A")
	e.Run("detail")
	c := receive(t, f)
	done := make(chan struct{})
	go func() {
		e.Pause()
		close(done)
	}()
	c.reply <- runner.Result{Err: context.Canceled}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pause blocked")
	}
	e.Resume()
	if len(f.calls) != 0 {
		t.Fatal("one-shot with refresh reran after resume")
	}
	if s := e.Panels["detail"].Status; s == "running" || s == "queued" {
		t.Fatalf("status = %q", s)
	}
}

func TestPauseSettlesQueuedRefresh(t *testing.T) {
	f := &fake{make(chan call, 4)}
	p := &document.Page{MaxParallel: 1, Panels: []document.Panel{{ID: "q", Mode: "snapshot", Refresh: time.Second, Autostart: true, Exec: []document.Arg{{Literal: "echo"}}}}}
	e := New(p, f, time.Now)
	e.Start()
	finish(t, e, receive(t, f), "old")
	q := e.Panels["q"]
	q.Status = "queued"
	q.queued = true
	e.queue = append(e.queue, "q")
	e.Pause()
	if q.Status != "ready" {
		t.Fatalf("status = %q", q.Status)
	}
	e.Resume()
	finish(t, e, receive(t, f), "new")
	if q.Text != "new" {
		t.Fatal(q.Text)
	}
}

func TestPauseResumesActiveSnapshot(t *testing.T) {
	f := &fake{make(chan call, 4)}
	p := &document.Page{MaxParallel: 1, Panels: []document.Panel{{ID: "snap", Mode: "snapshot", Autostart: true, Exec: []document.Arg{{Literal: "snap"}}}}}
	e := New(p, f, time.Now)
	e.Start()
	c := receive(t, f)
	done := make(chan struct{})
	go func() {
		e.Pause()
		close(done)
	}()
	c.reply <- runner.Result{Err: context.Canceled}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pause blocked")
	}
	e.Resume()
	again := receive(t, f)
	if again.spec.Args[0] != "snap" {
		t.Fatal(again.spec.Args)
	}
}

func TestPauseCommitsCompletedEvent(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	c := receive(t, f)
	c.reply <- runner.Result{Stdout: []byte("done")}
	deadline := time.Now().Add(time.Second)
	for len(e.Events) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("event not delivered")
		}
		time.Sleep(time.Millisecond)
	}
	e.Pause()
	if e.Panels["detail"].Text != "done" {
		t.Fatal(e.Panels["detail"].Text)
	}
	if e.Panels["detail"].Status != "ready" {
		t.Fatal(e.Panels["detail"].Status)
	}
}

func TestPauseDoesNotRestartCancelledStream(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	finish(t, e, receive(t, f), "A")
	e.Panels["detail"].Definition.Refresh = 0
	e.Panels["log"].Definition.Autostart = true
	e.Run("log")
	stream := receive(t, f)
	e.Cancel("log")
	finish(t, e, stream, "")
	e.Pause()
	e.Resume()
	if len(f.calls) != 0 {
		t.Fatal("cancelled stream restarted")
	}
}

func TestPauseDoesNotRestartNaturallyExitedStream(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	finish(t, e, receive(t, f), "A")
	e.Panels["detail"].Definition.Refresh = 0
	e.Request("log")
	stream := receive(t, f)
	stream.reply <- runner.Result{Stdout: []byte("done")}
	deadline := time.Now().Add(time.Second)
	for len(e.Events) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("event not delivered")
		}
		time.Sleep(time.Millisecond)
	}
	e.Pause()
	e.Resume()
	if len(f.calls) != 0 {
		t.Fatal("naturally exited stream restarted")
	}
	if s := e.Panels["log"].Status; s == "running" || s == "queued" {
		t.Fatalf("status = %q", s)
	}
}

func TestPauseDoesNotRestartCancelBeforeDrain(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	finish(t, e, receive(t, f), "A")
	e.Panels["detail"].Definition.Refresh = 0
	e.Request("log")
	stream := receive(t, f)
	e.Cancel("log")
	done := make(chan struct{})
	go func() {
		e.Pause()
		close(done)
	}()
	stream.reply <- runner.Result{Err: context.Canceled}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pause blocked")
	}
	e.Resume()
	if len(f.calls) != 0 {
		t.Fatal("cancelled stream restarted")
	}
	if s := e.Panels["log"].Status; s == "running" || s == "queued" {
		t.Fatalf("status = %q", s)
	}
}

func TestPauseResumesPendingRegularAfterOneShot(t *testing.T) {
	f := &fake{make(chan call, 4)}
	page := &document.Page{MaxParallel: 1, Panels: []document.Panel{{ID: "detail", Mode: "snapshot", Exec: []document.Arg{{Literal: "show"}}}}}
	e := New(page, f, time.Now)
	e.Start()
	e.Run("detail")
	first := receive(t, f)
	if reason := e.Request("detail"); reason == "" {
		t.Fatal("request not coalesced")
	}
	p := e.Panels["detail"]
	if !p.pending || p.pendingRun {
		t.Fatalf("pending = %v pendingRun = %v", p.pending, p.pendingRun)
	}
	done := make(chan struct{})
	go func() {
		e.Pause()
		close(done)
	}()
	first.reply <- runner.Result{Err: context.Canceled}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pause blocked")
	}
	e.Resume()
	finish(t, e, receive(t, f), "ok")
	if p.Text != "ok" {
		t.Fatal(p.Text)
	}
}

func TestPauseResumesPendingRegularAfterOneShotSuccess(t *testing.T) {
	f := &fake{make(chan call, 4)}
	page := &document.Page{MaxParallel: 1, Panels: []document.Panel{{ID: "detail", Mode: "snapshot", Exec: []document.Arg{{Literal: "show"}}}}}
	e := New(page, f, time.Now)
	e.Start()
	e.Run("detail")
	first := receive(t, f)
	if reason := e.Request("detail"); reason == "" {
		t.Fatal("request not coalesced")
	}
	done := make(chan struct{})
	go func() {
		e.Pause()
		close(done)
	}()
	first.reply <- runner.Result{Stdout: []byte("first")}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("pause blocked")
	}
	e.Resume()
	finish(t, e, receive(t, f), "ok")
	if e.Panels["detail"].Text != "ok" {
		t.Fatal(e.Panels["detail"].Text)
	}
}

func TestPauseKeepsStaleGenerationWork(t *testing.T) {
	e, f, _ := loaded(t)
	e.Select("list", "A")
	old := receive(t, f)
	e.Select("list", "B")
	old.reply <- runner.Result{Stdout: []byte("old")}
	e.Pause()
	e.Resume()
	next := receive(t, f)
	if len(next.spec.Args) != 2 || next.spec.Args[0] != "show" || next.spec.Args[1] != "B" {
		t.Fatalf("args = %v", next.spec.Args)
	}
	finish(t, e, next, "new")
	if e.Panels["detail"].Text != "new" {
		t.Fatal(e.Panels["detail"].Text)
	}
}

func TestPauseDoesNotRestartUserCancelBeforeDrainSuccess(t *testing.T) {
	f := &fake{make(chan call, 4)}
	page := &document.Page{MaxParallel: 2, Panels: []document.Panel{
		{ID: "list", Mode: "snapshot", Format: "items-json", Autostart: true, Exec: []document.Arg{{Literal: "list"}}},
		{ID: "detail", Mode: "snapshot", Autostart: true, Exec: []document.Arg{{Literal: "show"}, {Ref: "list.selected.id"}}},
	}}
	e := New(page, f, time.Now)
	e.Start()
	finish(t, e, receive(t, f), `{"version":1,"items":[{"id":"A","label":"one"},{"id":"B","label":"two"}]}`)
	e.Select("list", "A")
	c := receive(t, f)
	e.Cancel("detail")
	c.reply <- runner.Result{Stdout: []byte("late")}
	e.Pause()
	e.Resume()
	select {
	case call := <-f.calls:
		t.Fatalf("unexpected restart: %v", call.spec.Args)
	case <-time.After(100 * time.Millisecond):
	}
	if s := e.Panels["detail"].Status; s == "running" || s == "queued" {
		t.Fatalf("status = %q", s)
	}
}

func TestPauseDoesNotRestartInvalidatedManualPanel(t *testing.T) {
	f := &fake{make(chan call, 4)}
	page := &document.Page{MaxParallel: 2, Panels: []document.Panel{
		{ID: "list", Mode: "snapshot", Format: "items-json", Autostart: true, Exec: []document.Arg{{Literal: "list"}}},
		{ID: "detail", Mode: "snapshot", Exec: []document.Arg{{Literal: "show"}, {Ref: "list.selected.id"}}},
	}}
	e := New(page, f, time.Now)
	e.Start()
	finish(t, e, receive(t, f), `{"version":1,"items":[{"id":"A","label":"one"},{"id":"B","label":"two"}]}`)
	e.Select("list", "A")
	e.Request("detail")
	c := receive(t, f)
	e.Select("list", "B")
	c.reply <- runner.Result{Stdout: []byte("late")}
	e.Pause()
	e.Resume()
	select {
	case call := <-f.calls:
		t.Fatalf("unexpected restart: %v", call.spec.Args)
	case <-time.After(100 * time.Millisecond):
	}
	if s := e.Panels["detail"].Status; s == "running" || s == "queued" {
		t.Fatalf("status = %q", s)
	}
}

func TestCursorReferenceAndInvalidation(t *testing.T) {
	f := &fake{make(chan call, 8)}
	p := &document.Page{MaxParallel: 2, Panels: []document.Panel{
		{ID: "list", Mode: "snapshot", Format: "items-json", Autostart: true, Exec: []document.Arg{{Literal: "list"}}},
		{ID: "preview", Mode: "snapshot", Autostart: true, Exec: []document.Arg{{Literal: "show"}, {Ref: "list.cursor.id"}}},
	}}
	e := New(p, f, time.Now)
	e.Start()
	finish(t, e, receive(t, f), `{"version":1,"items":[{"id":"A","label":"A"},{"id":"B","label":"B"}]}`)
	if e.Panels["preview"].Status != "blocked" {
		t.Fatal("preview ran before cursor was available")
	}
	e.Panels["list"].Cursor = 1
	e.Cursor("list")
	first := receive(t, f)
	if len(first.spec.Args) != 2 || first.spec.Args[1] != "B" {
		t.Fatal(first.spec.Args)
	}
	finish(t, e, first, "B")
	e.Select("list", "A")
	if e.Panels["preview"].Text != "B" {
		t.Fatal("selection invalidated cursor preview")
	}
	e.Panels["list"].Cursor = 0
	e.Cursor("list")
	second := receive(t, f)
	if len(second.spec.Args) != 2 || second.spec.Args[1] != "A" {
		t.Fatal(second.spec.Args)
	}
	finish(t, e, second, "A")
	if e.Panels["preview"].Text != "A" {
		t.Fatal(e.Panels["preview"].Text)
	}
}
