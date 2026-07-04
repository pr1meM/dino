package editor

import "github.com/gdamore/tcell/v2"

func (e *Editor) handlePromptKey(ev *tcell.EventKey) {
	switch e.Mode {
	case ModeQuitConfirm:
		e.handleQuitConfirmKey(ev)
		return
	case ModeSaveAsPrompt, ModeQuitConfirmSaveAsPrompt:
		e.handleSaveAsKey(ev)
		return
	case ModeNewFilePrompt:
		e.handleNewFileKey(ev)
		return
	case ModeSearch:
		e.handleSearchKey(ev)
		return
	}
}

func (e *Editor) cancelPrompt() {
	e.Mode = ModeNormal
	e.PromptInput = ""
}

func (e *Editor) handleQuitConfirmKey(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		e.cancelPrompt()
		e.pendingQuitAll = false
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'y', 'Y':
			if err := e.saveAllDirty(); err != nil {
				e.PromptInput = ""
				e.Mode = ModeQuitConfirmSaveAsPrompt
				e.SetStatus("")
				return
			}
			e.cancelPrompt()
			if e.pendingQuitAll {
				e.quit = true
			}
		case 'n', 'N':
			e.cancelPrompt()
			if e.pendingQuitAll {
				e.quit = true
			}
		}
	}
}

// saveAllDirty tries to save every dirty tab. If any tab has no
// FilePath yet, it returns an error so the caller can prompt for a path.
func (e *Editor) saveAllDirty() error {
	for i, t := range e.Tabs {
		if !t.Buf.Dirty {
			continue
		}
		if t.Buf.FilePath == "" {
			e.Active = i
			return errNeedsPath
		}
		if err := t.Buf.Save(); err != nil {
			return err
		}
	}
	return nil
}

func (e *Editor) handleSaveAsKey(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		e.cancelPrompt()
		e.pendingQuitAll = false
	case tcell.KeyEnter:
		path := e.PromptInput
		wasQuitFlow := e.Mode == ModeQuitConfirmSaveAsPrompt
		t := e.CurTab()
		if t != nil && path != "" {
			if err := t.Buf.SaveAs(path); err != nil {
				e.SetStatus("Error: " + err.Error())
				e.cancelPrompt()
				return
			}
		}
		e.cancelPrompt()
		if wasQuitFlow {
			switch err := e.saveAllDirty(); err {
			case nil:
				if e.pendingQuitAll {
					e.quit = true
				}
			case errNeedsPath:
				e.Mode = ModeQuitConfirmSaveAsPrompt
				e.PromptInput = ""
			}
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.PromptInput) > 0 {
			e.PromptInput = e.PromptInput[:len(e.PromptInput)-1]
		}
	case tcell.KeyRune:
		e.PromptInput += string(ev.Rune())
	}
}

func (e *Editor) handleNewFileKey(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		e.cancelPrompt()
	case tcell.KeyEnter:
		path := e.PromptInput
		e.cancelPrompt()
		if path != "" {
			e.newFileNamed(path)
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.PromptInput) > 0 {
			e.PromptInput = e.PromptInput[:len(e.PromptInput)-1]
		}
	case tcell.KeyRune:
		e.PromptInput += string(ev.Rune())
	}
}
