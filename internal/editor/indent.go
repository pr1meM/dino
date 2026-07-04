package editor

import (
	"unicode"

	"dino/internal/buffer"
)

// autoIndentNewline inserts a newline and copies the indentation of the
// current line, adding one extra indent level after an opening bracket
// or a trailing ':' (Python-style blocks).
func (e *Editor) autoIndentNewline(t *Tab) {
	b := t.Buf
	line := b.Line(b.Cursor.Line)
	indent := leadingWhitespace(line)

	trimmed := trimTrailingSpace(line[:min(b.Cursor.Col, len(line))])
	extra := ""
	var opener rune
	if len(trimmed) > 0 {
		last := trimmed[len(trimmed)-1]
		if last == '{' || last == '(' || last == '[' || last == ':' {
			extra = indentUnit(e.Config)
			opener = last
		}
	}

	// If the cursor sits directly between a bracket pair (as left by
	// auto-pairing, e.g. "{|}"), also push the closer onto its own line
	// at the original indent, leaving the cursor on the indented line
	// in between: "{\n    |\n}" instead of "{\n    |}".
	if closer, ok := matchingCloser(opener); ok && b.Cursor.Col < len(line) && line[b.Cursor.Col] == closer {
		startLine := b.Cursor.Line
		b.InsertText("\n" + indent + extra + "\n" + indent)
		b.MoveTo(buffer.Position{Line: startLine + 1, Col: len(indent + extra)}, false)
		return
	}

	b.InsertText("\n" + indent + extra)
}

func matchingCloser(opener rune) (rune, bool) {
	switch opener {
	case '{':
		return '}', true
	case '(':
		return ')', true
	case '[':
		return ']', true
	}
	return 0, false
}

func leadingWhitespace(line []rune) string {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	return string(line[:i])
}

func trimTrailingSpace(line []rune) []rune {
	i := len(line)
	for i > 0 && unicode.IsSpace(line[i-1]) {
		i--
	}
	return line[:i]
}

// maybeDedentForCloser removes one indent level when the user types a
// closing bracket as the first non-whitespace character on a line,
// e.g. typing '}' on an empty, over-indented line dedents it to match
// the block it closes.
func (e *Editor) maybeDedentForCloser(t *Tab, r rune) {
	if r != ')' && r != ']' && r != '}' {
		return
	}
	b := t.Buf
	line := b.Line(b.Cursor.Line)
	lead := leadingWhitespace(line)
	if b.Cursor.Col != len(lead) {
		return
	}
	unit := len(indentUnit(e.Config))
	if unit <= 0 || len(lead) < unit {
		return
	}
	b.MoveTo(buffer.Position{Line: b.Cursor.Line, Col: 0}, false)
	b.MoveTo(buffer.Position{Line: b.Cursor.Line, Col: unit}, true)
	b.DeleteSelection()
}

func indentUnit(cfg Config) string {
	if cfg.UseSpaces {
		s := ""
		for i := 0; i < cfg.TabSize; i++ {
			s += " "
		}
		return s
	}
	return "\t"
}
