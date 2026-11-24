package tui

import (
	"image/color"
	"strings"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/look"
)

// The Outline's marks (DESIGN.md, Outline): the bar before the heading read, and the room it hangs in before the
// Outline's column, out in the page's margin, as the page hangs it
const (
	readMark    = "▌"
	outlineHang = 2
)

// The heading read where none is, one before the first, so the last of none is none
const noHeadingRead = -1

// The line gone to where the command line went to no section, above every line, so no page ever shows it
const noneGoneTo = -1

// laidOutline is a post's Outline laid out (DESIGN.md, Outline), standing in sight beside the post's body as it
// scrolls: the body's lines, from its first to the one past its last; each heading of the body's top level, in order;
// and the Outline's lines with no heading read, each opening with the room the bar hangs in.
type laidOutline struct {
	first    int
	end      int
	headings []outlinedHeading
	unread   []line
}

// outlinedHeading is a heading as the Outline lays it out: the line of the post it stands at, the first of the
// Outline's lines it takes, and those lines as they read once the heading is read, as many as unread.
type outlinedHeading struct {
	line  int
	first int
	read  []line
}

// noOutline is the Outline of a tab that has none, which nothing stands in sight for.
func noOutline() laidOutline {
	return laidOutline{first: 0, end: 0, headings: nil, unread: nil}
}

// outline lays out the Outline of a post's body, its blocks standing on lines from a first to an end, each heading
// at the line headings names it by: the headings of its top level, each title muted in the layout's width less the
// room the bar hangs in, and once read bold after the bar.
func (layout layout) outline(blocks []document.Block, headings map[string]int, first, end int) laidOutline {
	laid := laidOutline{first: first, end: end, headings: []outlinedHeading{}, unread: []line{}}
	for _, block := range blocks {
		section, isSection := block.(document.Section)
		if !isSection {
			continue
		}
		unread := layout.continued(line{gapRun(outlineHang)}, layout.words(section.Title, kindMuted), outlineHang)
		readLead := line{gapRun(0), plainRun(readMark, kindLookMark), gapRun(outlineHang - textWidth(readMark))}
		read := layout.continued(readLead, layout.words(section.Title, kindStrong), outlineHang)
		laid.headings = append(laid.headings, outlinedHeading{line: headings[section.Name], first: len(laid.unread), read: read})
		laid.unread = append(laid.unread, unread...)
	}
	return laid
}

// stuckLine is a line standing in sight over a row's left part, a group's name on the blog's list or a row of a
// post's Outline: its runs, from the room it hangs out by into the page's margin; that room; and the line painted
// from the screen's left edge, which a row the page fades paints again, as it paints the body's.
type stuckLine struct {
	marked  markedLine
	hang    int
	painted string
}

// stuckLineOf paints a line alone, unmarked by the search, to stand in sight over a row's left part: the page's
// margin but the room it hangs out by, then the line as wide as that room and the part, a split of the page's lines.
func (paint paint) stuckLineOf(laid line, hang, split int, frame frame) stuckLine {
	marked := markedLineOf(laid, "")
	return stuckLine{marked: marked, hang: hang, painted: paint.stuckRow(marked, hang, split, frame, nil)}
}

// stuckRow paints a line standing in sight from the screen's left edge, hanging out into the page's margin by a room,
// its text in a faded colour where one is given.
func (paint paint) stuckRow(marked markedLine, hang, split int, frame frame, faded color.Color) string {
	unmarked := lineMarks{named: noMatch, pointed: columnSpan{start: 0, end: 0}, split: 0}
	width := hang + split
	text, used := paint.line(marked.runs, 0, width, faded, unmarked)
	return strings.Repeat(space, max(frame.left-hang, 0)) + text + strings.Repeat(space, width-used)
}

// paintedOutline is a post's Outline painted alone, to stand in sight over its lines' left part: the lines it stands
// beside, from the first to the one past the last; each heading as painted read; and its rows with no heading read.
type paintedOutline struct {
	first    int
	end      int
	headings []paintedHeading
	unread   []stuckLine
}

