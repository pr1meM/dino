package editor

import (
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
)

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
		t.Buf.InsertText(reindentPaste(t.Buf, text))
	}
}

// handlePasteEvent marks the start/end of a terminal bracketed paste. The
// keys arriving between them are buffered by bufferPasteKey rather than
// dispatched to the normal key handlers; on End the buffered text is
// reindented (as with a clipboard paste) and inserted in one shot.
func (e *Editor) handlePasteEvent(ev *tcell.EventPaste) {
	if ev.Start() {
		e.pasting = true
		e.pasteBuf.Reset()
		return
	}
	e.pasting = false
	text := e.pasteBuf.String()
	e.pasteBuf.Reset()
	if text == "" {
		return
	}
	if e.Mode != ModeNormal {
		// Prompt inputs are single-line; drop embedded newlines rather
		// than letting one submit the prompt early.
		e.PromptInput += strings.ReplaceAll(text, "\n", "")
		return
	}
	t := e.CurTab()
	if t == nil {
		return
	}
	t.Buf.InsertText(reindentPaste(t.Buf, text))
}

// bufferPasteKey accumulates one key event belonging to an in-progress
// bracketed paste. Keys other than plain characters/newline/tab (e.g. any
// escape sequences embedded in the pasted bytes) are dropped rather than
// acted on, which also keeps a paste containing stray control codes from
// being interpreted as a command.
func (e *Editor) bufferPasteKey(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEnter, tcell.KeyLF:
		e.pasteBuf.WriteByte('\n')
	case tcell.KeyTab:
		e.pasteBuf.WriteByte('\t')
	case tcell.KeyRune:
		e.pasteBuf.WriteRune(ev.Rune())
	}
}
