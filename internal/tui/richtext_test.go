package tui

import (
	"strings"
	"testing"

	"github.com/rivo/uniseg"
	"mdtui/internal/content"
)

func TestRichMarkdownWrapsProseButKeepsWideCodeAndTables(t *testing.T) {
	long := strings.Repeat("中文", 14)
	rendered := content.Markdown("**" + long + "**\n\n```text\n" + long + "\n```\n\n| 名称 | 值 |\n| --- | --- |\n| " + long + " | 1 |")
	lines := wrapRichLines(strings.Split(rendered, "\n"), 12)
	proseCount, wideCount := 0, 0
	for _, line := range lines {
		plain := visibleRich(line)
		if strings.Contains(plain, "中文") && !richNoWrap(line) {
			proseCount++
			if width := uniseg.StringWidth(plain); width > 12 {
				t.Fatalf("prose width %d exceeds 12: %q", width, plain)
			}
		}
		if strings.Contains(plain, long) && richNoWrap(line) {
			wideCount++
		}
	}
	if proseCount < 2 || wideCount < 2 {
		t.Fatalf("expected wrapped prose and unwrapped code/table; prose=%d wide=%d", proseCount, wideCount)
	}
	if !strings.Contains(lines[1], "\x1b[1m") {
		t.Fatal("bold styling was lost after wrapping")
	}
}

func TestSearchMatchesUnicodeAndIgnoresStyles(t *testing.T) {
	line := content.Markdown("**网络** status")
	if !hasMatch(line, "网络") || !hasMatch(line, "STATUS") {
		t.Fatal("styled text was not searchable")
	}
	if hasMatch(line, "missing") || len(matchCells(line, "网络")) != 1 {
		t.Fatal("incorrect search results")
	}
}
