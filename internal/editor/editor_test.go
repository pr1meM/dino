package editor

import (
	"os"
	"testing"

	"github.com/gdamore/tcell/v2"

	"dino/internal/buffer"
)

func newTestEditor(t *testing.T) (*Editor, tcell.SimulationScreen) {
	t.Helper()
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(80, 24)
	e := New(s, DefaultConfig())
	e.OpenBuffer(buffer.New())
	return e, s
}

func key(k tcell.Key, r rune, mod tcell.ModMask) *tcell.EventKey {
	return tcell.NewEventKey(k, r, mod)
}

// typeString simulates typing s, one key event per rune. A literal '\n'
// is sent as a KeyEnter event (as a real terminal would for the Enter
// key), not as a raw KeyRune('\n') — tcell reserves rune 0x0A for KeyLF,
// a distinct key from KeyEnter (0x0D).
func typeString(e *Editor, s string) {
	for _, r := range s {
		if r == '\n' {
			e.handleKey(key(tcell.KeyEnter, 0, 0))
			continue
		}
		e.handleKey(key(tcell.KeyRune, r, 0))
	}
}

func TestTypingAndDraw(t *testing.T) {
	e, s := newTestEditor(t)
	typeString(e, "hello")
	e.Draw() // must not panic
	s.Sync()

	b := e.CurTab().Buf
	if string(b.Lines[0]) != "hello" {
		t.Fatalf("line = %q", string(b.Lines[0]))
	}
}

func TestEnterAndAutoIndent(t *testing.T) {
	e, _ := newTestEditor(t)
	typeString(e, "  if x:")
	e.handleKey(key(tcell.KeyEnter, 0, 0))
	typeString(e, "pass")
	b := e.CurTab().Buf
	if b.LineCount() != 2 {
		t.Fatalf("line count = %d", b.LineCount())
	}
	if string(b.Lines[1]) != "      pass" {
		t.Fatalf("indented line = %q", string(b.Lines[1]))
	}
}

func TestAutoIndentAfterBrace(t *testing.T) {
	e, _ := newTestEditor(t)
	typeString(e, "func main() {")
	e.handleKey(key(tcell.KeyEnter, 0, 0))
	typeString(e, "x")
	b := e.CurTab().Buf
	if string(b.Lines[1]) != "    x" {
		t.Fatalf("indented line = %q", string(b.Lines[1]))
	}
}

func TestConfigurableTabSize(t *testing.T) {
	e, _ := newTestEditor(t)
	e.Config.TabSize = 2
	typeString(e, "if x:")
	e.handleKey(key(tcell.KeyEnter, 0, 0))
	typeString(e, "y")
	b := e.CurTab().Buf
	if string(b.Lines[1]) != "  y" {
		t.Fatalf("indented line with tabsize=2 = %q", string(b.Lines[1]))
	}
}

func TestSmartBracketNewline(t *testing.T) {
	// Typing '{' auto-pairs to "{}" with the cursor in between; pressing
	// Enter there should split into three lines with the closer
	// dedented back to the opening line's indent.
	e, _ := newTestEditor(t)
	typeString(e, "func main() {")
	e.handleKey(key(tcell.KeyEnter, 0, 0))
	b := e.CurTab().Buf
	if b.LineCount() != 3 {
		t.Fatalf("line count = %d, want 3", b.LineCount())
	}
	if string(b.Lines[1]) != "    " || string(b.Lines[2]) != "}" {
		t.Fatalf("lines = %q / %q", string(b.Lines[1]), string(b.Lines[2]))
	}
	if b.Cursor.Line != 1 || b.Cursor.Col != 4 {
		t.Fatalf("cursor = %+v, want on the indented middle line", b.Cursor)
	}
}

func TestDedentOnClosingBrace(t *testing.T) {
	// Simulate an over-indented empty line (e.g. loaded from a file)
	// and confirm typing '}' as the first character dedents it.
	e, _ := newTestEditor(t)
	b := e.CurTab().Buf
	b.InsertText("if x {\n    y\n    ")
	e.handleKey(key(tcell.KeyRune, '}', 0))
	if string(b.Lines[2]) != "}" {
		t.Fatalf("closing line = %q, want dedented '}'", string(b.Lines[2]))
	}
}

