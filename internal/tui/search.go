package tui

import "fmt"

func (u *UI) focusContext() string {
	location := fmt.Sprintf("焦点 %d/%d", u.focus+1, u.count())
	if u.focus == 0 {
		context := fmt.Sprintf("%s · 页面说明 · 行 %d/%d · ↑↓ 滚动", location, min(u.proseScroll+1, len(u.proseLines)), len(u.proseLines))
		if u.searchFocus == u.focus && u.searchIndex >= 0 {
			context += fmt.Sprintf(" · 匹配 %d", u.searchIndex+1)
		}
		return context
	}
	if u.focus <= len(u.Engine.Page.Inputs) {
		input := u.Engine.Page.Inputs[u.focus-1]
		if u.editing {
			return location + " · 参数 " + input.Label + " · ←→ 移动 · Del 删除 · Enter 提交"
		}
		_, hint := inputDisplay(input, u.Engine.Inputs[input.ID])
		return location + " · 参数 " + input.Label + " · " + hint
	}
	if button := u.focusedButton(); button != nil {
		context := location + " · 按钮 " + button.Label + " · Enter 执行"
		if reason := u.Engine.ActionReason(*button); reason != "" {
			context += " · 禁用: " + reason
		}
		return context
	}
	if p := u.panel(); p != nil {
		context := location + " · " + p.Definition.Title + " · ↑↓ 滚动 · / 搜索"
		if p.Definition.Format == "items-json" {
			context += " · F3 过滤"
		}
		if u.searchFocus == u.focus && u.searchIndex >= 0 {
			context += fmt.Sprintf(" · 匹配 %d", u.searchIndex+1)
		}
		return context
	}
	return location
}

func (u *UI) focusedLines() ([]string, *int) {
	if u.focus == 0 {
		return u.proseLines, &u.proseScroll
	}
	if p := u.panel(); p != nil {
		if p.Definition.Format == "items-json" {
			return u.panelLines[p.Definition.ID], &p.Cursor
		}
		return u.panelLines[p.Definition.ID], &p.Scroll
	}
	return nil, nil
}

func (u *UI) searchNext(direction int, includeCurrent bool) {
	lines, position := u.focusedLines()
	if position == nil || len(lines) == 0 || u.searchQuery == "" {
		u.notice = "当前区域没有可搜索的内容"
		return
	}
	start := min(max(*position, 0), len(lines)-1)
	if p := u.panel(); p != nil && p.Definition.Format == "items-json" {
		for row, index := range u.visibleItems(p) {
			if index == p.Cursor {
				start = row
				break
			}
		}
	}
	if u.searchFocus == u.focus && u.searchIndex >= 0 {
		start = min(u.searchIndex, len(lines)-1)
	} else {
		includeCurrent = true
		u.searchFocus = u.focus
	}
	firstStep := 1
	if includeCurrent {
		firstStep = 0
	}
	for step := firstStep; step < len(lines)+firstStep; step++ {
		index := (start + direction*step%len(lines) + len(lines)) % len(lines)
		if hasMatch(lines[index], u.searchQuery) {
			u.searchIndex = index
			*position = index
			if p := u.panel(); p != nil {
				if p.Definition.Format == "items-json" {
					indices := u.visibleItems(p)
					if index < len(indices) {
						p.Cursor = indices[index]
					}
					u.syncItemFocus()
				}
				p.Follow = false
			}
			u.notice = ""
			return
		}
	}
	u.notice = "未找到“" + u.searchQuery + "”"
}
