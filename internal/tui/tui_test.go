package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"mdtui/internal/content"
	"mdtui/internal/document"
	"mdtui/internal/runner"
	"mdtui/internal/scheduler"
)

func testUI(t *testing.T) *UI {
	t.Helper()
	p := &document.Page{Title: "中文工作台", MaxParallel: 2, Panels: []document.Panel{{ID: "list", Title: "列表", Format: "items-json", Exec: []document.Arg{{Literal: "false"}}}, {ID: "detail", Title: "详情", Mode: "snapshot", Exec: []document.Arg{{Literal: "false"}, {Ref: "list.selected.id"}}}}, Inputs: []document.Input{{ID: "filter", Label: "过滤", MaxLength: 128}}, Actions: []document.Action{{ID: "run", Label: "运行", Run: "detail", Key: "d", Confirm: "确认？"}}}
	s := tcell.NewSimulationScreen("UTF-8")
	s.Init()
	s.SetSize(120, 40)
	t.Cleanup(s.Fini)
	e := scheduler.New(p, runner.OS{}, time.Now)
	e.Panels["list"].Items = []content.Item{{ID: "A", Label: "中文"}, {ID: "B", Label: "第二项"}}
	return New(s, e, false)
}
func TestCursorRequiresEnterAndResize(t *testing.T) {
	u := testUI(t)
	u.Key(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	if u.Engine.Panels["list"].Selected != "" {
		t.Fatal("moving selected")
	}
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	if u.Engine.Panels["list"].Selected != "B" {
		t.Fatal("enter not selected")
	}
	s := u.Screen.(tcell.SimulationScreen)
	for _, size := range [][2]int{{59, 15}, {80, 24}, {120, 40}} {
		s.SetSize(size[0], size[1])
		u.Draw()
		if u.Engine.Panels["list"].Selected != "B" {
			t.Fatal("resize lost selection")
		}
	}
}
func TestConfirmationDefaultsCancel(t *testing.T) {
	u := testUI(t)
	u.Engine.Panels["list"].Selected = "A"
	u.Key(tcell.NewEventKey(tcell.KeyRune, 'd', 0))
	if u.modal != "confirm" || u.confirm {
		t.Fatal("not default cancel")
	}
	u.Draw()
	s := u.Screen.(tcell.SimulationScreen)
	var modalText strings.Builder
	for y := 0; y < 40; y++ {
		for x := 0; x < 120; x++ {
			main, _, _, _ := s.GetContent(x, y)
			if main != ' ' && main != 0 {
				modalText.WriteRune(main)
			}
		}
	}
	if !strings.Contains(modalText.String(), "确认操作") || !strings.Contains(modalText.String(), "●取消○执行") {
		t.Fatalf("confirmation dialog was not drawn with cancel selected: %s", modalText.String())
	}
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	if u.Engine.Panels["detail"].Request != 0 {
		t.Fatal("cancel ran process")
	}
}
func TestInputDraftAndActionSuppression(t *testing.T) {
	u := testUI(t)
	u.focus = 1
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	u.Key(tcell.NewEventKey(tcell.KeyRune, 'q', 0))
	if u.Engine.Inputs["filter"] != "" {
		t.Fatal("draft committed early")
	}
	u.Key(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	for _, r := range "空格 ;$(x)" {
		u.Key(tcell.NewEventKey(tcell.KeyRune, r, 0))
	}
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	if u.Engine.Inputs["filter"] != "空格 ;$(x)" {
		t.Fatal("input changed")
	}
}
func TestStreamPause(t *testing.T) {
	u := testUI(t)
	p := u.Engine.Panels["list"]
	p.Definition.Mode = "stream"
	u.Key(tcell.NewEventKey(tcell.KeyUp, 0, 0))
	if p.Follow {
		t.Fatal("scroll did not pause")
	}
	u.Key(tcell.NewEventKey(tcell.KeyEnd, 0, 0))
	if !p.Follow {
		t.Fatal("end did not follow")
	}
}

func screenText(s tcell.SimulationScreen, width, height int) string {
	var out strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			ch, _, _, _ := s.GetContent(x, y)
			out.WriteRune(ch)
		}
		out.WriteByte('\n')
	}
	return out.String()
}

