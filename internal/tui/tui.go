package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
	"mdtui/internal/content"
	"mdtui/internal/document"
	"mdtui/internal/scheduler"
)

type UI struct {
	Screen                   tcell.Screen
	Engine                   *scheduler.Engine
	NoColor                  bool
	focus                    int
	notice, modal, draft     string
	editing                  bool
	draftCursor, inputX      int
	searching                bool
	searchQuery              string
	searchDraft              string
	searchIndex, searchFocus int
	menu                     int
	confirm                  bool
	action                   *document.Action
	proseScroll, proseX      int
	modalScroll, modalX      int
	cache                    map[string]cached
	proseLines               []string
	panelLines               map[string][]string
	zoom                     bool
	workspaceDetail          string
	sessions                 []session
	opened                   map[string]string
}

type session struct {
	Engine          *scheduler.Engine
	File            string
	Focus           int
	Zoom            bool
	WorkspaceDetail string
	ProseScroll     int
	ProseX          int
	SearchQuery     string
	SearchIndex     int
	SearchFocus     int
	Opened          map[string]string
}

type cached struct {
	source   string
	width    int
	rendered bool
	lines    []string
}

func New(s tcell.Screen, e *scheduler.Engine, noColor bool) *UI {
	focus := 1 + len(e.Page.Inputs)
	if len(e.Page.Panels) == 0 {
		focus = 0
		if len(e.Page.Inputs) > 0 {
			focus = 1
		}
	}
	detail := ""
	for _, panel := range e.Page.Panels {
		if panel.ID != e.Page.Sidebar {
			detail = panel.ID
			break
		}
	}
	return &UI{Screen: s, Engine: e, NoColor: noColor, focus: focus, workspaceDetail: detail, searchIndex: -1, searchFocus: -1, cache: map[string]cached{}, panelLines: map[string][]string{}, opened: map[string]string{}}
}

func (u *UI) depth() int { return len(u.sessions) + 1 }

func (u *UI) saveSession() session {
	file, err := document.Canonical(u.Engine.Page.File)
	if err != nil {
		file = u.Engine.Page.File
	}
	return session{
		Engine:          u.Engine,
		File:            file,
		Focus:           u.focus,
		Zoom:            u.zoom,
		WorkspaceDetail: u.workspaceDetail,
		ProseScroll:     u.proseScroll,
		ProseX:          u.proseX,
		SearchQuery:     u.searchQuery,
		SearchIndex:     u.searchIndex,
		SearchFocus:     u.searchFocus,
		Opened:          u.opened,
	}
}

func (u *UI) restoreSession(s session) {
	u.Engine = s.Engine
	u.focus = s.Focus
	u.zoom = s.Zoom
	u.workspaceDetail = s.WorkspaceDetail
	u.proseScroll, u.proseX = s.ProseScroll, s.ProseX
	u.searchQuery, u.searchIndex, u.searchFocus = s.SearchQuery, s.SearchIndex, s.SearchFocus
	u.opened = s.Opened
}

func (u *UI) resetView(page *document.Page) {
	u.focus = 1 + len(page.Inputs)
	if len(page.Panels) == 0 {
		u.focus = 0
		if len(page.Inputs) > 0 {
			u.focus = 1
		}
	}
	u.workspaceDetail = ""
	for _, p := range page.Panels {
		if p.ID != page.Sidebar {
			u.workspaceDetail = p.ID
			break
		}
	}
	u.zoom = false
	u.proseScroll, u.proseX = 0, 0
	u.draft, u.draftCursor, u.inputX = "", 0, 0
	u.modalScroll, u.modalX = 0, 0
	u.proseLines = nil
	u.notice, u.modal, u.editing, u.searching = "", "", false, false
	u.searchQuery, u.searchDraft = "", ""
	u.searchIndex, u.searchFocus = -1, -1
	u.action, u.confirm, u.menu = nil, false, 0
	u.cache = map[string]cached{}
	u.panelLines = map[string][]string{}
}

func (u *UI) pushChild(panelID string) {
	p := u.Engine.Panels[panelID]
	if p == nil || p.Definition.Open == "" || p.Selected == "" {
		return
	}
	if u.opened[panelID] == p.Selected {
		return
	}
	if u.depth() >= document.MaxDepth {
		u.notice = fmt.Sprintf("页面嵌套超过 %d 层", document.MaxDepth)
		return
	}
	target, err := document.Resolve(u.Engine.Page, p.Definition.Open)
	if err != nil {
		u.notice = err.Error()
		return
	}
	current, err := document.Canonical(u.Engine.Page.File)
	if err == nil {
		if target == current {
			u.notice = "页面不能引用自身"
			return
		}
		for _, s := range u.sessions {
			if s.File == target {
				u.notice = "检测到循环页面引用"
				return
			}
		}
	}
	page, err := document.Load(target)
	if err != nil {
		u.notice = err.Error()
		return
	}
	if page.Parent != "" && page.Parent != panelID {
		u.notice = fmt.Sprintf("子页 parent %q 与引用面板 %q 不一致", page.Parent, panelID)
		return
	}
	selected := p.Selected
	parent := u.Engine
	u.opened[panelID] = selected
	parent.Pause()
	u.sessions = append(u.sessions, u.saveSession())
	child := scheduler.New(page, parent.Runner, parent.Now)
	child.Log = parent.Log
	child.ParentSelected = selected
	u.Engine = child
	u.opened = map[string]string{}
	u.resetView(page)
	child.Start()
}

