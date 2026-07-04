package editor

import (
	"path/filepath"
	"strings"

	"dino/internal/buffer"
)

var lineCommentPrefixes = map[string]string{
	".py": "#", ".sh": "#", ".bash": "#", ".rb": "#", ".yaml": "#", ".yml": "#",
	".go": "//", ".c": "//", ".h": "//", ".cpp": "//", ".hpp": "//", ".cc": "//",
	".rs": "//", ".js": "//", ".ts": "//", ".java": "//", ".css": "//",
}

func commentPrefixFor(filename string) string {
	prefix, ok := lineCommentPrefixes[strings.ToLower(filepath.Ext(filename))]
	if !ok {
		return "#"
	}
	return prefix
}

// toggleComment line-comments (or uncomments, if already commented) the
// selected lines, or the current line when there is no selection.
func (e *Editor) toggleComment() {
	t := e.CurTab()
	if t == nil {
		return
	}
	b := t.Buf
	prefix := commentPrefixFor(t.Title())

	startLine, endLine := b.Cursor.Line, b.Cursor.Line
	hadSel := b.HasSel
	if b.HasSel {
		s, en := buffer.OrderedSelection(b.SelAnchor, b.Cursor)
		startLine, endLine = s.Line, en.Line
		if en.Col == 0 && endLine > startLine {
			endLine-- // selection ends right at a line start; don't touch that line
		}
	}

	allCommented := true
	for l := startLine; l <= endLine; l++ {
		line := b.Line(l)
		trimmed := strings.TrimLeft(string(line), " \t")
		if trimmed != "" && !strings.HasPrefix(trimmed, prefix) {
			allCommented = false
			break
		}
	}

	for l := startLine; l <= endLine; l++ {
		line := b.Line(l)
		lead := leadingWhitespace(line)
		rest := string(line[len(lead):])
		var newLine string
		if allCommented {
			newLine = lead + strings.TrimPrefix(strings.TrimPrefix(rest, prefix+" "), prefix)
		} else if rest == "" {
			newLine = string(line)
		} else {
			newLine = lead + prefix + " " + rest
		}
		b.MoveTo(buffer.Position{Line: l, Col: 0}, false)
		b.MoveTo(buffer.Position{Line: l, Col: len(line)}, true)
		b.DeleteSelection()
		b.InsertText(newLine)
	}

	if hadSel {
		b.ClearSelection()
	}
}