func TestWorkspaceAndZoomNavigation(t *testing.T) {
	u := testUI(t)
	u.Engine.Page.Layout = "workspace"
	u.Engine.Page.Sidebar = "list"
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	if u.workspaceDetail != "detail" {
		t.Fatal("sidebar selection did not open dependent detail")
	}
	u.Engine.Panels["detail"].Text = content.Markdown("# 详情标记")
	u.Draw()
	s := u.Screen.(tcell.SimulationScreen)
	if text := strings.ReplaceAll(screenText(s, 120, 40), " ", ""); !strings.Contains(text, "详情标记") || !strings.Contains(text, "中文") {
		t.Fatal("workspace did not show sidebar and detail")
	}
	u.Key(tcell.NewEventKey(tcell.KeyTab, 0, 0))
	if u.focus != 3 {
		t.Fatalf("workspace detail focus = %d", u.focus)
	}
	u.Key(tcell.NewEventKey(tcell.KeyRune, 'Z', 0))
	if !u.zoom {
		t.Fatal("Z did not zoom panel")
	}
	u.Draw()
	if text := strings.ReplaceAll(screenText(s, 120, 40), " ", ""); !strings.Contains(text, "详情标记") || strings.Contains(text, "第二项") {
		t.Fatal("zoom showed another panel")
	}
	u.Key(tcell.NewEventKey(tcell.KeyTab, 0, 0))
	if u.focus != 2 {
		t.Fatal("Tab did not cycle panels while zoomed")
	}
	u.Key(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	if u.zoom {
		t.Fatal("Esc did not exit zoom")
	}
	s.SetSize(80, 24)
	u.Draw()
	if text := strings.ReplaceAll(screenText(s, 80, 24), " ", ""); !strings.Contains(text, "中文") {
		t.Fatal("narrow workspace hid focused sidebar")
	}
}

func TestWorkspaceViewShortcuts(t *testing.T) {
	u := testUI(t)
	u.Engine.Page.Layout = "workspace"
	u.Engine.Page.Sidebar = "list"
	u.Engine.Page.Panels = append(u.Engine.Page.Panels, document.Panel{ID: "extra", Title: "附加视图", Format: "markdown"})
	u.Engine.Panels["extra"] = &scheduler.Panel{Definition: u.Engine.Page.Panels[2], Status: "idle"}
	u.Draw()
	u.Key(tcell.NewEventKey(tcell.KeyRune, ']', 0))
	if u.workspaceDetail != "extra" || u.focus != 4 {
		t.Fatalf("next view: %s, focus %d", u.workspaceDetail, u.focus)
	}
	u.Key(tcell.NewEventKey(tcell.KeyRune, '[', 0))
	if u.workspaceDetail != "detail" || u.focus != 3 {
		t.Fatalf("previous view: %s, focus %d", u.workspaceDetail, u.focus)
	}
}

func TestInputCursorEditsGraphemeClusters(t *testing.T) {
	u := testUI(t)
	u.Engine.Inputs["filter"] = "A中e\u0301B"
	u.focus = 1
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	u.Key(tcell.NewEventKey(tcell.KeyLeft, 0, 0))
	u.Key(tcell.NewEventKey(tcell.KeyLeft, 0, 0))
	if u.draftCursor != 2 {
		t.Fatalf("cursor split a grapheme: %d", u.draftCursor)
	}
	u.Key(tcell.NewEventKey(tcell.KeyDelete, 0, 0))
	u.Key(tcell.NewEventKey(tcell.KeyRune, '好', 0))
	u.Key(tcell.NewEventKey(tcell.KeyHome, 0, 0))
	u.Key(tcell.NewEventKey(tcell.KeyDelete, 0, 0))
	u.Key(tcell.NewEventKey(tcell.KeyEnd, 0, 0))
	u.Key(tcell.NewEventKey(tcell.KeyBackspace, 0, 0))
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	if got := u.Engine.Inputs["filter"]; got != "中好" {
		t.Fatalf("unexpected edited input: %q", got)
	}
}

func TestLongInputKeepsCursorVisible(t *testing.T) {
	u := testUI(t)
	u.Engine.Inputs["filter"] = strings.Repeat("中文", 25)
	u.focus = 1
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	s := u.Screen.(tcell.SimulationScreen)
	s.SetSize(60, 20)
	for _, key := range []tcell.Key{tcell.KeyEnd, tcell.KeyHome} {
		u.Key(tcell.NewEventKey(key, 0, 0))
		u.Draw()
		visible := false
		for x := 0; x < 60; x++ {
			main, _, _, _ := s.GetContent(x, 2)
			visible = visible || main == '▏'
		}
		if !visible {
			t.Fatalf("input cursor not visible after %v", key)
		}
	}
}

func TestSearchNavigatesWrappedMarkdownAndResize(t *testing.T) {
	u := testUI(t)
	u.focus = 3
	u.Engine.Panels["detail"].Definition.Format = "markdown"
	u.Engine.Panels["detail"].Text = content.Markdown("起点 " + strings.Repeat("中文", 50) + " 目标\n\n第二处目标")
	s := u.Screen.(tcell.SimulationScreen)
	s.SetSize(80, 24)
	u.Draw()
	narrow := len(u.panelLines["detail"])
	s.SetSize(120, 40)
	u.Draw()
	if len(u.panelLines["detail"]) >= narrow {
		t.Fatal("markdown did not reflow after widening terminal")
	}
	u.Key(tcell.NewEventKey(tcell.KeyRune, '/', 0))
	for _, r := range "目标" {
		u.Key(tcell.NewEventKey(tcell.KeyRune, r, 0))
	}
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	first := u.Engine.Panels["detail"].Scroll
	if first == 0 || !hasMatch(u.panelLines["detail"][first], "目标") {
		t.Fatalf("search did not jump to first match: %d", first)
	}
	u.Draw()
	highlighted := false
	for y := 0; y < 40; y++ {
		for x := 0; x < 120; x++ {
			main, _, style, _ := s.GetContent(x, y)
			_, _, attributes := style.Decompose()
			if main == '目' && attributes&tcell.AttrReverse != 0 {
				highlighted = true
			}
		}
	}
	if !highlighted {
		t.Fatal("search match was not highlighted")
	}
	u.Key(tcell.NewEventKey(tcell.KeyRune, 'n', 0))
	second := u.searchIndex
	if second <= first || !hasMatch(u.panelLines["detail"][second], "目标") {
		t.Fatalf("next match failed: %d -> %d", first, second)
	}
	u.Draw()
	u.Key(tcell.NewEventKey(tcell.KeyRune, 'N', 0))
	if u.searchIndex != first {
		t.Fatal("previous match failed")
	}
}

func TestDrawsWorkbenchChromeAndRichMarkdown(t *testing.T) {
	u := testUI(t)
	u.Engine.Page.ID = "network"
	u.Engine.Page.Markdown = "# 网络概览\n\n**设备在线**"
	u.Draw()
	s := u.Screen.(tcell.SimulationScreen)
	var rendered strings.Builder
	for y := 0; y < 40; y++ {
		for x := 0; x < 120; x++ {
			main, _, _, _ := s.GetContent(x, y)
			if main != ' ' && main != 0 {
				rendered.WriteRune(main)
			}
		}
		rendered.WriteByte('\n')
	}
	text := rendered.String()
	for _, want := range []string{"◆MDTUI", "PAGE·network", "▍网络概览", "设备在线", "╭", "╯", "Tab焦点"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing visual element %q\n%s", want, text)
		}
	}
	if strings.Contains(text, "**设备在线**") {
		t.Fatal("markdown markers were rendered literally")
	}
}

