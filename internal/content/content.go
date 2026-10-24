package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

type Item struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

func Items(data []byte) ([]Item, error) {
	var envelope struct {
		Version int             `json:"version"`
		Items   json.RawMessage `json:"items"`
	}
	if err := strictObject(data, &envelope, "version items"); err != nil {
		return nil, err
	}
	if envelope.Version != 1 {
		return nil, fmt.Errorf("items-json version 必须为 1")
	}
	if len(envelope.Items) == 0 || envelope.Items[0] != '[' {
		return nil, fmt.Errorf("items 必须为数组")
	}
	type rawItem struct {
		ID          *string         `json:"id"`
		Label       *string         `json:"label"`
		Description json.RawMessage `json:"description"`
	}
	var raw []json.RawMessage
	if err := strictJSON(envelope.Items, &raw); err != nil {
		return nil, err
	}
	if len(raw) > 10000 {
		return nil, fmt.Errorf("items 超过 10000 项")
	}
	result := make([]Item, 0, len(raw))
	seen := map[string]bool{}
	for _, encoded := range raw {
		var x rawItem
		if err := strictObject(encoded, &x, "id label description"); err != nil {
			return nil, err
		}
		if x.ID == nil || x.Label == nil || *x.ID == "" || utf8.RuneCountInString(*x.ID) > 256 || seen[*x.ID] {
			return nil, fmt.Errorf("item id/label 缺失、id 重复或超限")
		}
		for _, r := range *x.ID {
			if unicode.IsControl(r) {
				return nil, fmt.Errorf("item id 含控制字符")
			}
		}
		seen[*x.ID] = true
		description := ""
		if len(x.Description) > 0 {
			if x.Description[0] != '"' {
				return nil, fmt.Errorf("description 必须为字符串")
			}
			if err := json.Unmarshal(x.Description, &description); err != nil {
				return nil, err
			}
		}
		result = append(result, Item{*x.ID, Safe(*x.Label), Safe(description)})
	}
	return result, nil
}
func strictObject(data []byte, v any, fields string) error {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("需要 JSON 对象")
	}
	for key := range obj {
		if !strings.Contains(" "+fields+" ", " "+key+" ") {
			return fmt.Errorf("未知 JSON 字段 %q", key)
		}
	}
	return strictJSON(data, v)
}
func strictJSON(data []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	var x any
	if err := d.Decode(&x); err != io.EOF {
		return fmt.Errorf("JSON 后有多余数据")
	}
	return nil
}

// Filter consumes control sequences across write boundaries, including OSC/DCS.
type Filter struct {
	state   int
	pending []byte
}

func (f *Filter) Write(data []byte, final bool) []byte {
	// The common case needs no sanitization or intermediate byte slices.
	if f.state == 0 && len(f.pending) == 0 && plainText(data) {
		return data
	}
	out := make([]byte, 0, len(data))
	for _, b := range data {
		switch f.state {
		case 0:
			if b == 27 {
				f.state = 1
			} else if b >= 32 || b == '\n' || b == '\t' {
				out = append(out, b)
			}
		case 1:
			switch b {
			case '[':
				f.state = 2
			case ']', 'P', 'X', '^', '_':
				f.state = 3
			default:
				f.state = 0
			}
		case 2:
			if b >= 0x40 && b <= 0x7e {
				f.state = 0
			}
		case 3:
			if b == 7 {
				f.state = 0
			} else if b == 27 {
				f.state = 4
			}
		case 4:
			if b == '\\' {
				f.state = 0
			} else if b != 27 {
				f.state = 3
			}
		}
	}
	out = append(f.pending, out...)
	f.pending = nil
	clean := make([]byte, 0, len(out))
	for len(out) > 0 {
		if !final && !utf8.FullRune(out) {
			f.pending = append([]byte(nil), out...)
			break
		}
		r, n := utf8.DecodeRune(out)
		out = out[n:]
		if !unicode.IsControl(r) || r == '\n' || r == '\t' {
			clean = utf8.AppendRune(clean, r)
		}
	}
	return clean
}
func plainText(data []byte) bool {
	for len(data) > 0 {
		r, n := utf8.DecodeRune(data)
		if r == utf8.RuneError && n == 1 || unicode.IsControl(r) && r != '\n' && r != '\t' {
			return false
		}
		data = data[n:]
	}
	return true
}

func Safe(s string) string {
	if plainText([]byte(s)) {
		return s
	}
	var f Filter
	return string(f.Write([]byte(s), true))
}

const StreamBytes = 2 << 20
const StreamLines = 10000
const MaxLine = 64 << 10

type Buffer struct {
	mu     sync.Mutex
	filter Filter
	// Builder snapshots are immutable strings. Appending never modifies a
	// previously returned window; advancing start evicts without shifting bytes.
	data               strings.Builder
	start              int
	maxBytes, maxLines int
	lineBytes, lines   int
	dropping           bool
	Dropped, Truncated int
	revision           uint64
}
type Window struct {
	Text               string
	Dropped, Truncated int
	Revision           uint64
}

// NewBuffer configures a bounded stream window; zero values use the defaults.
func NewBuffer(maxBytes, maxLines int) *Buffer {
	return &Buffer{maxBytes: maxBytes, maxLines: maxLines}
}

func (b *Buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.add(b.filter.Write(p, false))
	return len(p), nil
}
func (b *Buffer) Finish() { b.mu.Lock(); defer b.mu.Unlock(); b.add(b.filter.Write(nil, true)) }
func (b *Buffer) add(p []byte) {
	if len(p) == 0 {
		return
	}
	b.revision++
	maxBytes, maxLines := b.maxBytes, b.maxLines
	if maxBytes <= 0 {
		maxBytes = StreamBytes
	}
	if maxLines <= 0 {
		maxLines = StreamLines
	}
	lineLimit := min(MaxLine, max(1, maxBytes-32))
	for len(p) > 0 {
		i := bytes.IndexByte(p, '\n')
		part := p
		if i >= 0 {
			part = p[:i]
		}
		if !b.dropping {
			n := min(len(part), max(0, lineLimit-b.lineBytes))
			// Filter has already validated UTF-8. Never split a rune at the cap.
			for n > 0 && n < len(part) && !utf8.RuneStart(part[n]) {
				n--
			}
			b.data.Write(part[:n])
			b.lineBytes += n
			if n < len(part) {
				b.dropping = true
				b.Truncated++
				b.data.WriteString(" [单行截断]")
			}
		}
		if i >= 0 {
			b.data.WriteByte('\n')
			b.lines++
			b.lineBytes = 0
			b.dropping = false
			p = p[i+1:]
		} else {
			p = nil
		}
		for b.data.Len()-b.start > maxBytes || b.lines > maxLines {
			idx := strings.IndexByte(b.data.String()[b.start:], '\n')
			if idx < 0 {
				break
			}
			b.start += idx + 1
			b.lines--
			b.Dropped++
		}
		// Compact periodically, not on every read or evicted line. Clone the
		// remaining window into a new builder so older snapshots stay valid.
		if b.start >= maxBytes {
			text := b.data.String()[b.start:]
			b.data.Reset()
			// Leave space for the retained window AND the next eviction cycle.
			// Otherwise Builder growth repeatedly copies the entire retained tail.
			b.data.Grow(2*maxBytes + lineLimit)
			b.data.WriteString(text)
			b.start = 0
		}
	}
}
func (b *Buffer) Read(after uint64) (Window, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if after == b.revision {
		return Window{}, false
	}
	return Window{b.data.String()[b.start:], b.Dropped, b.Truncated, b.revision}, true
}
