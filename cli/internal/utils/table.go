package utils

import (
	"strings"
	"unicode"
)

// Table represents a simple unicode table.
type Table struct {
	Headers []string
	Rows    [][]string
}

// Render returns the formatted table as a string using box-drawing characters.
func (t *Table) Render() string {
	widths := t.calculateWidths()
	var b strings.Builder

	top := buildBorder("┌", "┬", "┐", widths)
	headerSep := buildBorder("├", "┼", "┤", widths)
	bottom := buildBorder("└", "┴", "┘", widths)

	b.WriteString(top)
	if len(t.Headers) > 0 {
		b.WriteString(renderRow(t.Headers, widths))
		if len(t.Rows) > 0 {
			b.WriteString(headerSep)
		}
	}

	for i, row := range t.Rows {
		b.WriteString(renderRow(row, widths))
		if i < len(t.Rows)-1 {
			b.WriteString(headerSep)
		}
	}
	b.WriteString(bottom)

	return b.String()
}

func (t *Table) calculateWidths() []int {
	columns := len(t.Headers)
	if len(t.Rows) > 0 && len(t.Rows[0]) > columns {
		columns = len(t.Rows[0])
	}

	widths := make([]int, columns)
	for i, h := range t.Headers {
		if w := DisplayWidth(h); w > widths[i] {
			widths[i] = w
		}
	}

	for _, row := range t.Rows {
		for idx := 0; idx < columns && idx < len(row); idx++ {
			for _, line := range strings.Split(row[idx], "\n") {
				if w := DisplayWidth(line); w > widths[idx] {
					widths[idx] = w
				}
			}
		}
	}
	return widths
}

func buildBorder(start, sep, end string, widths []int) string {
	var b strings.Builder
	b.WriteString(start)
	for i, w := range widths {
		b.WriteString(strings.Repeat("─", w+2))
		if i == len(widths)-1 {
			b.WriteString(end)
		} else {
			b.WriteString(sep)
		}
	}
	b.WriteString("\n")
	return b.String()
}

func renderRow(row []string, widths []int) string {
	height := 1
	lines := make([][]string, len(widths))
	for i := range widths {
		if i < len(row) {
			lines[i] = strings.Split(row[i], "\n")
			if len(lines[i]) > height {
				height = len(lines[i])
			}
		} else {
			lines[i] = []string{""}
		}
	}

	var b strings.Builder
	for h := 0; h < height; h++ {
		b.WriteString("│")
		for idx, w := range widths {
			cellLine := ""
			if h < len(lines[idx]) {
				cellLine = lines[idx][h]
			}
			b.WriteString(" " + PadRight(cellLine, w) + " ")
			b.WriteString("│")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// PadRight pads the string to the desired display width.
func PadRight(s string, width int) string {
	diff := width - DisplayWidth(s)
	if diff > 0 {
		return s + strings.Repeat(" ", diff)
	}
	return s
}

// DisplayWidth calculates the printable width of a string with basic unicode support.
func DisplayWidth(s string) int {
	width := 0
	for _, r := range s {
		switch {
		case unicode.Is(unicode.Mn, r):
			// combining marks have zero width
		case isWideRune(r):
			width += 2
		default:
			width++
		}
	}
	return width
}

func isWideRune(r rune) bool {
	if unicode.In(r, unicode.Han, unicode.Hangul, unicode.Hiragana, unicode.Katakana) {
		return true
	}
	switch {
	case r >= 0x1100 && r <= 0x115F:
		return true
	case r == 0x2329 || r == 0x232A:
		return true
	case r >= 0x2E80 && r <= 0xA4CF && r != 0x303F:
		return true
	case r >= 0xAC00 && r <= 0xD7A3:
		return true
	case r >= 0xF900 && r <= 0xFAFF:
		return true
	case r >= 0xFE10 && r <= 0xFE19:
		return true
	case r >= 0xFE30 && r <= 0xFE6F:
		return true
	case r >= 0xFF00 && r <= 0xFF60:
		return true
	case r >= 0xFFE0 && r <= 0xFFE6:
		return true
	default:
		return false
	}
}
