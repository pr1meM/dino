package buffer

import "strings"

// insertTextAt splices text (which may contain '\n') into the buffer at
// pos, mutating Lines directly. It does not touch cursor or undo state.
func (b *Buffer) insertTextAt(pos Position, text string) {
	if text == "" {
		return
	}
	parts := strings.Split(text, "\n")
	line := b.Lines[pos.Line]
	tail := append([]rune{}, line[pos.Col:]...)

	if len(parts) == 1 {
		newLine := make([]rune, 0, len(line)+len(parts[0]))
		newLine = append(newLine, line[:pos.Col]...)
		newLine = append(newLine, []rune(parts[0])...)
		newLine = append(newLine, tail...)
		b.Lines[pos.Line] = newLine
		return
	}

	first := append(append([]rune{}, line[:pos.Col]...), []rune(parts[0])...)
	middle := make([][]rune, len(parts)-2)
	for i := 1; i < len(parts)-1; i++ {
		middle[i-1] = []rune(parts[i])
	}
	last := append([]rune(parts[len(parts)-1]), tail...)

	newLines := make([][]rune, 0, len(b.Lines)+len(parts)-1)
	newLines = append(newLines, b.Lines[:pos.Line]...)
	newLines = append(newLines, first)
	newLines = append(newLines, middle...)
	newLines = append(newLines, last)
	newLines = append(newLines, b.Lines[pos.Line+1:]...)
	b.Lines = newLines
}

// deleteRange removes [start, end) from the buffer and returns the
// removed text. start must be <= end in document order.
func (b *Buffer) deleteRange(start, end Position) string {
	if start.Equal(end) {
		return ""
	}
	if start.Line == end.Line {
		line := b.Lines[start.Line]
		removed := string(line[start.Col:end.Col])
		newLine := make([]rune, 0, len(line)-(end.Col-start.Col))
		newLine = append(newLine, line[:start.Col]...)
		newLine = append(newLine, line[end.Col:]...)
		b.Lines[start.Line] = newLine
		return removed
	}

	var sb strings.Builder
	sb.WriteString(string(b.Lines[start.Line][start.Col:]))
	for l := start.Line + 1; l < end.Line; l++ {
		sb.WriteByte('\n')
		sb.WriteString(string(b.Lines[l]))
	}
	sb.WriteByte('\n')
	sb.WriteString(string(b.Lines[end.Line][:end.Col]))

	merged := make([]rune, 0)
	merged = append(merged, b.Lines[start.Line][:start.Col]...)
	merged = append(merged, b.Lines[end.Line][end.Col:]...)

	newLines := make([][]rune, 0, len(b.Lines)-(end.Line-start.Line))
	newLines = append(newLines, b.Lines[:start.Line]...)
	newLines = append(newLines, merged)
	newLines = append(newLines, b.Lines[end.Line+1:]...)
	b.Lines = newLines

	return sb.String()
}

// advance computes the position reached after walking text (which may
// contain '\n') forward from start.
func advance(start Position, text string) Position {
	line, col := start.Line, start.Col
	for _, r := range text {
		if r == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return Position{Line: line, Col: col}
}
