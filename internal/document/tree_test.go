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
	if err == nil || !strings.Contains(err.Error(), "不一致") || !strings.Contains(err.Error(), "list") || !strings.Contains(err.Error(), "./child.md") {
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

func TestCheckDepthEightPasses(t *testing.T) {
	dir := t.TempDir()
	for i := 1; i <= 8; i++ {
		panels := ""
		if i < 8 {
			panels = fmt.Sprintf(":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: ./p%d.md\n:::\n", i+1)
		}
		writePage(t, dir, fmt.Sprintf("p%d.md", i), pageBody(fmt.Sprintf("p%d", i), "", panels))
	}
	pages, err := Check(filepath.Join(dir, "p1.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 8 {
		t.Fatalf("pages = %d, want 8", len(pages))
	}
}

func TestCheckMissingTargetContext(t *testing.T) {
	dir := t.TempDir()
	root := writePage(t, dir, "root.md", pageBody("root", "", ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: ./missing.md\n:::\n"))
	_, err := Check(root)
	if err == nil || !strings.Contains(err.Error(), "root.md") || !strings.Contains(err.Error(), "list") || !strings.Contains(err.Error(), "missing.md") {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckDuplicateTargetLoadedOnce(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "child.md", pageBody("child", "", ""))
	root := writePage(t, dir, "root.md", pageBody("root", "", ":::mdtui panel\nid: p1\nformat: items-json\nexec: [echo]\nopen: ./child.md\n:::\n:::mdtui panel\nid: p2\nformat: items-json\nexec: [echo]\nopen: ./child.md\n:::\n"))
	loads := 0
	pages, err := check(root, func(path string) (*Page, error) {
		loads++
		return Load(path)
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 {
		t.Fatalf("pages = %d, want 2", len(pages))
	}
	if loads != 2 {
		t.Fatalf("loads = %d, want 2", loads)
	}
}

func TestCheckCycleAtMaxDepth(t *testing.T) {
	dir := t.TempDir()
	for i := 1; i <= 8; i++ {
		target := fmt.Sprintf("./p%d.md", i+1)
		if i == 8 {
			target = "./p1.md"
		}
		panels := fmt.Sprintf(":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: %s\n:::\n", target)
		writePage(t, dir, fmt.Sprintf("p%d.md", i), pageBody(fmt.Sprintf("p%d", i), "", panels))
	}
	_, err := Check(filepath.Join(dir, "p1.md"))
	if err == nil || !strings.Contains(err.Error(), "循环页面引用") || strings.Contains(err.Error(), "超过 8 层") {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckSeenAtMaxDepthStillRejects(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "p9.md", pageBody("p9", "", ""))
	rootPanels := ":::mdtui panel\nid: a\nformat: items-json\nexec: [echo]\nopen: ./p9.md\n:::\n:::mdtui panel\nid: b\nformat: items-json\nexec: [echo]\nopen: ./p2.md\n:::\n"
	writePage(t, dir, "p1.md", pageBody("p1", "", rootPanels))
	for i := 2; i <= 8; i++ {
		target := fmt.Sprintf("./p%d.md", i+1)
		if i == 8 {
			target = "./p9.md"
		}
		panels := fmt.Sprintf(":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: %s\n:::\n", target)
		writePage(t, dir, fmt.Sprintf("p%d.md", i), pageBody(fmt.Sprintf("p%d", i), "", panels))
	}
	_, err := Check(filepath.Join(dir, "p1.md"))
	if err == nil || !strings.Contains(err.Error(), "超过 8 层") {
		t.Fatalf("err = %v", err)
	}
}

func TestResolveCanonicalizesSymlink(t *testing.T) {
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

func TestResolveSymlinkCycle(t *testing.T) {
	dir := t.TempDir()
	page := &Page{File: filepath.Join(dir, "root.md"), Dir: dir}
	if err := os.Symlink("loop.md", filepath.Join(dir, "loop.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Resolve(page, "./loop.md"); err == nil {
		t.Fatal("accepted symlink cycle")
	}
}
