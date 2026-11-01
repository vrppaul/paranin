package tui

import (
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

// How the command line says where a search stands, and that it found nothing, as Neovim says it
const (
	searchMark    = "/"
	matchOpening  = " ["
	matchOf       = "/"
	matchClosing  = "]"
	notFoundError = "E486: Pattern not found: "
)

// searchState is the search: what the reader is typing after the slash, the search confirmed last, which every
// painted tab marks and n and N go through, and the match each tab showed last.
type searchState struct {
	typing string
	query  string
	shown  []shownMatch
}

// shownMatch is the line of the match a tab showed last, and the scroll it left the tab at: n goes on from that
// match while the page stays there, and from the page's top once the reader scrolls it.
type shownMatch struct {
	line   int
	scroll int
	shown  bool
}

// commandMessage is what the command line says after a key, and whether it says something went wrong.
type commandMessage struct {
	text   string
	failed bool
}

// startSearch opens the search line, empty, after a slash.
func (model Model) startSearch(int) (Model, tea.Cmd) {
	model.keys.mode, model.search.typing = modeSearch, ""
	return model, nil
}

// endSearch takes the search away, and its marks with it.
func (model Model) endSearch(int) (Model, tea.Cmd) {
	if model.search.query == "" {
		return model, nil
	}
	model.search.query = ""
	return model.repainted(), nil
}

// typedInSearch is the model after a key while the reader types a search: Enter searches, Esc leaves unsought, a
// backspace takes a letter back or, on nothing, leaves, and a letter is typed. Anything else, a control character
// among it, is passed over, so it never reaches the screen.
func (model Model) typedInSearch(key tea.KeyPressMsg) Model {
	switch key.String() {
	case enterKey:
		return model.searched()
	case escapeKey:
		model.keys.mode = modeNormal
		return model
	case backspaceKey:
		if model.search.typing == "" {
			model.keys.mode = modeNormal
			return model
		}
		_, last := utf8.DecodeLastRuneInString(model.search.typing)
		model.search.typing = model.search.typing[:len(model.search.typing)-last]
		return model
	}
	if printable(key.Text) {
		model.search.typing += key.Text
	}
	return model
}

// searched is the model after the search typed is confirmed: every tab painted again with its matches marked, no
// match shown yet on any, and the page at the first match at or after its top line.
func (model Model) searched() Model {
	model.keys.mode = modeNormal
	if model.search.typing == "" {
		return model
	}
	model.search.query = model.search.typing
	model.search.shown = make([]shownMatch, len(model.tabs))
	model = model.repainted()
	return model.stepMatch(1, model.scrolls[model.chosen]-1, matchAfter)
}

// nextMatch goes to the next match: after the one the tab showed last, or at or after the page's top line, as Enter
// does, where it shows none.
func (model Model) nextMatch(count int) (Model, tea.Cmd) {
	return model.stepMatch(max(count, 1), model.matchReference(model.scrolls[model.chosen]-1), matchAfter), nil
}

// previousMatch goes to the match before: before the one the tab showed last, or before the page's top line.
func (model Model) previousMatch(count int) (Model, tea.Cmd) {
	return model.stepMatch(max(count, 1), model.matchReference(model.scrolls[model.chosen]), matchBefore), nil
}

// matchReference is the line n and N go on from: the match the tab showed last while the page stays there, or a
// line by the page's top where it shows none, or the reader has scrolled since.
func (model Model) matchReference(byTheTop int) int {
	if last := model.search.shown[model.chosen]; last.shown && last.scroll == model.scrolls[model.chosen] {
		return last.line
	}
	return byTheTop
}

// stepMatch shows the match a number of steps from a line, each step the next match after the line or the one
// before it, round from the last to the first; it says which line of how many holding a match it is, or that the
// search found nothing in the tab.
func (model Model) stepMatch(steps, from int, step func(matches []int, line int) int) Model {
	query, matches := model.search.query, model.drawn.tabs[model.chosen].matches
	if query == "" {
		return model
	}
	if len(matches) == 0 {
		model.said = commandMessage{text: notFoundError + query, failed: true}
		return model
	}
	place := step(matches, from)
	for range steps - 1 {
		place = step(matches, matches[place])
	}
	model.said = commandMessage{
		text:   searchMark + query + matchOpening + strconv.Itoa(place+1) + matchOf + strconv.Itoa(len(matches)) + matchClosing,
		failed: false,
	}
	model = model.scrolledTo(matches[place])
	model.search.shown = slices.Clone(model.search.shown)
	model.search.shown[model.chosen] = shownMatch{line: matches[place], scroll: model.scrolls[model.chosen], shown: true}
	return model
}

// matchAfter is the place of the first match after a line, or the first of all after the last.
func matchAfter(matches []int, line int) int {
	return max(slices.IndexFunc(matches, func(matching int) bool { return matching > line }), 0)
}

// matchBefore is the place of the last match before a line, or the last of all before the first.
func matchBefore(matches []int, line int) int {
	for place, matching := range slices.Backward(matches) {
		if matching < line {
			return place
		}
	}
	return len(matches) - 1
}

// matchedRun is a run of a line, and whether it is a match of the search.
type matchedRun struct {
	run
	matched bool
}

// matchedRuns are a line's runs with each match of a query, whatever its case, parted into a run of its own; with
// no query, the runs as they are.
func matchedRuns(shown line, query string) []matchedRun {
	letters := []rune{}
	for _, piece := range shown {
		letters = append(letters, []rune(piece.text)...)
	}
	marks := matchMarks(letters, []rune(query))
	matched := []matchedRun{}
	place := 0
	for _, piece := range shown {
		pieceLetters := []rune(piece.text)
		start := 0
		for index := range pieceLetters {
			if index > start && marks[place+index] != marks[place+start] {
				matched = append(matched, partOf(piece, pieceLetters[start:index], marks[place+start]))
				start = index
			}
		}
		if start < len(pieceLetters) {
			matched = append(matched, partOf(piece, pieceLetters[start:], marks[place+start]))
		}
		place += len(pieceLetters)
	}
	return matched
}

// matchMarks are whether each letter lies in a match of a query, whatever its case; matches never overlap.
func matchMarks(letters, query []rune) []bool {
	marks := make([]bool, len(letters))
	if len(query) == 0 {
		return marks
	}
	for start := 0; start+len(query) <= len(letters); {
		if !strings.EqualFold(string(letters[start:start+len(query)]), string(query)) {
			start++
			continue
		}
		for index := range query {
			marks[start+index] = true
		}
		start += len(query)
	}
	return marks
}

func partOf(piece run, letters []rune, matched bool) matchedRun {
	piece.text = string(letters)
	return matchedRun{run: piece, matched: matched}
}