func (u *UI) popChild() bool {
	if len(u.sessions) == 0 {
		return false
	}
	u.Engine.Close()
	last := u.sessions[len(u.sessions)-1]
	u.sessions = u.sessions[:len(u.sessions)-1]
	u.restoreSession(last)
	u.cache = map[string]cached{}
	u.panelLines = map[string][]string{}
	u.draft, u.draftCursor, u.inputX = "", 0, 0
	u.modalScroll, u.modalX = 0, 0
	u.proseLines = nil
	u.notice, u.modal, u.editing, u.searching, u.searchDraft = "", "", false, false, ""
	u.action, u.confirm, u.menu = nil, false, 0
	u.Engine.Resume()
	return true
}

func (u *UI) Close() {
	u.Engine.Close()
	for i := len(u.sessions) - 1; i >= 0; i-- {
		u.sessions[i].Engine.Close()
	}
}

func (u *UI) Run(ctx context.Context) error {
	defer u.Close()
	events := make(chan tcell.Event, 32)
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		defer close(events)
		for {
			ev := u.Screen.PollEvent()
			if ev == nil {
				return
			}
			select {
			case events <- ev:
			case <-stop:
				return
			}
		}
	}()
	u.Engine.Start()
	u.Draw()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-events:
			if !ok {
				return fmt.Errorf("终端输入已关闭")
			}
			switch v := ev.(type) {
			case *tcell.EventError:
				return v
			case *tcell.EventKey:
				if u.Key(v) {
					return nil
				}
			case *tcell.EventResize:
				u.Screen.Sync()
			}
			u.Draw()
		case ev := <-u.Engine.Events:
			u.Engine.Consume(ev)
			u.Draw()
		case <-tick.C:
			if u.Engine.Tick() {
				u.Draw()
			}
		}
	}
}
func (u *UI) count() int { return 1 + len(u.Engine.Page.Inputs) + len(u.Engine.Page.Panels) }
func (u *UI) panel() *scheduler.Panel {
	i := u.focus - 1 - len(u.Engine.Page.Inputs)
	if i < 0 || i >= len(u.Engine.Page.Panels) {
		return nil
	}
	return u.Engine.Panels[u.Engine.Page.Panels[i].ID]
}
func (u *UI) panelFocus(id string) int {
	for i, p := range u.Engine.Page.Panels {
		if p.ID == id {
			return 1 + len(u.Engine.Page.Inputs) + i
		}
	}
	return 0
}
func (u *UI) moveFocus(step int) {
	if !u.zoom {
		u.focus = (u.focus + step + u.count()) % u.count()
	} else {
		first := 1 + len(u.Engine.Page.Inputs)
		n := len(u.Engine.Page.Panels)
		if n == 0 {
			return
		}
		u.focus = first + (u.focus-first+step+n)%n
	}
	if p := u.panel(); p != nil && p.Definition.ID != u.Engine.Page.Sidebar {
		u.workspaceDetail = p.Definition.ID
	}
}
func (u *UI) selectWorkspaceView(source string) {
	if u.Engine.Page.Layout != "workspace" || source != u.Engine.Page.Sidebar {
		return
	}
	ref := source + ".selected.id"
	for _, panel := range u.Engine.Page.Panels {
		for _, arg := range panel.Exec {
			if arg.Ref == ref {
				u.workspaceDetail = panel.ID
				return
			}
		}
	}
}
func (u *UI) cycleWorkspaceView(step int) {
	if u.Engine.Page.Layout != "workspace" || u.zoom {
		return
	}
	var ids []string
	for _, panel := range u.Engine.Page.Panels {
		if panel.ID != u.Engine.Page.Sidebar {
			ids = append(ids, panel.ID)
		}
	}
	if len(ids) == 0 {
		return
	}
	index := 0
	for i, id := range ids {
		if id == u.workspaceDetail {
			index = i
			break
		}
	}
	u.workspaceDetail = ids[(index+step+len(ids))%len(ids)]
	u.focus = u.panelFocus(u.workspaceDetail)
}
func (u *UI) Key(k *tcell.EventKey) bool {
	if u.modal != "" {
		switch k.Key() {
		case tcell.KeyEscape:
			u.modal = ""
			u.modalScroll, u.modalX = 0, 0
			u.action = nil
			return false
		case tcell.KeyUp:
			if u.modal == "actions" {
				u.menu = max(0, u.menu-1)
			} else {
				u.modalScroll = max(0, u.modalScroll-1)
			}
		case tcell.KeyDown:
			if u.modal == "actions" {
				u.menu = min(len(u.Engine.Page.Actions)-1, u.menu+1)
			} else {
				u.modalScroll++
			}
		case tcell.KeyPgUp:
			u.modalScroll = max(0, u.modalScroll-8)
		case tcell.KeyPgDn:
			u.modalScroll += 8
		case tcell.KeyHome:
			u.modalScroll, u.modalX = 0, 0
		case tcell.KeyEnd:
			u.modalScroll = 1 << 30
		case tcell.KeyLeft, tcell.KeyRight, tcell.KeyTab:
			if u.modal == "confirm" {
				u.confirm = !u.confirm
			} else if k.Key() == tcell.KeyRight {
				u.modalX += 8
			} else if k.Key() == tcell.KeyLeft {
				u.modalX = max(0, u.modalX-8)
			}
		case tcell.KeyEnter:
			if u.modal == "actions" {
				if len(u.Engine.Page.Actions) > 0 {
					u.activate(u.Engine.Page.Actions[u.menu])
				}
			} else if u.modal == "confirm" {
				a := u.action
				u.modal = ""
				if u.confirm && a != nil {
					u.perform(*a)
				}
			} else {
				u.modal = ""
			}
		}
		return false
	}
	if u.searching {
		switch k.Key() {
		case tcell.KeyEscape:
			u.searching = false
		case tcell.KeyEnter:
			u.searchQuery = u.searchDraft
			u.searching = false
			u.searchIndex = -1
			u.searchFocus = u.focus
			if u.searchQuery != "" {
				u.searchNext(1, true)
			}
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			runes := []rune(u.searchDraft)
			if len(runes) > 0 {
				u.searchDraft = string(runes[:len(runes)-1])
			}
		case tcell.KeyRune:
			if !unicode.IsControl(k.Rune()) && utf8.RuneCountInString(u.searchDraft) < 256 {
				u.searchDraft += string(k.Rune())
			}
		}
		return false
	}
	if u.editing {
		input := u.Engine.Page.Inputs[u.focus-1]
		switch k.Key() {
		case tcell.KeyEscape:
			u.editing = false
		case tcell.KeyEnter:
			u.Engine.Submit(input.ID, u.draft)
			u.editing = false
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			previous := previousGrapheme(u.draft, u.draftCursor)
			u.draft = replaceRunes(u.draft, previous, u.draftCursor, "")
			u.draftCursor = previous
		case tcell.KeyDelete:
			u.draft = replaceRunes(u.draft, u.draftCursor, nextGrapheme(u.draft, u.draftCursor), "")
		case tcell.KeyLeft:
			u.draftCursor = previousGrapheme(u.draft, u.draftCursor)
		case tcell.KeyRight:
			u.draftCursor = nextGrapheme(u.draft, u.draftCursor)
		case tcell.KeyHome, tcell.KeyCtrlA:
			u.draftCursor = 0
		case tcell.KeyEnd, tcell.KeyCtrlE:
			u.draftCursor = utf8.RuneCountInString(u.draft)
		case tcell.KeyRune:
			if !unicode.IsControl(k.Rune()) && utf8.RuneCountInString(u.draft) < input.MaxLength {
				u.draft = replaceRunes(u.draft, u.draftCursor, u.draftCursor, string(k.Rune()))
				u.draftCursor++
			}
		}
		return false
	}
	switch k.Key() {
	case tcell.KeyTab:
		u.moveFocus(1)
	case tcell.KeyBacktab:
		u.moveFocus(-1)
	case tcell.KeyF2:
		u.modal = "actions"
		u.menu = 0
	case tcell.KeyEscape:
		if u.zoom {
			u.zoom = false
		} else {
			u.popChild()
		}
	case tcell.KeyEnter:
		if u.focus > 0 && u.focus <= len(u.Engine.Page.Inputs) {
			u.editing = true
			u.draft = u.Engine.Inputs[u.Engine.Page.Inputs[u.focus-1].ID]
			u.draftCursor = utf8.RuneCountInString(u.draft)
			u.inputX = 0
		} else if p := u.panel(); p != nil && p.Definition.Format == "items-json" && len(p.Items) > 0 {
			u.Engine.Select(p.Definition.ID, p.Items[p.Cursor].ID)
			u.selectWorkspaceView(p.Definition.ID)
		}
	case tcell.KeyCtrlC:
		if p := u.panel(); p != nil && (p.Status == "running" || p.Status == "queued") {
			u.Engine.Cancel(p.Definition.ID)
		} else {
			u.notice = "使用 q 退出工作台"
		}
	case tcell.KeyUp, tcell.KeyDown, tcell.KeyPgUp, tcell.KeyPgDn, tcell.KeyHome, tcell.KeyEnd, tcell.KeyLeft, tcell.KeyRight:
		u.scroll(k.Key())
	case tcell.KeyRune:
		if k.Rune() == 'Z' {
			if len(u.Engine.Page.Panels) > 0 {
				if u.panel() == nil {
					u.focus = u.panelFocus(u.Engine.Page.Panels[0].ID)
				}
				u.zoom = !u.zoom
			}
			return false
		}
		if k.Rune() == '[' || k.Rune() == ']' {
			step := 1
			if k.Rune() == '[' {
				step = -1
			}
			u.cycleWorkspaceView(step)
			return false
		}
		if k.Rune() == '/' {
			if u.focus > 0 && u.focus <= len(u.Engine.Page.Inputs) {
				u.notice = "请先聚焦说明区或面板，再按 / 搜索"
			} else {
				u.searching = true
				u.searchDraft = ""
				u.notice = ""
			}
			return false
		}
		if u.searchQuery != "" && (k.Rune() == 'n' || k.Rune() == 'N') {
			direction := 1
			if k.Rune() == 'N' {
				direction = -1
			}
			u.searchNext(direction, false)
			return false
		}
		if k.Rune() == 'q' {
			if u.popChild() {
				return false
			}
			return true
		}
		if k.Rune() == '?' {
			u.modal = "help"
			return false
		}
		if k.Rune() == ' ' {
			if p := u.panel(); p != nil && p.Definition.Mode == "stream" {
				p.Follow = !p.Follow
			}
			return false
		}
		if k.Rune() == '!' {
			u.modal = "errors"
			return false
		}
		for _, a := range u.Engine.Page.Actions {
			if a.Key == string(k.Rune()) {
				u.activate(a)
				break
			}
		}
	}
	return false
}
func (u *UI) activate(a document.Action) {
	u.modalScroll, u.modalX = 0, 0
	if reason := u.Engine.ActionReason(a); reason != "" {
		u.notice = reason
		u.modal = ""
		return
	}
	if a.Confirm != "" {
		u.action = &a
		u.confirm = false
		u.modal = "confirm"
	} else {
		u.modal = ""
		u.perform(a)
	}
}
func (u *UI) perform(a document.Action) {
	targets := scheduler.Targets(a)
	if u.Engine.Page.Layout == "workspace" && len(targets) == 1 && targets[0] != u.Engine.Page.Sidebar {
		u.workspaceDetail = targets[0]
	}
	for _, id := range targets {
		request := u.Engine.Request
		if a.Run != "" {
			request = u.Engine.Run
		}
		if s := request(id); s != "" {
			u.notice = s
		}
	}
}
func (u *UI) scroll(key tcell.Key) {
	p := u.panel()
	pos, x := &u.proseScroll, &u.proseX
	if p != nil {
		pos = &p.Scroll
		x = &p.X
		if p.Definition.Format == "items-json" {
			pos = &p.Cursor
		}
		if key != tcell.KeyLeft && key != tcell.KeyRight {
			p.Follow = false
		}
	}
	switch key {
	case tcell.KeyUp:
		*pos = max(0, *pos-1)
	case tcell.KeyDown:
		*pos++
	case tcell.KeyPgUp:
		*pos = max(0, *pos-8)
	case tcell.KeyPgDn:
		*pos += 8
	case tcell.KeyHome:
		*pos = 0
	case tcell.KeyEnd:
		*pos = 1 << 30
		if p != nil {
			p.Follow = true
		}
	case tcell.KeyLeft:
		*x = max(0, *x-4)
	case tcell.KeyRight:
		*x += 4
	}
	if p != nil && p.Definition.Format == "items-json" {
		p.Cursor = min(p.Cursor, max(0, len(p.Items)-1))
	}
}
func (u *UI) style(active bool) tcell.Style {
	s := tcell.StyleDefault
	if active {
		s = s.Bold(true)
		if !u.NoColor {
			s = s.Foreground(tcell.ColorLightSkyBlue)
		}
	}
	return s
}

