package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"mdtui/internal/content"
)

func mouseRegionFor(t *testing.T, u *UI, kind string, index int) mouseRegion {
	t.Helper()
	u.Draw()
	for _, region := range u.mouseRegions {
		if region.kind == kind && (index < 0 || region.index == index) {
			return region
		}
	}
	t.Fatalf("missing mouse target %s[%d]", kind, index)
	return mouseRegion{}
}

func clickAt(u *UI, x, y int) {
	u.Mouse(tcell.NewEventMouse(x, y, tcell.Button1, 0))
	u.Draw()
	u.Mouse(tcell.NewEventMouse(x, y, tcell.ButtonNone, 0))
	u.Draw()
}

func clickRegion(u *UI, r mouseRegion) { clickAt(u, r.x+r.w/2, r.y) }

func TestMouseNavigationAndHeldButton(t *testing.T) {
	u := navUI(t, childPage)
	t.Cleanup(u.Close)
	r := mouseRegionFor(t, u, "item", 1)
	u.Mouse(tcell.NewEventMouse(r.x, r.y, tcell.Button1, 0))
	u.Mouse(tcell.NewEventMouse(r.x+1, r.y, tcell.Button1, 0))
	if u.Engine.Page.ID != "root" || u.Engine.Panels["list"].Cursor != 1 || u.Engine.Panels["list"].Selected != "" {
		t.Fatal("press/drag confirmed a list item")
	}
	u.Mouse(tcell.NewEventMouse(r.x, r.y, tcell.ButtonNone, 0))
	clickRegion(u, r)
	if u.Engine.Page.ID != "child" || u.Engine.ParentSelected != "B" {
		t.Fatal("double click did not navigate using item ID")
	}
	back := mouseRegionFor(t, u, "back", -1)
	clickRegion(u, back)
	if u.Engine.Page.ID != "root" || u.Engine.Panels["list"].Selected != "B" || u.Engine.Paused() {
		t.Fatal("mouse back did not restore parent")
	}
	// A stale child-page target must not act on the restored parent.
	u.Mouse(tcell.NewEventMouse(back.x, back.y, tcell.Button1, 0))
	u.Mouse(tcell.NewEventMouse(back.x, back.y, tcell.ButtonNone, 0))
	if u.Engine.Page.ID != "root" {
		t.Fatal("stale back target changed page")
	}
}

func TestMouseFilteredRowsAndRefreshIdentity(t *testing.T) {
	u := testUI(t)
	t.Cleanup(u.Close)
	applyFilter(u, "B")
	r := mouseRegionFor(t, u, "item", 1)
	clickRegion(u, r)
	if u.Engine.Panels["list"].Cursor != 1 {
		t.Fatal("filter row used visible index as item index")
	}
	clickRegion(u, r)
	if u.Engine.Panels["list"].Selected != "B" {
		t.Fatal("double click confirmed a hidden/wrong filtered item")
	}
}

func TestMouseDoubleClickRequiresSameObject(t *testing.T) {
	u := testUI(t)
	t.Cleanup(u.Close)
	r := mouseRegionFor(t, u, "item", 0)
	clickRegion(u, r)
	u.Engine.Panels["list"].Items[0].ID = "changed"
	r = mouseRegionFor(t, u, "item", 0)
	clickRegion(u, r)
	if u.Engine.Panels["list"].Selected != "" {
		t.Fatal("refresh turned clicks on two different objects into a double click")
	}
	u.lastClick.at = time.Now().Add(-time.Second)
	clickRegion(u, r)
	if u.Engine.Panels["list"].Selected != "" {
		t.Fatal("slow consecutive clicks confirmed an item")
	}
}

func TestMouseDoubleClickInterruptedByOtherInput(t *testing.T) {
	for _, interrupt := range []string{"wheel", "keyboard"} {
		t.Run(interrupt, func(t *testing.T) {
			u := testUI(t)
			t.Cleanup(u.Close)
			r := mouseRegionFor(t, u, "item", 0)
			clickRegion(u, r)
			if interrupt == "wheel" {
				u.Mouse(tcell.NewEventMouse(r.x, r.y, tcell.WheelDown, 0))
			} else {
				press(u, tcell.KeyDown)
			}
			clickRegion(u, r)
			if u.Engine.Panels["list"].Selected != "" {
				t.Fatal("interrupted clicks confirmed an item")
			}
		})
	}
}

func TestMouseWheelAndScrolledRows(t *testing.T) {
	u := testUI(t)
	t.Cleanup(u.Close)
	p := u.Engine.Panels["list"]
	p.Items = nil
	for i := 0; i < 40; i++ {
		p.Items = append(p.Items, content.Item{ID: fmt.Sprintf("item%d", i), Label: fmt.Sprintf("row%d", i)})
	}
	p.Cursor = 39
	r := mouseRegionFor(t, u, "item", 39)
	if p.Scroll == 0 {
		t.Fatal("test list did not scroll")
	}
	clickRegion(u, r)
	if p.Cursor != 39 {
		t.Fatal("scrolled row used wrong item index")
	}
	u.Mouse(tcell.NewEventMouse(r.x+r.w/2, r.y, tcell.WheelUp, 0))
	if p.Cursor != 38 {
		t.Fatal("list wheel did not move cursor")
	}
	detail := u.Engine.Panels["detail"]
	detail.Text = strings.Repeat("line\n", 100)
	detail.Definition.Format = "text"
	r = mouseRegionFor(t, u, "panel", -1)
	for _, region := range u.mouseRegions {
		if region.kind == "panel" && region.focus == u.panelFocus("detail") {
			r = region
			break
		}
	}
	u.Mouse(tcell.NewEventMouse(r.x, r.y, tcell.WheelDown, 0))
	if u.panel() != detail || detail.Scroll != 3 || detail.Follow {
		t.Fatal("wheel did not focus and scroll hovered content")
	}
}

