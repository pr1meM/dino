package editor

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/gdamore/tcell/v2"

	"dino/internal/buffer"
)

// Tab holds one open buffer plus its own viewport scroll state.
type Tab struct {
	Buf *buffer.Buffer

	// TopLine/LeftCol are the first visible line/column in the text area.
	TopLine int
	LeftCol int

	lexer         chroma.Lexer
	lexerFilePath string

	// highlightCache avoids re-tokenizing the whole buffer on every
	// redraw; it's invalidated whenever Buf.Version changes.
	highlightVersion int
	highlightCache   [][]tcell.Style
}

func NewTab(b *buffer.Buffer) *Tab {
	return &Tab{Buf: b}
}

// Title returns the short display name for the tab bar / status line.
func (t *Tab) Title() string {
	if t.Buf.FilePath == "" {
		return "[New]"
	}
	return shortPath(t.Buf.FilePath)
}

func shortPath(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[i+1:]
		}
	}
	return p
}
