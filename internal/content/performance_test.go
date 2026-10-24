package content

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestStreamSnapshotSurvivesAppendAndCompaction(t *testing.T) {
	b := NewBuffer(64<<10, 3)
	b.Write([]byte("初始\n"))
	old, _ := b.Read(0)
	for i := 0; i < 300; i++ {
		b.Write([]byte(strings.Repeat("中", 200) + "\n"))
	}
	b.Write([]byte("末尾"))
	w, changed := b.Read(old.Revision)
	if !changed || old.Text != "初始\n" || !strings.HasSuffix(w.Text, "末尾") || w.Dropped != 298 {
		t.Fatalf("invalid snapshots: old=%q dropped=%d", old.Text, w.Dropped)
	}
	if _, changed := b.Read(w.Revision); changed {
		t.Fatal("unchanged window copied")
	}
	if b.data.Len() > (64<<10)*2+MaxLine {
		t.Fatal("evicted prefix retained indefinitely")
	}
}

func TestConfiguredStreamBoundsAndPartialRune(t *testing.T) {
	b := NewBuffer(64<<10, 4)
	raw := []byte(strings.Repeat("界", MaxLine) + "\nlast\n")
	// Split a rune at the write boundary, including the final partial line.
	b.Write(raw[:1])
	b.Write(raw[1:])
	b.Finish()
	w, _ := b.Read(0)
	if len(w.Text) > 64<<10 || w.Truncated != 1 || !utf8.ValidString(w.Text) || !strings.HasSuffix(w.Text, "last\n") {
		t.Fatalf("bad bounded window: bytes=%d truncated=%d", len(w.Text), w.Truncated)
	}
}

func TestSanitizerFastPathMatchesSplitWrites(t *testing.T) {
	for _, value := range []string{"中文 é\n\t文字", "a\x1b[31m红色\x1b[0m", "a\x00\x7f\u0085b", "a\xffb", "\x1b]52;c;secret\aok"} {
		var filter Filter
		var out strings.Builder
		for _, b := range []byte(value) {
			out.Write(filter.Write([]byte{b}, false))
		}
		out.Write(filter.Write(nil, true))
		if out.String() != Safe(value) {
			t.Fatalf("split sanitization differs for %q", value)
		}
	}
}

func BenchmarkStreamRead(b *testing.B) {
	var window Buffer
	window.Write([]byte(strings.Repeat("diagnostic\n", 10000)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		window.Read(0)
	}
}

func BenchmarkStreamWriteRead(b *testing.B) {
	var window Buffer
	block := []byte(strings.Repeat("中文 diagnostic "+strings.Repeat("x", 180)+"\n", 100))
	b.SetBytes(int64(len(block)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		window.Write(block)
		window.Read(0)
	}
}
