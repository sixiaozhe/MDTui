package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePage(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func pageBody(id, extra, panels string) string {
	return ":::mdtui page\nversion: 1\nid: " + id + "\n" + extra + ":::\n" + panels
}

func TestCheckFollowsOpen(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "child.md", pageBody("child", "parent: list\n", ":::mdtui panel\nid: detail\nexec: [echo]\n:::\n"))
	root := writePage(t, dir, "root.md", pageBody("root", "", ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: ./child.md\n:::\n"))
	pages, err := Check(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 || pages[0].ID != "root" || pages[1].ID != "child" {
		t.Fatalf("pages = %#v", pages)
	}
}

func TestCheckParentMismatch(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "child.md", pageBody("child", "parent: other\n", ""))
	root := writePage(t, dir, "root.md", pageBody("root", "", ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: ./child.md\n:::\n"))
	_, err := Check(root)
	if err == nil || !strings.Contains(err.Error(), "不一致") {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckCycle(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "a.md", pageBody("a", "", ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: ./b.md\n:::\n"))
	writePage(t, dir, "b.md", pageBody("b", "", ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: ./a.md\n:::\n"))
	_, err := Check(filepath.Join(dir, "a.md"))
	if err == nil || !strings.Contains(err.Error(), "循环页面引用") {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckDepth(t *testing.T) {
	dir := t.TempDir()
	for i := 1; i <= 9; i++ {
		panels := ""
		if i < 9 {
			panels = fmt.Sprintf(":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: ./p%d.md\n:::\n", i+1)
		}
		writePage(t, dir, fmt.Sprintf("p%d.md", i), pageBody(fmt.Sprintf("p%d", i), "", panels))
	}
	if _, err := Check(filepath.Join(dir, "p1.md")); err == nil || !strings.Contains(err.Error(), "超过 8 层") {
		t.Fatalf("err = %v", err)
	}
}

func TestResolveRejectsAbsoluteAndSymlinkCycle(t *testing.T) {
	dir := t.TempDir()
	page := &Page{File: filepath.Join(dir, "root.md"), Dir: dir}
	if _, err := Resolve(page, filepath.Join(dir, "child.md")); err == nil {
		t.Fatal("accepted absolute open path")
	}
	writePage(t, dir, "real.md", pageBody("real", "", ""))
	if err := os.Symlink(filepath.Join(dir, "real.md"), filepath.Join(dir, "link.md")); err != nil {
		t.Fatal(err)
	}
	got, err := Resolve(page, "./link.md")
	if err != nil {
		t.Fatal(err)
	}
	want, err := Canonical(filepath.Join(dir, "real.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("resolve = %q, want %q", got, want)
	}
}
