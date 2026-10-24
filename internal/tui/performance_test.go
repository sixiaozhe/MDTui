package tui

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"mdtui/internal/content"
)

type recordingScreen struct {
	tcell.Screen
	writes [][2]int
}

func (s *recordingScreen) SetContent(x, y int, ch rune, combining []rune, style tcell.Style) {
	s.writes = append(s.writes, [2]int{x, y})
	s.Screen.SetContent(x, y, ch, combining, style)
}

func screenCells(u *UI) []string {
	w, h := u.Screen.Size()
	cells := make([]string, 0, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			ch, comb, style, width := u.Screen.GetContent(x, y)
			cells = append(cells, fmt.Sprintf("%c/%s/%v/%d", ch, string(comb), style, width))
		}
	}
	return cells
}

func mouseTargets(u *UI) []string {
	var targets []string
	for _, r := range u.mouseRegions {
		targets = append(targets, fmt.Sprintf("%s %d %d %d %d %d %d", r.kind, r.x, r.y, r.w, r.h, r.focus, r.index))
	}
	sort.Strings(targets)
	return targets
}

func TestPartialPaintMatchesFullPaintAndMouseTargets(t *testing.T) {
	for _, layout := range []string{"stack", "master-detail", "workspace"} {
		t.Run(layout, func(t *testing.T) {
			u := testUI(t)
			u.Engine.Page.Layout, u.Engine.Page.Sidebar = layout, "list"
			p := u.Engine.Panels["detail"]
			p.Definition.Mode, p.Definition.Format = "stream", "markdown"
			p.Status, p.Text = "running", "# 开始\n\n"+strings.Repeat("旧行\n", 40)
			u.workspaceDetail = "detail"
			u.Draw()
			for _, text := range []string{"# 新内容\n\n```text\n中英文\n", "短行", "| 名称 | 值 |\n|---|---|\n|中文|1|"} {
				p.Text = text
				u.drawUpdates()
				partial, targets := screenCells(u), mouseTargets(u)
				u.Draw()
				if !reflect.DeepEqual(partial, screenCells(u)) {
					t.Fatal("partial frame differs from full redraw")
				}
				if !reflect.DeepEqual(targets, mouseTargets(u)) {
					t.Fatal("mouse targets lost or duplicated")
				}
			}
		})
	}
}

func TestStreamPaintSkipsUnchangedCellsAndOpenModal(t *testing.T) {
	u := testUI(t)
	u.Draw()
	recorder := &recordingScreen{Screen: u.Screen}
	u.Screen = recorder
	p := u.Engine.Panels["detail"]
	box := u.painted["detail"]
	p.Text = "new output"
	u.drawUpdates()
	if len(recorder.writes) == 0 {
		t.Fatal("changed panel was not painted")
	}
	for _, at := range recorder.writes {
		if at[0] < box.x || at[0] >= box.x+box.w || at[1] < box.y || at[1] >= box.y+box.h {
			t.Fatal("unrelated region repainted")
		}
	}
	recorder.writes = nil
	u.drawUpdates()
	if len(recorder.writes) != 0 {
		t.Fatal("unchanged frame repainted")
	}
	u.modal = "help"
	u.Draw()
	before := screenCells(u)
	p.Text = "while modal open"
	u.drawUpdates()
	if !reflect.DeepEqual(before, screenCells(u)) {
		t.Fatal("stream painted over modal")
	}
	u.modal = ""
	u.Draw()
	if !strings.Contains(strings.Join(u.panelLines["detail"], "\n"), "while modal open") {
		t.Fatal("modal hid updates permanently")
	}
}

func TestItemCacheInvalidatesOnFilterRefreshAndSelection(t *testing.T) {
	u := testUI(t)
	p := u.Engine.Panels["list"]
	u.Draw()
	initial := u.visibleItems(p)
	if allocations := testing.AllocsPerRun(100, func() { u.visibleItems(p) }); allocations != 0 {
		t.Fatalf("cached filtering allocated %v", allocations)
	}
	u.filters = map[string]string{"list": "B"}
	u.Draw()
	if allocations := testing.AllocsPerRun(100, func() { u.visibleItems(p) }); allocations != 0 {
		t.Fatalf("cached case folding allocated %v", allocations)
	}
	if got := u.visibleItems(p); len(got) != 1 || got[0] != 1 || len(initial) != 2 {
		t.Fatal("filter cache stale")
	}
	p.Items = []content.Item{{ID: "B", Label: "replacement"}}
	p.Cursor, p.Selected = 0, "B"
	u.Draw()
	if text := strings.Join(u.panelLines["list"], "\n"); !strings.Contains(text, "replacement") || !strings.Contains(text, "›✓") {
		t.Fatalf("refresh/selection stale: %s", text)
	}
	p.Items = nil
	u.Draw()
	if len(u.panelLines["list"]) != 0 {
		t.Fatal("empty list retained rows")
	}
}

func TestChildInheritsResourceSettings(t *testing.T) {
	u := navUI(t, childPage)
	defer u.Close()
	u.Engine.StreamBytes, u.Engine.StreamLines = 256<<10, 2000
	u.Engine.StreamInterval, u.Engine.MarkdownInterval = 500*time.Millisecond, time.Second
	u.ParentSnapshotBytes = 1 << 20
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	if u.Engine.StreamBytes != 256<<10 || u.Engine.StreamLines != 2000 {
		t.Fatal("child lost stream limits")
	}
	if u.Engine.StreamInterval != 500*time.Millisecond || u.Engine.MarkdownInterval != time.Second {
		t.Fatal("child lost stream intervals")
	}
}
