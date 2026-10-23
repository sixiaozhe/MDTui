// Package scheduler owns all mutable application state on one event loop.
package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mdtui/internal/content"
	"mdtui/internal/document"
	"mdtui/internal/runner"
)

type Panel struct {
	Definition                                      document.Panel
	Status, Reason, Text, Selected, Stderr          string
	Items                                           []content.Item
	Generation, Request                             uint64
	Stale                                           bool
	Failures, Cursor, Scroll, X, Dropped, Truncated int
	Follow                                          bool
	Due, Updated                                    time.Time
	ExitCode                                        int
	LastProgram                                     string
	Duration                                        time.Duration
	pending, queued                                 bool
	queuedRun, pendingRun                           bool
	active                                          *job
}
type job struct {
	request, generation uint64
	cancel              context.CancelFunc
	buffer              *content.Buffer
	revision            uint64
	once                bool
}
type Event struct {
	ID                  string
	Request, Generation uint64
	Result              runner.Result
	Items               []content.Item
	Text                string
}
type Engine struct {
	Page           *document.Page
	Panels         map[string]*Panel
	Inputs         map[string]string
	ParentSelected string
	Events         chan Event
	Now            func() time.Time
	Runner         runner.Executor
	Log            func(string, ...any)
	queue          []string
	sequence       uint64
	running        int
	closing        bool
}

func New(p *document.Page, r runner.Executor, now func() time.Time) *Engine {
	e := &Engine{Page: p, Panels: map[string]*Panel{}, Inputs: map[string]string{}, Events: make(chan Event, 16), Now: now, Runner: r, Log: func(string, ...any) {}}
	for _, v := range p.Inputs {
		e.Inputs[v.ID] = v.Default
	}
	for _, v := range p.Panels {
		e.Panels[v.ID] = &Panel{Definition: v, Status: "idle", Follow: true}
	}
	return e
}
func (e *Engine) Start() {
	for _, d := range e.Page.Panels {
		p := e.Panels[d.ID]
		if _, reason := e.Args(d.ID); reason != "" {
			p.Status = "blocked"
			p.Reason = reason
		} else if d.Autostart {
			e.enqueue(p)
		}
	}
	e.dispatch()
}
func (e *Engine) Args(id string) ([]string, string) {
	p := e.Panels[id]
	var result []string
	for _, a := range p.Definition.Exec {
		if a.Ref == "" {
			result = append(result, a.Literal)
			continue
		}
		parts := strings.Split(a.Ref, ".")
		if len(parts) == 3 && parts[0] == "parent" && parts[1] == "selected" && parts[2] == "id" {
			if e.ParentSelected == "" {
				return nil, "请选择父页对象"
			}
			result = append(result, e.ParentSelected)
			continue
		}
		if len(parts) == 2 {
			value := e.Inputs[parts[0]]
			for _, input := range e.Page.Inputs {
				if input.ID == parts[0] && input.Required && value == "" {
					return nil, "请填写参数: " + input.Label
				}
			}
			result = append(result, value)
		} else {
			value := e.Panels[parts[0]].Selected
			if value == "" {
				return nil, "请选择对象: " + parts[0]
			}
			result = append(result, value)
		}
	}
	return result, ""
}
func (e *Engine) Request(id string) string {
	return e.request(id, false)
}

