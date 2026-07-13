// Package editor implements the dino terminal UI: screen layout,
// rendering, input handling, file tree, search and the embedded
// terminal. It builds on top of internal/buffer for text storage.
package editor

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	"dino/internal/buffer"
)

type Focus int

const (
	FocusEditor Focus = iota
	FocusTree
)

// Mode selects what the bottom input line is currently doing.
type Mode int

const (
	ModeNormal Mode = iota
	ModeSearch
	ModeSaveAsPrompt
	ModeNewFilePrompt
	ModeQuitConfirm
	ModeQuitConfirmSaveAsPrompt
)

// Config holds user-tunable editor settings.
type Config struct {
	TabSize   int
	UseSpaces bool
}

func DefaultConfig() Config {
	return Config{TabSize: 4, UseSpaces: true}
}

// Editor is the top-level application state.
type Editor struct {
	Screen tcell.Screen
	Config Config

	Tabs   []*Tab
	Active int

	Focus Focus
	Mode  Mode

	Tree *FileTree

	// ShowHelp toggles the keybinding overlay (F1).
	ShowHelp bool

	// PromptInput accumulates text typed while Mode != ModeNormal.
	PromptInput string
	// StatusMsg is a transient one-line message shown in the status bar.
	StatusMsg string

	// pendingQuitTab is set when Ctrl+Q's save-prompt is answered "y" and
	// we need to know which tab (and whether to quit afterwards) to save.
	pendingQuitAll bool

	Search SearchState

	// internalClipboard is used as a fallback when the OS clipboard is
	// unavailable (e.g. headless/no clipboard utility installed).
	internalClipboard string

	// pasting is true between a bracketed-paste Start/End pair reported
	// by the terminal. Incoming key events are buffered into pasteBuf
	// instead of running through the normal per-key handlers, which
	// would otherwise re-trigger auto-indent/auto-pair once for every
	// pasted character and produce runaway "staircase" indentation.
	pasting  bool
	pasteBuf strings.Builder

	quit bool
}

func New(screen tcell.Screen, cfg Config) *Editor {
	e := &Editor{
		Screen: screen,
		Config: cfg,
		Focus:  FocusEditor,
	}
	wd := "."
	e.Tree = NewFileTree(wd)
	return e
}

// CurTab returns the active tab, or nil if none are open.
func (e *Editor) CurTab() *Tab {
	if e.Active < 0 || e.Active >= len(e.Tabs) {
		return nil
	}
	return e.Tabs[e.Active]
}

// OpenBuffer adds a new tab for the given buffer and focuses it. If a tab
// with the same display title (e.g. "main.c") is already open, that tab is
// focused instead so the tab bar never shows the same name twice.
func (e *Editor) OpenBuffer(b *buffer.Buffer) {
	title := shortPath(b.FilePath)
	for i, t := range e.Tabs {
		if b.FilePath != "" && t.Buf.FilePath != "" && t.Title() == title {
			e.Active = i
			return
		}
	}
	b.TabSize = e.Config.TabSize
	e.Tabs = append(e.Tabs, NewTab(b))
	e.Active = len(e.Tabs) - 1
}

// OpenFile loads path (or focuses it if already open) into a new tab.
func (e *Editor) OpenFile(path string) error {
	title := shortPath(path)
	for i, t := range e.Tabs {
		if t.Buf.FilePath == path || (t.Buf.FilePath != "" && t.Title() == title) {
			e.Active = i
			return nil
		}
	}
	b, err := buffer.Load(path)
	if err != nil {
		return err
	}
	b.ExpandTabs(e.Config.TabSize)
	e.OpenBuffer(b)
	return nil
}

func (e *Editor) SetStatus(msg string) {
	e.StatusMsg = msg
}

// Run drives the main event loop until the user quits.
func (e *Editor) Run() error {
	if len(e.Tabs) == 0 {
		e.OpenBuffer(buffer.New())
	}
	e.Draw()
	for !e.quit {
		ev := e.Screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			e.Screen.Sync()
		case *tcell.EventKey:
			e.handleKey(ev)
		case *tcell.EventMouse:
			e.handleMouse(ev)
		case *tcell.EventPaste:
			e.handlePasteEvent(ev)
		}
		if e.quit {
			break
		}
		e.Draw()
	}
	return nil
}