func TestAutoPairBrackets(t *testing.T) {
	e, _ := newTestEditor(t)
	typeString(e, "(")
	b := e.CurTab().Buf
	if string(b.Lines[0]) != "()" {
		t.Fatalf("line = %q", string(b.Lines[0]))
	}
	if b.Cursor.Col != 1 {
		t.Fatalf("cursor col = %d, want 1 (between brackets)", b.Cursor.Col)
	}
}

func TestFileTreeToggleAndFocus(t *testing.T) {
	e, _ := newTestEditor(t)
	if e.Tree.Visible {
		t.Fatal("tree should start hidden")
	}
	e.handleKey(key(tcell.KeyCtrlB, 0, 0))
	if !e.Tree.Visible || e.Focus != FocusTree {
		t.Fatalf("after Ctrl+B: visible=%v focus=%v", e.Tree.Visible, e.Focus)
	}
	e.handleKey(key(tcell.KeyCtrlW, 0, 0))
	if e.Focus != FocusEditor {
		t.Fatalf("after Ctrl+W: focus=%v, want FocusEditor", e.Focus)
	}
	e.Draw()
}

func TestQuitWithoutDirtyBuffer(t *testing.T) {
	e, _ := newTestEditor(t)
	e.handleKey(key(tcell.KeyCtrlQ, 0, 0))
	if !e.quit {
		t.Fatal("expected quit=true for clean buffer")
	}
}

func TestQuitWithDirtyBufferPromptsThenDiscards(t *testing.T) {
	e, _ := newTestEditor(t)
	typeString(e, "x")
	e.handleKey(key(tcell.KeyCtrlQ, 0, 0))
	if e.Mode != ModeQuitConfirm {
		t.Fatalf("mode = %v, want ModeQuitConfirm", e.Mode)
	}
	if e.quit {
		t.Fatal("should not quit yet, awaiting confirmation")
	}
	e.handleKey(key(tcell.KeyRune, 'n', 0))
	if !e.quit {
		t.Fatal("expected quit=true after choosing 'n'")
	}
}

func TestQuitEscCancels(t *testing.T) {
	e, _ := newTestEditor(t)
	typeString(e, "x")
	e.handleKey(key(tcell.KeyCtrlQ, 0, 0))
	e.handleKey(key(tcell.KeyEscape, 0, 0))
	if e.Mode != ModeNormal || e.quit {
		t.Fatalf("mode=%v quit=%v, want ModeNormal/false", e.Mode, e.quit)
	}
}

func TestSaveExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/f.txt"
	e, _ := newTestEditor(t)
	e.CurTab().Buf.FilePath = path
	typeString(e, "content")
	e.handleKey(key(tcell.KeyCtrlS, 0, 0))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "content\n" {
		t.Fatalf("saved content = %q", string(data))
	}
	if e.CurTab().Buf.Dirty {
		t.Fatal("expected buffer to be clean after save")
	}
}

func TestSaveAsPromptForUntitledBuffer(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/new.txt"
	e, _ := newTestEditor(t)
	typeString(e, "hi")
	e.handleKey(key(tcell.KeyCtrlS, 0, 0))
	if e.Mode != ModeSaveAsPrompt {
		t.Fatalf("mode = %v, want ModeSaveAsPrompt", e.Mode)
	}
	typeString(e, path)
	e.handleKey(key(tcell.KeyEnter, 0, 0))
	if e.Mode != ModeNormal {
		t.Fatalf("mode after save-as = %v", e.Mode)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hi\n" {
		t.Fatalf("saved content = %q", string(data))
	}
}