func (u *UI) themed(fg, bg int) tcell.Style {
	s := tcell.StyleDefault
	if u.NoColor {
		return s
	}
	if fg >= 0 {
		s = s.Foreground(tcell.PaletteColor(fg))
	}
	if bg >= 0 {
		s = s.Background(tcell.PaletteColor(bg))
	}
	return s
}

func (u *UI) statusStyle(status string) tcell.Style {
	if u.NoColor {
		if status == "running" || status == "error" {
			return tcell.StyleDefault.Bold(true)
		}
		return tcell.StyleDefault
	}
	switch status {
	case "ready":
		return u.themed(78, 22).Bold(true)
	case "running", "queued":
		return u.themed(222, 58).Bold(true)
	case "error":
		return u.themed(210, 52).Bold(true)
	case "blocked":
		return u.themed(245, 236)
	case "cancelled":
		return u.themed(208, 52)
	default:
		return u.themed(250, 236)
	}
}

func statusLabel(status string) string {
	switch status {
	case "ready":
		return " ✓ 就绪 "
	case "running":
		return " ● 运行中 "
	case "queued":
		return " ◷ 排队 "
	case "error":
		return " ! 错误 "
	case "blocked":
		return " ◌ 等待参数 "
	case "cancelled":
		return " × 已取消 "
	default:
		return " ○ 空闲 "
	}
}

