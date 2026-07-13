package editor

import (
	"os"

	"github.com/gdamore/tcell/v2"

	"dino/internal/buffer"
)

func (e *Editor) handleKey(ev *tcell.EventKey) {
	if e.pasting {
		e.bufferPasteKey(ev)
		return
	}
	if e.ShowHelp {
		e.ShowHelp = false
		return
	}
	if e.Mode != ModeNormal {
		e.handlePromptKey(ev)
		return
	}
	if e.Focus == FocusTree {
		if e.handleTreeKey(ev) {
			return
		}
	}
	if e.handleGlobalKey(ev) {
		return
	}
	if e.Focus == FocusEditor {
		e.handleEditorKey(ev)
	}
}

// handleGlobalKey handles bindings active regardless of focus. Returns
// true if the event was consumed.
func (e *Editor) handleGlobalKey(ev *tcell.EventKey) bool {
	if ev.Modifiers()&tcell.ModAlt != 0 {
		switch ev.Key() {
		case tcell.KeyLeft:
			e.prevTab()
			return true
		case tcell.KeyRight:
			e.nextTab()
			return true
		}
	}
	switch ev.Key() {
	case tcell.KeyCtrlH:
		e.ShowHelp = true
		return true
	case tcell.KeyCtrlN:
		e.Mode = ModeNewFilePrompt
		e.PromptInput = ""
		return true
	case tcell.KeyCtrlB:
		e.Tree.Toggle()
		if e.Tree.Visible {
			e.Focus = FocusTree
		} else {
			e.Focus = FocusEditor
		}
		return true
	case tcell.KeyCtrlW, tcell.KeyTab:
		if e.Tree.Visible {
			if e.Focus == FocusEditor {
				e.Focus = FocusTree
			} else {
				e.Focus = FocusEditor
			}
			return true
		}
	case tcell.KeyCtrlQ:
		e.requestQuit()
		return true
	case tcell.KeyCtrlS:
		e.saveCurrent()
		return true
	case tcell.KeyCtrlUnderscore:
		e.toggleComment()
		return true
	case tcell.KeyCtrlF:
		e.startSearch()
		return true
	case tcell.KeyCtrlC:
		e.copySelection()
		return true
	case tcell.KeyCtrlX:
		e.cutSelection()
		return true
	case tcell.KeyCtrlV:
		e.pasteClipboard()
		return true
	case tcell.KeyCtrlA:
		if t := e.CurTab(); t != nil {
			t.Buf.SelectAll()
		}
		return true
	case tcell.KeyCtrlZ:
		if t := e.CurTab(); t != nil {
			t.Buf.Undo()
		}
		return true
	case tcell.KeyCtrlY:
		if t := e.CurTab(); t != nil {
			t.Buf.Redo()
		}
		return true
	}
	return false
}

func (e *Editor) handleTreeKey(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyUp:
		e.Tree.MoveUp()
		return true
	case tcell.KeyDown:
		e.Tree.MoveDown()
		return true
	case tcell.KeyEnter:
		if path := e.Tree.Activate(); path != "" {
			if err := e.OpenFile(path); err != nil {
				e.SetStatus("Error: " + err.Error())
			} else {
				e.Focus = FocusEditor
			}
		}
		return true
	}
	return false
}

func (e *Editor) handleEditorKey(ev *tcell.EventKey) {
	t := e.CurTab()
	if t == nil {
		return
	}
	b := t.Buf
	shift := ev.Modifiers()&tcell.ModShift != 0

	switch ev.Key() {
	case tcell.KeyLeft:
		if ev.Modifiers()&tcell.ModCtrl != 0 {
			b.MoveWordLeft(shift)
		} else {
			b.MoveLeft(shift)
		}
	case tcell.KeyRight:
		if ev.Modifiers()&tcell.ModCtrl != 0 {
			b.MoveWordRight(shift)
		} else {
			b.MoveRight(shift)
		}
	case tcell.KeyUp:
		b.MoveUp(shift)
	case tcell.KeyDown:
		b.MoveDown(shift)
	case tcell.KeyHome:
		b.MoveHome(shift)
	case tcell.KeyEnd:
		b.MoveEnd(shift)
	case tcell.KeyPgUp:
		e.movePage(t, -1, shift)
	case tcell.KeyPgDn:
		e.movePage(t, 1, shift)
	case tcell.KeyEnter:
		e.insertNewline(t)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		b.Backspace()
	case tcell.KeyDelete:
		b.DeleteForward()
	case tcell.KeyTab:
		e.insertTab(t)
	case tcell.KeyRune:
		e.insertAutoPaired(t, ev.Rune())
	}
}

func (e *Editor) movePage(t *Tab, dir int, extend bool) {
	l := e.computeLayout()
	rows := l.textY1 - l.textY0 + 1
	for i := 0; i < rows; i++ {
		if dir < 0 {
			t.Buf.MoveUp(extend)
		} else {
			t.Buf.MoveDown(extend)
		}
	}
}

func (e *Editor) insertNewline(t *Tab) {
	e.autoIndentNewline(t)
}

func (e *Editor) insertTab(t *Tab) {
	if e.Config.UseSpaces {
		n := e.Config.TabSize - (t.Buf.Cursor.Col % e.Config.TabSize)
		for i := 0; i < n; i++ {
			t.Buf.InsertRune(' ')
		}
	} else {
		t.Buf.InsertRune('\t')
	}
}

func (e *Editor) nextTab() {
	if len(e.Tabs) == 0 {
		return
	}
	e.Active = (e.Active + 1) % len(e.Tabs)
}

func (e *Editor) prevTab() {
	if len(e.Tabs) == 0 {
		return
	}
	e.Active = (e.Active - 1 + len(e.Tabs)) % len(e.Tabs)
}

// newFileNamed opens path in a new tab: an existing file is loaded as
// usual, a path that doesn't exist yet becomes a fresh untitled buffer
// pre-named path (created on disk on first save, same as passing an
// unwritten path on the command line).
func (e *Editor) newFileNamed(path string) {
	if err := e.OpenFile(path); err != nil {
		if !os.IsNotExist(err) {
			e.SetStatus("Error: " + err.Error())
			return
		}
		b := buffer.New()
		b.FilePath = path
		e.OpenBuffer(b)
	}
	e.Focus = FocusEditor
}

func (e *Editor) saveCurrent() {
	t := e.CurTab()
	if t == nil {
		return
	}
	if t.Buf.FilePath == "" {
		e.Mode = ModeSaveAsPrompt
		e.PromptInput = ""
		return
	}
	if err := t.Buf.Save(); err != nil {
		e.SetStatus("Error saving: " + err.Error())
		return
	}
	e.SetStatus("Saved: " + t.Buf.FilePath)
}

func (e *Editor) requestQuit() {
	if dirty := e.dirtyTabs(); len(dirty) > 0 {
		e.Mode = ModeQuitConfirm
		e.pendingQuitAll = true
		return
	}
	e.quit = true
}

func (e *Editor) dirtyTabs() []*Tab {
	var out []*Tab
	for _, t := range e.Tabs {
		if t.Buf.Dirty {
			out = append(out, t)
		}
	}
	return out
}
