package editor

import (
	"github.com/gdamore/tcell/v2"

	"dino/internal/highlight"
)

// highlightVisible returns, for each visible row, a slice of per-column
// styles for that line (or nil to fall back to the default style). The
// whole buffer is tokenized (not just the visible slice) so that
// multi-line constructs like block comments highlight correctly; the
// result is cached per-tab and only recomputed when the buffer changes.
func (e *Editor) highlightVisible(t *Tab, rows int) [][]tcell.Style {
	if t.lexer == nil || t.lexerFilePath != t.Buf.FilePath {
		t.lexer = highlight.LexerFor(t.Title())
		t.lexerFilePath = t.Buf.FilePath
		t.highlightCache = nil
	}
	if t.highlightCache == nil || t.highlightVersion != t.Buf.Version {
		t.highlightCache = highlight.LineStyles(t.lexer, t.Buf.String(), styleDefault)
		t.highlightVersion = t.Buf.Version
	}

	out := make([][]tcell.Style, rows)
	for row := 0; row < rows; row++ {
		lineIdx := t.TopLine + row
		if lineIdx >= 0 && lineIdx < len(t.highlightCache) {
			out[row] = t.highlightCache[lineIdx]
		}
	}
	return out
}
