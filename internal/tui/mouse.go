package tui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"mdtui/internal/scheduler"
)

// Hit targets are recorded by the renderer, so wrapped text, filtering,
// scrolling and responsive layouts share exactly the same coordinates.
type mouseRegion struct {
	x, y, w, h   int
	kind         string
	focus, index int
	engine       *scheduler.Engine
}

type mouseClick struct {
	region mouseRegion
	at     time.Time
	itemID string
}

func (u *UI) addMouse(region mouseRegion) {
	w, h := u.Screen.Size()
	right, bottom := min(w, region.x+region.w), min(h, region.y+region.h)
	region.x, region.y = max(0, region.x), max(0, region.y)
	region.w, region.h = right-region.x, bottom-region.y
	region.engine = u.Engine
	if region.w > 0 && region.h > 0 {
		u.mouseRegions = append(u.mouseRegions, region)
	}
}

func (u *UI) mouseTarget(x, y int) (mouseRegion, bool) {
	for i := len(u.mouseRegions) - 1; i >= 0; i-- {
		r := u.mouseRegions[i]
		if r.engine == u.Engine && x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h {
			return r, true
		}
	}
	return mouseRegion{}, false
}

func (u *UI) mouseFocus(focus int) {
	u.focus = focus
	if p := u.panel(); p != nil && p.Definition.ID != u.Engine.Page.Sidebar {
		u.workspaceDetail = p.Definition.ID
	}
}

func (u *UI) mouseKey(key tcell.Key) { u.Key(tcell.NewEventKey(key, 0, 0)) }

func (u *UI) Mouse(event *tcell.EventMouse) {
	if u.NoMouse {
		return
	}
	buttons := event.Buttons()
	pressed := buttons&tcell.Button1 != 0 && u.mouseButtons&tcell.Button1 == 0
	u.mouseButtons = buttons & (tcell.Button1 | tcell.Button2 | tcell.Button3)
	x, y := event.Position()
	region, ok := u.mouseTarget(x, y)
	if !ok {
		if pressed {
			u.lastClick = mouseClick{}
		}
		return
	}
	if buttons&(tcell.WheelUp|tcell.WheelDown) != 0 {
		u.lastClick = mouseClick{}
		direction := 1
		if buttons&tcell.WheelUp != 0 {
			direction = -1
		}
		u.mouseWheel(region, direction)
		return
	}
	if !pressed {
		return
	} // Releases and held-button motion never activate twice.
	if u.editing || u.searching || u.filtering {
		u.notice = "请先按 Enter 提交或 Esc 取消当前编辑"
		return
	}
	if region.kind != "item" {
		u.lastClick = mouseClick{}
	}
	switch region.kind {
	case "back":
		u.popChild()
	case "close":
		u.mouseKey(tcell.KeyEscape)
	case "menu":
		u.menu = region.index
		u.mouseKey(tcell.KeyEnter)
	case "cancel", "execute":
		u.confirm = region.kind == "execute"
		u.mouseKey(tcell.KeyEnter)
	case "prose":
		u.mouseFocus(0)
	case "panel", "tab", "buttonbar":
		u.mouseFocus(region.focus)
	case "button":
		u.mouseFocus(region.focus)
		u.mouseKey(tcell.KeyEnter)
	case "input":
		u.mouseFocus(1 + region.index)
		u.mouseKey(tcell.KeyEnter)
	case "input-prev", "input-next":
		direction := 1
		if region.kind == "input-prev" {
			direction = -1
		}
		u.cycleMouseInput(region.index, direction)
	case "button-prev", "button-next":
		direction := 1
		if region.kind == "button-prev" {
			direction = -1
		}
		u.cycleMouseButton(region.index, direction)
	case "item":
		u.mouseFocus(region.focus)
		p := u.panel()
		if p == nil || region.index < 0 || region.index >= len(p.Items) {
			return
		}
		p.Cursor = region.index
		u.syncItemFocus()
		last := u.lastClick
		elapsed := event.When().Sub(last.at)
		if last.region.engine == region.engine && last.region.focus == region.focus && last.region.index == region.index && last.itemID == p.Items[region.index].ID && elapsed >= 0 && elapsed <= 400*time.Millisecond {
			u.lastClick = mouseClick{}
			u.mouseKey(tcell.KeyEnter)
		} else {
			u.lastClick = mouseClick{region: region, at: event.When(), itemID: p.Items[region.index].ID}
		}
	}
	u.syncItemFocus()
}

func (u *UI) cycleMouseInput(index, direction int) {
	n := len(u.Engine.Page.Inputs)
	if n > 0 {
		u.mouseFocus(1 + (index+direction+n)%n)
	}
}

func (u *UI) cycleMouseButton(index, direction int) {
	n := len(u.buttons())
	if n > 0 {
		u.mouseFocus(1 + len(u.Engine.Page.Inputs) + len(u.Engine.Page.Panels) + (index+direction+n)%n)
	}
}

func (u *UI) mouseWheel(region mouseRegion, direction int) {
	if u.modal != "" {
		if u.modal == "select" {
			key := tcell.KeyDown
			if direction < 0 {
				key = tcell.KeyUp
			}
			u.mouseKey(key)
		} else if u.modal == "actions" {
			u.menu = min(max(0, u.menu+direction), max(0, len(u.Engine.Page.Actions)-1))
		} else {
			u.modalScroll = max(0, u.modalScroll+3*direction)
		}
		return
	}
	if u.editing {
		if region.kind == "input" && u.Engine.Page.Inputs[u.focus-1].Type == "number" {
			key := tcell.KeyUp
			if direction > 0 {
				key = tcell.KeyDown
			}
			u.mouseKey(key)
		}
		return
	}
	if u.searching || u.filtering {
		return
	}
	switch region.kind {
	case "input", "input-prev", "input-next":
		u.cycleMouseInput(region.index, direction)
	case "button", "buttonbar", "button-prev", "button-next":
		u.cycleMouseButton(region.index, direction)
	case "prose", "panel", "item":
		if region.kind == "prose" {
			u.mouseFocus(0)
		} else {
			u.mouseFocus(region.focus)
		}
		key := tcell.KeyDown
		if direction < 0 {
			key = tcell.KeyUp
		}
		steps := 3
		if p := u.panel(); p != nil && p.Definition.Format == "items-json" {
			steps = 1
		}
		for i := 0; i < steps; i++ {
			u.scroll(key)
		}
		u.syncItemFocus()
	}
}
