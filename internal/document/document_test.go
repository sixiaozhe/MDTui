package document

import (
	"strings"
	"testing"
)

const header = ":::mdtui page\nversion: 1\nid: test\n:::\n"

func TestCompilerContracts(t *testing.T) {
	cases := []struct {
		name, body string
		valid      bool
	}{
		{"minimal", header, true},
		{"panel-max-height", header + ":::mdtui panel\nid: x\nexec: [echo]\nmax_height: 5\n:::\n", true},
		{"panel-max-height-min", header + ":::mdtui panel\nid: x\nexec: [echo]\nmax_height: 3\n:::\n", true},
		{"panel-max-height-small", header + ":::mdtui panel\nid: x\nexec: [echo]\nmax_height: 2\n:::\n", false},
		{"panel-max-height-zero", header + ":::mdtui panel\nid: x\nexec: [echo]\nmax_height: 0\n:::\n", false},
		{"panel-max-height-negative", header + ":::mdtui panel\nid: x\nexec: [echo]\nmax_height: -1\n:::\n", false},
		{"panel-max-height-string", header + ":::mdtui panel\nid: x\nexec: [echo]\nmax_height: '5'\n:::\n", false},
		{"panel-max-height-fraction", header + ":::mdtui panel\nid: x\nexec: [echo]\nmax_height: 4.5\n:::\n", false},
		{"workspace-sidebar", strings.Replace(header, "id: test", "id: test\nlayout: workspace\nsidebar: list", 1) + ":::mdtui panel\nid: list\nexec: [echo]\n:::\n", true},
		{"workspace-missing-sidebar", strings.Replace(header, "id: test", "id: test\nlayout: workspace", 1), false},
		{"code", "```text\n:::mdtui panel\ninvalid\n```\n" + header, true},
		{"tilde", "~~~\n:::mdtui page\n~~~\n" + header, true},
		{"unknown", header + ":::mdtui panel\nid: x\nexec: [echo]\ntypo: true\n:::", false},
		{"duplicate", header + ":::mdtui panel\nid: test\nexec: [echo]\n:::", false},
		{"alias", header + ":::mdtui panel\nid: x\nexec: [echo, &a literal, *a]\n:::", false},
		{"duplicate-key", strings.Replace(header, "version: 1", "version: 1\nversion: 1", 1), false},
		{"numeric-arg", header + ":::mdtui panel\nid: x\nexec: [echo, 12]\n:::", false},
		{"missing-ref", header + ":::mdtui panel\nid: x\nexec: [echo, {ref: 'bad.value'}]\n:::", false},
		{"cycle", header + ":::mdtui panel\nid: x\nformat: items-json\nexec: [echo, {ref: 'x.selected.id'}]\n:::", false},
		{"unclosed", header + ":::mdtui panel\nid: x", false},
		{"nested", header + ":::mdtui panel\n:::mdtui input\n:::", false},
		{"stream-items", header + ":::mdtui panel\nid: x\nmode: stream\nformat: items-json\nexec: [echo]\n:::", false},
		{"refresh-zero", header + ":::mdtui panel\nid: x\nrefresh: 0\nexec: [echo]\n:::", false},
		{"open-valid", header + ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: ./child.md\n:::", true},
		{"open-wrong-format", header + ":::mdtui panel\nid: x\nexec: [echo]\nopen: ./child.md\n:::", false},
		{"open-absolute", header + ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: /tmp/child.md\n:::", false},
		{"parent-ref", strings.Replace(header, "id: test", "id: test\nparent: list", 1) + ":::mdtui panel\nid: x\nexec: [echo, {ref: 'parent.selected.id'}]\n:::", true},
		{"panel-named-parent", header + ":::mdtui panel\nid: parent\nformat: items-json\nexec: [echo]\n:::\n:::mdtui panel\nid: detail\nexec: [echo, {ref: 'parent.selected.id'}]\n:::", true},
		{"parent-ref-without-parent", header + ":::mdtui panel\nid: x\nexec: [echo, {ref: 'parent.selected.id'}]\n:::", false},
		{"parent-invalid", strings.Replace(header, "id: test", "id: test\nparent: 1bad", 1), false},
		{"open-empty", header + ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: \"\"\n:::", false},
		{"open-blank", header + ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: \"   \"\n:::", false},
		{"open-map-valid", header + ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: {system: ./system.md, network: ./network.md}\n:::", true},
		{"open-map-empty-path", header + ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: {system: \"\"}\n:::", false},
		{"open-map-non-string", header + ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: {system: 1}\n:::", false},
		{"open-map-empty", header + ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: {}\n:::", false},
		{"open-map-wrong-format", header + ":::mdtui panel\nid: list\nexec: [echo]\nopen: {system: ./system.md}\n:::", false},
		{"on-select-valid", header + ":::mdtui panel\nid: list\nformat: items-json\non_select: inspect\nexec: [echo]\n:::\n:::mdtui action\nid: inspect\nrun: list\n:::", true},
		{"on-focus-valid", header + ":::mdtui panel\nid: list\nformat: items-json\non_focus: refresh_list\nexec: [echo]\n:::\n:::mdtui action\nid: refresh_list\nrefresh: [list]\n:::", true},
		{"on-focus-not-items", header + ":::mdtui panel\nid: detail\non_focus: refresh_detail\nexec: [echo]\n:::\n:::mdtui action\nid: refresh_detail\nrefresh: [detail]\n:::", false},
		{"on-select-not-items", header + ":::mdtui panel\nid: detail\non_select: inspect\nexec: [echo]\n:::\n:::mdtui action\nid: inspect\nrun: detail\n:::", false},
		{"cursor-ref-valid", header + ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\n:::\n:::mdtui panel\nid: preview\nexec: [echo, {ref: 'list.cursor.id'}]\n:::", true},
		{"cursor-ref-not-list", header + ":::mdtui panel\nid: detail\nexec: [echo]\n:::\n:::mdtui panel\nid: preview\nexec: [echo, {ref: 'detail.cursor.id'}]\n:::", false},
		{"on-select-unknown-action", header + ":::mdtui panel\nid: list\nformat: items-json\non_select: missing\nexec: [echo]\n:::", false},
		{"on-select-confirm", header + ":::mdtui panel\nid: list\nformat: items-json\non_select: inspect\nexec: [echo]\n:::\n:::mdtui action\nid: inspect\nrun: list\nconfirm: \"确认？\"\n:::", false},
		{"on-select-open", header + ":::mdtui panel\nid: list\nformat: items-json\non_select: inspect\nopen: ./child.md\nexec: [echo]\n:::\n:::mdtui action\nid: inspect\nrun: list\n:::", false},
		{"reserved-key", header + ":::mdtui panel\nid: x\nexec: [echo]\n:::\n:::mdtui action\nid: a\nrun: x\nkey: q\n:::", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, e := Parse("fixture.md", []byte(tc.body))
			if (e == nil) != tc.valid {
				t.Fatalf("valid=%v err=%v", tc.valid, e)
			}
			if e != nil && !strings.Contains(e.Error(), "fixture.md:") {
				t.Fatal(e)
			}
		})
	}
}
func TestOpenPathIsTrimmed(t *testing.T) {
	body := header + ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen: \"  ./child.md  \"\n:::"
	p, err := Parse("fixture.md", []byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Panels) != 1 || p.Panels[0].Open != "./child.md" {
		t.Fatalf("open = %q", p.Panels[0].Open)
	}
}

func TestPanelMaxHeightValue(t *testing.T) {
	for _, height := range []int{0, 5} {
		field := ""
		if height > 0 {
			field = "max_height: 5\n"
		}
		p, err := Parse("height.md", []byte(header+":::mdtui panel\nid: overview\nexec: [echo]\n"+field+":::\n"))
		if err != nil {
			t.Fatal(err)
		}
		if p.Panels[0].MaxHeight != height {
			t.Fatalf("max_height=%d want=%d", p.Panels[0].MaxHeight, height)
		}
	}
}

func TestOpenTargets(t *testing.T) {
	body := header + ":::mdtui panel\nid: list\nformat: items-json\nexec: [echo]\nopen:\n  network: ./network.md\n  system: \"./system.md\"\n:::"
	p, err := Parse("fixture.md", []byte(body))
	if err != nil {
		t.Fatal(err)
	}
	panel := p.Panels[0]
	if !panel.HasOpen() {
		t.Fatal("map open not parsed")
	}
	if target, ok := panel.OpenTarget("network"); !ok || target != "./network.md" {
		t.Fatalf("network target = %q, %v", target, ok)
	}
	if _, ok := panel.OpenTarget("missing"); ok {
		t.Fatal("unexpected target for missing item")
	}
	targets := panel.OpenTargets()
	if len(targets) != 2 || targets[0] != "./network.md" || targets[1] != "./system.md" {
		t.Fatalf("targets = %#v", targets)
	}
}

func TestLimits(t *testing.T) {
	if _, e := Parse("large.md", make([]byte, (1<<20)+1)); e == nil {
		t.Fatal("accepted oversized file")
	}
	if _, e := Parse("utf8.md", []byte{255}); e == nil {
		t.Fatal("accepted invalid UTF-8")
	}
}
func FuzzParser(f *testing.F) {
	f.Add([]byte(header))
	f.Fuzz(func(t *testing.T, b []byte) { Parse("fuzz.md", b) })
}
