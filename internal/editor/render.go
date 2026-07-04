package editor

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"

	"dino/internal/buffer"
)

var (
	colorBg          = tcell.NewRGBColor(0x2e, 0x2e, 0x2e)
	colorFg          = tcell.NewRGBColor(0xd8, 0xd8, 0xd8)
	colorTabBarBg    = tcell.NewRGBColor(0x1e, 0x1e, 0x1e)
	colorTabActiveBg = tcell.NewRGBColor(0x3c, 0x3c, 0x3c)
	styleDefault     = tcell.StyleDefault.Background(colorBg).Foreground(colorFg)
	styleLineNum     = styleDefault.Foreground(tcell.NewRGBColor(0xa8, 0xa8, 0xa8))
	styleStatusBar   = tcell.StyleDefault.Background(tcell.ColorNavy).Foreground(tcell.ColorWhite)
	styleTabBar      = tcell.StyleDefault.Background(colorTabBarBg).Foreground(colorFg)
	styleTabActive   = tcell.StyleDefault.Background(colorTabActiveBg).Foreground(tcell.NewRGBColor(0xf0, 0xf0, 0xf0)).Bold(true)
	styleTabPassive  = tcell.StyleDefault.Background(colorTabBarBg).Foreground(tcell.NewRGBColor(0x88, 0x88, 0x88))
	styleTabSep      = tcell.StyleDefault.Background(colorTabBarBg).Foreground(tcell.NewRGBColor(0x45, 0x45, 0x45))
	styleTreeDir     = styleDefault.Foreground(tcell.ColorTeal).Bold(true)
	styleTreeFile    = styleDefault
	styleTreeSel     = tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)
	styleSelection   = tcell.StyleDefault.Background(tcell.ColorSteelBlue)
	stylePrompt      = tcell.StyleDefault.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorWhite)
	styleQuitPrompt  = tcell.StyleDefault.Background(tcell.NewRGBColor(0x50, 0x50, 0x50)).Foreground(tcell.ColorWhite)
	styleTreeHeader  = styleDefault
	styleHelpBox     = tcell.StyleDefault.Background(colorTabActiveBg).Foreground(colorFg)
	styleHelpTitle   = styleHelpBox.Bold(true)
	styleHelpKey     = styleHelpBox.Foreground(tcell.NewRGBColor(0xf0, 0xf0, 0xf0)).Bold(true)
)

// helpEntries lists the global keybindings shown in the Ctrl+H overlay.
var helpEntries = []struct{ key, desc string }{
	{"Ctrl+H", "Toggle this help"},
	{"Ctrl+N", "New file"},
	{"Ctrl+B", "Toggle file tree"},
	{"Ctrl+W / Tab", "Switch focus tree <-> editor"},
	{"Ctrl+S", "Save"},
	{"Ctrl+Q", "Quit"},
	{"Ctrl+F", "Search"},
	{"Ctrl+/", "Toggle comment"},
	{"Ctrl+C / X / V", "Copy / Cut / Paste"},
	{"Ctrl+A", "Select all"},
	{"Ctrl+Z / Y", "Undo / Redo"},
	{"Alt+Left / Right", "Switch to previous/next tab"},
}

type layout struct {
	width, height  int
	treeWidth      int
	textX0         int
	lineNumWidth   int
	textX1         int // first column of actual text (after gutter)
	textY0, textY1 int // inclusive row range of the text/tree body
	statusY        int
}

func (e *Editor) computeLayout() layout {
	w, h := e.Screen.Size()
	l := layout{width: w, height: h}
	l.textY0 = 1
	l.textY1 = h - 2
	l.statusY = h - 1
	if e.Tree.Visible {
		l.treeWidth = e.Tree.Width
	}
	l.textX0 = l.treeWidth
	lineCount := 1
	if t := e.CurTab(); t != nil {
		lineCount = t.Buf.LineCount()
	}
	digits := len(fmt.Sprintf("%d", lineCount))
	l.lineNumWidth = digits + 1
	l.textX1 = l.textX0 + l.lineNumWidth
	return l
}

