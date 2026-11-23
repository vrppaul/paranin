package tui

import (
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// How the command line says where a search stands, and that it found nothing, as Neovim says it
const (
	searchMark    = "/"
	matchOpening  = "["
	matchOf       = "/"
	matchClosing  = "]"
	notFoundError = "E486: Pattern not found: "
)

// searchState is the search: the line the chosen tab's page stood at when the reader typed the slash, which what
// they type seeks its first match from, and the page goes back to while what is typed matches nothing and once the
// reader leaves the search line unconfirmed; the search confirmed last, which every painted tab marks and n and N
// go through; and the match each tab showed last.
type searchState struct {
	begunAt int
	query   string
	shown   []shownMatch
}

// shownMatch is the place among a tab's matches of the one it showed last, and the scroll it left the tab at: n
// goes on from that match while the page stays there, and from the page's top once the reader scrolls it.
type shownMatch struct {
	place  int
	scroll int
	shown  bool
}

// commandMessage is what the command line says after a key, and whether it says something went wrong.
type commandMessage struct {
	text   string
	failed bool
}

// The steps n and N take through the matches: on to the next, and back to the one before
const (
	nextStep     = 1
	previousStep = -1
)

// matchWay is the way n or N goes through the matches: the step it takes from the match the tab showed last, and
// the match it goes to by the page's top line where the tab shows none.
type matchWay struct {
	step     int
	byTheTop func(matches []int, line int) int
}

// onward is n's way: to the next match, or the first at or after the page's top line.
func onward() matchWay {
	return matchWay{step: nextStep, byTheTop: matchAtOrAfter}
}

// back is N's way: to the match before, or the last before the page's top line.
func back() matchWay {
	return matchWay{step: previousStep, byTheTop: matchBefore}
}

// startSearch opens the search line, empty, after a slash: the chosen tab marks nothing until the reader types,
// where Neovim keeps the marks of the search before until then.
func (model Model) startSearch() (Model, tea.Cmd) {
	model.keys.mode, model.keys.typing, model.search.begunAt = modeSearch, "", model.scrolls[model.chosen]
	return model.repaintedChosen(), nil
}

// endSearch takes the search away, its marks and the matches it showed with it.
func (model Model) endSearch() (Model, tea.Cmd) {
	if model.search.query == "" {
		return model, nil
	}
	model.search.query, model.search.shown = "", make([]shownMatch, len(model.tabs))
	return model.repainted(), nil
}

// typedInSearch is the model after a key while the reader types a search: Enter searches; any other key edits the
// line, or leaves it unsought.
func (model Model) typedInSearch(key tea.KeyPressMsg) Model {
	if key.String() == enterKey {
		return model.searched()
	}
	typing, edit := edited(model.keys.typing, key)
	switch edit {
	case lineChanged:
		model.keys.typing = typing
		return model.retyped()
	case lineLeft:
		return model.searchLeft()
	case lineKept:
	}
	return model
}

// retyped is the model after what is typed changes, as Neovim's incsearch shows it: the chosen tab painted again
// with what is typed marked, its page at the first match at or after the line the search began at, or back at
// that line where nothing matches.
func (model Model) retyped() Model {
	model = model.repaintedChosen()
	drawn := model.drawn.tabs[model.chosen]
	if named := drawn.marks.named; named.occurrence != noMatch {
		return model.scrolledToward(named.line, drawn.spotOf(named)).withMatchTileTop(named.line)
	}
	return model.scrolledTo(model.search.begunAt)
}

// spotOf is where a match a tab's count names stands: its line, and the column its letters start at.
func (page paintedPage) spotOf(named namedMatch) spot {
	column := 0
	for _, piece := range page.marked[named.line].runs {
		if piece.occurrence == named.occurrence {
			break
		}
		column += textWidth(piece.text)
	}
	return spot{line: named.line, column: column}
}

// searchLeft is the model after the reader leaves the search line unconfirmed: the chosen tab marks the search
// before it again, and its page is back where it stood.
func (model Model) searchLeft() Model {
	model.keys.mode = modeNormal
	return model.repaintedChosen().scrolledTo(model.search.begunAt)
}

// searched is the model after the search typed is confirmed: every tab painted again with its matches marked, no
// match shown yet on any, and the page at the first match at or after its top line, the one the typing showed.
// Enter on nothing typed leaves the search line as Esc does.
func (model Model) searched() Model {
	if model.keys.typing == "" {
		return model.searchLeft()
	}
	model.keys.mode = modeNormal
	model.search.query = model.keys.typing
	model.search.shown = make([]shownMatch, len(model.tabs))
	model = model.repainted()
	return model.toMatch(func(matches []int) int { return matchAtOrAfter(matches, model.scrolls[model.chosen]) })
}

// nextMatch goes a match on, from the one the tab showed last, or from the page's top line, as Enter does, where it
// shows none.
func (model Model) nextMatch() (Model, tea.Cmd) {
	return model.toMatch(func(matches []int) int { return model.stepped(matches, onward()) }), nil
}

// previousMatch goes a match back, from the one the tab showed last, or from the page's top line.
func (model Model) previousMatch() (Model, tea.Cmd) {
	return model.toMatch(func(matches []int) int { return model.stepped(matches, back()) }), nil
}

// stepped is the place of the match a step one way goes to: a step from the match the tab showed last, while the page
// stays there, round from the last to the first; with none, the match the way picks by the page's top line.
func (model Model) stepped(matches []int, way matchWay) int {
	if last, standing := model.standingMatch(); standing {
		return wrapped(last.place+way.step, len(matches))
	}
	return way.byTheTop(matches, model.scrolls[model.chosen])
}

// standingMatch is the match the chosen tab showed last, while its page stays where the match left it, and whether
// there is one: n steps on from it, and y copies from its line.
func (model Model) standingMatch() (shownMatch, bool) {
	last := model.search.shown[model.chosen]
	return last, last.shown && last.scroll == model.scrolls[model.chosen]
}

// toMatch shows a match of the chosen tab, the one a choice picks among its matches, and says which of how many it
// is; or says the search found nothing in the tab.
func (model Model) toMatch(choose func(matches []int) int) Model {
	query, matches := model.search.query, model.drawn.tabs[model.chosen].matches
	if query == "" {
		return model
	}
	if len(matches) == 0 {
		model.said = commandMessage{text: notFoundError + query, failed: true}
		return model
	}
	place := choose(matches)
	model.said = commandMessage{text: searchMark + query + space + matchCount(place, len(matches)), failed: false}
	model = model.scrolledToward(matches[place], model.drawn.tabs[model.chosen].spotOf(namedAt(matches, place))).withMatchTileTop(matches[place])
	model.search.shown = slices.Clone(model.search.shown)
	model.search.shown[model.chosen] = shownMatch{place: place, scroll: model.scrolls[model.chosen], shown: true}
	return model.renamed()
}

// renamed is the model with the chosen tab's match its count names named anew: the line of the match named before
// and the line of the one named now painted again, and no other.
func (model Model) renamed() Model {
	drawn := model.drawn.tabs[model.chosen]
	marks := drawn.marks
	marks.named = model.namedMatchOf(model.chosen, drawn.matches)
	changed := []int{}
	for _, named := range []namedMatch{drawn.marks.named, marks.named} {
		if named.occurrence != noMatch {
			changed = append(changed, named.line)
		}
	}
	return model.repaintedLines(changed, marks)
}

// repaintedLines is the model with lines of the chosen tab painted again, each once, as new marks set their runs apart,
// and no other line.
func (model Model) repaintedLines(changed []int, marks pageMarks) Model {
	drawn := model.drawn.tabs[model.chosen]
	lines := slices.Clone(drawn.lines)
	slices.Sort(changed)
	for _, line := range slices.Compact(changed) {
		lines[line] = model.paint.pageLine(drawn.marked[line], model.frame, nil, marks.on(line))
	}
	model.drawn.tabs = slices.Clone(model.drawn.tabs)
	model.drawn.tabs[model.chosen].lines, model.drawn.tabs[model.chosen].marks = lines, marks
	return model
}

// matchCount says which match of how many a place among them is, as Neovim counts them: "[2/5]".
func matchCount(place, count int) string {
	return matchOpening + strconv.Itoa(place+1) + matchOf + strconv.Itoa(count) + matchClosing
}

// markedQuery is the search a tab marks: what the reader types, on the chosen tab while they type it; the search
// confirmed last everywhere else.
func (model Model) markedQuery(place int) string {
	if model.keys.mode == modeSearch && place == model.chosen {
		return model.keys.typing
	}
	return model.search.query
}

// matchAtOrAfter is the place of the first match on a line or after it, or the first of all after the last.
func matchAtOrAfter(matches []int, line int) int {
	return max(slices.IndexFunc(matches, func(matching int) bool { return matching >= line }), 0)
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

// matchedRun is a run of a line, and which match of the search on its line it is part of, counted from the line's
// first, noMatch where it is none.
type matchedRun struct {
	run
	occurrence int
}

// noMatch is the occurrence of a run that is no match, and of a line's named match where its line names none.
const noMatch = -1

// namedMatch is the match a tab's count names: its place among the tab's matches, its line, and which match of the
// line it is, noMatch where the tab names none.
type namedMatch struct {
	place      int
	line       int
	occurrence int
}

// noneNamed is a tab's named match where it names none.
func noneNamed() namedMatch {
	return namedMatch{place: noMatch, line: noMatch, occurrence: noMatch}
}

// occurrenceOn is which match of a line the count names, noMatch where it names none there.
func (named namedMatch) occurrenceOn(line int) int {
	if named.occurrence == noMatch || named.line != line {
		return noMatch
	}
	return named.occurrence
}

// namedMatchOf is the match a tab's count names among its matches: on the chosen tab while the reader types, the
// first at or after the line the search began at, as the line counts it; else the one the tab showed last.
func (model Model) namedMatchOf(place int, matches []int) namedMatch {
	if model.keys.mode == modeSearch && place == model.chosen {
		return namedAt(matches, matchAtOrAfter(matches, model.search.begunAt))
	}
	if place < len(model.search.shown) && model.search.shown[place].shown {
		return namedAt(matches, model.search.shown[place].place)
	}
	return noneNamed()
}

// namedAt is the match at a place among a tab's matches, in order, named; none where no match stands there. A line's
// matches stand side by side among them, so which of its line a match is counts from the line's first.
func namedAt(matches []int, place int) namedMatch {
	if place < 0 || place >= len(matches) {
		return noneNamed()
	}
	line := matches[place]
	return namedMatch{place: place, line: line, occurrence: place - slices.Index(matches, line)}
}

// markedLine is a line with each match of the search parted into a run of its own, and how many matches it holds.
type markedLine struct {
	runs       []matchedRun
	matchCount int
}

// markedLineOf is a line with each match of a query, whatever its case, marked; with no query, the line as it is.
func markedLineOf(shown line, query string) markedLine {
	letters := []rune{}
	for _, piece := range shown {
		letters = append(letters, []rune(piece.text)...)
	}
	queryLetters := []rune(query)
	starts := matchStarts(letters, queryLetters)
	return markedLine{runs: partedRuns(shown, matchMarks(starts, len(letters), len(queryLetters))), matchCount: len(starts)}
}

// matchStarts are the places of the letters each match of a query starts at, whatever its case. Matches never
// overlap, as Neovim's do not: the next is sought after the end of the one before.
func matchStarts(letters, query []rune) []int {
	starts := []int{}
	if len(query) == 0 {
		return starts
	}
	for start := 0; start+len(query) <= len(letters); {
		if !strings.EqualFold(string(letters[start:start+len(query)]), string(query)) {
			start++
			continue
		}
		starts = append(starts, start)
		start += len(query)
	}
	return starts
}

// matchMarks are which match each of a count of letters lies in, noMatch where none, the matches starting at places
// and each as long as the query.
func matchMarks(starts []int, letterCount, queryLength int) []int {
	marks := slices.Repeat([]int{noMatch}, letterCount)
	for occurrence, start := range starts {
		for index := range queryLength {
			marks[start+index] = occurrence
		}
	}
	return marks
}

// partedRuns are a line's runs parted where a letter's mark differs from the one before it, each part marked as
// its letters are.
func partedRuns(shown line, marks []int) []matchedRun {
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

func partOf(piece run, letters []rune, occurrence int) matchedRun {
	piece.text = string(letters)
	return matchedRun{run: piece, occurrence: occurrence}
}
