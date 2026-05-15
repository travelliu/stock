package analysis

import (
	"strings"
	"unicode"
)

// DisplayWidth returns the terminal cell width, treating CJK wide chars as 2.
func DisplayWidth(s string) int {
	w := 0
	for _, r := range s {
		if isWide(r) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func isWide(r rune) bool {
	if r >= 0x1100 && r <= 0x115F { // Hangul Jamo
		return true
	}
	if r >= 0x2E80 && r <= 0x303E { // CJK Radicals etc.
		return true
	}
	if r >= 0x3041 && r <= 0x33FF { // Hiragana/Katakana/CJK Sym & Punct
		return true
	}
	if r >= 0x3400 && r <= 0x4DBF { // CJK Ext A
		return true
	}
	if r >= 0x4E00 && r <= 0x9FFF { // CJK Unified
		return true
	}
	if r >= 0xA000 && r <= 0xA4CF { // Yi
		return true
	}
	if r >= 0xAC00 && r <= 0xD7A3 { // Hangul Syllables
		return true
	}
	if r >= 0xF900 && r <= 0xFAFF { // CJK Compatibility
		return true
	}
	if r >= 0xFE30 && r <= 0xFE4F { // CJK Compatibility Forms
		return true
	}
	if r >= 0xFF00 && r <= 0xFF60 { // Fullwidth ASCII
		return true
	}
	if r >= 0xFFE0 && r <= 0xFFE6 {
		return true
	}
	if unicode.Is(unicode.Han, r) {
		return true
	}
	return false
}

// Rpad right-pads `s` with spaces to the requested display width.
func Rpad(s string, width int) string {
	if d := width - DisplayWidth(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

// FormatTable renders a CJK-aware ASCII table.
func FormatTable(headers []string, rows [][]string) string {
	colW := make([]int, len(headers))
	for i, h := range headers {
		colW[i] = DisplayWidth(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colW) && DisplayWidth(cell) > colW[i] {
				colW[i] = DisplayWidth(cell)
			}
		}
	}
	sep := "+"
	for _, w := range colW {
		sep += strings.Repeat("-", w+2) + "+"
	}
	var b strings.Builder
	b.WriteString(sep + "\n")
	b.WriteString("|")
	for i, h := range headers {
		b.WriteString(" " + Rpad(h, colW[i]) + " |")
	}
	b.WriteString("\n" + sep + "\n")
	for _, row := range rows {
		b.WriteString("|")
		for i, cell := range row {
			if i < len(colW) {
				b.WriteString(" " + Rpad(cell, colW[i]) + " |")
			}
		}
		b.WriteString("\n")
	}
	b.WriteString(sep)
	return b.String()
}
