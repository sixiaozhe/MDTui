package tui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"mdtui/internal/document"
	"mdtui/internal/runner"
	"mdtui/internal/scheduler"
)

func controlsUI(t *testing.T) *UI {
	t.Helper()
	p, err := document.Parse("controls.md", []byte(`:::mdtui page
version: 1
id: controls
:::
:::mdtui input
id: verbose
type: toggle
:::
:::mdtui input
id: protocol
type: select
options: [{value: tcp, label: TCP}, {value: udp, label: UDP}]
:::
:::mdtui input
id: timeout
type: number
default: 5
min: 1
max: 60
required: true
:::
:::mdtui panel
id: result
autostart: false
exec: ["false", {ref: verbose.value}, {ref: protocol.value}, {ref: timeout.value}]
:::
:::mdtui action
id: apply
button: true
run: result
:::
:::mdtui action
id: confirmed
button: true
run: result
confirm: Continue?
:::
`))
	if err != nil {
		t.Fatal(err)
	}
	u := testUI(t)
	u.Engine = scheduler.New(p, runner.OS{}, time.Now)
	u.focus = 1
	t.Cleanup(u.Close)
	return u
}

func press(u *UI, key tcell.Key) { u.Key(tcell.NewEventKey(key, 0, 0)) }

func TestToggleAndSelectControls(t *testing.T) {
	u := controlsUI(t)
	u.Key(tcell.NewEventKey(tcell.KeyRune, ' ', 0))
	if u.Engine.Inputs["verbose"] != "true" || u.editing {
		t.Fatal("Space did not toggle")
	}
	press(u, tcell.KeyEnter)
	if u.Engine.Inputs["verbose"] != "false" {
		t.Fatal("Enter did not toggle")
	}
	press(u, tcell.KeyTab)
	press(u, tcell.KeyEnter)
	press(u, tcell.KeyDown)
	u.Draw()
	press(u, tcell.KeyEscape)
	if u.Engine.Inputs["protocol"] != "tcp" {
		t.Fatal("cancel changed select")
	}
	press(u, tcell.KeyEnter)
	press(u, tcell.KeyEnd)
	press(u, tcell.KeyEnter)
	if u.Engine.Inputs["protocol"] != "udp" || u.modal != "" {
		t.Fatal("select did not commit underlying value")
	}
	args, reason := u.Engine.Args("result")
	if reason != "" || strings.Join(args[1:], ",") != "false,udp,5" {
		t.Fatalf("args=%v reason=%s", args, reason)
	}
}

func TestNumberControlValidationAndCancel(t *testing.T) {
	u := controlsUI(t)
	u.focus = 3
	press(u, tcell.KeyEnter)
	u.draft = "61"
	press(u, tcell.KeyEnter)
	if !u.editing || u.notice == "" || u.Engine.Inputs["timeout"] != "5" {
		t.Fatal("invalid number committed")
	}
	press(u, tcell.KeyDown)
	if u.draft != "60" {
		t.Fatal("numeric step did not clamp")
	}
	press(u, tcell.KeyEnter)
	if u.editing || u.Engine.Inputs["timeout"] != "60" {
		t.Fatal("valid number did not commit")
	}
	press(u, tcell.KeyEnter)
	u.draft = "not-an-integer"
	press(u, tcell.KeyEnter)
	if !u.editing {
		t.Fatal("invalid numeric string accepted")
	}
	press(u, tcell.KeyEscape)
	if u.Engine.Inputs["timeout"] != "60" {
		t.Fatal("cancel changed number")
	}
	input := document.Input{Type: "number"}
	for _, value := range []int64{math.MinInt64, math.MaxInt64} {
		u.draft = strconv.FormatInt(value, 10)
		direction := 1
		if value < 0 {
			direction = -1
		}
		u.stepNumber(input, direction)
		if u.draft != strconv.FormatInt(value, 10) {
			t.Fatal("step overflowed integer")
		}
	}
}

func TestButtonFocusConfirmationAndDisabled(t *testing.T) {
	u := controlsUI(t)
	u.focus = 4 // result panel, followed by the two buttons
	press(u, tcell.KeyTab)
	if u.focusedButton() == nil || u.panel() != nil {
		t.Fatal("button not focusable")
	}
	press(u, tcell.KeyEnter)
	if u.Engine.Panels["result"].Request == 0 {
		t.Fatal("button did not execute action")
	}
	press(u, tcell.KeyTab)
	press(u, tcell.KeyEnter)
	if u.modal != "confirm" || u.confirm {
		t.Fatal("button skipped default-cancel confirmation")
	}
	before := u.Engine.Panels["result"].Generation
	press(u, tcell.KeyEnter)
	if u.Engine.Panels["result"].Generation != before {
		t.Fatal("cancel executed action")
	}
	select {
	case ev := <-u.Engine.Events:
		u.Engine.Consume(ev)
	case <-time.After(time.Second):
		t.Fatal("initial button action did not complete")
	}
	request := u.Engine.Panels["result"].Request
	press(u, tcell.KeyEnter)
	press(u, tcell.KeyTab)
	press(u, tcell.KeyEnter)
	if u.modal != "" || u.Engine.Panels["result"].Request <= request {
		t.Fatal("confirmed button did not execute")
	}
	u.Engine.Inputs["timeout"] = ""
	press(u, tcell.KeyEnter)
	if u.modal != "" || !strings.Contains(u.notice, "timeout") {
		t.Fatal("disabled button allowed action")
	}
	press(u, tcell.KeyTab)
	if u.focus != 0 {
		t.Fatal("button focus did not wrap to prose")
	}
}

func TestControlsRenderAndLongOptions(t *testing.T) {
	u := controlsUI(t)
	for i := 0; i < 40; i++ {
		u.Engine.Page.Inputs[1].Options = append(u.Engine.Page.Inputs[1].Options, document.Option{Value: fmt.Sprintf("v%d", i), Label: fmt.Sprintf("选项 %d", i)})
	}
	u.focus = 2
	press(u, tcell.KeyEnter)
	press(u, tcell.KeyEnd)
	for _, size := range [][2]int{{60, 16}, {80, 24}, {120, 40}} {
		u.Screen.(tcell.SimulationScreen).SetSize(size[0], size[1])
		u.Draw()
		var text strings.Builder
		for y := 0; y < size[1]; y++ {
			for x := 0; x < size[0]; x++ {
				ch, _, _, width := u.Screen.GetContent(x, y)
				text.WriteRune(ch)
				x += max(1, width) - 1
			}
		}
		if !strings.Contains(text.String(), "选项 39") {
			t.Fatalf("last dropdown option not visible at %v: %q", size, text.String())
		}
	}
	press(u, tcell.KeyEnter)
	if u.Engine.Inputs["protocol"] != "v39" {
		t.Fatal("long dropdown selection failed")
	}
}
