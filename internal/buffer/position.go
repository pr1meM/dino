package buffer

// Position is a cursor location expressed as a line index and a rune
// (not byte) column index within that line.
type Position struct {
	Line int
	Col  int
}

func (p Position) Less(o Position) bool {
	if p.Line != o.Line {
		return p.Line < o.Line
	}
	return p.Col < o.Col
}

func (p Position) Equal(o Position) bool {
	return p.Line == o.Line && p.Col == o.Col
}

// OrderedSelection returns the two positions in document order.
func OrderedSelection(a, b Position) (Position, Position) {
	if b.Less(a) {
		return b, a
	}
	return a, b
}
