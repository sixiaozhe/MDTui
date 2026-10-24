package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"mdtui/internal/content"
)

func applyFilter(u *UI, query string) {
	u.Key(tcell.NewEventKey(tcell.KeyF3, 0, 0))
	u.searchDraft = query
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
}

func TestFilterSelectionAndSearch(t *testing.T) {
	u := testUI(t)
	defer u.Close()
	p := u.Engine.Panels["list"]
	p.Items = []content.Item{{ID: "A", Label: "first"}, {ID: "B", Label: "网络", Description: "Ethernet"}, {ID: "C", Label: "third", Description: "Ethernet"}}
	p.Selected = "A"
	applyFilter(u, "eThErNeT")
	if len(p.Items) != 3 || p.Selected != "A" || p.Cursor != 1 {
		t.Fatalf("filter changed data/selection or failed to move cursor: %+v", p)
	}
	u.Key(tcell.NewEventKey(tcell.KeyDown, 0, 0))
	if p.Cursor != 2 {
		t.Fatal("filtered down did not reach C")
	}
	u.Key(tcell.NewEventKey(tcell.KeyHome, 0, 0))
	if p.Cursor != 1 {
		t.Fatal("filtered home did not reach B")
	}
	u.Draw()
	u.searchQuery = "third"
	u.searchFocus = -1
	u.searchNext(1, true)
	if p.Cursor != 2 {
		t.Fatal("search used filtered row number as engine cursor")
	}
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	if p.Selected != "C" {
		t.Fatal("Enter selected hidden/wrong item")
	}
	applyFilter(u, "no-match")
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	u.Draw()
	if p.Selected != "C" || len(u.visibleItems(p)) != 0 {
		t.Fatal("empty filter changed selection")
	}
	u.Key(tcell.NewEventKey(tcell.KeyF3, 0, 0))
	u.searchDraft = "A"
	u.Key(tcell.NewEventKey(tcell.KeyEscape, 0, 0))
	if u.filters["list"] != "no-match" {
		t.Fatal("cancel applied draft")
	}
	applyFilter(u, "")
	if len(u.visibleItems(p)) != 3 {
		t.Fatal("clear did not restore list")
	}
	applyFilter(u, "网络")
	if len(u.visibleItems(p)) != 1 || p.Cursor != 1 {
		t.Fatal("Unicode label filter failed")
	}
}

func TestNavigationContextAndFilterRestore(t *testing.T) {
	u := navUI(t, childPage)
	defer u.Close()
	applyFilter(u, "one")
	u.Key(tcell.NewEventKey(tcell.KeyEnter, 0, 0))
	if got := u.navigationContext(); !strings.Contains(got, "root → child") || !strings.Contains(got, "父页对象: A") {
		t.Fatal(got)
	}
	if len(u.filters) != 0 {
		t.Fatal("parent filter leaked into child")
	}
	u.popChild()
	if u.filters["list"] != "one" {
		t.Fatal("parent filter was not restored")
	}
}

func TestExportSnapshot(t *testing.T) {
	u := testUI(t)
	now := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	u.Engine.Now = func() time.Time { return now }
	u.Engine.Inputs["filter"] = "设备甲"
	p := u.Engine.Panels["detail"]
	p.Text = "```\noutput\n````"
	p.Stderr = "permission denied"
	p.Status = "error"
	p.Reason = "exit status 1"
	p.Updated = now
	p.Stale = true
	u.filters = map[string]string{"list": "B"}
	dir := t.TempDir()
	first, err := u.exportSnapshot(dir)
	if err != nil {
		t.Fatal(err)
	}
	second, err := u.exportSnapshot(dir)
	if err != nil || first == second {
		t.Fatalf("export collision: %s %v", second, err)
	}
	data, err := os.ReadFile(first)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"2026-10-24T12:00:00Z", "设备甲", "permission denied", "exit status 1", "结果已过期", "`````text", "A\t中文", "B\t第二项"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("report missing %q", want)
		}
	}
	if !filepath.IsAbs(first) {
		t.Fatal("export path is not absolute")
	}
	if _, err = u.exportSnapshot(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing destination did not report failure")
	}
}
