package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"mdtui/internal/content"
	"mdtui/internal/scheduler"
)

func (u *UI) navigationContext() string {
	var parts []string
	for _, s := range u.sessions {
		title := s.Engine.Page.Title
		if s.Engine.ParentSelected != "" {
			title += " [" + s.Engine.ParentSelected + "]"
		}
		parts = append(parts, title)
	}
	parts = append(parts, u.Engine.Page.Title)
	result := strings.Join(parts, " → ")
	if u.Engine.ParentSelected != "" {
		result += " · 父页对象: " + u.Engine.ParentSelected
	}
	return result
}

// Keep engine cursors indexed into the full list so references and selection
// retain their stable IDs even when rows are hidden by a filter.
func (u *UI) visibleItems(p *scheduler.Panel) []int {
	return u.filtered(p).indices
}

func (u *UI) ensureVisibleCursor(p *scheduler.Panel, indices []int) {
	for _, i := range indices {
		if p.Cursor == i {
			return
		}
	}
	if len(indices) > 0 {
		p.Cursor = indices[0]
	}
}

func (u *UI) scrollFiltered(p *scheduler.Panel, key tcell.Key) {
	p.Follow = key == tcell.KeyEnd
	indices := u.visibleItems(p)
	if len(indices) == 0 {
		return
	}
	row := 0
	for i, index := range indices {
		if index == p.Cursor {
			row = i
			break
		}
	}
	switch key {
	case tcell.KeyUp:
		row--
	case tcell.KeyDown:
		row++
	case tcell.KeyPgUp:
		row -= 8
	case tcell.KeyPgDn:
		row += 8
	case tcell.KeyHome:
		row = 0
	case tcell.KeyEnd:
		row = len(indices) - 1
	}
	row = min(max(row, 0), len(indices)-1)
	p.Cursor = indices[row]
	u.syncItemFocus()
}

// Use a fence longer than any run in the content, preserving arbitrary command
// output as text rather than letting it become report markup.
func reportBlock(value string) string {
	longest, run := 0, 0
	for _, ch := range value {
		if ch == '`' {
			run++
			longest = max(longest, run)
		} else {
			run = 0
		}
	}
	fence := strings.Repeat("`", max(3, longest+1))
	return fence + "text\n" + value + "\n" + fence + "\n\n"
}

func (u *UI) snapshotMarkdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# MDTUI 诊断快照\n\n采集时间: %s\n\n", u.Engine.Now().Format(time.RFC3339))
	b.WriteString(reportBlock("导航: " + u.navigationContext() + "\n页面: " + u.Engine.Page.File))
	if len(u.Engine.Page.Inputs) > 0 {
		b.WriteString("## 页面参数\n\n")
		for _, input := range u.Engine.Page.Inputs {
			b.WriteString(reportBlock(input.Label + " (" + input.ID + "): " + u.Engine.Inputs[input.ID]))
		}
	}
	for _, def := range u.Engine.Page.Panels {
		p := u.Engine.Panels[def.ID]
		b.WriteString("## 面板\n\n")
		meta := fmt.Sprintf("%s (%s)\n状态: %s\n选中对象: %s", def.Title, def.ID, statusLabel(p.Status), p.Selected)
		if !p.Updated.IsZero() {
			meta += "\n结果更新时间: " + p.Updated.Format(time.RFC3339)
		}
		if p.Stale {
			meta += "\n结果已过期"
		}
		if p.Reason != "" {
			meta += "\n原因: " + p.Reason
		}
		if p.Dropped+p.Truncated > 0 {
			meta += fmt.Sprintf("\n丢弃 %d 行 / 截断 %d 行", p.Dropped, p.Truncated)
		}
		b.WriteString(reportBlock(meta))
		if def.Format == "items-json" {
			var items strings.Builder
			for _, item := range p.Items {
				fmt.Fprintf(&items, "%s\t%s\t%s\n", item.ID, item.Label, item.Description)
			}
			b.WriteString(reportBlock(items.String()))
		} else {
			b.WriteString(reportBlock(content.Safe(p.Text)))
		}
		if p.Stderr != "" {
			b.WriteString("错误输出:\n\n" + reportBlock(content.Safe(p.Stderr)))
		}
	}
	return b.String()
}

func (u *UI) exportSnapshot(dir string) (string, error) {
	f, err := os.CreateTemp(dir, "mdtui-report-"+u.Engine.Now().Format("20060102-150405")+"-*.md")
	if err != nil {
		return "", err
	}
	name := f.Name()
	_, writeErr := f.WriteString(u.snapshotMarkdown())
	closeErr := f.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(name)
		if writeErr != nil {
			return "", writeErr
		}
		return "", closeErr
	}
	return filepath.Abs(name)
}
