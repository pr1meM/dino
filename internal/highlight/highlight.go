// Package highlight adapts the chroma tokenizer/lexer library to
// produce per-rune tcell styles for the editor's text view.
package highlight

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gdamore/tcell/v2"
)

var chromaStyle = styles.Get("monokai")

// colourOverrides fixes token categories where the theme's color makes
// them hard to read or indistinguishable from unrelated tokens. Monokai
// paints preprocessor directives (e.g. "#include") the same muted gray
// as comments; bump them to a distinct, more visible color instead.
var colourOverrides = map[chroma.TokenType]tcell.Color{
	chroma.CommentPreproc:     tcell.NewRGBColor(0xf9, 0x26, 0x72),
	chroma.CommentPreprocFile: tcell.NewRGBColor(0xf9, 0x26, 0x72),
}

// LexerFor picks a lexer by filename (extension/glob match), falling
// back to a plain-text lexer for unknown or empty names.
func LexerFor(filename string) chroma.Lexer {
	l := lexers.Match(filename)
	if l == nil {
		l = lexers.Fallback
	}
	return chroma.Coalesce(l)
}

// LineStyles tokenizes text with lexer and returns, for each line, one
// tcell.Style per rune (matching how buffer.Buffer splits lines on
// '\n'). Tokenizing the whole text (rather than line by line) lets
// chroma correctly handle constructs that span multiple lines, such as
// block comments or triple-quoted strings.
func LineStyles(lexer chroma.Lexer, text string, base tcell.Style) [][]tcell.Style {
	iter, err := lexer.Tokenise(nil, text)
	if err != nil {
		return nil
	}
	var lines [][]tcell.Style
	var cur []tcell.Style
	for _, tok := range iter.Tokens() {
		st := styleFor(tok.Type, base)
		for _, r := range tok.Value {
			if r == '\n' {
				lines = append(lines, cur)
				cur = nil
				continue
			}
			cur = append(cur, st)
		}
	}
	lines = append(lines, cur)
	return lines
}

// styleFor builds a style for the given token type on top of base, so
// that tokens without an explicit chroma color (or without a chroma
// background) still use the editor's own background instead of
// tcell's undefined default (which terminals typically render as
// black).
func styleFor(tt chroma.TokenType, base tcell.Style) tcell.Style {
	entry := chromaStyle.Get(tt)
	st := base
	if override, ok := colourOverrides[tt]; ok {
		st = st.Foreground(override)
	} else if entry.Colour.IsSet() {
		st = st.Foreground(tcell.NewRGBColor(
			int32(entry.Colour.Red()),
			int32(entry.Colour.Green()),
			int32(entry.Colour.Blue()),
		))
	}
	if entry.Bold == chroma.Yes {
		st = st.Bold(true)
	}
	if entry.Italic == chroma.Yes {
		st = st.Italic(true)
	}
	if entry.Underline == chroma.Yes {
		st = st.Underline(true)
	}
	return st
}