// put clips by grapheme cell width, never writing a wide glyph into its neighbor.
func (u *UI) put(x, y, width, offset int, s string, style tcell.Style) {
	w, h := u.Screen.Size()
	if y < 0 || y >= h || width <= 0 {
		return
	}
	end := min(w, x+width)
	col := 0
	g := uniseg.NewGraphemes(strings.ReplaceAll(content.Safe(s), "\t", "    "))
	for g.Next() {
		cluster := g.Str()
		cw := g.Width()
		if col >= offset && x+col-offset >= 0 && x+col-offset+cw <= end {
			rs := []rune(cluster)
			if len(rs) > 0 {
				u.Screen.SetContent(x+col-offset, y, rs[0], rs[1:], style)
			}
		}
		col += cw
		if x+col-offset >= end {
			break
		}
	}
}

// putRich renders the small SGR subset generated by content.Markdown. It does
// not write escape bytes to the terminal and ignores colors in --no-color mode.
func (u *UI) putRich(x, y, width, offset int, value string, base tcell.Style) {
	w, h := u.Screen.Size()
	if y < 0 || y >= h || width <= 0 {
		return
	}
	end := min(w, x+width)
	style := base
	column := 0
	for i := 0; i < len(value); {
		if value[i] == 0x1b && i+1 < len(value) && value[i+1] == '[' {
			finish := i + 2
			for finish < len(value) && value[finish] != 'm' {
				finish++
			}
			if finish < len(value) {
				style = u.applySGR(base, style, value[i+2:finish])
				i = finish + 1
				continue
			}
		}
		next := strings.IndexByte(value[i:], 0x1b)
		if next < 0 {
			next = len(value) - i
		}
		chunk := strings.ReplaceAll(value[i:i+next], "\t", "    ")
		g := uniseg.NewGraphemes(chunk)
		for g.Next() {
			cluster := g.Str()
			cw := g.Width()
			position := x + column - offset
			if column >= offset && position >= x && position+cw <= end {
				runes := []rune(cluster)
				if len(runes) > 0 {
					u.Screen.SetContent(position, y, runes[0], runes[1:], style)
				}
			}
			column += cw
			if x+column-offset >= end {
				return
			}
		}
		i += next
	}
}

