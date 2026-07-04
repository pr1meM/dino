package buffer

import "unicode"

// beginMove prepares the selection anchor before a cursor move: extend
// keeps/starts a selection, otherwise any existing selection is cleared.
func (b *Buffer) beginMove(extend bool) {
	if extend {
		if !b.HasSel {
			b.SelAnchor = b.Cursor
			b.HasSel = true
		}
	} else {
		b.HasSel = false
	}
}

func (b *Buffer) MoveLeft(extend bool) {
	b.beginMove(extend)
	p := b.Cursor
	if p.Col > 0 {
		p.Col--
	} else if p.Line > 0 {
		p.Line--
		p.Col = len(b.Lines[p.Line])
	}
	b.Cursor = p
}

func (b *Buffer) MoveRight(extend bool) {
	b.beginMove(extend)
	p := b.Cursor
	if p.Col < len(b.Lines[p.Line]) {
		p.Col++
	} else if p.Line < len(b.Lines)-1 {
		p.Line++
		p.Col = 0
	}
	b.Cursor = p
}

func (b *Buffer) MoveUp(extend bool) {
	b.beginMove(extend)
	if b.Cursor.Line == 0 {
		b.Cursor.Col = 0
		return
	}
	b.Cursor.Line--
	b.Cursor = b.clampPosition(b.Cursor)
}

func (b *Buffer) MoveDown(extend bool) {
	b.beginMove(extend)
	if b.Cursor.Line == len(b.Lines)-1 {
		b.Cursor.Col = len(b.Lines[b.Cursor.Line])
		return
	}
	b.Cursor.Line++
	b.Cursor = b.clampPosition(b.Cursor)
}

func (b *Buffer) MoveHome(extend bool) {
	b.beginMove(extend)
	line := b.Lines[b.Cursor.Line]
	firstNonBlank := 0
	for firstNonBlank < len(line) && unicode.IsSpace(line[firstNonBlank]) {
		firstNonBlank++
	}
	if b.Cursor.Col == firstNonBlank {
		b.Cursor.Col = 0
	} else {
		b.Cursor.Col = firstNonBlank
	}
}

func (b *Buffer) MoveEnd(extend bool) {
	b.beginMove(extend)
	b.Cursor.Col = len(b.Lines[b.Cursor.Line])
}

func (b *Buffer) MoveTo(p Position, extend bool) {
	b.beginMove(extend)
	b.Cursor = b.clampPosition(p)
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// MoveWordLeft jumps to the start of the previous word.
func (b *Buffer) MoveWordLeft(extend bool) {
	b.beginMove(extend)
	p := b.Cursor
	for {
		if p.Col == 0 {
			if p.Line == 0 {
				b.Cursor = p
				return
			}
			p.Line--
			p.Col = len(b.Lines[p.Line])
			if p.Col > 0 {
				break
			}
			continue
		}
		break
	}
	line := b.Lines[p.Line]
	// skip whitespace
	for p.Col > 0 && unicode.IsSpace(line[p.Col-1]) {
		p.Col--
	}
	if p.Col > 0 {
		word := isWordRune(line[p.Col-1])
		for p.Col > 0 && isWordRune(line[p.Col-1]) == word && !unicode.IsSpace(line[p.Col-1]) {
			p.Col--
		}
	}
	b.Cursor = p
}

// MoveWordRight jumps to the start of the next word.
func (b *Buffer) MoveWordRight(extend bool) {
	b.beginMove(extend)
	p := b.Cursor
	line := b.Lines[p.Line]
	if p.Col == len(line) {
		if p.Line < len(b.Lines)-1 {
			p.Line++
			p.Col = 0
		}
		b.Cursor = p
		return
	}
	word := isWordRune(line[p.Col])
	for p.Col < len(line) && isWordRune(line[p.Col]) == word && !unicode.IsSpace(line[p.Col]) {
		p.Col++
	}
	for p.Col < len(line) && unicode.IsSpace(line[p.Col]) {
		p.Col++
	}
	b.Cursor = p
}
