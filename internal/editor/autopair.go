package editor

// insertAutoPaired inserts r, additionally inserting a matching closer
// for brackets/quotes, or skipping over an already-present closer when
// the user types it themselves right after the cursor.
func (e *Editor) insertAutoPaired(t *Tab, r rune) {
	b := t.Buf
	closers := map[rune]rune{'(': ')', '[': ']', '{': '}', '"': '"', '\'': '\''}

	if closer, isOpener := closers[r]; isOpener && !b.HasSel {
		b.InsertRune(r)
		b.InsertRune(closer)
		b.MoveLeft(false)
		return
	}

	if isCloser(r) && !b.HasSel {
		line := b.Line(b.Cursor.Line)
		if b.Cursor.Col < len(line) && line[b.Cursor.Col] == r {
			b.MoveRight(false)
			return
		}
		e.maybeDedentForCloser(t, r)
	}

	b.InsertRune(r)
}

func isCloser(r rune) bool {
	switch r {
	case ')', ']', '}', '"', '\'':
		return true
	}
	return false
}
