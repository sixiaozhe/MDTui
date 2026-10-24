package tui

import (
	"strings"
	"time"

	"mdtui/internal/content"
	"mdtui/internal/scheduler"
)

type filteredItems struct {
	first     *content.Item
	count     int
	query     string
	indices   []int
	lines     []string
	cursor    int
	selected  string
	cursorRow int
}

func firstItem(p *scheduler.Panel) *content.Item {
	if len(p.Items) == 0 {
		return nil
	}
	return &p.Items[0]
}

// Item slices are replaced atomically by the scheduler, never edited in place.
// Keep filtering and line construction out of unrelated stream redraws.
func (u *UI) itemLines(p *scheduler.Panel, visible []int) ([]string, int) {
	c := u.itemCache[p.Definition.ID]
	if c.lines != nil && c.cursor == p.Cursor && c.selected == p.Selected {
		return c.lines, c.cursorRow
	}
	c.lines = make([]string, len(visible))
	c.cursor, c.selected, c.cursorRow = p.Cursor, p.Selected, 0
	for row, i := range visible {
		item := p.Items[i]
		prefix := "   "
		if item.ID == p.Selected {
			prefix = " ✓ "
		}
		if i == p.Cursor {
			c.cursorRow = row
			prefix = " › "
			if item.ID == p.Selected {
				prefix = "›✓ "
			}
		}
		c.lines[row] = prefix + item.Label + "  " + item.Description
	}
	u.itemCache[p.Definition.ID] = c
	return c.lines, c.cursorRow
}

type panelVisual struct {
	text, status, reason, selected                   string
	items                                            *content.Item
	itemCount, cursor, scroll, x, dropped, truncated int
	follow, stale                                    bool
	updated                                          time.Time
}

func panelState(p *scheduler.Panel) panelVisual {
	return panelVisual{p.Text, p.Status, p.Reason, p.Selected, firstItem(p), len(p.Items), p.Cursor, p.Scroll, p.X, p.Dropped, p.Truncated, p.Follow, p.Stale, p.Updated}
}

type paintedPanel struct {
	state      panelVisual
	x, y, w, h int
	index      string
}

// Layout/input changes still use a full Draw. A stream-only tick repaints just
// changed, visible panels, retaining the other cells and their mouse targets.
func (u *UI) drawUpdates() {
	if u.modal != "" {
		return
	} // Never paint through an open modal.
	w, h := u.Screen.Size()
	if w != u.paintWidth || h != u.paintHeight || len(u.painted) == 0 {
		u.Draw()
		return
	}
	for id, p := range u.Engine.Panels {
		if u.paintStatus[id] != p.Status {
			u.Draw()
			return
		}
	}
	changed := false
	for id, last := range u.painted {
		p := u.Engine.Panels[id]
		if panelState(p) == last.state {
			continue
		}
		focus := u.panelFocus(id)
		kept := u.mouseRegions[:0]
		for _, region := range u.mouseRegions {
			if (region.kind == "panel" || region.kind == "item") && region.focus == focus {
				continue
			}
			kept = append(kept, region)
		}
		u.mouseRegions = kept
		u.drawPanel(id, last.x, last.y, last.w, last.h, last.index)
		changed = true
	}
	if changed {
		u.Screen.Show()
	}
}

func (u *UI) filtered(p *scheduler.Panel) filteredItems {
	if u.itemCache == nil {
		u.itemCache = map[string]filteredItems{}
	}
	id := p.Definition.ID
	query := u.filters[id]
	c, ok := u.itemCache[id]
	if ok && c.first == firstItem(p) && c.count == len(p.Items) && c.query == query {
		return c
	}
	c = filteredItems{first: firstItem(p), count: len(p.Items), query: query, indices: make([]int, 0, len(p.Items))}
	match := strings.ToLower(strings.TrimSpace(query))
	for i, item := range p.Items {
		if match == "" || strings.Contains(strings.ToLower(item.ID+" "+item.Label+" "+item.Description), match) {
			c.indices = append(c.indices, i)
		}
	}
	u.itemCache[id] = c
	return c
}
