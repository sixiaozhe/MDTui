package tui

import (
	"reflect"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"mdtui/internal/document"
	"mdtui/internal/scheduler"
)

func TestPanelHeights(t *testing.T) {
	for _, tc := range []struct {
		name       string
		caps, want []int
		available  int
	}{
		{"legacy", []int{0, 0, 0}, []int{9, 9, 10}, 30},
		{"redistribute", []int{5, 0, 0}, []int{5, 11, 12}, 30},
		{"last-capped", []int{0, 0, 3}, []int{12, 13, 3}, 30},
		{"all-capped", []int{5, 5, 5}, []int{5, 5, 5}, 30},
		{"large-cap", []int{1000, 0, 0}, []int{9, 9, 10}, 30},
		{"single", []int{5}, []int{5}, 30},
		{"empty", nil, []int{}, 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			panels := make([]document.Panel, len(tc.caps))
			for i, cap := range tc.caps {
				panels[i].MaxHeight = cap
			}
			if got := panelHeights(panels, tc.available); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("heights=%v want=%v", got, tc.want)
			}
		})
	}
}

func heightUI(t *testing.T, layout string, width, height int) *UI {
	t.Helper()
	u := testUI(t)
	t.Cleanup(u.Close)
	u.Engine.Page.Inputs = nil
	u.Engine.Page.Layout = layout
	u.Engine.Page.Sidebar = "list"
	u.Engine.Page.Panels[0].MaxHeight = 5
	u.Engine.Panels["list"].Definition.MaxHeight = 5
	u.Engine.Page.Panels[1].MaxHeight = 6
	u.Engine.Panels["detail"].Definition.MaxHeight = 6
	if layout != "workspace" {
		def := document.Panel{ID: "log", Title: "log", Format: "text"}
		u.Engine.Page.Panels = append(u.Engine.Page.Panels, def)
		u.Engine.Panels["log"] = &scheduler.Panel{Definition: def, Text: strings.Repeat("line\n", 100), Status: "ready"}
	}
	u.focus = u.panelFocus("list")
	u.Screen.(tcell.SimulationScreen).SetSize(width, height)
	u.Draw()
	return u
}

func panelRegion(t *testing.T, u *UI, id string) mouseRegion {
	t.Helper()
	for _, region := range u.mouseRegions {
		if region.kind == "panel" && region.focus == u.panelFocus(id) {
			return region
		}
	}
	t.Fatalf("missing panel %s", id)
	return mouseRegion{}
}

func TestMaxHeightLayoutsAndMouseBounds(t *testing.T) {
	for _, layout := range []string{"stack", "master-detail", "workspace"} {
		t.Run(layout, func(t *testing.T) {
			u := heightUI(t, layout, 120, 40)
			list, detail := panelRegion(t, u, "list"), panelRegion(t, u, "detail")
			if list.h != 5 || detail.h != 6 {
				t.Fatalf("capped regions=%v %v", list, detail)
			}
			if layout == "stack" {
				log := panelRegion(t, u, "log")
				if detail.y != list.y+list.h+1 || log.y != detail.y+detail.h+1 || log.y+log.h != 38 {
					t.Fatalf("space not redistributed: %v %v %v", list, detail, log)
				}
			} else if layout == "master-detail" {
				log := panelRegion(t, u, "log")
				if log.y != detail.y+detail.h+1 || log.y+log.h != 38 {
					t.Fatal("right column did not redistribute capped space")
				}
			}
			// Blank space below a capped sidebar must not focus that sidebar.
			if layout != "stack" {
				u.focus = u.panelFocus("detail")
				clickAt(u, list.x+2, list.y+list.h+1)
				if u.focus != u.panelFocus("detail") {
					t.Fatal("blank capped area retained a mouse target")
				}
			}
			// The cap also applies to zoom and the small-screen single-panel view.
			u.focus = u.panelFocus("detail")
			u.zoom = true
			u.Draw()
			if panelRegion(t, u, "detail").h != 6 {
				t.Fatal("zoom exceeded height cap")
			}
			u.zoom = false
			u.Screen.(tcell.SimulationScreen).SetSize(80, 16)
			u.Draw()
			if panelRegion(t, u, "detail").h > 6 {
				t.Fatal("narrow layout exceeded height cap")
			}
		})
	}
}

func TestCappedPanelScrollKeepsOutput(t *testing.T) {
	u := heightUI(t, "stack", 120, 40)
	p := u.Engine.Panels["detail"]
	p.Definition.Format = "text"
	p.Text = strings.Repeat("line\n", 20) + "TAIL"
	u.focus = u.panelFocus("detail")
	press(u, tcell.KeyEnd)
	u.Draw()
	r := panelRegion(t, u, "detail")
	lines := u.panelLines["detail"]
	if !strings.Contains(p.Text, "TAIL") || lines[len(lines)-1] != "TAIL" || p.Scroll != len(lines)-(r.h-2) {
		t.Fatal("height cap truncated output or broke End scrolling")
	}
}
