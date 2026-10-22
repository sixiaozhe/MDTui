package tui

import (
	"github.com/gdamore/tcell/v2"
	"mdtui/internal/content"
	"mdtui/internal/document"
	"mdtui/internal/runner"
	"mdtui/internal/scheduler"
	"strings"
	"testing"
	"time"
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
