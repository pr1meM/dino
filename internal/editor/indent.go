package editor

import (
	"strings"
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

// reindentPaste re-aligns multi-line pasted text to the destination line's
// indentation, so code pasted from elsewhere (a different nesting depth, a
// different indent style) lines up with its new surroundings instead of
// dragging its old indentation along verbatim. The relative indentation
// between pasted lines (nested blocks etc.) is preserved.
//
// Single-line pastes are returned unchanged: there's nothing to realign.
// Pastes where the cursor isn't at the start of a line (i.e. there's
// non-whitespace before it) are also left untouched, since inserting text
// mid-line gives no unambiguous destination indent to apply.
func reindentPaste(b *buffer.Buffer, text string) string {
	if !strings.Contains(text, "\n") {
		return text
	}
	line := b.Line(b.Cursor.Line)
	prefix := line[:min(b.Cursor.Col, len(line))]
	if strings.TrimSpace(string(prefix)) != "" {
		return text
	}
	destIndent := string(prefix)

	lines := strings.Split(text, "\n")
	minIndent := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		n := len(leadingWhitespace([]rune(l)))
		if minIndent == -1 || n < minIndent {
			minIndent = n
		}
	}
	if minIndent < 0 {
		minIndent = 0
	}

	out := make([]string, len(lines))
	for i, l := range lines {
		r := []rune(l)
		stripped := string(r[min(minIndent, len(r)):])
		if strings.TrimSpace(stripped) == "" {
			// Keep blank lines empty rather than trailing whitespace.
			out[i] = ""
			continue
		}
		if i == 0 {
			out[i] = stripped
		} else {
			out[i] = destIndent + stripped
		}
	}
	return strings.Join(out, "\n")
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
