// Package document implements the non-executing MDTUI v1 page compiler.
package document

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

type Arg struct{ Literal, Ref string }
type Panel struct {
	ID, Title, Mode, Format, Open string
	OpenByID                      map[string]string
	OnFocus, OnSelect             string
	Exec                          []Arg
	Refresh, Timeout              time.Duration
	Autostart                     bool
}

func (p Panel) HasOpen() bool { return p.Open != "" || len(p.OpenByID) > 0 }

func (p Panel) OpenTarget(id string) (string, bool) {
	if len(p.OpenByID) > 0 {
		target, ok := p.OpenByID[id]
		return target, ok
	}
	return p.Open, p.Open != ""
}

func (p Panel) OpenTargets() []string {
	if len(p.OpenByID) == 0 {
		if p.Open == "" {
			return nil
		}
		return []string{p.Open}
	}
	keys := make([]string, 0, len(p.OpenByID))
	for key := range p.OpenByID {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	targets := make([]string, 0, len(keys))
	for _, key := range keys {
		targets = append(targets, p.OpenByID[key])
	}
	return targets
}

type Input struct {
	ID, Label, Default string
	Required           bool
	MaxLength          int
}
type Action struct {
	ID, Label, Key, Run, Confirm string
	Refresh                      []string
}
type Page struct {
	File, Dir, ID, Title, Layout, Sidebar, Parent, Markdown string
	MaxParallel                                             int
	Panels                                                  []Panel
	Inputs                                                  []Input
	Actions                                                 []Action
}
type block struct {
	kind   string
	line   int
	fields map[string]*yaml.Node
}
type compiler struct {
	file   string
	blocks []block
	ids    map[string]bool
}

var identifier = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}$`)
var fence = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})(.*)$")

func Load(path string) (*Page, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, (1<<20)+1))
	if e != nil {
		return nil, e
	}
	return Parse(path, b)
}
func Parse(file string, data []byte) (p *Page, err error) {
	// All compiler errors carry source position; panics are private typed errors only.
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(compileError); ok {
				p = nil
				err = e
			} else {
				panic(r)
			}
		}
	}()
	c := compiler{file: file, ids: map[string]bool{}}
	if len(data) > 1<<20 {
		c.fail(1, "file", "文件超过 1 MiB，请拆分页面")
	}
	if !utf8.Valid(data) {
		c.fail(1, "file", "需要 UTF-8 编码")
	}
	p = &Page{File: file, Dir: filepath.Dir(file), Layout: "stack", MaxParallel: 2}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	var prose []string
	var marker byte
	var count int
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if m := fence.FindStringSubmatch(line); m != nil {
			if marker == 0 {
				marker = m[1][0]
				count = len(m[1])
			} else if m[1][0] == marker && len(m[1]) >= count && strings.TrimSpace(m[2]) == "" {
				marker = 0
			}
			prose = append(prose, line)
			continue
		}
		if marker != 0 || !strings.HasPrefix(line, ":::mdtui") {
			prose = append(prose, line)
			continue
		}
		if !strings.HasPrefix(line, ":::mdtui ") {
			c.fail(i+1, "block", "使用 :::mdtui TYPE")
		}
		kind := strings.TrimPrefix(line, ":::mdtui ")
		start := i + 1
		var body []string
		for i++; i < len(lines) && lines[i] != ":::"; i++ {
			if strings.HasPrefix(lines[i], ":::mdtui") {
				c.fail(i+1, "block", "扩展不允许嵌套")
			}
			body = append(body, lines[i])
		}
		if i == len(lines) {
			c.fail(start, "block", "缺少独占一行的 ::: 结束标记")
		}
		var n yaml.Node
		dec := yaml.NewDecoder(strings.NewReader(strings.Join(body, "\n")))
		if e := dec.Decode(&n); e != nil {
			c.fail(start+1, "yaml", e.Error()+"；请修正 YAML")
		}
		var extra yaml.Node
		if e := dec.Decode(&extra); e != io.EOF {
			c.fail(start+1, "yaml", "只允许一个 YAML 映射")
		}
		if len(n.Content) != 1 || n.Content[0].Kind != yaml.MappingNode {
			c.fail(start+1, "yaml", "块内容必须为映射")
		}
		c.inspect(n.Content[0], 0, start)
		b := block{kind: kind, line: start, fields: map[string]*yaml.Node{}}
		for j := 0; j < len(n.Content[0].Content); j += 2 {
			b.fields[n.Content[0].Content[j].Value] = n.Content[0].Content[j+1]
		}
		c.blocks = append(c.blocks, b)
	}
	p.Markdown = strings.TrimSpace(strings.Join(prose, "\n"))
	if len(c.blocks) == 0 || c.blocks[0].kind != "page" {
		c.fail(1, "page", "请在其他声明之前添加一个 page 块")
	}
	for index, b := range c.blocks {
		id := c.str(b, "id", "", true)
		if !identifier.MatchString(id) {
			c.fail(b.line, "id", "使用字母开头的 1～64 位 ASCII 标识")
		}
		if c.ids[id] {
			c.fail(b.line, "id", "重复 ID，请改名: "+id)
		}
		c.ids[id] = true
		switch b.kind {
		case "page":
			c.allowed(b, "version id title layout sidebar parent max_parallel")
			if index != 0 {
				c.fail(b.line, "page", "每页只能有一个 page 块")
			}
			if c.integer(b, "version", 0) != 1 {
				c.fail(b.line, "version", "仅支持整数 1")
			}
			p.ID = id
			p.Title = c.str(b, "title", id, false)
			p.Layout = c.str(b, "layout", "stack", false)
			p.Sidebar = c.str(b, "sidebar", "", false)
			p.Parent = c.str(b, "parent", "", false)
			if p.Parent != "" && !identifier.MatchString(p.Parent) {
				c.fail(b.line, "parent", "使用字母开头的 1～64 位 ASCII 标识")
			}
			p.MaxParallel = c.integer(b, "max_parallel", 2)
			if p.Layout != "stack" && p.Layout != "master-detail" && p.Layout != "workspace" {
				c.fail(b.line, "layout", "使用 stack、master-detail 或 workspace")
			}
			if p.MaxParallel < 1 || p.MaxParallel > 8 {
				c.fail(b.line, "max_parallel", "取值范围 1～8")
			}
		case "panel":
			c.allowed(b, "id title mode format exec refresh timeout autostart open on_focus on_select")
			v := Panel{ID: id, Title: c.str(b, "title", id, false), Mode: c.str(b, "mode", "snapshot", false), Format: c.str(b, "format", "markdown", false), OnFocus: c.str(b, "on_focus", "", false), OnSelect: c.str(b, "on_select", "", false), Autostart: c.boolean(b, "autostart", true)}
			v.Open, v.OpenByID = c.openTargets(b)
			if v.Mode != "snapshot" && v.Mode != "stream" {
				c.fail(b.line, "mode", "使用 snapshot 或 stream")
			}
			if v.Format != "markdown" && v.Format != "text" && v.Format != "items-json" {
				c.fail(b.line, "format", "使用 markdown、text 或 items-json")
			}
			if v.Mode == "stream" && v.Format == "items-json" {
				c.fail(b.line, "format", "items-json 只支持 snapshot")
			}
			if v.HasOpen() {
				if v.Format != "items-json" {
					c.fail(b.line, "open", "仅 items-json 面板可跳转子页")
				}
				for _, target := range v.OpenTargets() {
					if filepath.IsAbs(target) {
						c.fail(b.line, "open", "使用相对当前页面的路径")
					}
				}
			}
			d := "5s"
			if v.Mode == "stream" {
				d = "0"
			}
			v.Timeout = c.duration(b, "timeout", d)
			v.Refresh = c.duration(b, "refresh", "0")
			if v.Timeout < 0 || (v.Timeout == 0 && v.Mode != "stream") {
				c.fail(b.line, "timeout", "snapshot 必须设置正超时")
			}
			if _, ok := b.fields["refresh"]; ok && (v.Mode != "snapshot" || v.Refresh < 250*time.Millisecond) {
				c.fail(b.line, "refresh", "仅 snapshot 可用，最小 250ms")
			}
			n, ok := b.fields["exec"]
			if !ok || n.Kind != yaml.SequenceNode || len(n.Content) == 0 {
				c.fail(b.line, "exec", "提供非空参数数组")
			}
			for j, a := range n.Content {
				if a.Kind == yaml.ScalarNode && a.Tag == "!!str" {
					v.Exec = append(v.Exec, Arg{Literal: a.Value})
				} else if j > 0 && a.Kind == yaml.MappingNode && len(a.Content) == 2 && a.Content[0].Value == "ref" && a.Content[1].Tag == "!!str" {
					v.Exec = append(v.Exec, Arg{Ref: a.Content[1].Value})
					if a.Content[1].Value == "" {
						c.fail(b.line+a.Line, "exec.ref", "引用不能为空")
					}
				} else {
					c.fail(b.line+a.Line, "exec", "参数必须是字符串或单独 ref 映射，程序必须是字符串")
				}
			}
			if v.Exec[0].Literal == "" {
				c.fail(b.line, "exec[0]", "程序不能为空")
			}
			p.Panels = append(p.Panels, v)
		case "input":
			c.allowed(b, "id label default required max_length")
			v := Input{ID: id, Label: c.str(b, "label", id, false), Default: c.str(b, "default", "", false), Required: c.boolean(b, "required", false), MaxLength: c.integer(b, "max_length", 256)}
			if v.MaxLength < 1 || v.MaxLength > 4096 || utf8.RuneCountInString(v.Default) > v.MaxLength || strings.ContainsAny(v.Default, "\r\n") {
				c.fail(b.line, "input", "单行默认值不得超出 1～4096 的 max_length")
			}
			p.Inputs = append(p.Inputs, v)
		case "action":
			c.allowed(b, "id label key run refresh confirm")
			v := Action{ID: id, Label: c.str(b, "label", id, false), Key: c.str(b, "key", "", false), Run: c.str(b, "run", "", false), Confirm: c.str(b, "confirm", "", false)}
			if _, ok := b.fields["confirm"]; ok && strings.TrimSpace(v.Confirm) == "" {
				c.fail(b.line, "confirm", "确认文案不能为空")
			}
			if n, ok := b.fields["refresh"]; ok {
				if n.Kind != yaml.SequenceNode || len(n.Content) == 0 {
					c.fail(b.line, "refresh", "提供非空面板 ID 数组")
				}
				for _, x := range n.Content {
					if x.Tag != "!!str" {
						c.fail(b.line+x.Line, "refresh", "面板 ID 必须是字符串")
					}
					v.Refresh = append(v.Refresh, x.Value)
				}
			}
			if (v.Run == "") == (len(v.Refresh) == 0) {
				c.fail(b.line, "action", "只填写 run 或 refresh 中的一种")
			}
			if _, a := b.fields["run"]; a {
				if _, z := b.fields["refresh"]; z {
					c.fail(b.line, "action", "run 和 refresh 互斥")
				}
			}
			if v.Key != "" && (len(v.Key) != 1 || v.Key[0] < 'a' || v.Key[0] > 'z' || v.Key == "q") {
				c.fail(b.line, "key", "使用除 q 外的一个小写 ASCII 字母")
			}
			p.Actions = append(p.Actions, v)
		default:
			c.fail(b.line, "type", "未知块类型，请使用 page、panel、input 或 action")
		}
	}
	if len(p.Panels) > 16 || len(p.Inputs) > 16 || len(p.Actions) > 32 {
		c.fail(1, "page", "最多 16 面板、16 输入和 32 动作")
	}
	panels := map[string]Panel{}
	inputs := map[string]bool{}
	for _, v := range p.Panels {
		panels[v.ID] = v
	}
	for _, v := range p.Inputs {
		inputs[v.ID] = true
	}
	lineOf := func(id string) int {
		for _, b := range c.blocks {
			if n := b.fields["id"]; n != nil && n.Value == id {
				return b.line
			}
		}
		return 1
	}
	if p.Layout == "master-detail" || p.Layout == "workspace" {
		if _, ok := panels[p.Sidebar]; !ok {
			c.fail(lineOf(p.ID), "sidebar", "主从布局需要有效面板 ID")
		}
	}
	graph := map[string][]string{}
	for _, v := range p.Panels {
		for _, a := range v.Exec {
			if a.Ref == "" {
				if strings.ContainsRune(a.Literal, 0) {
					c.fail(lineOf(v.ID), "exec", "参数不得包含 NUL")
				}
				continue
			}
			parts := strings.Split(a.Ref, ".")
			valid := false
			if p.Parent != "" && len(parts) == 3 && parts[0] == "parent" && parts[1] == "selected" && parts[2] == "id" {
				valid = true
			} else {
				if len(parts) == 2 && parts[1] == "value" && inputs[parts[0]] {
					valid = true
				}
				if len(parts) == 3 && (parts[1] == "selected" || parts[1] == "cursor") && parts[2] == "id" {
					if x, ok := panels[parts[0]]; ok && x.Format == "items-json" {
						valid = true
						graph[v.ID] = append(graph[v.ID], x.ID)
					}
				}
			}
			if !valid {
				c.fail(lineOf(v.ID), "exec.ref", "无效引用 "+a.Ref+"；使用列表.selected.id、列表.cursor.id、输入.value 或 parent.selected.id")
			}
		}
	}
	visited := map[string]int{}
	var visit func(string)
	visit = func(id string) {
		if visited[id] == 1 {
			c.fail(lineOf(id), "exec.ref", "依赖成环，请移除自引用或循环")
		}
		if visited[id] == 2 {
			return
		}
		visited[id] = 1
		for _, dep := range graph[id] {
			visit(dep)
		}
		visited[id] = 2
	}
	for id := range panels {
		visit(id)
	}
	keys := map[string]bool{}
	actions := map[string]Action{}
	for _, a := range p.Actions {
		actions[a.ID] = a
		if a.Key != "" {
			if keys[a.Key] {
				c.fail(lineOf(a.ID), "key", "重复快捷键，请更换")
			}
			keys[a.Key] = true
		}
		targets := a.Refresh
		if a.Run != "" {
			targets = []string{a.Run}
		}
		for _, id := range targets {
			v, ok := panels[id]
			if !ok || (a.Run == "" && v.Mode != "snapshot") {
				c.fail(lineOf(a.ID), "action.target", "目标必须存在，refresh 只支持 snapshot")
			}
		}
	}
	for _, panel := range p.Panels {
		check := func(field, id string) {
			if id == "" {
				return
			}
			action, ok := actions[id]
			if !ok {
				c.fail(lineOf(panel.ID), field, "动作不存在: "+id)
			}
			if action.Confirm != "" {
				c.fail(lineOf(panel.ID), field, "自动触发动作不能设置 confirm")
			}
		}
		check("on_focus", panel.OnFocus)
		check("on_select", panel.OnSelect)
		if panel.OnFocus != "" && panel.Format != "items-json" {
			c.fail(lineOf(panel.ID), "on_focus", "仅 items-json 面板可绑定焦点事件")
		}
		if panel.OnSelect != "" && panel.Format != "items-json" {
			c.fail(lineOf(panel.ID), "on_select", "仅 items-json 面板可绑定选中事件")
		}
		if panel.OnSelect != "" && panel.HasOpen() {
			c.fail(lineOf(panel.ID), "on_select", "on_select 与 open 不能同时使用")
		}
	}
	return p, nil
}

type compileError string

func (e compileError) Error() string { return string(e) }
func (c *compiler) fail(line int, path, msg string) {
	panic(compileError(fmt.Sprintf("%s:%d: %s: %s", c.file, line, path, msg)))
}
func (c *compiler) inspect(n *yaml.Node, depth, offset int) {
	if depth > 8 {
		c.fail(offset+n.Line, "yaml", "嵌套深度不得超过 8")
	}
	if n.Kind == yaml.AliasNode || n.Tag == "!!merge" || (!strings.HasPrefix(n.Tag, "!!") && n.Tag != "") {
		c.fail(offset+n.Line, "yaml", "禁止自定义标签、别名与合并键")
	}
	if n.Kind == yaml.MappingNode {
		seen := map[string]bool{}
		for i := 0; i < len(n.Content); i += 2 {
			k := n.Content[i]
			if k.Tag != "!!str" || k.Kind != yaml.ScalarNode || k.Value == "<<" || seen[k.Value] {
				c.fail(offset+k.Line, "yaml", "键必须为唯一字符串，禁止合并键")
			}
			seen[k.Value] = true
		}
	}
	for _, v := range n.Content {
		c.inspect(v, depth+1, offset)
	}
}
func (c *compiler) allowed(b block, keys string) {
	for k, n := range b.fields {
		if !strings.Contains(" "+keys+" ", " "+k+" ") {
			c.fail(b.line+n.Line, k, "未知字段，请检查拼写")
		}
	}
}
func (c *compiler) openTargets(b block) (string, map[string]string) {
	n, ok := b.fields["open"]
	if !ok {
		return "", nil
	}
	switch {
	case n.Kind == yaml.ScalarNode && n.Tag == "!!str":
		ref := strings.TrimSpace(n.Value)
		if ref == "" {
			c.fail(b.line+n.Line, "open", "路径不能为空")
		}
		return ref, nil
	case n.Kind == yaml.MappingNode:
		targets := map[string]string{}
		for i := 0; i < len(n.Content); i += 2 {
			key, value := n.Content[i], n.Content[i+1]
			if key.Tag != "!!str" || key.Kind != yaml.ScalarNode || strings.TrimSpace(key.Value) == "" {
				c.fail(b.line+key.Line, "open", "选项 ID 必须是非空字符串")
			}
			if value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
				c.fail(b.line+value.Line, "open."+key.Value, "子页路径必须是字符串")
			}
			ref := strings.TrimSpace(value.Value)
			if ref == "" {
				c.fail(b.line+value.Line, "open."+key.Value, "路径不能为空")
			}
			targets[key.Value] = ref
		}
		if len(targets) == 0 {
			c.fail(b.line, "open", "至少配置一个选项")
		}
		return "", targets
	default:
		c.fail(b.line, "open", "open 必须是路径字符串或 ID 到路径的映射")
	}
	return "", nil
}

func (c *compiler) str(b block, k, def string, required bool) string {
	n, ok := b.fields[k]
	if !ok {
		if required {
			c.fail(b.line, k, "缺少必填字段")
		}
		return def
	}
	if n.Kind != yaml.ScalarNode || n.Tag != "!!str" {
		c.fail(b.line+n.Line, k, "必须为字符串，请给数值加引号")
	}
	return n.Value
}
func (c *compiler) integer(b block, k string, def int) int {
	n, ok := b.fields[k]
	if !ok {
		return def
	}
	if n.Tag != "!!int" {
		c.fail(b.line+n.Line, k, "必须为整数")
	}
	v, e := strconv.Atoi(n.Value)
	if e != nil {
		c.fail(b.line+n.Line, k, "无效整数")
	}
	return v
}
func (c *compiler) boolean(b block, k string, def bool) bool {
	n, ok := b.fields[k]
	if !ok {
		return def
	}
	if n.Tag != "!!bool" {
		c.fail(b.line+n.Line, k, "必须为布尔值")
	}
	return strings.EqualFold(n.Value, "true")
}
func (c *compiler) duration(b block, k, def string) time.Duration {
	n, ok := b.fields[k]
	s := def
	if ok {
		if n.Kind != yaml.ScalarNode || (n.Tag != "!!str" && !(n.Tag == "!!int" && n.Value == "0")) {
			c.fail(b.line+n.Line, k, "使用时长，例如 5s")
		}
		s = n.Value
	}
	v, e := time.ParseDuration(s)
	if e != nil {
		c.fail(b.line, k, "无效时长，请使用 5s、250ms 或 0")
	}
	return v
}
