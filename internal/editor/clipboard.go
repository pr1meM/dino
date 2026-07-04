package editor

import "github.com/atotto/clipboard"

// setClipboard writes to the OS clipboard, falling back to an
// in-process clipboard when no OS clipboard utility is available.
func (e *Editor) setClipboard(text string) {
	e.internalClipboard = text
	_ = clipboard.WriteAll(text)
}

// getClipboard reads the OS clipboard, falling back to the in-process
// clipboard when the OS clipboard is unavailable or empty.
func (e *Editor) getClipboard() string {
	if text, err := clipboard.ReadAll(); err == nil && text != "" {
		return text
	}
	return e.internalClipboard
}

func (e *Editor) copySelection() {
	t := e.CurTab()
	if t == nil {
		return
	}
	if text := t.Buf.SelectedText(); text != "" {
		e.setClipboard(text)
	}
}

func (e *Editor) cutSelection() {
	t := e.CurTab()
	if t == nil {
		return
	}
	if text := t.Buf.SelectedText(); text != "" {
		e.setClipboard(text)
		t.Buf.DeleteSelection()
	}
}

func (e *Editor) pasteClipboard() {
	t := e.CurTab()
	if t == nil {
		return
	}
	if text := e.getClipboard(); text != "" {
		t.Buf.InsertText(text)
	}
}