func (e *Editor) Draw() {
	e.Screen.SetStyle(styleDefault)
	e.Screen.Clear()
	l := e.computeLayout()

	e.drawTabBar(l)
	if e.Tree.Visible {
		e.drawTree(l)
	}
	e.drawText(l)
	e.drawStatusBar(l)
	if e.ShowHelp {
		e.drawHelpOverlay(l)
	}
	e.positionCursor(l)

	e.Screen.Show()
}

func (e *Editor) drawTabBar(l layout) {
	x := l.treeWidth
	for i, t := range e.Tabs {
		active := i == e.Active
		style := styleTabPassive
		if active {
			style = styleTabActive
		}
		label := " " + t.Title()
		if t.Buf.Dirty {
			label += "*"
		}
		label += " "
		x = drawStr(e.Screen, x, 0, l.width, style, label)
		if x < l.width && i < len(e.Tabs)-1 {
			e.Screen.SetContent(x, 0, tcell.RuneVLine, nil, styleTabSep)
			x++
		}
	}
	for ; x < l.width; x++ {
		e.Screen.SetContent(x, 0, ' ', nil, styleTabBar)
	}
}

func (e *Editor) drawTree(l layout) {
	headerY := l.textY0
	for x := 0; x < l.treeWidth-1; x++ {
		e.Screen.SetContent(x, headerY, ' ', nil, styleTreeHeader)
	}
	drawStr(e.Screen, 0, headerY, l.treeWidth-1, styleTreeHeader, " "+e.Tree.Root.path)

	listY0 := headerY + 1
	rows := l.textY1 - listY0 + 1
	if rows < 1 {
		rows = 1
	}
	e.Tree.ensureVisible(rows)
	for row := 0; row < rows; row++ {
		y := listY0 + row
		idx := e.Tree.Scroll + row
		for x := 0; x < l.treeWidth-1; x++ {
			e.Screen.SetContent(x, y, ' ', nil, styleDefault)
		}
		if idx < 0 || idx >= len(e.Tree.flat) {
			continue
		}
		n := e.Tree.flat[idx]
		style := styleTreeFile
		if n.isDir {
			style = styleTreeDir
		}
		if idx == e.Tree.Cursor && e.Focus == FocusTree {
			style = styleTreeSel
		}
		prefix := strings.Repeat("  ", n.depth)
		icon := "  "
		if n.isDir {
			if n.expanded {
				icon = "v "
			} else {
				icon = "> "
			}
		}
		text := prefix + icon + n.name
		drawStr(e.Screen, 0, y, l.treeWidth-1, style, text)
	}
	for y := l.textY0; y <= l.textY1; y++ {
		e.Screen.SetContent(l.treeWidth-1, y, tcell.RuneVLine, nil, styleLineNum)
	}
}

