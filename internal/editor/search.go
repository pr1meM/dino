package editor

import (
	"strings"

	"github.com/gdamore/tcell/v2"

	"dino/internal/buffer"
)

// SearchState holds the state of an in-progress search session.
type SearchState struct {
	Query   string
	Matches []matchRange
	Current int
}

type matchRange struct {
	Start, End buffer.Position
}

func (e *Editor) startSearch() {
	e.Mode = ModeSearch
	e.PromptInput = ""
}

func findAll(t *Tab, query string) []matchRange {
	if query == "" {
		return nil
	}
	var matches []matchRange
	for i, line := range t.Buf.Lines {
		s := string(line)
		start := 0
		for {
			idx := strings.Index(s[start:], query)
			if idx < 0 {
				break
			}
			col := len([]rune(s[:start+idx]))
			endCol := col + len([]rune(query))
			matches = append(matches, matchRange{
				Start: buffer.Position{Line: i, Col: col},
				End:   buffer.Position{Line: i, Col: endCol},
			})
			start += idx + len(query)
			if start > len(s) {
				break
			}
		}
	}
	return matches
}

func (e *Editor) jumpToMatch(t *Tab, m matchRange) {
	t.Buf.MoveTo(m.Start, false)
	t.Buf.MoveTo(m.End, true)
}

func (e *Editor) handleSearchKey(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		e.cancelPrompt()
		e.Search = SearchState{}
	case tcell.KeyEnter:
		t := e.CurTab()
		if t == nil {
			e.cancelPrompt()
			return
		}
		if e.Search.Query != e.PromptInput {
			e.Search.Query = e.PromptInput
			e.Search.Matches = findAll(t, e.PromptInput)
			e.Search.Current = -1
		}
		if len(e.Search.Matches) == 0 {
			e.SetStatus("No matches for \"" + e.PromptInput + "\"")
			return
		}
		if ev.Modifiers()&tcell.ModShift != 0 {
			e.Search.Current--
		} else {
			e.Search.Current++
		}
		n := len(e.Search.Matches)
		e.Search.Current = ((e.Search.Current % n) + n) % n
		e.jumpToMatch(t, e.Search.Matches[e.Search.Current])
		e.SetStatus("Match")
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.PromptInput) > 0 {
			e.PromptInput = e.PromptInput[:len(e.PromptInput)-1]
		}
	case tcell.KeyRune:
		e.PromptInput += string(ev.Rune())
	}
}