func navUI(t *testing.T, child string) *UI {
	t.Helper()
	dir := t.TempDir()
	rootPath := filepath.Join(dir, "root.md")
	root := ":::mdtui page\nversion: 1\nid: root\n:::\n:::mdtui panel\nid: list\nformat: items-json\nautostart: false\nexec: [echo]\nopen: ./child.md\n:::\n"
	if err := os.WriteFile(rootPath, []byte(root), 0600); err != nil {
		t.Fatal(err)
	}
	if child != "" {
		if err := os.WriteFile(filepath.Join(dir, "child.md"), []byte(child), 0600); err != nil {
			t.Fatal(err)
		}
	}
	page, err := document.Load(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	s := tcell.NewSimulationScreen("UTF-8")
	s.Init()
	s.SetSize(120, 40)
	t.Cleanup(s.Fini)
	e := scheduler.New(page, runner.OS{}, time.Now)
	e.Panels["list"].Items = []content.Item{{ID: "A", Label: "one"}, {ID: "B", Label: "two"}}
	return New(s, e, false)
}

const childPage = ":::mdtui page\nversion: 1\nid: child\nparent: list\n:::\n:::mdtui panel\nid: detail\nautostart: false\nexec: [echo]\n:::\n"

func TestPageNavigationPushPop(t *testing.T) {
	u := navUI(t, childPage)
	u.Engine.Select("list", "A")
	u.pushChild("list")
	if len(u.sessions) != 1 || u.Engine.Page.ID != "child" {
		t.Fatal("child not pushed")
	}
	if u.Engine.ParentSelected != "A" {
		t.Fatalf("parent selected = %q", u.Engine.ParentSelected)
	}
	if !u.sessions[0].Engine.Paused() {
		t.Fatal("parent not paused")
	}
	if u.sessions[0].Engine.Panels["list"].Selected != "A" {
		t.Fatal("parent selection lost")
	}
	u.Key(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	if len(u.sessions) != 0 || u.Engine.Page.ID != "root" {
		t.Fatal("child not popped")
	}
	if u.Engine.Paused() {
		t.Fatal("parent not resumed")
	}
	if u.Engine.Panels["list"].Selected != "A" {
		t.Fatal("selection lost after return")
	}
	u.pushChild("list")
	if len(u.sessions) != 0 {
		t.Fatal("same selection reopened child")
	}
	u.Engine.Select("list", "B")
	u.pushChild("list")
	if len(u.sessions) != 1 || u.Engine.ParentSelected != "B" {
		t.Fatal("new selection did not open child")
	}
}

func TestPageNavigationQuitKeys(t *testing.T) {
	u := navUI(t, childPage)
	u.Engine.Select("list", "A")
	u.pushChild("list")
	if u.Key(tcell.NewEventKey(tcell.KeyRune, 'q', 0)) {
		t.Fatal("q exited from child")
	}
	if len(u.sessions) != 0 {
		t.Fatal("q did not return")
	}
	if !u.Key(tcell.NewEventKey(tcell.KeyRune, 'q', 0)) {
		t.Fatal("q did not exit root")
	}
}

func TestPageNavigationDepthLimit(t *testing.T) {
	u := navUI(t, childPage)
	u.sessions = make([]session, document.MaxDepth-1)
	u.Engine.Select("list", "A")
	u.pushChild("list")
	if len(u.sessions) != document.MaxDepth-1 {
		t.Fatal("depth limit not enforced")
	}
	if !strings.Contains(u.notice, "嵌套") {
		t.Fatalf("notice = %q", u.notice)
	}
}

func writePage(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPageNavigationNestedPushPop(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "root.md", ":::mdtui page\nversion: 1\nid: root\n:::\n:::mdtui panel\nid: list\nformat: items-json\nautostart: false\nexec: [echo]\nopen: ./child.md\n:::\n")
	writePage(t, dir, "child.md", ":::mdtui page\nversion: 1\nid: child\nparent: list\n:::\n:::mdtui panel\nid: list\nformat: items-json\nautostart: false\nexec: [echo]\nopen: ./grand.md\n:::\n")
	writePage(t, dir, "grand.md", ":::mdtui page\nversion: 1\nid: grand\nparent: list\n:::\n:::mdtui panel\nid: detail\nautostart: false\nexec: [echo]\n:::\n")
	page, err := document.Load(filepath.Join(dir, "root.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := tcell.NewSimulationScreen("UTF-8")
	s.Init()
	s.SetSize(120, 40)
	t.Cleanup(s.Fini)
	e := scheduler.New(page, runner.OS{}, time.Now)
	e.Panels["list"].Items = []content.Item{{ID: "A", Label: "one"}, {ID: "B", Label: "two"}}
	u := New(s, e, false)
	u.Engine.Select("list", "A")
	u.pushChild("list")
	u.Engine.Panels["list"].Items = []content.Item{{ID: "A", Label: "one"}, {ID: "B", Label: "two"}}
	u.Engine.Select("list", "A")
	u.pushChild("list")
	if len(u.sessions) != 2 || u.Engine.Page.ID != "grand" {
		t.Fatal("grandchild not pushed")
	}
	if u.sessions[0].Engine.Page.ID != "root" || u.sessions[1].Engine.Page.ID != "child" {
		t.Fatal("session stack order wrong")
	}
	if !u.sessions[0].Engine.Paused() || !u.sessions[1].Engine.Paused() {
		t.Fatal("parents not paused")
	}
	u.Key(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	u.Key(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	if u.Engine.Page.ID != "root" || len(u.sessions) != 0 {
		t.Fatal("not back at root")
	}
	if u.Engine.Paused() {
		t.Fatal("root still paused")
	}
	if u.Engine.Panels["list"].Selected != "A" {
		t.Fatal("root selection lost")
	}
}

func TestPageNavigationFailedLoadLeavesParentRunning(t *testing.T) {
	u := navUI(t, "")
	u.Engine.Select("list", "A")
	u.pushChild("list")
	if len(u.sessions) != 0 {
		t.Fatal("session pushed for failed load")
	}
	if u.Engine.Paused() {
		t.Fatal("parent paused after failed load")
	}
	if u.notice == "" || !strings.Contains(u.notice, "child.md") {
		t.Fatalf("notice = %q", u.notice)
	}
}

type holdRunner struct {
	started   chan struct{}
	stopped   chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
}

func (r *holdRunner) Run(ctx context.Context, _ runner.Spec, _ *content.Buffer) runner.Result {
	r.startOnce.Do(func() { close(r.started) })
	<-ctx.Done()
	r.stopOnce.Do(func() { close(r.stopped) })
	return runner.Result{Err: ctx.Err()}
}

func TestCloseStopsChildEngine(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "root.md", ":::mdtui page\nversion: 1\nid: root\n:::\n:::mdtui panel\nid: list\nformat: items-json\nautostart: false\nexec: [echo]\nopen: ./child.md\n:::\n")
	writePage(t, dir, "child.md", ":::mdtui page\nversion: 1\nid: child\nparent: list\n:::\n:::mdtui panel\nid: job\nmode: stream\nautostart: true\nexec: [job]\n:::\n")
	page, err := document.Load(filepath.Join(dir, "root.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := tcell.NewSimulationScreen("UTF-8")
	s.Init()
	s.SetSize(120, 40)
	t.Cleanup(s.Fini)
	r := &holdRunner{started: make(chan struct{}), stopped: make(chan struct{})}
	e := scheduler.New(page, r, time.Now)
	e.Panels["list"].Items = []content.Item{{ID: "A", Label: "one"}}
	u := New(s, e, false)
	u.Engine.Select("list", "A")
	u.pushChild("list")
	select {
	case <-r.started:
	case <-time.After(time.Second):
		t.Fatal("child job did not start")
	}
	u.Close()
	select {
	case <-r.stopped:
	case <-time.After(time.Second):
		t.Fatal("child job did not stop")
	}
}
