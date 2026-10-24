package tui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func screenRow(u *UI, y int) string {
	w, _ := u.Screen.Size()
	var text strings.Builder
	for x := 0; x < w; x++ {
		ch, _, _, width := u.Screen.GetContent(x, y)
		text.WriteRune(ch)
		x += max(1, width) - 1
	}
	return text.String()
}

func TestNarrowChromePreservesPageAndExit(t *testing.T) {
	u := testUI(t)
	u.Engine.Page.ID = strings.Repeat("long-id-", 7)
	u.Engine.Page.Title = strings.Repeat("很长的页面标题", 15)
	for _, size := range [][2]int{{60, 16}, {80, 24}, {120, 40}} {
		u.Screen.(tcell.SimulationScreen).SetSize(size[0], size[1])
		u.Draw()
		header := screenRow(u, 0)
		if !strings.Contains(header, "◆ MDTUI") || !strings.Contains(header, "PAGE ·") || !strings.Contains(header, "…") {
			t.Fatalf("header overlap at %v: %s", size, header)
		}
		if footer := screenRow(u, size[1]-1); !strings.Contains(footer, "q 退出") {
			t.Fatalf("exit shortcut clipped at %v: %s", size, footer)
		}
	}
}

func TestThemeHonorsNoColor(t *testing.T) {
	u := testUI(t)
	u.NoColor = true
	u.Engine.Page.Markdown = "# 说明\n\n**强调**"
	u.Draw()
	w, h := u.Screen.Size()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			_, _, style, _ := u.Screen.GetContent(x, y)
			fg, bg, _ := style.Decompose()
			if fg != tcell.ColorDefault || bg != tcell.ColorDefault {
				t.Fatalf("no-color mode forced a color at %d,%d", x, y)
			}
		}
	}
}
