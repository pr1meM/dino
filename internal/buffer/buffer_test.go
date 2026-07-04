package buffer

import (
	"os"
	"testing"
)

func TestInsertAndString(t *testing.T) {
	b := New()
	b.InsertText("hello")
	if got := b.String(); got != "hello\n" {
		t.Fatalf("got %q", got)
	}
	if b.Cursor != (Position{0, 5}) {
		t.Fatalf("cursor = %+v", b.Cursor)
	}
}

func TestInsertNewlineSplitsLine(t *testing.T) {
	b := New()
	b.InsertText("ab")
	b.MoveLeft(false) // between a and b
	b.InsertText("\n")
	if len(b.Lines) != 2 || string(b.Lines[0]) != "a" || string(b.Lines[1]) != "b" {
		t.Fatalf("lines = %v", linesToStrings(b.Lines))
	}
	if b.Cursor != (Position{1, 0}) {
		t.Fatalf("cursor = %+v", b.Cursor)
	}
}

func TestBackspaceMergesLines(t *testing.T) {
	b := New()
	b.InsertText("ab\ncd")
	b.MoveTo(Position{1, 0}, false)
	b.Backspace()
	if len(b.Lines) != 1 || string(b.Lines[0]) != "abcd" {
		t.Fatalf("lines = %v", linesToStrings(b.Lines))
	}
	if b.Cursor != (Position{0, 2}) {
		t.Fatalf("cursor = %+v", b.Cursor)
	}
}

func TestDeleteForward(t *testing.T) {
	b := New()
	b.InsertText("abc")
	b.MoveTo(Position{0, 0}, false)
	b.DeleteForward()
	if string(b.Lines[0]) != "bc" {
		t.Fatalf("line = %q", string(b.Lines[0]))
	}
}

func TestUndoRedoTyping(t *testing.T) {
	b := New()
	for _, r := range "abc" {
		b.InsertRune(r)
	}
	if string(b.Lines[0]) != "abc" {
		t.Fatalf("line = %q", string(b.Lines[0]))
	}
	b.Undo()
	if string(b.Lines[0]) != "" {
		t.Fatalf("after undo, line = %q", string(b.Lines[0]))
	}
	b.Redo()
	if string(b.Lines[0]) != "abc" {
		t.Fatalf("after redo, line = %q", string(b.Lines[0]))
	}
}

func TestUndoRedoBackspaceChain(t *testing.T) {
	b := New()
	b.InsertText("abc")
	b.Backspace()
	b.Backspace()
	b.Backspace()
	if string(b.Lines[0]) != "" {
		t.Fatalf("line = %q", string(b.Lines[0]))
	}
	b.Undo() // undoes the whole backspace chain
	if string(b.Lines[0]) != "abc" {
		t.Fatalf("after undo, line = %q", string(b.Lines[0]))
	}
	if b.Cursor != (Position{0, 3}) {
		t.Fatalf("cursor after undo = %+v", b.Cursor)
	}
}

func TestUndoRedoInsertThenDeleteAreSeparateGroups(t *testing.T) {
	b := New()
	b.InsertText("abc")
	b.Backspace()
	// undo stack: [insert "abc", delete "c"]
	b.Undo()
	if string(b.Lines[0]) != "abc" {
		t.Fatalf("after first undo, line = %q", string(b.Lines[0]))
	}
	b.Undo()
	if string(b.Lines[0]) != "" {
		t.Fatalf("after second undo, line = %q", string(b.Lines[0]))
	}
}

func TestSelectionDeleteAndCopy(t *testing.T) {
	b := New()
	b.InsertText("hello world")
	b.MoveTo(Position{0, 0}, false)
	b.MoveTo(Position{0, 5}, true)
	if got := b.SelectedText(); got != "hello" {
		t.Fatalf("selected = %q", got)
	}
	b.DeleteSelection()
	if string(b.Lines[0]) != " world" {
		t.Fatalf("line = %q", string(b.Lines[0]))
	}
}

func TestSelectAll(t *testing.T) {
	b := New()
	b.InsertText("ab\ncd")
	b.SelectAll()
	if got := b.SelectedText(); got != "ab\ncd" {
		t.Fatalf("selected = %q", got)
	}
}

func TestLoadDetectsCRLF(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/f.txt"
	writeFile(t, path, "a\r\nb\r\n")
	b, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if b.LineEnding != "\r\n" {
		t.Fatalf("line ending = %q", b.LineEnding)
	}
	if len(b.Lines) != 2 {
		t.Fatalf("lines = %v", linesToStrings(b.Lines))
	}
	if b.String() != "a\r\nb\r\n" {
		t.Fatalf("roundtrip = %q", b.String())
	}
}

func TestNoFinalNewlinePreserved(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/f.txt"
	writeFile(t, path, "abc")
	b, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if b.String() != "abc" {
		t.Fatalf("roundtrip = %q", b.String())
	}
}

func linesToStrings(lines [][]rune) []string {
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = string(l)
	}
	return out
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