func (e *Editor) drawText(l layout) {
	t := e.CurTab()
	rows := l.textY1 - l.textY0 + 1
	if t == nil {
		// No tab is open yet (e.g. the file tree is active but nothing
		// has been selected from it). Still paint the text area and the
		// line-number gutter for the buffer's single empty line instead
		// of leaving it blank.
		for row := 0; row < rows; row++ {
			y := l.textY0 + row
			for x := l.textX0; x < l.width; x++ {
				e.Screen.SetContent(x, y, ' ', nil, styleDefault)
			}
			if row == 0 {
				numStr := fmt.Sprintf("%-*d ", l.lineNumWidth-1, 1)
				drawStr(e.Screen, l.textX0, y, l.textX0+l.lineNumWidth, styleLineNum, numStr)
			}
		}
		return
	}
	e.scrollToCursor(t, l)

	selStart, selEnd, hasSel := buffer.Position{}, buffer.Position{}, t.Buf.HasSel
	if hasSel {
		selStart, selEnd = buffer.OrderedSelection(t.Buf.SelAnchor, t.Buf.Cursor)
	}

	tokenLines := e.highlightVisible(t, rows)

	for row := 0; row < rows; row++ {
		y := l.textY0 + row
		lineIdx := t.TopLine + row
		for x := l.textX0; x < l.width; x++ {
			e.Screen.SetContent(x, y, ' ', nil, styleDefault)
		}
		if lineIdx >= t.Buf.LineCount() {
			continue
		}
		numStr := fmt.Sprintf("%-*d ", l.lineNumWidth-1, lineIdx+1)
		drawStr(e.Screen, l.textX0, y, l.textX0+l.lineNumWidth, styleLineNum, numStr)

		line := t.Buf.Line(lineIdx)
		styles := tokenLines[row]
		x := l.textX1
		for col := t.LeftCol; col < len(line); col++ {
			if x >= l.width {
				break
			}
			st := styleDefault
			if styles != nil && col < len(styles) {
				st = styles[col]
			}
			if hasSel && inSelection(lineIdx, col, selStart, selEnd) {
				st = st.Background(tcell.ColorSteelBlue)
			}
			e.Screen.SetContent(x, y, line[col], nil, st)
			x++
		}
		if hasSel && inSelection(lineIdx, len(line), selStart, selEnd) && x < l.width {
			e.Screen.SetContent(x, y, ' ', nil, styleSelection)
		}
	}
}

func inSelection(line, col int, start, end buffer.Position) bool {
	p := buffer.Position{Line: line, Col: col}
	return !p.Less(start) && p.Less(end)
}

func (e *Editor) scrollToCursor(t *Tab, l layout) {
	rows := l.textY1 - l.textY0 + 1
	cols := l.width - l.textX1
	if cols < 1 {
		cols = 1
	}
	if t.Buf.Cursor.Line < t.TopLine {
		t.TopLine = t.Buf.Cursor.Line
	}
	if t.Buf.Cursor.Line >= t.TopLine+rows {
		t.TopLine = t.Buf.Cursor.Line - rows + 1
	}
	if t.TopLine < 0 {
		t.TopLine = 0
	}
	if t.Buf.Cursor.Col < t.LeftCol {
		t.LeftCol = t.Buf.Cursor.Col
	}
	if t.Buf.Cursor.Col >= t.LeftCol+cols {
		t.LeftCol = t.Buf.Cursor.Col - cols + 1
	}
	if t.LeftCol < 0 {
		t.LeftCol = 0
	}
}

func (e *Editor) drawStatusBar(l layout) {
	y := l.statusY
	for x := 0; x < l.width; x++ {
		e.Screen.SetContent(x, y, ' ', nil, styleStatusBar)
	}
	const helpHint = "Ctrl+H Help "
	if e.Mode != ModeNormal {
		style := stylePrompt
		if e.Mode == ModeQuitConfirm || e.Mode == ModeQuitConfirmSaveAsPrompt {
			style = styleQuitPrompt
		}
		prompt := e.promptLabel() + e.PromptInput
		drawStr(e.Screen, 0, y, l.width, style, prompt)
		for x := len(prompt); x < l.width; x++ {
			e.Screen.SetContent(x, y, ' ', nil, style)
		}
		drawStr(e.Screen, l.width-len(helpHint), y, l.width, style, helpHint)
		return
	}
	t := e.CurTab()
	if t == nil {
		drawStr(e.Screen, l.width-len(helpHint), y, l.width, styleStatusBar, helpHint)
		return
	}
	dirty := ""
	if t.Buf.Dirty {
		dirty = " [+]"
	}
	left := fmt.Sprintf(" Line %d/%d, Col %d", t.Buf.Cursor.Line+1, t.Buf.LineCount(), t.Buf.Cursor.Col+1)
	right := fmt.Sprintf("%s%s  %s", t.Title(), dirty, helpHint)
	if e.StatusMsg != "" {
		left = " " + e.StatusMsg
	}
	drawStr(e.Screen, 0, y, l.width, styleStatusBar, left)
	drawStr(e.Screen, l.width-len(right), y, l.width, styleStatusBar, right)
}

