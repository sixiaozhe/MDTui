package document

import (
	"fmt"
	"path/filepath"
	"strings"
)

const MaxDepth = 8

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
	rootFile, err := Canonical(path)
	if err != nil {
		return nil, err
	}
	root, err := Load(rootFile)
	if err != nil {
		return nil, err
	}
	pages := []*Page{root}
	seen := map[string]bool{rootFile: true}
	var walk func(current *Page, depth int, chain []string) error
	walk = func(current *Page, depth int, chain []string) error {
		if depth > MaxDepth {
			return fmt.Errorf("%s: 页面嵌套超过 %d 层", current.File, MaxDepth)
		}
		for _, panel := range current.Panels {
			if panel.Open == "" {
				continue
			}
			target, err := Resolve(current, panel.Open)
			if err != nil {
				return err
			}
			for _, file := range chain {
				if file == target {
					cycle := append(append([]string{}, chain...), target)
					return fmt.Errorf("循环页面引用: %s", strings.Join(cycle, " -> "))
				}
			}
			child, err := Load(target)
			if err != nil {
				return err
			}
			if child.Parent != "" && child.Parent != panel.ID {
				return fmt.Errorf("%s: parent %q 与引用面板 %q 不一致", child.File, child.Parent, panel.ID)
			}
			if seen[target] {
				continue
			}
			if len(pages) >= 256 {
				return fmt.Errorf("%s: 页面引用超过 256 个", current.File)
			}
			seen[target] = true
			pages = append(pages, child)
			if err := walk(child, depth+1, append(chain, target)); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(root, 1, []string{rootFile}); err != nil {
		return nil, err
	}
	return pages, nil
}
