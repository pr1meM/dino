package buffer

import "time"

type opKind int

const (
	opInsert opKind = iota
	opDelete
)

// editOp is one undoable group. For opDelete, growLeft distinguishes a
// backspace chain (the deletion boundary moves left as the group grows,
// e.g. repeated Backspace) from a forward-delete chain (the boundary
// stays fixed, e.g. repeated Delete).
type editOp struct {
	kind     opKind
	pos      Position
	text     string
	growLeft bool
	at       time.Time
}

const mergeWindow = 2 * time.Second

func (b *Buffer) pushUndo(op editOp) {
	b.redo = nil
	b.undo = append(b.undo, op)
}

func (b *Buffer) lastOp() *editOp {
	if len(b.undo) == 0 {
		return nil
	}
	return &b.undo[len(b.undo)-1]
}

// recordInsert registers a single-position text insertion, merging into
// the previous op when it directly extends it.
func (b *Buffer) recordInsert(pos Position, text string) {
	if last := b.lastOp(); last != nil && last.kind == opInsert &&
		time.Since(last.at) < mergeWindow &&
		advance(last.pos, last.text).Equal(pos) &&
		text != "\n" {
		last.text += text
		last.at = time.Now()
		b.redo = nil
		return
	}
	b.pushUndo(editOp{kind: opInsert, pos: pos, text: text, at: time.Now()})
}

// recordDelete registers a deletion of text that originally lived at pos
// (document position before the delete). growLeft indicates a
// backspace-style chain; see editOp.
func (b *Buffer) recordDelete(pos Position, text string, growLeft bool) {
	if last := b.lastOp(); last != nil && last.kind == opDelete && last.growLeft == growLeft &&
		time.Since(last.at) < mergeWindow {
		if growLeft && advance(pos, text).Equal(last.pos) {
			last.pos = pos
			last.text = text + last.text
			last.at = time.Now()
			b.redo = nil
			return
		}
		if !growLeft && pos.Equal(last.pos) {
			last.text += text
			last.at = time.Now()
			b.redo = nil
			return
		}
	}
	b.pushUndo(editOp{kind: opDelete, pos: pos, text: text, growLeft: growLeft, at: time.Now()})
}

// CanUndo/CanRedo report whether there is anything to undo/redo.
func (b *Buffer) CanUndo() bool { return len(b.undo) > 0 }
func (b *Buffer) CanRedo() bool { return len(b.redo) > 0 }

// Undo reverts the most recent edit group.
func (b *Buffer) Undo() {
	if len(b.undo) == 0 {
		return
	}
	op := b.undo[len(b.undo)-1]
	b.undo = b.undo[:len(b.undo)-1]

	switch op.kind {
	case opInsert:
		b.deleteRange(op.pos, advance(op.pos, op.text))
		b.setCursorNoSel(op.pos)
	case opDelete:
		b.insertTextAt(op.pos, op.text)
		if op.growLeft {
			b.setCursorNoSel(advance(op.pos, op.text))
		} else {
			b.setCursorNoSel(op.pos)
		}
	}
	b.redo = append(b.redo, op)
	b.Dirty = len(b.undo) > 0
	b.Version++
}

// Redo re-applies the most recently undone edit group.
func (b *Buffer) Redo() {
	if len(b.redo) == 0 {
		return
	}
	op := b.redo[len(b.redo)-1]
	b.redo = b.redo[:len(b.redo)-1]

	switch op.kind {
	case opInsert:
		b.insertTextAt(op.pos, op.text)
		b.setCursorNoSel(advance(op.pos, op.text))
	case opDelete:
		b.deleteRange(op.pos, advance(op.pos, op.text))
		b.setCursorNoSel(op.pos)
	}
	b.undo = append(b.undo, op)
	b.Dirty = true
	b.Version++
}
