package editor

import (
	"github.com/gdamore/tcell/v2"

	"dino/internal/buffer"
)

func (e *Editor) handleMouse(ev *tcell.EventMouse) {
	l := e.computeLayout()
	x, y := ev.Position()
	btn := ev.Buttons()

	if btn&tcell.WheelUp != 0 {
		e.scrollAt(x, l, -3)
		return
	}
	if btn&tcell.WheelDown != 0 {
		e.scrollAt(x, l, 3)
		return
	}
	if btn&tcell.Button1 == 0 {
		return
	}

	if e.Tree.Visible && x < l.treeWidth {
		e.Focus = FocusTree
		row := y - l.textY0
		idx := e.Tree.Scroll + row
		if idx >= 0 && idx < len(e.Tree.flat) {
			e.Tree.Cursor = idx
			if path := e.Tree.Activate(); path != "" {
				if err := e.OpenFile(path); err == nil {
					e.Focus = FocusEditor
				}
			}
		}
		return
	}

	if y >= l.textY0 && y <= l.textY1 {
		e.Focus = FocusEditor
		t := e.CurTab()
		if t == nil {
			return
		}
		line := t.TopLine + (y - l.textY0)
		col := t.LeftCol + (x - l.textX1)
		if line >= t.Buf.LineCount() {
			line = t.Buf.LineCount() - 1
		}
		if col < 0 {
			col = 0
		}
		t.Buf.MoveTo(buffer.Position{Line: line, Col: col}, false)
	}
}

func (e *Editor) scrollAt(x int, l layout, delta int) {
	if e.Tree.Visible && x < l.treeWidth {
		e.Tree.Scroll += delta
		if e.Tree.Scroll < 0 {
			e.Tree.Scroll = 0
		}
		return
	}
	t := e.CurTab()
	if t == nil {
		return
	}
	t.TopLine += delta
	if t.TopLine < 0 {
		t.TopLine = 0
	}
	if t.TopLine > t.Buf.LineCount()-1 {
		t.TopLine = t.Buf.LineCount() - 1
	}
}
