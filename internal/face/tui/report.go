package tui

import (
	"maps"
	"slices"
	"strconv"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/look"
)

// How the health report stands: the room between its title and the meter beside it, how far a check stands in under
// its group, and how far a check's advice stands in under the check's mark
const (
	meterGap     = 4
	checkIndent  = nestIndent
	adviceIndent = 2
)

// As Neovim's :checkhealth writes them: the mark after a group's name, and the word before a check's advice
const (
	groupMark  = "~"
	adviceWord = "ADVICE:"
)

// levelLook is how a check's level reads: the mark before it and the meter's cell for it, each told apart from the
// other levels' by its shape where colour cannot tell them, the word Neovim's :checkhealth writes it in, and the kind
// of all three.
type levelLook struct {
	mark string
	cell string
	word string
	kind kind
}

// levelLooks are how each level of a check reads. exhaustive holds the literal to every level, so a level the
// document gains reads its own way here too.
func levelLooks() map[document.CheckLevel]levelLook {
	return map[document.CheckLevel]levelLook{
		document.CheckOK:      {mark: look.DoneMark, cell: "█", word: "OK", kind: kindPassed},
		document.CheckWarning: {mark: "⚠", cell: "▒", word: "WARNING", kind: kindWarned},
		document.CheckError:   {mark: "✗", cell: "░", word: "ERROR", kind: kindFailed},
	}
}

// report lays out the health report's head: the title in the titles' font, and beside its last row, where the page
// has the room, the meter, a cell for each check of the opening and the checks counted by level after the cells;
// under the title where it has not.
func (layout layout) report(head document.Head, blocks []document.Block) []line {
	title := layout.titleRows(document.PlainText(head.Title))
	meter := meterWords(checksIn(blocks))
	if len(title) == 0 || len(meter) == 0 || linesWidth(title)+meterGap+wordsWidth(meter) > layout.width {
		return stacked(title, layout.fill(meter, 0))
	}
	last := len(title) - 1
	title[last] = slices.Concat(title[last], line{gapRun(linesWidth(title) - runsWidth(title[last]) + meterGap)}, wordRuns(meter))
	return title
}

// meterWords are the meter of checks: a cell for each, in the shape and the kind of its level, then the checks
// counted by level; nothing for no checks.
func meterWords(checks []document.Check) []word {
	looks, words := levelLooks(), []word{}
	for _, check := range checks {
		look := looks[check.Level]
		words = append(words, wordOf(look.cell, look.kind))
	}
	return slices.Concat(words, tallyWords(checks))
}

// checks lays out a group of checks as Neovim's :checkhealth does: the group's name and its mark, the group's checks
// counted by level at the page's right edge, or under the name where the row has not the room; then each check after
// the mark and the word of its level, its words wrapped under their start, and its advice after Neovim's word for it.
func (layout layout) checks(checks document.Checks, indent int) []line {
	lines := layout.fill(append(layout.words(checks.Title, kindStrong), textWords(groupMark, kindFaint)...), indent)
	tally := tallyWords(checks.List)
	if room := layout.width - runsWidth(lines[0]) - wordsWidth(tally); room >= columnGap {
		lines[0] = slices.Concat(lines[0], line{gapRun(room)}, wordRuns(tally))
	} else {
		lines = append(lines, layout.fill(tally, indent)...)
	}
	looks := levelLooks()
	for _, check := range checks.List {
		look := looks[check.Level]
		lead := line{gapRun(indent + checkIndent), plainRun(look.mark+space+look.word+space, look.kind)}
		lines = append(lines, layout.continued(lead, layout.words(check.Spans, kindText), runsWidth(lead))...)
		if len(check.Advice) > 0 {
			advice := line{gapRun(indent + checkIndent + adviceIndent), plainRun(adviceWord+space, kindMuted)}
			lines = append(lines, layout.continued(advice, layout.words(check.Advice, kindText), runsWidth(advice))...)
		}
	}
	return lines
}

// tallyWords are checks counted by level, in the order the document gives the levels, each level's count before
// Neovim's word for it, a word in its level's kind; each after the first opens with the place's mark, so a line
// breaks between two counts, never inside one nor before a mark.
func tallyWords(checks []document.Check) []word {
	looks, words := levelLooks(), []word{}
	for _, level := range slices.Sorted(maps.Keys(looks)) {
		count := checksOfLevel(checks, level)
		if count == 0 {
			continue
		}
		counted := plainRun(strconv.Itoa(count)+space+looks[level].word, looks[level].kind)
		if len(words) == 0 {
			words = append(words, wordOf(counted.text, counted.kind))
			continue
		}
		mark := plainRun(placeMark+space, kindMuted)
		words = append(words, word{runs: []run{mark, counted}, highlightBefore: false, spaceBefore: kindText})
	}
	return words
}

// checksIn are the checks blocks hold, the blocks of their sections included, in their order.
func checksIn(blocks []document.Block) []document.Check {
	checks := []document.Check{}
	for _, block := range blocks {
		switch typed := block.(type) {
		case document.Section:
			checks = append(checks, checksIn(typed.Blocks)...)
		case document.Reference:
			checks = append(checks, checksIn(typed.Blocks())...)
		case document.Checks:
			checks = append(checks, typed.List...)
		case document.Paragraph, document.Table, document.CodeBlock, document.List, document.Quote, document.Pause,
			document.Figure, document.Widget, document.Facts, document.ExperienceHead, document.Entry, document.Points, document.Side,
			document.Card, document.Contents, document.Skills, document.Posts:
		}
	}
	return checks
}

// checksOfLevel is how many of checks read at a level.
func checksOfLevel(checks []document.Check, level document.CheckLevel) int {
	count := 0
	for _, check := range checks {
		if check.Level == level {
			count++
		}
	}
	return count
}
