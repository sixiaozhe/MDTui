package document

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	MaxDepth = 8
	maxPages = 256
)

// Canonical 返回解析符号链接后的绝对路径。
func Canonical(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// Resolve 把页面中的相对 open 路径解析为规范绝对路径。
func Resolve(page *Page, ref string) (string, error) {
	if filepath.IsAbs(ref) {
		return "", fmt.Errorf("%s: open: 使用相对路径", page.File)
	}
	return Canonical(filepath.Join(page.Dir, ref))
}

// Check 递归加载并校验页面树，返回按首次访问顺序排列的页面。
func Check(path string) ([]*Page, error) {
	return check(path, Load)
}

func check(path string, loadPage func(string) (*Page, error)) ([]*Page, error) {
	rootFile, err := Canonical(path)
	if err != nil {
		return nil, err
	}
	root, err := loadPage(rootFile)
	if err != nil {
		return nil, err
	}
	cache := map[string]*Page{rootFile: root}
	load := func(file string) (*Page, error) {
		if page, ok := cache[file]; ok {
			return page, nil
		}
		page, err := loadPage(file)
		if err != nil {
			return nil, err
		}
		cache[file] = page
		return page, nil
	}
	pages := []*Page{root}
	seen := map[string]bool{rootFile: true}
	var walk func(current *Page, depth int, chain []string) error
	walk = func(current *Page, depth int, chain []string) error {
		for _, panel := range current.Panels {
			if !panel.HasOpen() {
				continue
			}
			for _, ref := range panel.OpenTargets() {
				target, err := Resolve(current, ref)
				if err != nil {
					return fmt.Errorf("%s: panel %q: open %q: %w", current.File, panel.ID, ref, err)
				}
				for _, file := range chain {
					if file == target {
						cycle := append(append([]string{}, chain...), target)
						return fmt.Errorf("循环页面引用: %s", strings.Join(cycle, " -> "))
					}
				}
				if depth >= MaxDepth {
					return fmt.Errorf("%s: panel %q: open %q: 页面嵌套超过 %d 层", current.File, panel.ID, ref, MaxDepth)
				}
				if seen[target] {
					child := cache[target]
					if child.Parent != "" && child.Parent != panel.ID {
						return fmt.Errorf("%s: panel %q: open %q: %s: parent %q 不一致", current.File, panel.ID, ref, child.File, child.Parent)
					}
					continue
				}
				if len(pages) >= maxPages {
					return fmt.Errorf("%s: panel %q: open %q: 页面引用超过 %d 个", current.File, panel.ID, ref, maxPages)
				}
				child, err := load(target)
				if err != nil {
					return fmt.Errorf("%s: panel %q: open %q: %w", current.File, panel.ID, ref, err)
				}
				if child.Parent != "" && child.Parent != panel.ID {
					return fmt.Errorf("%s: panel %q: open %q: %s: parent %q 不一致", current.File, panel.ID, ref, child.File, child.Parent)
				}
				seen[target] = true
				pages = append(pages, child)
				if err := walk(child, depth+1, append(chain[:len(chain):len(chain)], target)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(root, 1, []string{rootFile}); err != nil {
		return nil, err
	}
	return pages, nil
}