// Run is an explicit one-shot action, never a source of automatic retries.
func (e *Engine) Run(id string) string {
	return e.request(id, true)
}
func (e *Engine) request(id string, once bool) string {
	p := e.Panels[id]
	if e.closing {
		return "正在退出"
	}
	if _, reason := e.Args(id); reason != "" {
		p.Status = "blocked"
		p.Reason = reason
		return reason
	}
	if p.Definition.Mode == "stream" {
		if e.Page.MaxParallel < 2 {
			return "stream 至少需要 max_parallel: 2，以保留查询槽位"
		}
		for key, other := range e.Panels {
			if key != id && other.Definition.Mode == "stream" && (other.active != nil || other.queued) {
				return "同时只允许一个 stream"
			}
		}
	}
	if p.active != nil {
		p.pending = true
		p.pendingRun = p.pendingRun || once
		return "已合并为一次后续执行"
	}
	p.queuedRun = p.queuedRun || once
	e.enqueue(p)
	e.dispatch()
	return ""
}
func (e *Engine) enqueue(p *Panel) {
	if e.closing || p.queued {
		return
	}
	p.queued = true
	p.Status = "queued"
	p.Reason = ""
	p.Due = time.Time{}
	e.queue = append(e.queue, p.Definition.ID)
}
func (e *Engine) dispatch() {
	if e.closing {
		return
	}
	for index := 0; index < len(e.queue) && e.running < e.Page.MaxParallel; {
		id := e.queue[index]
		p := e.Panels[id]
		if p.active != nil {
			index++
			continue
		}
		if p.Definition.Mode == "stream" {
			reason := ""
			if e.Page.MaxParallel < 2 {
				reason = "stream 需要至少两个并发槽位"
			}
			for _, other := range e.Panels {
				if other != p && other.Definition.Mode == "stream" && other.active != nil {
					reason = "已有 stream 正在运行"
				}
			}
			if reason != "" {
				p.Status = "error"
				p.Reason = reason
				p.queued = false
				e.queue = append(e.queue[:index], e.queue[index+1:]...)
				continue
			}
		}
		e.queue = append(e.queue[:index], e.queue[index+1:]...)
		p.queued = false
		args, reason := e.Args(id)
		if reason != "" {
			p.Status = "blocked"
			p.Reason = reason
			continue
		}
		e.sequence++
		ctx, cancel := context.WithCancel(context.Background())
		j := &job{request: e.sequence, generation: p.Generation, cancel: cancel, buffer: &content.Buffer{}, once: p.queuedRun}
		p.queuedRun = false
		p.active = j
		p.Request = j.request
		p.Status = "running"
		p.Reason = ""
		p.LastProgram = args[0]
		if p.Definition.Mode == "stream" {
			p.Text = ""
			p.Scroll = 0
			p.Follow = true
			p.Dropped = 0
			p.Truncated = 0
		}
		e.running++
		spec := runner.Spec{Args: args, Dir: e.Page.Dir, Timeout: p.Definition.Timeout, Stream: p.Definition.Mode == "stream"}
		e.Log("start panel=%s request=%d generation=%d", id, j.request, j.generation)
		go func(id string, j *job, s runner.Spec, format string) {
			r := e.Runner.Run(ctx, s, j.buffer)
			ev := Event{ID: id, Request: j.request, Generation: j.generation, Result: r}
			if r.Err == nil && !s.Stream {
				if format == "items-json" {
					ev.Items, ev.Result.Err = content.Items(r.Stdout)
				} else if format == "markdown" {
					ev.Text = content.Markdown(string(r.Stdout))
				} else {
					ev.Text = content.Safe(string(r.Stdout))
				}
			}
			ev.Result.Stdout = nil
			e.Events <- ev
		}(id, j, spec, p.Definition.Format)
	}
}
func (e *Engine) Consume(ev Event) {
	p := e.Panels[ev.ID]
	if p == nil || p.active == nil || p.active.request != ev.Request {
		return
	}
	j := p.active
	if ev.Generation == p.Generation {
		e.readStream(p)
		p.ExitCode = ev.Result.ExitCode
		p.Stderr = ev.Result.Stderr
		p.Duration = ev.Result.Finished.Sub(ev.Result.Started)
		p.Updated = e.Now()
		err := ev.Result.Err
		if err == nil && p.Definition.Format == "items-json" {
			items := ev.Items
			{
				p.Items = items
				if p.Cursor >= len(items) {
					p.Cursor = max(0, len(items)-1)
				}
				if p.Selected != "" {
					found := false
					for _, it := range items {
						if it.ID == p.Selected {
							found = true
							break
						}
					}
					if !found {
						p.Selected = ""
						e.invalidate(p.Definition.ID)
					}
				}
			}
		}
		if err != nil {
			p.Status = "error"
			p.Reason = err.Error()
			p.Stale = true
			p.Failures++
		} else {
			p.Status = "ready"
			p.Reason = ""
			p.Stale = false
			p.Failures = 0
			if p.Definition.Mode != "stream" && p.Definition.Format != "items-json" {
				p.Text = ev.Text
			}
		}
		if p.Definition.Refresh > 0 && !j.once {
			p.Due = e.Now().Add(p.Definition.Refresh)
		}
	}
	j.cancel()
	p.active = nil
	e.running--
	e.Log("finish panel=%s request=%d status=%s elapsed_ms=%d", ev.ID, ev.Request, p.Status, ev.Result.Finished.Sub(ev.Result.Started).Milliseconds())
	if p.pending && !e.closing {
		p.pending = false
		p.queuedRun = p.pendingRun
		p.pendingRun = false
		if _, reason := e.Args(ev.ID); reason == "" {
			e.enqueue(p)
		}
	}
	e.dispatch()
}
func (e *Engine) readStream(p *Panel) bool {
	j := p.active
	if j == nil || p.Definition.Mode != "stream" || j.generation != p.Generation {
		return false
	}
	w, changed := j.buffer.Read(j.revision)
	if changed {
		if !p.Follow {
			p.Scroll = max(0, p.Scroll-(w.Dropped-p.Dropped))
		}
		p.Text = w.Text
		p.Dropped = w.Dropped
		p.Truncated = w.Truncated
		j.revision = w.Revision
	}
	return changed
}
func (e *Engine) Tick() bool {
	changed := false
	for _, d := range e.Page.Panels {
		p := e.Panels[d.ID]
		if e.readStream(p) {
			changed = true
		}
		if !p.Due.IsZero() && !e.Now().Before(p.Due) {
			p.Due = time.Time{}
			e.Request(d.ID)
			changed = true
		}
	}
	return changed
}
func (e *Engine) Select(id, value string) {
	p := e.Panels[id]
	if p.Selected == value {
		return
	}
	found := false
	for _, v := range p.Items {
		if v.ID == value {
			found = true
		}
	}
	if !found {
		return
	}
	p.Selected = value
	e.invalidate(id)
	e.dispatch()
}
func (e *Engine) Submit(id, value string) {
	if e.Inputs[id] == value {
		return
	}
	e.Inputs[id] = value
	e.invalidate(id)
	e.dispatch()
}
func (e *Engine) invalidate(source string) {
	// Compute transitive closure first, then clear every selection before resolving.
	affected := map[string]bool{}
	var visit func(string)
	visit = func(src string) {
		for _, d := range e.Page.Panels {
			if affected[d.ID] {
				continue
			}
			for _, a := range d.Exec {
				if strings.HasPrefix(a.Ref, src+".") {
					affected[d.ID] = true
					visit(d.ID)
					break
				}
			}
		}
	}
	visit(source)
	for id := range affected {
		p := e.Panels[id]
		p.Generation++
		p.Selected = ""
		p.Items = nil
		p.Text = ""
		p.Stderr = ""
		p.Stale = false
		p.Failures = 0
		p.Scroll = 0
		p.X = 0
		p.Cursor = 0
		p.Dropped, p.Truncated = 0, 0
		p.Updated = time.Time{}
		p.pending = false
		p.queuedRun, p.pendingRun = false, false
		p.queued = false
		p.Due = time.Time{}
		p.Status = "idle"
		p.Reason = ""
		if p.active != nil {
			p.active.cancel()
		}
	}
	filtered := e.queue[:0]
	for _, id := range e.queue {
		if !affected[id] {
			filtered = append(filtered, id)
		}
	}
	e.queue = filtered
	for _, d := range e.Page.Panels {
		if !affected[d.ID] {
			continue
		}
		p := e.Panels[d.ID]
		if _, reason := e.Args(d.ID); reason != "" {
			p.Status = "blocked"
			p.Reason = reason
		} else if d.Autostart {
			e.enqueue(p)
		}
	}
}
func (e *Engine) Cancel(id string) {
	p := e.Panels[id]
	p.Generation++
	if p.active != nil {
		p.active.cancel()
	}
	p.pending = false
	p.queuedRun, p.pendingRun = false, false
	p.queued = false
	p.Due = time.Time{}
	p.Status = "cancelled"
	p.Reason = "用户已取消"
	q := e.queue[:0]
	for _, v := range e.queue {
		if v != id {
			q = append(q, v)
		}
	}
	e.queue = q
}
func (e *Engine) ActionReason(a document.Action) string {
	for _, id := range Targets(a) {
		if _, r := e.Args(id); r != "" {
			return r
		}
		if e.Panels[id].Definition.Mode == "stream" {
			if e.Page.MaxParallel < 2 {
				return "stream 需要 max_parallel ≥ 2"
			}
			for other, p := range e.Panels {
				if other != id && p.Definition.Mode == "stream" && (p.active != nil || p.queued) {
					return "已有 stream"
				}
			}
		}
	}
	return ""
}
func Targets(a document.Action) []string {
	if a.Run != "" {
		return []string{a.Run}
	}
	return a.Refresh
}
func (e *Engine) Preview(a document.Action) string {
	var lines []string
	for _, id := range Targets(a) {
		args, reason := e.Args(id)
		if reason != "" {
			lines = append(lines, reason)
			continue
		}
		lines = append(lines, fmt.Sprintf("%s: %q", id, args))
	}
	return strings.Join(lines, "\n")
}
func (e *Engine) Close() {
	e.closing = true
	e.queue = nil
	for _, p := range e.Panels {
		p.Due = time.Time{}
		p.pending = false
		if p.active != nil {
			p.active.cancel()
		}
	}
	for e.running > 0 {
		e.Consume(<-e.Events)
	}
}
