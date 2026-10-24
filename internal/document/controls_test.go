package document

import (
	"strings"
	"testing"
)

func TestInputControlContracts(t *testing.T) {
	for _, tc := range []struct {
		name, fields string
		valid        bool
		value        string
	}{
		{"legacy", "default: hello", true, "hello"},
		{"toggle-default", "type: toggle", true, "false"},
		{"toggle-bool", "type: toggle\ndefault: true", true, "true"},
		{"toggle-string", "type: toggle\ndefault: 'false'", true, "false"},
		{"toggle-invalid", "type: toggle\ndefault: yes", false, ""},
		{"number", "type: number\ndefault: 5\nmin: 1\nmax: 60", true, "5"},
		{"number-string", "type: number\ndefault: '-2'\nmin: -3", true, "-2"},
		{"number-empty", "type: number\nrequired: true", true, ""},
		{"number-low", "type: number\ndefault: 0\nmin: 1", false, ""},
		{"number-high", "type: number\ndefault: 61\nmax: 60", false, ""},
		{"number-fraction", "type: number\ndefault: 1.5", false, ""},
		{"number-overflow", "type: number\ndefault: 9223372036854775808", false, ""},
		{"bounds-reversed", "type: number\nmin: 10\nmax: 1", false, ""},
		{"bounds-fraction", "type: number\nmin: 1.5", false, ""},
		{"bounds-overflow", "type: number\nmax: 9223372036854775808", false, ""},
		{"text-bounds", "min: 0", false, ""},
		{"select", "type: select\noptions: [{value: tcp, label: TCP}, {value: udp, label: UDP}]", true, "tcp"},
		{"select-default", "type: select\ndefault: udp\noptions: [{value: tcp}, {value: udp}]", true, "udp"},
		{"select-unknown", "type: select\ndefault: sctp\noptions: [{value: tcp}]", false, ""},
		{"select-missing", "type: select", false, ""},
		{"select-empty", "type: select\noptions: []", false, ""},
		{"select-duplicate", "type: select\noptions: [{value: tcp}, {value: tcp}]", false, ""},
		{"select-number", "type: select\noptions: [{value: 3}]", false, ""},
		{"select-empty-value", "type: select\noptions: [{value: ''}]", false, ""},
		{"select-empty-label", "type: select\noptions: [{value: tcp, label: ''}]", false, ""},
		{"select-unknown-field", "type: select\noptions: [{value: tcp, typo: true}]", false, ""},
		{"text-options", "options: [{value: tcp}]", false, ""},
		{"unknown-type", "type: slider", false, ""},
		{"length", "type: number\ndefault: 12\nmax_length: 1", false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse("controls.md", []byte(header+":::mdtui input\nid: value\n"+tc.fields+"\n:::\n"))
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v err=%v", tc.valid, err)
			}
			if err == nil && p.Inputs[0].Default != tc.value {
				t.Fatalf("default=%q", p.Inputs[0].Default)
			}
		})
	}
}

func TestInputOptionLimit(t *testing.T) {
	_, err := Parse("controls.md", []byte(header+":::mdtui input\nid: choice\ntype: select\noptions:\n"+strings.Repeat("  - {value: a}\n", 257)+":::\n"))
	if err == nil {
		t.Fatal("accepted too many options")
	}
}

func TestActionButtonContract(t *testing.T) {
	body := header + ":::mdtui panel\nid: result\nexec: [echo]\n:::\n:::mdtui action\nid: apply\nrun: result\nbutton: true\nconfirm: 'Continue?'\n:::\n"
	p, err := Parse("controls.md", []byte(body))
	if err != nil || !p.Actions[0].Button {
		t.Fatalf("button parse: %v", err)
	}
	if _, err := Parse("controls.md", []byte(strings.Replace(body, "button: true", "button: yes", 1))); err == nil {
		t.Fatal("accepted nonboolean button")
	}
}
