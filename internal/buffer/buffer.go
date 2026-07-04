// Package buffer implements the in-memory text buffer used by the editor:
// line storage, cursor/selection tracking and undo/redo.
package buffer

import (
	"os"
	"strings"
)

// Buffer holds the content of a single open file.
type Buffer struct {
	Lines [][]rune

	Cursor    Position
	SelAnchor Position
	HasSel    bool

	FilePath string
	Dirty    bool

	// LineEnding is "\n" or "\r\n", detected from the loaded file.
	LineEnding string
	// NoFinalNewline is true if the source file did not end with a
	// trailing newline; preserved so Save() round-trips exactly.
	NoFinalNewline bool

	TabSize int

	// Version increments on every content mutation. Callers can use it
	// to cheaply detect "did the text change" for caching (e.g. syntax
	// highlighting) without diffing content.
	Version int

	undo []editOp
	redo []editOp
}

// New creates an empty, untitled buffer.
func New() *Buffer {
	return &Buffer{
		Lines:      [][]rune{{}},
		LineEnding: "\n",
		TabSize:    4,
	}
}

// Load reads a file from disk into a new buffer.
func Load(path string) (*Buffer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	b := New()
	b.FilePath = path

	text := string(data)
	if strings.Contains(text, "\r\n") {
		b.LineEnding = "\r\n"
		text = strings.ReplaceAll(text, "\r\n", "\n")
	}

	if text == "" {
		b.Lines = [][]rune{{}}
		return b, nil
	}

	b.NoFinalNewline = !strings.HasSuffix(text, "\n")
	lines := strings.Split(text, "\n")
	if !b.NoFinalNewline {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		lines = []string{""}
	}
	b.Lines = make([][]rune, len(lines))
	for i, l := range lines {
		b.Lines[i] = []rune(l)
	}
	return b, nil
}

// Save writes the buffer content back to FilePath.
func (b *Buffer) Save() error {
	return b.SaveAs(b.FilePath)
}

// SaveAs writes the buffer content to the given path and, on success,
// adopts that path as FilePath and clears the dirty flag.
func (b *Buffer) SaveAs(path string) error {
	content := b.String()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	b.FilePath = path
	b.Dirty = false
	return nil
}

// String renders the full buffer content as it would be saved to disk.
func (b *Buffer) String() string {
	parts := make([]string, len(b.Lines))
	for i, l := range b.Lines {
		parts[i] = string(l)
	}
	joined := strings.Join(parts, "\n")
	if b.LineEnding == "\r\n" {
		joined = strings.ReplaceAll(joined, "\n", "\r\n")
	}
	if !b.NoFinalNewline {
		joined += b.LineEnding
	}
	return joined
}

// LineCount returns the number of lines in the buffer.
func (b *Buffer) LineCount() int { return len(b.Lines) }

// Line returns the runes of a given line, or nil if out of range.
func (b *Buffer) Line(i int) []rune {
	if i < 0 || i >= len(b.Lines) {
		return nil
	}
	return b.Lines[i]
}

func (b *Buffer) clampPosition(p Position) Position {
	if p.Line < 0 {
		p.Line = 0
	}
	if p.Line >= len(b.Lines) {
		p.Line = len(b.Lines) - 1
	}
	lineLen := len(b.Lines[p.Line])
	if p.Col < 0 {
		p.Col = 0
	}
	if p.Col > lineLen {
		p.Col = lineLen
	}
	return p
}
