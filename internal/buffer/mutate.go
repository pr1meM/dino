package buffer

// setCursorNoSel moves the cursor and clears any active selection.
func (b *Buffer) setCursorNoSel(p Position) {
	b.Cursor = b.clampPosition(p)
	b.HasSel = false
}

// DeleteSelection removes the currently selected text, if any, and
// returns the removed text. It is a no-op returning "" when there is no
// selection.
func (b *Buffer) DeleteSelection() string {
	if !b.HasSel {
		return ""
	}
	start, end := OrderedSelection(b.SelAnchor, b.Cursor)
	removed := b.deleteRange(start, end)
	b.recordDelete(start, removed, false)
	b.setCursorNoSel(start)
	b.Dirty = true
	b.Version++
	return removed
}

// InsertText inserts arbitrary text (may contain '\n') at the cursor,
// replacing the selection first if one is active.
func (b *Buffer) InsertText(text string) {
	if text == "" {
		return
	}
	if b.HasSel {
		b.DeleteSelection()
	}
	pos := b.Cursor
	b.insertTextAt(pos, text)
	if len(text) == 1 {
		b.recordInsert(pos, text)
	} else {
		b.pushUndo(editOp{kind: opInsert, pos: pos, text: text})
	}
	b.setCursorNoSel(advance(pos, text))
	b.Dirty = true
	b.Version++
}

// InsertRune inserts a single rune at the cursor.
func (b *Buffer) InsertRune(r rune) {
	b.InsertText(string(r))
}

// Backspace deletes the character before the cursor (or the selection).
func (b *Buffer) Backspace() {
	if b.HasSel {
		b.DeleteSelection()
		return
	}
	pos := b.Cursor
	var start Position
	if pos.Col > 0 {
		start = Position{Line: pos.Line, Col: pos.Col - 1}
	} else if pos.Line > 0 {
		start = Position{Line: pos.Line - 1, Col: len(b.Lines[pos.Line-1])}
	} else {
		return
	}
	removed := b.deleteRange(start, pos)
	b.recordDelete(start, removed, true)
	b.setCursorNoSel(start)
	b.Dirty = true
	b.Version++
}

// DeleteForward deletes the character after the cursor (or the
// selection); this is the classic "Delete" key.
func (b *Buffer) DeleteForward() {
	if b.HasSel {
		b.DeleteSelection()
		return
	}
	pos := b.Cursor
	var end Position
	if pos.Col < len(b.Lines[pos.Line]) {
		end = Position{Line: pos.Line, Col: pos.Col + 1}
	} else if pos.Line < len(b.Lines)-1 {
		end = Position{Line: pos.Line + 1, Col: 0}
	} else {
		return
	}
	removed := b.deleteRange(pos, end)
	b.recordDelete(pos, removed, false)
	b.setCursorNoSel(pos)
	b.Dirty = true
	b.Version++
}

// SelectedText returns the text currently selected, or "" if none.
func (b *Buffer) SelectedText() string {
	if !b.HasSel {
		return ""
	}
	start, end := OrderedSelection(b.SelAnchor, b.Cursor)
	if start.Line == end.Line {
		return string(b.Lines[start.Line][start.Col:end.Col])
	}
	var out []rune
	out = append(out, b.Lines[start.Line][start.Col:]...)
	for l := start.Line + 1; l < end.Line; l++ {
		out = append(out, '\n')
		out = append(out, b.Lines[l]...)
	}
	out = append(out, '\n')
	out = append(out, b.Lines[end.Line][:end.Col]...)
	return string(out)
}

// SelectAll selects the entire buffer content.
func (b *Buffer) SelectAll() {
	b.SelAnchor = Position{Line: 0, Col: 0}
	last := len(b.Lines) - 1
	b.Cursor = Position{Line: last, Col: len(b.Lines[last])}
	b.HasSel = true
}

// ClearSelection drops any active selection without changing the cursor.
func (b *Buffer) ClearSelection() {
	b.HasSel = false
}