func (u *UI) applySGR(base, current tcell.Style, encoded string) tcell.Style {
	if encoded == "" {
		encoded = "0"
	}
	parts := strings.Split(encoded, ";")
	for i := 0; i < len(parts); i++ {
		code, err := strconv.Atoi(parts[i])
		if err != nil {
			continue
		}
		switch code {
		case 0:
			current = base
		case 1:
			current = current.Bold(true)
		case 2:
			current = current.Dim(true)
		case 3:
			current = current.Italic(true)
		case 4:
			current = current.Underline(true)
		case 9:
			current = current.StrikeThrough(true)
		case 38, 48:
			if i+2 < len(parts) && parts[i+1] == "5" {
				index, parseErr := strconv.Atoi(parts[i+2])
				if parseErr == nil && !u.NoColor {
					if code == 38 {
						current = current.Foreground(tcell.PaletteColor(index))
					} else {
						current = current.Background(tcell.PaletteColor(index))
					}
				}
				i += 2
			}
		}
	}
	return current
}
func (u *UI) lines(id, source string, renderMarkdown bool, wrapWidth int) []string {
	c, ok := u.cache[id]
	if ok && c.source == source && c.width == wrapWidth && c.rendered == renderMarkdown {
		return c.lines
	}
	if source == "" {
		u.cache[id] = cached{source: source, width: wrapWidth, rendered: renderMarkdown}
		return nil
	}
	s := source
	if renderMarkdown {
		s = content.Markdown(source)
	}
	lines := strings.Split(s, "\n")
	if wrapWidth > 0 {
		lines = wrapRichLines(lines, wrapWidth)
	}
	u.cache[id] = cached{source: source, width: wrapWidth, rendered: renderMarkdown, lines: lines}
	return lines
}
func (u *UI) Draw() {
	s := u.Screen
	s.Clear()
	w, h := s.Size()
	if w < 60 || h < 16 {
		u.fill(0, 0, w, h, ' ', u.themed(252, 234))
		u.put(2, 2, w-4, 0, "◆ MDTUI", u.themed(81, -1).Bold(true))
		u.put(2, 4, w-4, 0, "终端空间不足", u.themed(210, -1).Bold(true))
		u.put(2, 6, w-4, 0, fmt.Sprintf("当前 %d×%d，需要至少 60×16", w, h), u.themed(245, -1))
		hint := "调整窗口尺寸，或按 q 安全退出"
		if len(u.sessions) > 0 {
			hint = "调整窗口尺寸，或按 q 返回上一层"
		}
		u.put(2, 8, w-4, 0, hint, u.themed(250, -1))
		s.Show()
		return
	}
	page := u.Engine.Page
	header := u.themed(231, 24).Bold(true)
	if u.NoColor {
		header = tcell.StyleDefault.Reverse(true).Bold(true)
	}
	u.fill(0, 0, w, 1, ' ', header)
	u.put(2, 0, w-4, 0, "◆ MDTUI   "+page.Title, header)
	pageID := "PAGE · " + page.ID
	u.put(max(2, w-uniseg.StringWidth(pageID)-2), 0, uniseg.StringWidth(pageID), 0, pageID, header.Dim(true))
	y := 2
	if u.zoom {
		y = 2
	} else {
		prose := u.lines("@prose", page.Markdown, true, w-5)
		u.proseLines = prose
		ph := min(len(prose), max(1, (h-4)/4))
		if strings.TrimSpace(page.Markdown) == "" {
			ph = 0
		}
		u.proseScroll = min(u.proseScroll, max(0, len(prose)-ph))
		if ph > 0 {
			position := fmt.Sprintf("说明 %d/%d", min(u.proseScroll+1, len(prose)), len(prose))
			u.put(max(2, w-uniseg.StringWidth(position)-2), 1, uniseg.StringWidth(position), 0, position, u.themed(245, -1))
		}
		for i := 0; i < ph; i++ {
			bar := u.themed(81, -1)
			if u.focus == 0 {
				bar = bar.Bold(true)
			}
			rail := "│"
			if u.focus == 0 && u.searchFocus == 0 && u.proseScroll+i == u.searchIndex {
				rail = "┃"
			}
			u.put(1, y+i, 1, 0, rail, bar)
			line := prose[u.proseScroll+i]
			offset := 0
			if richNoWrap(line) {
				offset = u.proseX
			}
			u.putRich(3, y+i, w-5, offset, line, tcell.StyleDefault)
			if u.focus == 0 {
				u.highlightMatches(3, y+i, w-5, offset, line)
			}
		}
		if ph > 0 {
			y += ph + 1
		}
		// Inputs occupy a compact focusable strip, preserving room for panel content.
		if len(page.Inputs) > 0 {
			idx := 0
			if u.focus > 0 && u.focus <= len(page.Inputs) {
				idx = u.focus - 1
			}
			input := page.Inputs[idx]
			value := u.Engine.Inputs[input.ID]
			active := u.focus == idx+1
			inputStyle := u.themed(252, 236)
			if active {
				inputStyle = u.themed(231, 24).Bold(true)
			}
			u.fill(1, y, w-2, 1, ' ', inputStyle)
			mode := "Enter 编辑"
			if u.editing {
				mode = "Enter 提交 · Esc 放弃"
			}
			label := fmt.Sprintf("⌘ 参数 %d/%d   %s   › ", idx+1, len(page.Inputs), input.Label)
			modeX := max(3, w-uniseg.StringWidth(mode)-3)
			valueX := min(modeX-10, 3+uniseg.StringWidth(label))
			u.put(3, y, max(0, valueX-3), 0, label, inputStyle)
			valueWidth := max(0, modeX-valueX-2)
			if u.editing && active {
				runes := []rune(u.draft)
				cursor := min(u.draftCursor, len(runes))
				cursorCell := uniseg.StringWidth(string(runes[:cursor]))
				if cursorCell < u.inputX {
					u.inputX = cursorCell
				}
				if cursorCell >= u.inputX+valueWidth {
					u.inputX = cursorCell - valueWidth + 1
				}
				value = string(runes[:cursor]) + "▏" + string(runes[cursor:])
			} else {
				u.inputX = 0
			}
			u.put(valueX, y, valueWidth, u.inputX, value, inputStyle)
			u.put(modeX, y, uniseg.StringWidth(mode), 0, mode, inputStyle.Dim(true))
			y += 2
		}
	}
	avail := h - y - 2
	n := len(page.Panels)
	if n > 0 {
		if u.zoom {
			if p := u.panel(); p != nil {
				u.drawPanel(p.Definition.ID, 0, y, w, avail, "  ·  放大视图")
			}
		} else if page.Layout == "workspace" && n > 1 {
			u.drawWorkspace(0, y, w, avail)
		} else if avail < 4*n {
			idx := max(0, u.focus-1-len(page.Inputs))
			idx = min(idx, n-1)
			u.drawPanel(page.Panels[idx].ID, 0, y, w, avail, fmt.Sprintf(" %d/%d", idx+1, n))
		} else if page.Layout == "master-detail" && w >= 100 && n > 1 {
			sw := max(24, w*30/100)
			u.drawPanel(page.Sidebar, 0, y, sw, avail, "")
			row := 0
			rightX := sw + 1
			rightW := w - rightX
			rightAvailable := avail - (n - 2)
			for _, p := range page.Panels {
				if p.ID == page.Sidebar {
					continue
				}
				height := rightAvailable / (n - 1)
				if row == n-2 {
					height = rightAvailable - (rightAvailable/(n-1))*row
				}
				u.drawPanel(p.ID, rightX, y+(rightAvailable/(n-1)+1)*row, rightW, height, "")
				row++
			}
		} else {
			panelAvailable := avail - (n - 1)
			for i, p := range page.Panels {
				height := panelAvailable / n
				if i == n-1 {
					height = panelAvailable - (panelAvailable/n)*i
				}
				u.drawPanel(p.ID, 0, y+(panelAvailable/n+1)*i, w, height, "")
			}
		}
	}
	context := u.focusContext()
	if u.notice != "" {
		context = "◆ " + u.notice
	}
	if u.searching {
		context = "/ " + u.searchDraft + "▏    Enter 查找 · Esc 取消"
	}
	u.put(2, h-2, w-4, 0, context, u.themed(222, -1).Bold(true))
	footer := u.themed(250, 235)
	if u.NoColor {
		footer = tcell.StyleDefault.Reverse(true)
	}
	u.fill(0, h-1, w, 1, ' ', footer)
	shortcuts := "Tab 焦点   [ ] 视图   / 搜索   Enter 确认   F2 动作   Z 放大   ? 帮助   q 退出"
	if len(u.sessions) > 0 {
		shortcuts = fmt.Sprintf("子页 %d/%d   Esc/q 返回   Tab 焦点   [ ] 视图   Enter 确认   F2 动作   Z 放大", u.depth(), document.MaxDepth)
	}
	if w < 90 {
		shortcuts = "Tab 焦点   Enter 确认   F2 动作   Z 放大   ? 帮助   q 退出"
		if len(u.sessions) > 0 {
			shortcuts = fmt.Sprintf("子页 %d/%d   Esc/q 返回   Tab 焦点   Enter 确认", u.depth(), document.MaxDepth)
		}
	}
	u.put(2, h-1, w-4, 0, shortcuts, footer)
	if u.modal != "" {
		u.drawModal(w, h)
	}
	s.Show()
}
func (u *UI) drawWorkspace(x, y, w, h int) {
	page := u.Engine.Page
	if u.workspaceDetail == "" || u.workspaceDetail == page.Sidebar {
		for _, p := range page.Panels {
			if p.ID != page.Sidebar {
				u.workspaceDetail = p.ID
				break
			}
		}
	}
	leftW, rightX := 0, x
	if w >= 90 {
		leftW = max(26, w*28/100)
		rightX = x + leftW + 1
	}
	rightW := w - leftW
	if leftW > 0 {
		rightW--
		u.drawPanel(page.Sidebar, x, y, leftW, h, "")
	}
	if leftW == 0 {
		if p := u.panel(); p != nil && p.Definition.ID == page.Sidebar {
			u.drawPanel(page.Sidebar, x, y, w, h, "")
			return
		}
	}
	var tabs []string
	for i, p := range page.Panels {
		if p.ID == page.Sidebar {
			continue
		}
		label := fmt.Sprintf(" %d %s ", i+1, p.Title)
		if p.ID == u.workspaceDetail {
			label = "▸" + label
		}
		tabs = append(tabs, label)
	}
	u.put(rightX+1, y, rightW-2, 0, strings.Join(tabs, "  "), u.themed(81, -1).Bold(true))
	u.drawPanel(u.workspaceDetail, rightX, y+1, rightW, h-1, "")
}
func (u *UI) drawPanel(id string, x, y, w, h int, index string) {
	if h < 3 || w < 8 {
		return
	}
	p := u.Engine.Panels[id]
	active := u.panel() == p
	border := u.themed(240, -1)
	if active {
		border = u.themed(81, -1).Bold(true)
	}
	u.put(x, y, w, 0, "╭"+strings.Repeat("─", w-2)+"╮", border)
	panelTitle := " " + p.Definition.Title + index + " "
	panelTitleWidth := min(w-4, uniseg.StringWidth(panelTitle))
	u.fill(x+2, y, panelTitleWidth, 1, ' ', border)
	u.put(x+2, y, panelTitleWidth, 0, panelTitle, border.Bold(true))
	badge := statusLabel(p.Status)
	if p.Stale {
		badge = " ⚠ 过期 "
	}
	badgeWidth := uniseg.StringWidth(badge)
	if badgeWidth+4 < w {
		badgeStyle := u.statusStyle(p.Status)
		u.fill(x+w-badgeWidth-2, y, badgeWidth, 1, ' ', badgeStyle)
		u.put(x+w-badgeWidth-2, y, badgeWidth, 0, badge, badgeStyle)
	}
	for row := y + 1; row < y+h-1; row++ {
		u.put(x, row, 1, 0, "│", border)
		u.put(x+w-1, row, 1, 0, "│", border)
	}
	// Snapshot Markdown is rendered off the UI loop. Streams are re-rendered
	// from their bounded window at the 100 ms scheduler cadence.
	wrapWidth := 0
	if p.Definition.Format == "markdown" {
		wrapWidth = w - 4
	}
	lines := u.lines(id, p.Text, p.Definition.Format == "markdown" && p.Definition.Mode == "stream", wrapWidth)
	rows := h - 2
	if p.Definition.Format == "items-json" {
		lines = nil
		for i, it := range p.Items {
			prefix := "  "
			if it.ID == p.Selected {
				prefix = "✓ "
			}
			if i == p.Cursor {
				prefix = ">" + prefix
			}
			lines = append(lines, prefix+it.Label+"  "+it.Description)
		}
		if p.Cursor < p.Scroll {
			p.Scroll = p.Cursor
		}
		if p.Cursor >= p.Scroll+rows {
			p.Scroll = p.Cursor - rows + 1
		}
	}
	u.panelLines[id] = lines
	if p.Definition.Mode == "stream" && p.Follow {
		p.Scroll = max(0, len(lines)-rows)
	}
	p.Scroll = min(p.Scroll, max(0, len(lines)-rows))
	if len(lines) > rows && rows > 0 {
		thumb := y + 1 + min(rows-1, p.Scroll*rows/len(lines))
		u.put(x+w-1, thumb, 1, 0, "┃", u.themed(81, -1).Bold(true))
	}
	for i := 0; i < rows && p.Scroll+i < len(lines); i++ {
		if active && u.searchFocus == u.focus && p.Scroll+i == u.searchIndex {
			u.put(x, y+1+i, 1, 0, "┃", u.themed(81, -1).Bold(true))
		}
		rowStyle := tcell.StyleDefault
		if p.Definition.Format == "items-json" {
			itemIndex := p.Scroll + i
			if itemIndex == p.Cursor {
				rowStyle = u.themed(231, 24).Bold(true)
				u.fill(x+1, y+1+i, w-2, 1, ' ', rowStyle)
			} else if itemIndex < len(p.Items) && p.Items[itemIndex].ID == p.Selected {
				rowStyle = u.themed(78, -1).Bold(true)
			}
		}
		if p.Definition.Format == "markdown" {
			line := lines[p.Scroll+i]
			offset := 0
			if richNoWrap(line) {
				offset = p.X
			}
			u.putRich(x+2, y+1+i, w-4, offset, line, rowStyle)
			if active {
				u.highlightMatches(x+2, y+1+i, w-4, offset, line)
			}
		} else {
			u.put(x+2, y+1+i, w-4, p.X, lines[p.Scroll+i], rowStyle)
			if active {
				u.highlightMatches(x+2, y+1+i, w-4, p.X, lines[p.Scroll+i])
			}
		}
	}
	foot := p.Reason
	if foot == "" {
		if !p.Updated.IsZero() {
			foot = "更新 " + p.Updated.Format("15:04:05")
		}
		if p.Selected != "" {
			foot += "  ◆ " + p.Selected
		}
		if p.Dropped+p.Truncated > 0 {
			foot += fmt.Sprintf(" 丢弃%d行 / 截断%d行", p.Dropped, p.Truncated)
		}
	}
	if p.Definition.Mode == "stream" {
		if p.Follow {
			foot += "  ↓ 跟随"
		} else {
			foot += "  ‖ 已暂停跟随"
		}
	}
	u.put(x, y+h-1, w, 0, "╰"+strings.Repeat("─", w-2)+"╯", border)
	if foot != "" {
		u.put(x+2, y+h-1, w-16, 0, " "+foot+" ", u.themed(245, -1))
	}
	position := 0
	if len(lines) > 0 {
		position = min(p.Scroll+1, len(lines))
		if p.Definition.Format == "items-json" {
			position = min(p.Cursor+1, len(lines))
		}
	}
	location := fmt.Sprintf(" %d/%d ", position, len(lines))
	locationWidth := uniseg.StringWidth(location)
	if locationWidth+4 < w {
		u.fill(x+w-locationWidth-2, y+h-1, locationWidth, 1, ' ', border)
		u.put(x+w-locationWidth-2, y+h-1, locationWidth, 0, location, border)
	}
}

