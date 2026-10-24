package tui

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
	"mdtui/internal/document"
)

const (
	colorCanvas    = 0x101923
	colorSurface   = 0x16222f
	colorElevated  = 0x1e2e3e
	colorSelection = 0x28485d
	colorText      = 0xd9e4ef
	colorMuted     = 0x8a9bae
	colorBorder    = 0x35495c
	colorAccent    = 0x80c8de
	colorSuccess   = 0x98c9a8
	colorWarning   = 0xe4c38c
	colorDanger    = 0xe8a1a3
)

func (u *UI) palette(fg, bg int) tcell.Style {
	if u.NoColor {
		return tcell.StyleDefault
	}
	return tcell.StyleDefault.Foreground(tcell.NewHexColor(int32(fg))).Background(tcell.NewHexColor(int32(bg)))
}
func (u *UI) baseStyle() tcell.Style    { return u.palette(colorText, colorCanvas) }
func (u *UI) surfaceStyle() tcell.Style { return u.palette(colorText, colorSurface) }
func (u *UI) mutedStyle() tcell.Style   { return u.palette(colorMuted, colorCanvas) }
func (u *UI) accentStyle() tcell.Style  { return u.palette(colorAccent, colorCanvas) }
func (u *UI) controlStyle(active bool) tcell.Style {
	style := u.palette(colorText, colorElevated)
	if active {
		if u.NoColor {
			return style.Bold(true).Reverse(true)
		}
		return u.palette(colorText, colorSelection).Bold(true)
	}
	return style
}
func (u *UI) panelBorder(active bool) tcell.Style {
	if active {
		return u.palette(colorAccent, colorSurface).Bold(true)
	}
	return u.palette(colorBorder, colorSurface)
}

// Clip titles and badges before drawing so long names cannot overlap nearby UI.
func fitText(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if uniseg.StringWidth(value) <= width {
		return value
	}
	g := uniseg.NewGraphemes(value)
	result, used := "", 0
	for g.Next() {
		if used+g.Width() > width-1 {
			break
		}
		result += g.Str()
		used += g.Width()
	}
	return result + "…"
}

func (u *UI) drawShortcuts(width, y int) {
	style := u.palette(colorMuted, colorSurface)
	if u.NoColor {
		style = style.Reverse(true)
	}
	u.fill(0, y, width, 1, ' ', style)
	type hint struct{ key, label string }
	hints := []hint{{"Tab", "焦点"}, {"Enter", "确认"}, {"F3", "过滤"}, {"F4", "导出"}, {"?", "帮助"}}
	exit := hint{"q", "退出"}
	x := 2
	if len(u.sessions) > 0 {
		exit.label = "返回"
		hints = []hint{{"Esc", "返回"}, {"Enter", "确认"}, {"F3", "过滤"}, {"F4", "导出"}}
		u.put(x, y, 3, 0, fmt.Sprintf("%d/%d", u.depth(), document.MaxDepth), style)
		x += 5
	}
	if width >= 100 {
		hints = []hint{{"Tab", "焦点"}, {"Enter", "确认"}, {"/", "搜索"}, {"F2", "动作"}, {"F3", "过滤"}, {"F4", "导出"}, {"Z", "放大"}, {"?", "帮助"}}
		if len(u.sessions) > 0 {
			hints = append([]hint{{"Esc", "返回"}}, hints...)
		}
	}
	exitWidth := uniseg.StringWidth(exit.key) + 1 + uniseg.StringWidth(exit.label)
	exitX := width - 2 - exitWidth
	draw := func(x int, h hint) {
		keyStyle := u.palette(colorText, colorElevated).Bold(true)
		if u.NoColor {
			keyStyle = keyStyle.Reverse(true)
		}
		keyWidth := uniseg.StringWidth(h.key)
		u.put(x, y, keyWidth, 0, h.key, keyStyle)
		u.put(x+keyWidth+1, y, uniseg.StringWidth(h.label), 0, h.label, style)
	}
	for _, h := range hints {
		size := uniseg.StringWidth(h.key) + 1 + uniseg.StringWidth(h.label)
		if x+size > exitX-2 {
			break
		}
		draw(x, h)
		x += size + 2
	}
	draw(exitX, exit)
}