// drawHelpOverlay renders the Ctrl+H keybinding overlay, centered on screen,
// on top of everything else.
func (e *Editor) drawHelpOverlay(l layout) {
	title := "Keybindings"
	keyColW, descColW := 0, 0
	for _, en := range helpEntries {
		if len(en.key) > keyColW {
			keyColW = len(en.key)
		}
		if len(en.desc) > descColW {
			descColW = len(en.desc)
		}
	}
	innerW := keyColW + 2 + descColW
	if len(title) > innerW {
		innerW = len(title)
	}
	boxW := innerW + 4
	boxH := len(helpEntries) + 4
	if boxW > l.width {
		boxW = l.width
	}
	if boxH > l.height {
		boxH = l.height
	}
	x0 := (l.width - boxW) / 2
	y0 := (l.height - boxH) / 2

	for row := 0; row < boxH; row++ {
		y := y0 + row
		for col := 0; col < boxW; col++ {
			x := x0 + col
			r := ' '
			switch {
			case row == 0 && col == 0:
				r = tcell.RuneULCorner
			case row == 0 && col == boxW-1:
				r = tcell.RuneURCorner
			case row == boxH-1 && col == 0:
				r = tcell.RuneLLCorner
			case row == boxH-1 && col == boxW-1:
				r = tcell.RuneLRCorner
			case row == 0 || row == boxH-1:
				r = tcell.RuneHLine
			case col == 0 || col == boxW-1:
				r = tcell.RuneVLine
			}
			e.Screen.SetContent(x, y, r, nil, styleHelpBox)
		}
	}
	drawStr(e.Screen, x0+(boxW-len(title))/2, y0, x0+boxW-1, styleHelpTitle, title)
	for i, en := range helpEntries {
		y := y0 + 2 + i
		if y >= y0+boxH-1 {
			break
		}
		keyStr := fmt.Sprintf("%-*s", keyColW, en.key)
		x := drawStr(e.Screen, x0+2, y, x0+boxW-2, styleHelpKey, keyStr)
		drawStr(e.Screen, x+2, y, x0+boxW-2, styleHelpBox, en.desc)
	}
}

func (e *Editor) promptLabel() string {
	switch e.Mode {
	case ModeSearch:
		return "Search: "
	case ModeSaveAsPrompt, ModeQuitConfirmSaveAsPrompt:
		return "Save as: "
	case ModeNewFilePrompt:
		return "New file: "
	case ModeQuitConfirm:
		return "Unsaved changes. Save? [y]es/[n]o/[esc] cancel: "
	}
	return ""
}

func (e *Editor) positionCursor(l layout) {
	if e.ShowHelp {
		e.Screen.HideCursor()
		return
	}
	if e.Mode != ModeNormal {
		e.Screen.ShowCursor(len(e.promptLabel())+len(e.PromptInput), l.statusY)
		return
	}
	if e.Focus == FocusTree {
		e.Screen.HideCursor()
		return
	}
	t := e.CurTab()
	if t == nil {
		e.Screen.HideCursor()
		return
	}
	y := l.textY0 + (t.Buf.Cursor.Line - t.TopLine)
	x := l.textX1 + (t.Buf.Cursor.Col - t.LeftCol)
	if y < l.textY0 || y > l.textY1 || x < l.textX1 || x >= l.width {
		e.Screen.HideCursor()
		return
	}
	e.Screen.ShowCursor(x, y)
}

func drawStr(s tcell.Screen, x, y, maxX int, style tcell.Style, str string) int {
	for _, r := range str {
		if x >= maxX {
			break
		}
		s.SetContent(x, y, r, nil, style)
		x++
	}
	return x
}
