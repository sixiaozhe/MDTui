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
