package highlight

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestLineStylesMatchesLineCount(t *testing.T) {
	src := "def foo():\n    return 1\n"
	lexer := LexerFor("test.py")
	lines := LineStyles(lexer, src, tcell.StyleDefault)
	if len(lines) != 3 { // "def foo():", "    return 1", ""
		t.Fatalf("got %d lines, want 3: %#v", len(lines), lines)
	}
	if len(lines[0]) != len("def foo():") {
		t.Fatalf("line 0 style count = %d, want %d", len(lines[0]), len("def foo():"))
	}
}

func TestKeywordGetsNonDefaultStyle(t *testing.T) {
	src := "def foo():\n    pass\n"
	lexer := LexerFor("test.py")
	lines := LineStyles(lexer, src, tcell.StyleDefault)
	// "def" occupies columns 0-2 on line 0.
	defStyle := lines[0][0]
	spaceStyle := lines[1][0] // leading indentation space before "pass"
	if defStyle == tcell.StyleDefault {
		t.Fatal("expected 'def' keyword to have a non-default style")
	}
	_ = spaceStyle
}

func TestLexerSelectionByExtension(t *testing.T) {
	cases := map[string]string{
		"a.py":   "def x():\n    pass\n",
		"a.go":   "package main\nfunc main() {}\n",
		"a.c":    "int main() { return 0; }\n",
		"a.rs":   "fn main() {}\n",
		"a.json": "{\"a\": 1}\n",
		"a.md":   "# Title\n",
		"a.sh":   "#!/bin/bash\necho hi\n",
	}
	for name, src := range cases {
		lexer := LexerFor(name)
		if lexer == nil {
			t.Fatalf("no lexer for %s", name)
		}
		lines := LineStyles(lexer, src, tcell.StyleDefault)
		if len(lines) == 0 {
			t.Fatalf("no styles produced for %s", name)
		}
	}
}