func TestMouseControlsAndDraftProtection(t *testing.T) {
	u := controlsUI(t)
	clickRegion(u, mouseRegionFor(t, u, "input", 0))
	if u.Engine.Inputs["verbose"] != "true" {
		t.Fatal("click did not toggle")
	}
	clickRegion(u, mouseRegionFor(t, u, "input-next", 0))
	clickRegion(u, mouseRegionFor(t, u, "input", 1))
	if u.modal != "select" {
		t.Fatal("click did not open dropdown")
	}
	clickAt(u, 0, 0)
	if u.modal != "select" || u.Engine.Inputs["protocol"] != "tcp" {
		t.Fatal("outside click dismissed or changed dropdown")
	}
	clickRegion(u, mouseRegionFor(t, u, "menu", 1))
	if u.modal != "" || u.Engine.Inputs["protocol"] != "udp" {
		t.Fatal("click did not commit option")
	}
	r := mouseRegionFor(t, u, "input", 1)
	u.Mouse(tcell.NewEventMouse(r.x, r.y, tcell.WheelDown, 0))
	clickRegion(u, mouseRegionFor(t, u, "input", 2))
	r = mouseRegionFor(t, u, "input", 2)
	u.Mouse(tcell.NewEventMouse(r.x+r.w/2, r.y, tcell.WheelUp, 0))
	if !u.editing || u.draft != "6" {
		t.Fatal("wheel did not step numeric draft")
	}
	clickRegion(u, mouseRegionFor(t, u, "button", 0))
	if !u.editing || u.focus != 3 || u.Engine.Panels["result"].Request != 0 {
		t.Fatal("mouse discarded a draft or executed during editing")
	}
	press(u, tcell.KeyEnter)
	clickRegion(u, mouseRegionFor(t, u, "button", 0))
	if u.Engine.Panels["result"].Request == 0 {
		t.Fatal("mouse button did not execute")
	}
	clickRegion(u, mouseRegionFor(t, u, "button-next", 0))
	clickRegion(u, mouseRegionFor(t, u, "button", 1))
	if u.modal != "confirm" || u.confirm {
		t.Fatal("mouse bypassed confirmation")
	}
	clickRegion(u, mouseRegionFor(t, u, "cancel", -1))
	if u.modal != "" || u.confirm {
		t.Fatal("mouse cancel failed")
	}
	clickRegion(u, mouseRegionFor(t, u, "button", 1))
	clickRegion(u, mouseRegionFor(t, u, "execute", -1))
	if u.modal != "" || !u.confirm {
		t.Fatal("mouse execute failed")
	}
}

func TestMouseModalIsolationResizeAndOptOut(t *testing.T) {
	u := controlsUI(t)
	u.Engine.Page.Actions[1].Confirm = "● 取消      ○ 执行" // prompt text is not a hit target
	u.focus = 6
	press(u, tcell.KeyEnter)
	u.Draw()
	count := 0
	for _, r := range u.mouseRegions {
		if r.kind == "execute" {
			count++
		}
		if r.kind == "input" || r.kind == "button" {
			t.Fatal("modal exposed underlying controls")
		}
	}
	if count != 1 {
		t.Fatal("prompt text created executable mouse targets")
	}
	clickRegion(u, mouseRegionFor(t, u, "close", -1))
	if u.modal != "" {
		t.Fatal("modal close did not work")
	}
	u.focus = 1
	r := mouseRegionFor(t, u, "input", 0)
	u.NoMouse = true
	clickRegion(u, r)
	if u.Engine.Inputs["verbose"] != "false" {
		t.Fatal("disabled mouse handled click")
	}
	u.NoMouse = false
	u.Screen.(tcell.SimulationScreen).SetSize(59, 15)
	u.Draw()
	clickRegion(u, r)
	if len(u.mouseRegions) != 0 || u.Engine.Inputs["verbose"] != "false" {
		t.Fatal("small-screen layout retained stale targets")
	}
}

func TestMouseWorkspaceTabs(t *testing.T) {
	u := testUI(t)
	u.Engine.Page.Layout = "workspace"
	u.Engine.Page.Sidebar = "list"
	r := mouseRegionFor(t, u, "tab", -1)
	clickRegion(u, r)
	if u.panel() != u.Engine.Panels["detail"] || u.workspaceDetail != "detail" {
		t.Fatal("mouse tab did not switch workspace view")
	}
}
