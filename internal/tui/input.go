package tui

import (
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

func previousGrapheme(value string, cursor int) int {
	previous, position := 0, 0
	graphemes := uniseg.NewGraphemes(value)
	for graphemes.Next() {
		position += utf8.RuneCountInString(graphemes.Str())
		if position >= cursor {
			return previous
		}
		previous = position
	}
	return previous
}

func nextGrapheme(value string, cursor int) int {
	position := 0
	graphemes := uniseg.NewGraphemes(value)
	for graphemes.Next() {
		position += utf8.RuneCountInString(graphemes.Str())
		if position > cursor {
			return position
		}
	}
	return utf8.RuneCountInString(value)
}

func replaceRunes(value string, start, end int, replacement string) string {
	runes := []rune(value)
	start = max(0, min(start, len(runes)))
	end = max(start, min(end, len(runes)))
	return string(runes[:start]) + replacement + string(runes[end:])
}
