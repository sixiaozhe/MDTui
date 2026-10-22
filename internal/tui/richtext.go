package tui

import (
	"strings"
	"unicode"

	"github.com/rivo/uniseg"
)

// visibleRich removes only the SGR sequences produced by the Markdown renderer.
func visibleRich(line string) string {
	var out strings.Builder
	for i := 0; i < len(line); {
		if line[i] == '\x1b' && i+1 < len(line) && line[i+1] == '[' {
			if end := strings.IndexByte(line[i+2:], 'm'); end >= 0 {
				i += end + 3
				continue
			}
		}
		out.WriteByte(line[i])
		i++
	}
	return out.String()
}

func richNoWrap(line string) bool {
	plain := strings.TrimLeft(visibleRich(line), " ")
	if plain == "" {
		return false
	}
	if strings.HasPrefix(plain, "╭") || strings.HasPrefix(plain, "├") || strings.HasPrefix(plain, "╰") {
		return true
	}
	return strings.HasPrefix(plain, "│") && (strings.HasSuffix(plain, "│") || strings.Contains(line, "\x1b[48;5;236m"))
}

func continuationIndent(line string) string {
	plain := strings.TrimLeft(visibleRich(line), " ")
	if strings.HasPrefix(plain, "│ ") || strings.HasPrefix(plain, "• ") || strings.HasPrefix(plain, "☑ ") || strings.HasPrefix(plain, "☐ ") {
		return "  "
	}
	for _, r := range plain {
		if unicode.IsDigit(r) {
			continue
		}
		if r == '.' {
			return "   "
		}
		break
	}
	return ""
}

// wrapRichLine wraps styled prose by terminal cell width, carrying active SGR
// styling across visual lines. Tables and code retain horizontal scrolling.
func wrapRichLine(line string, width int) []string {
	if width <= 0 || richNoWrap(line) {
		return []string{line}
	}
	indent := continuationIndent(line)
	if uniseg.StringWidth(indent) >= width {
		indent = ""
	}
	var result []string
	var current strings.Builder
	active := ""
	used := 0
	for i := 0; i < len(line); {
		if line[i] == '\x1b' && i+1 < len(line) && line[i+1] == '[' {
			if end := strings.IndexByte(line[i+2:], 'm'); end >= 0 {
				sequence := line[i : i+end+3]
				current.WriteString(sequence)
				if sequence == "\x1b[0m" {
					active = ""
				} else {
					active += sequence
				}
				i += end + 3
				continue
			}
		}
		next := strings.IndexByte(line[i:], '\x1b')
		if next < 0 {
			next = len(line) - i
		}
		if next == 0 {
			// Rendered Markdown cannot contain an unsupported escape, but progress
			// if a malformed sequence is passed by a caller.
			i++
			continue
		}
		graphemes := uniseg.NewGraphemes(strings.ReplaceAll(line[i:i+next], "\t", "    "))
		for graphemes.Next() {
			cluster := graphemes.Str()
			cellWidth := graphemes.Width()
			if used > 0 && used+cellWidth > width {
				result = append(result, current.String())
				current.Reset()
				current.WriteString(active)
				current.WriteString(indent)
				used = uniseg.StringWidth(indent)
			}
			current.WriteString(cluster)
			used += cellWidth
		}
		i += next
	}
	result = append(result, current.String())
	return result
}

func wrapRichLines(lines []string, width int) []string {
	var wrapped []string
	for _, line := range lines {
		wrapped = append(wrapped, wrapRichLine(line, width)...)
	}
	return wrapped
}

func (u *UI) highlightMatches(x, y, width, offset int, line string) {
	if u.searchQuery == "" || width <= 0 {
		return
	}
	for _, match := range matchCells(line, u.searchQuery) {
		left, right := match[0], match[1]
		for cell := left; cell < right; cell++ {
			column := x + cell - offset
			if cell < offset || column < x || column >= x+width {
				continue
			}
			main, combining, style, _ := u.Screen.GetContent(column, y)
			if main != 0 && main != ' ' {
				u.Screen.SetContent(column, y, main, combining, style.Reverse(true))
			}
		}
	}
}

func hasMatch(line, query string) bool {
	return query != "" && strings.Contains(strings.ToLower(visibleRich(line)), strings.ToLower(query))
}

func matchCells(line, query string) [][2]int {
	if query == "" {
		return nil
	}
	plain := []rune(visibleRich(line))
	needle := []rune(query)
	if len(needle) > len(plain) {
		return nil
	}
	var matches [][2]int
	for start := 0; start+len(needle) <= len(plain); {
		end := start + len(needle)
		if strings.EqualFold(string(plain[start:end]), query) {
			matches = append(matches, [2]int{
				uniseg.StringWidth(string(plain[:start])),
				uniseg.StringWidth(string(plain[:end])),
			})
			start = end
		} else {
			start++
		}
	}
	return matches
}
