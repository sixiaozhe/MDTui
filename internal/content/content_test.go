package content

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestItems(t *testing.T) {
	for _, s := range []string{`{"version":2,"items":[]}`, `{"version":1,"items":null}`, `{"version":1,"items":[{"id":"a","label":"a"},{"id":"a","label":"b"}]}`, `{"version":1,"items":[{"id":"a"}]}`, `{"version":1,"items":[],"extra":1}`, `{"version":1,"items":[]} {}`} {
		if _, e := Items([]byte(s)); e == nil {
			t.Fatal(s)
		}
	}
	items, e := Items([]byte(`{"version":1,"items":[{"id":"空 格;$()","label":"中文"}]}`))
	if e != nil || items[0].ID != "空 格;$()" {
		t.Fatal(items, e)
	}
}
func TestSplitUTF8AndEscape(t *testing.T) {
	var b Buffer
	raw := []byte("中文é\x1b]52;c;evil\x07\x1b[2J安全\n")
	for _, x := range raw {
		b.Write([]byte{x})
	}
	b.Finish()
	w, _ := b.Read(0)
	if w.Text != "中文é安全\n" {
		t.Fatalf("%q", w.Text)
	}
}
func TestWindowBounds(t *testing.T) {
	var b Buffer
	for i := 0; i < 12000; i++ {
		b.Write([]byte(strings.Repeat("中", 100) + "\n"))
	}
	w, _ := b.Read(0)
	if len(w.Text) > StreamBytes || strings.Count(w.Text, "\n") > StreamLines || w.Dropped == 0 {
		t.Fatalf("invalid window bytes=%d dropped=%d", len(w.Text), w.Dropped)
	}
	b.Write([]byte(strings.Repeat("长", MaxLine)))
	b.Finish()
	w, _ = b.Read(0)
	if w.Truncated != 1 || !utf8.ValidString(w.Text) {
		t.Fatal(w.Truncated)
	}
}
func TestMarkdown(t *testing.T) {
	s := Markdown("# 标题\n\n> 引用\n\n**粗体** [文档](https://example.test)\n\n- [x] 完成\n\n| 名称 | 值 |\n| :--- | ---: |\n| 中文 | 1 |\n\n```go\n未闭合\x1b[31m恶意")
	plain := stripSGR(s)
	for _, part := range []string{"▍ 标题", "│ 引用", "粗体", "https://example.test", "☑", "╭", "┼", "╯", "code · go", "未闭合", "恶意"} {
		if !strings.Contains(plain, part) {
			t.Fatalf("missing %s in %s", part, s)
		}
	}
	if strings.Contains(s, "\x1b[31m") {
		t.Fatal("untrusted SGR escaped the sanitizer")
	}
	if !strings.Contains(s, "\x1b[1m") {
		t.Fatal("semantic styling was not generated")
	}
}

func TestItemsLimitAndUnknownItemField(t *testing.T) {
	if _, e := Items([]byte(`{"Version":1,"items":[]}`)); e == nil {
		t.Fatal("case-insensitive field accepted")
	}
	if _, e := Items([]byte(`{"version":1,"items":[{"id":"x","label":"x","description":null}]}`)); e == nil {
		t.Fatal("null description accepted")
	}
	if _, e := Items([]byte(`{"version":1,"items":[{"id":"x","label":"x","action":"evil"}]}`)); e == nil {
		t.Fatal("unknown field accepted")
	}
	var s strings.Builder
	s.WriteString(`{"version":1,"items":[`)
	for i := 0; i < 10001; i++ {
		if i > 0 {
			s.WriteByte(',')
		}
		s.WriteString(`{"id":"x","label":"x"}`)
	}
	s.WriteString(`]}`)
	if _, e := Items([]byte(s.String())); e == nil {
		t.Fatal("item count limit not enforced")
	}
}