// paintedHeading is a heading of the Outline painted: the line of the post it stands at, the first of the Outline's
// rows it takes, and those rows painted read.
type paintedHeading struct {
	line  int
	first int
	read  []stuckLine
}

// outline paints a post's Outline alone, its rows unread once and each heading's read once, each row hanging out into
// the page's margin by the room the bar hangs in, as wide as that room and a split of the post's lines.
func (paint paint) outline(laid laidOutline, split int, frame frame) paintedOutline {
	painted := paintedOutline{first: laid.first, end: laid.end, headings: make([]paintedHeading, 0, len(laid.headings)), unread: paint.stuckLines(laid.unread, outlineHang, split, frame)}
	for _, heading := range laid.headings {
		painted.headings = append(painted.headings, paintedHeading{line: heading.line, first: heading.first, read: paint.stuckLines(heading.read, outlineHang, split, frame)})
	}
	return painted
}

// stuckLines paints lines alone to stand in sight, each hanging out into the page's margin by a room.
func (paint paint) stuckLines(lines []line, hang, split int, frame frame) []stuckLine {
	stuck := make([]stuckLine, 0, len(lines))
	for _, laid := range lines {
		stuck = append(stuck, paint.stuckLineOf(laid, hang, split, frame))
	}
	return stuck
}

// readerView is where a reader stands in a tab, as its Outline reads it: how far the page is scrolled, how many rows
// it shows, how many lines the tab has, and the line of the section the command line last went to, noneGoneTo where
// it went to none.
type readerView struct {
	scroll int
	rows   int
	lines  int
	goneTo int
}

// read is the heading the Outline marks read where a reader stands, as DESIGN.md's Outline says, by look.ReadingLine;
// noHeadingRead where none is.
func (outline paintedOutline) read(view readerView) int {
	atEnd, reading := view.scroll+view.rows >= view.lines, view.scroll+view.rows*look.ReadingLine/wholePercent
	reach := reading
	if atEnd {
		reach = view.scroll + view.rows
	}
	if view.goneTo >= view.scroll && view.goneTo < reach {
		return outline.lastBefore(view.goneTo + 1)
	}
	if atEnd {
		return len(outline.headings) - 1
	}
	return outline.lastBefore(reading)
}

// lastBefore is the last heading whose line stands before a line, noHeadingRead where none does.
func (outline paintedOutline) lastBefore(line int) int {
	last := noHeadingRead
	for index, heading := range outline.headings {
		if heading.line < line {
			last = index
		}
	}
	return last
}

// rowAt is the Outline's row of a place among its rows with a heading read: the read heading's own painted read, any
// other unread.
func (outline paintedOutline) rowAt(place, read int) stuckLine {
	if read == noHeadingRead {
		return outline.unread[place]
	}
	if heading := outline.headings[read]; place >= heading.first && place < heading.first+len(heading.read) {
		return heading.read[place-heading.first]
	}
	return outline.unread[place]
}

// stuckRows are, for each row of a tab's page where a reader stands, what stands in sight over the row's left part,
// none where nothing does: on the blog's list a group's name, as stuckNames has it; on a post its Outline, with the
// heading read marked, at the row a group's name keeps from the page's top, or lower where its body starts lower, and
// rising with its body's end, beside its body's lines alone.
func stuckRows(painted stuckPaint, view readerView) []stuckLine {
	stuck := stuckNames(painted.groups, view.scroll, view.rows)
	outline := painted.outline
	read := outline.read(view)
	top := max(min(view.scroll+stuckRow, outline.end-len(outline.unread)), outline.first)
	// the Outline's rows on the page and beside the body's lines
	for place := max(view.scroll-top, 0); place < min(len(outline.unread), view.scroll+view.rows-top, outline.end-top); place++ {
		stuck[top+place-view.scroll] = outline.rowAt(place, read)
	}
	return stuck
}

// goneToLine is the line of the section the command line last went to in the chosen tab, noneGoneTo where it went to
// none, which no tab's sections are named by.
func (model Model) goneToLine() int {
	line, laid := model.laid[model.chosen].sections[model.positions[model.chosen].goneTo]
	if !laid {
		return noneGoneTo
	}
	return line
}