func (u *UI) fill(x, y, width, height int, ch rune, style tcell.Style) {
	screenWidth, screenHeight := u.Screen.Size()
	for row := max(0, y); row < min(screenHeight, y+height); row++ {
		for column := max(0, x); column < min(screenWidth, x+width); column++ {
			u.Screen.SetContent(column, row, ch, nil, style)
		}
	}
}
func (u *UI) drawModal(w, h int) {
	title := ""
	var lines []string
	switch u.modal {
	case "help":
		title = "? 快捷键帮助"
		lines = []string{"Tab / Shift+Tab   切换说明、输入和面板", "[ / ]             切换工作区详情视图", "↑↓ / PgUp PgDn    移动、翻页（列表需 Enter 确认）", "←→ / Home End     横向查看代码和宽表、跳到首尾", "/ / n / N         搜索当前区域、下一个、上一个匹配", "输入：←→ Home End 移动；Backspace / Del 删除", "Enter / Esc       编辑、提交或放弃输入", "F2 / !            动作列表、当前面板错误详情", "Z / Esc          放大当前面板、退出放大", "Space / End       暂停跟随、恢复日志跟随", "Ctrl+C / q        取消当前任务、安全退出", "", "链接地址会显示在链接文字之后。"}
	case "errors":
		title = "! 错误详情"
		if p := u.panel(); p != nil {
			lines = append(lines, fmt.Sprintf("程序: %s  exit=%d  耗时=%s", p.LastProgram, p.ExitCode, p.Duration), "时间: "+p.Updated.Format(time.RFC3339), p.Reason)
			lines = append(lines, strings.Split(p.Stderr, "\n")...)
		}
	case "actions":
		title = "◆ 可用动作"
		for i, a := range u.Engine.Page.Actions {
			prefix := "  "
			if i == u.menu {
				prefix = "› "
			}
			line := prefix + a.Label + " [" + a.Key + "]"
			if reason := u.Engine.ActionReason(a); reason != "" {
				line += " 禁用: " + reason
			}
			lines = append(lines, line)
		}
	case "confirm":
		title = "◆ 确认操作"
		lines = []string{u.action.Confirm, ""}
		lines = append(lines, strings.Split(u.Engine.Preview(*u.action), "\n")...)
		choice := "● 取消      ○ 执行"
		if u.confirm {
			choice = "○ 取消      ● 执行"
		}
		lines = append(lines, "", choice, "←→ 或 Tab 切换   Enter 确认   Esc 取消")
	}
	mw := min(w-6, 100)
	if u.modal == "confirm" {
		var wrapped []string
		for _, line := range lines {
			var part strings.Builder
			width := 0
			g := uniseg.NewGraphemes(line)
			for g.Next() {
				if width+g.Width() > mw-6 {
					wrapped = append(wrapped, part.String())
					part.Reset()
					width = 0
				}
				part.WriteString(g.Str())
				width += g.Width()
			}
			wrapped = append(wrapped, part.String())
		}
		lines = wrapped
	}
	mh := min(h-4, len(lines)+4)
	x := (w - mw) / 2
	y := (h - mh) / 2
	shadow := u.themed(0, 0)
	u.fill(x+2, y+1, mw, mh, ' ', shadow)
	panel := u.themed(252, 235)
	border := u.themed(81, 235).Bold(true)
	u.fill(x, y, mw, mh, ' ', panel)
	u.put(x, y, mw, 0, "╭"+strings.Repeat("─", mw-2)+"╮", border)
	u.put(x, y+mh-1, mw, 0, "╰"+strings.Repeat("─", mw-2)+"╯", border)
	for row := y + 1; row < y+mh-1; row++ {
		u.put(x, row, 1, 0, "│", border)
		u.put(x+mw-1, row, 1, 0, "│", border)
	}
	modalTitle := " " + title + " "
	modalTitleWidth := min(mw-17, uniseg.StringWidth(modalTitle))
	u.fill(x+2, y, modalTitleWidth, 1, ' ', panel)
	u.put(x+2, y, modalTitleWidth, 0, modalTitle, border)
	u.fill(x+mw-13, y, 11, 1, ' ', panel)
	u.put(x+mw-13, y, 11, 0, " Esc 关闭 ", u.themed(245, 235))
	u.modalScroll = min(u.modalScroll, max(0, len(lines)-(mh-4)))
	start := u.modalScroll
	if u.modal == "actions" && u.menu >= mh-4 {
		start = u.menu - (mh - 5)
	}
	for j := 0; j < mh-4 && j+start < len(lines); j++ {
		lineIndex := j + start
		style := panel
		if u.modal == "actions" && lineIndex == u.menu {
			style = u.themed(231, 24).Bold(true)
			u.fill(x+2, y+2+j, mw-4, 1, ' ', style)
		}
		if u.modal == "confirm" && (strings.HasPrefix(lines[lineIndex], "●") || strings.HasPrefix(lines[lineIndex], "○")) {
			style = u.themed(222, 235).Bold(true)
		}
		u.put(x+3, y+2+j, mw-6, u.modalX, lines[lineIndex], style)
	}
	u.put(x+3, y+mh-2, mw-6, 0, "↑↓ 滚动", u.themed(245, 235))
}