func TestQuitConfirmSavesUntitledBufferThenQuits(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/out.txt"
	e, _ := newTestEditor(t)
	typeString(e, "unsaved")
	e.handleKey(key(tcell.KeyCtrlQ, 0, 0))
	if e.Mode != ModeQuitConfirm {
		t.Fatalf("mode = %v, want ModeQuitConfirm", e.Mode)
	}
	e.handleKey(key(tcell.KeyRune, 'y', 0))
	if e.Mode != ModeQuitConfirmSaveAsPrompt {
		t.Fatalf("mode = %v, want ModeQuitConfirmSaveAsPrompt", e.Mode)
	}
	typeString(e, path)
	e.handleKey(key(tcell.KeyEnter, 0, 0))
	if !e.quit {
		t.Fatal("expected quit=true after saving via quit-confirm flow")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "unsaved\n" {
		t.Fatalf("saved content = %q", string(data))
	}
}

func TestCopyCutPaste(t *testing.T) {
	e, _ := newTestEditor(t)
	typeString(e, "hello world")
	e.CurTab().Buf.MoveTo(buffer.Position{Line: 0, Col: 0}, false)
	e.CurTab().Buf.MoveTo(buffer.Position{Line: 0, Col: 5}, true)

	e.handleKey(key(tcell.KeyCtrlC, 0, 0))
	if e.internalClipboard != "hello" {
		t.Fatalf("clipboard = %q", e.internalClipboard)
	}
	if string(e.CurTab().Buf.Lines[0]) != "hello world" {
		t.Fatalf("copy should not modify text, got %q", string(e.CurTab().Buf.Lines[0]))
	}

	e.CurTab().Buf.MoveTo(buffer.Position{Line: 0, Col: 0}, false)
	e.CurTab().Buf.MoveTo(buffer.Position{Line: 0, Col: 5}, true)
	e.handleKey(key(tcell.KeyCtrlX, 0, 0))
	if string(e.CurTab().Buf.Lines[0]) != " world" {
		t.Fatalf("after cut = %q", string(e.CurTab().Buf.Lines[0]))
	}

	e.handleKey(key(tcell.KeyCtrlV, 0, 0))
	if string(e.CurTab().Buf.Lines[0]) != "hello world" {
		t.Fatalf("after paste = %q", string(e.CurTab().Buf.Lines[0]))
	}
}

func TestSelectAllKeybinding(t *testing.T) {
	e, _ := newTestEditor(t)
	typeString(e, "ab\ncd")
	e.handleKey(key(tcell.KeyCtrlA, 0, 0))
	if !e.CurTab().Buf.HasSel {
		t.Fatal("expected selection after Ctrl+A")
	}
	if got := e.CurTab().Buf.SelectedText(); got != "ab\ncd" {
		t.Fatalf("selected = %q", got)
	}
}

func TestUndoRedoKeybindings(t *testing.T) {
	e, _ := newTestEditor(t)
	typeString(e, "abc")
	e.handleKey(key(tcell.KeyCtrlZ, 0, 0))
	if string(e.CurTab().Buf.Lines[0]) != "" {
		t.Fatalf("after Ctrl+Z = %q", string(e.CurTab().Buf.Lines[0]))
	}
	e.handleKey(key(tcell.KeyCtrlY, 0, 0))
	if string(e.CurTab().Buf.Lines[0]) != "abc" {
		t.Fatalf("after Ctrl+Y = %q", string(e.CurTab().Buf.Lines[0]))
	}
}

func TestAltArrowsSwitchTabs(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/a.txt", []byte("A\n"), 0o644)
	os.WriteFile(dir+"/b.txt", []byte("B\n"), 0o644)
	os.WriteFile(dir+"/c.txt", []byte("C\n"), 0o644)

	e, _ := newTestEditor(t) // starts with one untitled tab
	if err := e.OpenFile(dir + "/a.txt"); err != nil {
		t.Fatal(err)
	}
	if err := e.OpenFile(dir + "/b.txt"); err != nil {
		t.Fatal(err)
	}
	if err := e.OpenFile(dir + "/c.txt"); err != nil {
		t.Fatal(err)
	}
	if len(e.Tabs) != 4 {
		t.Fatalf("tab count = %d, want 4", len(e.Tabs))
	}
	if e.CurTab().Title() != "c.txt" {
		t.Fatalf("active tab = %q, want c.txt (most recently opened)", e.CurTab().Title())
	}

	e.handleKey(key(tcell.KeyLeft, 0, tcell.ModAlt))
	if e.CurTab().Title() != "b.txt" {
		t.Fatalf("after Alt+Left, active = %q, want b.txt", e.CurTab().Title())
	}
	e.handleKey(key(tcell.KeyLeft, 0, tcell.ModAlt))
	if e.CurTab().Title() != "a.txt" {
		t.Fatalf("after 2x Alt+Left, active = %q, want a.txt", e.CurTab().Title())
	}
	e.handleKey(key(tcell.KeyRight, 0, tcell.ModAlt))
	if e.CurTab().Title() != "b.txt" {
		t.Fatalf("after Alt+Right, active = %q, want b.txt", e.CurTab().Title())
	}

	// wraparound
	for i := 0; i < 3; i++ {
		e.handleKey(key(tcell.KeyLeft, 0, tcell.ModAlt))
	}
	if e.CurTab().Title() != "c.txt" {
		t.Fatalf("wraparound: active = %q, want c.txt", e.CurTab().Title())
	}
}

func TestOpeningSameFileTwiceReusesTab(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/a.txt", []byte("A\n"), 0o644)
	e, _ := newTestEditor(t)
	if err := e.OpenFile(dir + "/a.txt"); err != nil {
		t.Fatal(err)
	}
	if err := e.OpenFile(dir + "/a.txt"); err != nil {
		t.Fatal(err)
	}
	if len(e.Tabs) != 2 { // untitled + a.txt, opened twice but reused
		t.Fatalf("tab count = %d, want 2", len(e.Tabs))
	}
}

func TestToggleCommentPythonSingleLine(t *testing.T) {
	e, _ := newTestEditor(t)
	e.CurTab().Buf.FilePath = "script.py"
	typeString(e, "x = 1")
	e.CurTab().Buf.MoveTo(buffer.Position{Line: 0, Col: 0}, false)
	e.handleKey(key(tcell.KeyCtrlUnderscore, 0, 0))
	if got := string(e.CurTab().Buf.Lines[0]); got != "# x = 1" {
		t.Fatalf("commented = %q", got)
	}
	e.handleKey(key(tcell.KeyCtrlUnderscore, 0, 0))
	if got := string(e.CurTab().Buf.Lines[0]); got != "x = 1" {
		t.Fatalf("uncommented = %q", got)
	}
}

func TestToggleCommentGoMultiLineSelection(t *testing.T) {
	e, _ := newTestEditor(t)
	e.CurTab().Buf.FilePath = "main.go"
	b := e.CurTab().Buf
	b.InsertText("a := 1\nb := 2")
	b.MoveTo(buffer.Position{Line: 0, Col: 0}, false)
	b.MoveTo(buffer.Position{Line: 1, Col: 6}, true)
	e.handleKey(key(tcell.KeyCtrlUnderscore, 0, 0))
	if string(b.Lines[0]) != "// a := 1" || string(b.Lines[1]) != "// b := 2" {
		t.Fatalf("lines = %q / %q", string(b.Lines[0]), string(b.Lines[1]))
	}
}

func TestWordJumpNavigation(t *testing.T) {
	e, _ := newTestEditor(t)
	typeString(e, "foo bar baz")
	b := e.CurTab().Buf
	b.MoveTo(buffer.Position{Line: 0, Col: 0}, false)
	e.handleKey(key(tcell.KeyRight, 0, tcell.ModCtrl))
	if b.Cursor.Col != 4 {
		t.Fatalf("cursor col after Ctrl+Right = %d, want 4", b.Cursor.Col)
	}
	e.handleKey(key(tcell.KeyRight, 0, tcell.ModCtrl))
	if b.Cursor.Col != 8 {
		t.Fatalf("cursor col after 2x Ctrl+Right = %d, want 8", b.Cursor.Col)
	}
	e.handleKey(key(tcell.KeyLeft, 0, tcell.ModCtrl))
	if b.Cursor.Col != 4 {
		t.Fatalf("cursor col after Ctrl+Left = %d, want 4", b.Cursor.Col)
	}
}

func TestShiftArrowSelection(t *testing.T) {
	e, _ := newTestEditor(t)
	typeString(e, "hello")
	b := e.CurTab().Buf
	b.MoveTo(buffer.Position{Line: 0, Col: 0}, false)
	for i := 0; i < 3; i++ {
		e.handleKey(key(tcell.KeyRight, 0, tcell.ModShift))
	}
	if !b.HasSel {
		t.Fatal("expected selection after shift+right")
	}
	if got := b.SelectedText(); got != "hel" {
		t.Fatalf("selected = %q", got)
	}
}

func TestFileTreeNavigationOpensFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(dir+"/sub", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/a.txt", []byte("alpha\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dir+"/sub/b.txt", []byte("beta\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	e, _ := newTestEditor(t)
	e.Tree = NewFileTree(dir)
	// Entries are sorted dirs-first: "sub" (dir), then "a.txt" (file).
	if len(e.Tree.flat) != 2 {
		t.Fatalf("flat tree = %v", e.Tree.flat)
	}
	if e.Tree.flat[0].name != "sub" || !e.Tree.flat[0].isDir {
		t.Fatalf("expected first entry to be dir 'sub', got %+v", e.Tree.flat[0])
	}

	e.handleKey(key(tcell.KeyCtrlB, 0, 0))
	if !e.Tree.Visible || e.Focus != FocusTree {
		t.Fatalf("after Ctrl+B: visible=%v focus=%v", e.Tree.Visible, e.Focus)
	}

	// Enter on "sub" expands it instead of opening a file.
	e.handleKey(key(tcell.KeyEnter, 0, 0))
	if len(e.Tree.flat) != 3 || e.Tree.flat[1].name != "b.txt" {
		t.Fatalf("after expand, flat = %v", e.Tree.flat)
	}
	if e.Focus != FocusTree {
		t.Fatal("expanding a directory should keep tree focus")
	}

	e.Tree.MoveDown() // -> b.txt
	e.handleKey(key(tcell.KeyEnter, 0, 0))
	if e.Focus != FocusEditor {
		t.Fatal("opening a file should switch focus to the editor")
	}
	if got := e.CurTab().Title(); got != "b.txt" {
		t.Fatalf("opened tab = %q, want b.txt", got)
	}
	if string(e.CurTab().Buf.Lines[0]) != "beta" {
		t.Fatalf("opened content = %q", string(e.CurTab().Buf.Lines[0]))
	}

	e.handleKey(key(tcell.KeyCtrlW, 0, 0))
	if e.Focus != FocusTree {
		t.Fatal("Ctrl+W should switch focus back to tree")
	}
}

func TestSearchFindsMatch(t *testing.T) {
	e, _ := newTestEditor(t)
	typeString(e, "foo bar foo")
	e.CurTab().Buf.MoveTo(buffer.Position{Line: 0, Col: 0}, false)
	e.handleKey(key(tcell.KeyCtrlF, 0, 0))
	typeString(e, "bar")
	e.handleKey(key(tcell.KeyEnter, 0, 0))
	if e.Mode != ModeSearch {
		t.Fatalf("mode = %v, want ModeSearch", e.Mode)
	}
	sel := e.CurTab().Buf.SelectedText()
	if sel != "bar" {
		t.Fatalf("selected = %q, want bar", sel)
	}
	e.handleKey(key(tcell.KeyEscape, 0, 0))
	if e.Mode != ModeNormal {
		t.Fatalf("mode after Esc = %v", e.Mode)
	}
}

func TestSearchCyclesThroughMultipleMatches(t *testing.T) {
	e, _ := newTestEditor(t)
	typeString(e, "foo bar foo")
	e.handleKey(key(tcell.KeyCtrlF, 0, 0))
	typeString(e, "foo")
	e.handleKey(key(tcell.KeyEnter, 0, 0))
	first := e.CurTab().Buf.SelAnchor
	e.handleKey(key(tcell.KeyEnter, 0, 0))
	second := e.CurTab().Buf.SelAnchor
	if first == second {
		t.Fatal("second Enter should jump to the next match")
	}
	e.handleKey(key(tcell.KeyEnter, 0, 0)) // wraps back to first match
	third := e.CurTab().Buf.SelAnchor
	if third != first {
		t.Fatalf("expected wraparound to first match, got %+v vs %+v", third, first)
	}
}

func TestCtrlHTogglesHelpOverlay(t *testing.T) {
	e, _ := newTestEditor(t)
	before := e.CurTab().Buf.String()
	e.handleKey(key(tcell.KeyCtrlH, 0, 0))
	if !e.ShowHelp {
		t.Fatal("Ctrl+H should open the help overlay")
	}
	e.handleKey(key(tcell.KeyRune, 'x', 0))
	if e.ShowHelp {
		t.Fatal("any key should close the help overlay")
	}
	if got := e.CurTab().Buf.String(); got != before {
		t.Fatalf("the key that closed the overlay should not have been inserted into the buffer, got %q, want %q", got, before)
	}
}
