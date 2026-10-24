package tui

import (
	"fmt"
	"math"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
	"mdtui/internal/document"
)

func (u *UI) buttons() []document.Action {
	var actions []document.Action
	for _, action := range u.Engine.Page.Actions {
		if action.Button {
			actions = append(actions, action)
		}
	}
	return actions
}

func (u *UI) focusedButton() *document.Action {
	index := u.focus - 1 - len(u.Engine.Page.Inputs) - len(u.Engine.Page.Panels)
	buttons := u.buttons()
	if index < 0 || index >= len(buttons) {
		return nil
	}
	return &buttons[index]
}

func (u *UI) activateInput(input document.Input) {
	u.notice = ""
	switch input.Type {
	case "toggle":
		value := "true"
		if u.Engine.Inputs[input.ID] == "true" {
			value = "false"
		}
		u.notice = u.Engine.Submit(input.ID, value)
	case "select":
		u.menu = 0
		for i, option := range input.Options {
			if option.Value == u.Engine.Inputs[input.ID] {
				u.menu = i
				break
			}
		}
		u.modal = "select"
		u.modalScroll, u.modalX = 0, 0
	default:
		u.editing = true
		u.draft = u.Engine.Inputs[input.ID]
		u.draftCursor = len([]rune(u.draft))
		u.inputX = 0
	}
}

func (u *UI) selectKey(k *tcell.EventKey) {
	input := u.Engine.Page.Inputs[u.focus-1]
	switch k.Key() {
	case tcell.KeyEscape:
		u.modal = ""
	case tcell.KeyUp:
		u.menu--
	case tcell.KeyDown:
		u.menu++
	case tcell.KeyPgUp:
		u.menu -= 8
	case tcell.KeyPgDn:
		u.menu += 8
	case tcell.KeyHome:
		u.menu = 0
	case tcell.KeyEnd:
		u.menu = len(input.Options) - 1
	case tcell.KeyLeft:
		u.modalX = max(0, u.modalX-4)
	case tcell.KeyRight:
		u.modalX += 4
	case tcell.KeyEnter:
		if u.menu >= 0 && u.menu < len(input.Options) {
			u.notice = u.Engine.Submit(input.ID, input.Options[u.menu].Value)
			u.modal = ""
		}
	}
	u.menu = min(max(0, u.menu), max(0, len(input.Options)-1))
}

func (u *UI) stepNumber(input document.Input, direction int) {
	n, err := strconv.ParseInt(u.draft, 10, 64)
	if err != nil {
		n = 0
	} else if direction > 0 && n < math.MaxInt64 {
		n++
	} else if direction < 0 && n > math.MinInt64 {
		n--
	}
	if input.Min != nil {
		n = max(n, *input.Min)
	}
	if input.Max != nil {
		n = min(n, *input.Max)
	}
	u.draft = strconv.FormatInt(n, 10)
	u.draftCursor = len(u.draft)
}

func inputDisplay(input document.Input, value string) (string, string) {
	switch input.Type {
	case "toggle":
		if value == "true" {
			return "[✓] 开启", "Space / Enter 切换"
		}
		return "[ ] 关闭", "Space / Enter 切换"
	case "select":
		for _, option := range input.Options {
			if option.Value == value {
				return option.Label + " ▾", "Enter 选择"
			}
		}
	case "number":
		return value, "Enter 编辑整数"
	}
	return value, "Enter 编辑"
}

func (u *UI) drawButtons(y, width int) {
	buttons := u.buttons()
	if len(buttons) == 0 {
		return
	}
	index := max(0, u.focus-1-len(u.Engine.Page.Inputs)-len(u.Engine.Page.Panels))
	index = min(index, len(buttons)-1)
	button := buttons[index]
	style := u.palette(colorMuted, colorSurface)
	buttonStyle := u.controlStyle(u.focusedButton() != nil)
	label := fmt.Sprintf("操作 %d/%d  [ %s ]  Tab 切换 · Enter 执行", index+1, len(buttons), button.Label)
	if reason := u.Engine.ActionReason(button); reason != "" {
		label += " · 禁用: " + reason
		style = style.Dim(true)
		buttonStyle = u.palette(colorMuted, colorSurface).Dim(true)
	}
	u.fill(1, y, width-2, 1, ' ', style)
	u.put(3, y, width-6, 0, label, style)
	base := 1 + len(u.Engine.Page.Inputs) + len(u.Engine.Page.Panels)
	u.addMouse(mouseRegion{x: 1, y: y, w: width - 2, h: 1, kind: "buttonbar", index: index, focus: base + index})
	buttonX := 3 + uniseg.StringWidth(fmt.Sprintf("操作 %d/%d  ", index+1, len(buttons)))
	buttonWidth := min(uniseg.StringWidth("[ "+button.Label+" ]"), width-3-buttonX)
	u.fill(buttonX, y, buttonWidth, 1, ' ', buttonStyle)
	u.put(buttonX, y, buttonWidth, 0, fitText("[ "+button.Label+" ]", buttonWidth), buttonStyle)
	u.addMouse(mouseRegion{x: buttonX, y: y, w: min(uniseg.StringWidth("[ "+button.Label+" ]"), width-3-buttonX), h: 1, kind: "button", index: index, focus: base + index})
	if len(buttons) > 1 {
		u.put(1, y, 1, 0, "‹", style)
		u.put(width-2, y, 1, 0, "›", style)
		u.addMouse(mouseRegion{x: 1, y: y, w: 1, h: 1, kind: "button-prev", index: index})
		u.addMouse(mouseRegion{x: width - 2, y: y, w: 1, h: 1, kind: "button-next", index: index})
	}
}
