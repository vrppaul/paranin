package tui

import (
	"math"
	"slices"
	"strings"

	"github.com/vrppaul/paranin/internal/document"
)

// As lazy.nvim marks a plugin loaded or not, the mark of a skill in use and of one still being learnt
const (
	usedMark     = "●"
	learningMark = "○"
)

// How the list of skills draws each skill's months, as bars on one scale: the widest bar, the longest skill's, which
// stands beside the names only where it keeps that width, else under them; the shortest, of a span however short; the
// cell a bar is drawn in, thin so that two rows' bars never touch; and the mark that stands in for a bar where a skill
// has no measured time of its own
const (
	widestBar      = 30
	shortestBar    = 1
	barCell        = "━"
	unmeasuredMark = "┄┄┄"
)

// skillColumns are where the parts of the list of skills stand on the page: the names, after each skill's mark, as
// wide as the widest; the bars, beside the names or under them, as wide as the widest bar on the scale of the longest
// skill's months; how long each was used as the profile writes it, after the bars, as wide as the widest; and where
// each was proven, beside them or under the names.
type skillColumns struct {
	nameColumn    int
	nameWidth     int
	barsBeside    bool
	barWidth      int
	longestMonths int
	durationWidth int
	sourcesColumn int
	sourcesBeside bool
}

// skills lays out each skill as lazy.nvim lays out a plugin, after the mark of whether the author uses it or is still
// learning it: its name, a bar as long as its months on the scale of the longest skill's, or the mark of no measured
// time where it has none, how long as the profile writes it, and where it was proven, beside them where the page has
// the room and under the name where it has not. Where the row leaves the bars less than their widest beside the
// names, the bars stand under them, so a narrower page never draws them longer.
func (layout layout) skills(skills document.Skills, indent int) []line {
	columns := layout.skillColumns(skills.List, indent)
	lines := []line{}
	for _, skill := range skills.List {
		lines = append(lines, layout.skill(skill, indent, columns)...)
	}
	return lines
}

// skillColumns works out where the parts of the list of skills stand for the skills of a list at an indent.
func (layout layout) skillColumns(skills []document.Skill, indent int) skillColumns {
	nameWidth, durationWidth, longestMonths := 0, 0, 0
	for _, skill := range skills {
		nameWidth = max(nameWidth, wordsWidth(layout.words(skill.Name, kindStrong)))
		durationWidth = max(durationWidth, wordsWidth(layout.words(skill.Duration, kindMuted)))
		longestMonths = max(longestMonths, skill.Months)
	}
	nameColumn := indent + textWidth(usedMark+space)
	afterBar := columnGap + durationWidth
	barsBeside := nameColumn+nameWidth+columnGap+widestBar+afterBar <= layout.width
	barColumn := nameColumn
	if barsBeside {
		barColumn += nameWidth + columnGap
	}
	// the mark of no measured time stands in the bars' column, which is never narrower than the mark
	barWidth := min(widestBar, max(layout.width-barColumn-afterBar, textWidth(unmeasuredMark)))
	sourcesColumn := barColumn + barWidth + afterBar + columnGap
	return skillColumns{
		nameColumn: nameColumn, nameWidth: nameWidth, barsBeside: barsBeside, barWidth: barWidth,
		longestMonths: longestMonths, durationWidth: durationWidth, sourcesColumn: sourcesColumn,
		sourcesBeside: layout.width-sourcesColumn >= narrowestLastColumn,
	}
}

// skill lays out a skill in the list's columns: its mark and name, its bar and how long, beside the name or on a row
// of their own under it, and where it was proven, beside them or under the name.
func (layout layout) skill(skill document.Skill, indent int, columns skillColumns) []line {
	nameKind := kindStrong
	if skill.Learning {
		nameKind = kindMuted
	}
	name := layout.words(skill.Name, nameKind)
	lead := line{gapRun(indent), skillMark(skill.Learning)}
	lines := []line{}
	if columns.barsBeside {
		lead = slices.Concat(lead, wordRuns(name), line{gapRun(columns.nameWidth - wordsWidth(name) + columnGap)})
	} else {
		lines = layout.continued(lead, name, columns.nameColumn)
		lead = line{gapRun(columns.nameColumn)}
	}
	duration := layout.words(skill.Duration, kindMuted)
	timed := slices.Concat(lead, barRuns(skill.Months, columns), wordRuns(duration))
	sources := layout.words(skill.Sources, kindMuted)
	if columns.sourcesBeside {
		timed = append(timed, gapRun(columns.durationWidth-wordsWidth(duration)+columnGap))
		return append(lines, layout.continued(timed, sources, columns.sourcesColumn)...)
	}
	return slices.Concat(lines, []line{timed}, layout.fill(sources, columns.nameColumn))
}

// barRuns are a skill's bar on the list's scale and the gap after it to how long it was used: as many cells of the
// widest bar as its months are of the longest skill's, to the nearest cell and the shortest bar at least; the mark of
// no measured time where it has no months.
func barRuns(months int, columns skillColumns) line {
	if months == 0 {
		return line{
			{text: unmeasuredMark, kind: kindFaint, highlighted: false, address: "", copied: "", tone: 0},
			gapRun(columns.barWidth - textWidth(unmeasuredMark) + columnGap),
		}
	}
	cells := max(shortestBar, int(math.Round(float64(months)/float64(columns.longestMonths)*float64(columns.barWidth))))
	return line{
		{text: strings.Repeat(barCell, cells), kind: kindBar, highlighted: false, address: "", copied: "", tone: 0},
		gapRun(columns.barWidth - cells + columnGap),
	}
}

// skillMark is the mark before a skill: lazy.nvim's mark of a plugin loaded for one in use, of one not loaded for
// one still being learnt.
func skillMark(learning bool) run {
	if learning {
		return run{text: learningMark + space, kind: kindFaint, highlighted: false, address: "", copied: "", tone: 0}
	}
	return run{text: usedMark + space, kind: kindPassed, highlighted: false, address: "", copied: "", tone: 0}
}
