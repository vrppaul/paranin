package tui

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/vrppaul/paranin/internal/document"
)

// The blog's list's columns (DESIGN.md, The TUI): the groups' names at the left, narrower on a page under the widest
// and none on a page under the least, where a group's name stands over its tiles; the gap after them; the least width
// of a tile, its left edge with it, as the page's tiles keep thirty cells; the room inside a tile at each side and
// above and below its words; and how far into its top edge a tile's date stands
const (
	nameColumnWide   = 22
	nameColumnNarrow = 18
	nameColumnLeast  = 100
	nameGap          = 4
	leastTile        = 31
	tilePadding      = 2
	tileRowsAround   = 1
	tileLabelInset   = 2
)

// The list's head and groups: the width the blog's line wraps at, as the page holds its lead to a measure; the gap
// between the way the list is grouped and its tags, and between two tags; the rows between the head and the first
// group and between two groups; and the row a group's name keeps from the page's top while its tiles scroll under it,
// as a post's Outline keeps while its body scrolls
const (
	blogLineMeasure = 52
	waysGap         = 4
	tagGap          = 2
	groupRows       = 2
	stuckRow        = 1
)

// The mark before a post's tag, as readers know a tag that gathers posts (DESIGN.md, Content)
const tagMark = "#"

// laidLink is a link of the blog's list a key goes to (DESIGN.md, The TUI), a way its head offers or a post's tile:
// the page Enter opens; its first line and its last, which a key shows whole, a tile's edges among them; and the
// ground the keys' mark lays on it. The zero link is none.
type laidLink struct {
	page   document.Page
	top    int
	bottom int
	ground linkGround
}

// linkGround is the cells a link's mark lays the panel's ground on: a way's words, or the rows between a tile's edges
// over the columns between them.
type linkGround struct {
	lines   lineSpan
	columns columnSpan
}

// tileLink is a post's tile as a link: from its top edge to its bottom edge, its ground the rows between them over
// columns.
func tileLink(page document.Page, top, bottom int, columns columnSpan) laidLink {
	return laidLink{page: page, top: top, bottom: bottom, ground: linkGround{lines: lineSpan{first: top + 1, end: bottom}, columns: columns}}
}

// wordsLink is words on a line that open a page as a link, its ground the columns they take.
func wordsLink(page document.Page, line int, columns columnSpan) laidLink {
	return laidLink{page: page, top: line, bottom: line, ground: linkGround{lines: lineSpan{first: line, end: line + 1}, columns: columns}}
}

// groundOn is the columns the link's ground takes on a line: none on a line it does not cross.
func (link laidLink) groundOn(line int) columnSpan {
	if !link.ground.lines.holds(line) {
		return columnSpan{start: 0, end: 0}
	}
	return link.ground.columns
}

// shifted is the link a number of lines lower and of columns further right.
func (link laidLink) shifted(lines, columns int) laidLink {
	ground := link.ground
	link.top, link.bottom = link.top+lines, link.bottom+lines
	link.ground = linkGround{
		lines:   lineSpan{first: ground.lines.first + lines, end: ground.lines.end + lines},
		columns: columnSpan{start: ground.columns.start + columns, end: ground.columns.end + columns},
	}
	return link
}

// spot is a place on a tab's lines, its line and its column, which a move went to or a click fell on.
type spot struct {
	line   int
	column int
}

// nearer is whether the link lies nearer a spot than another: fewer lines between them, or as few and fewer columns.
func (link laidLink) nearer(than laidLink, near spot) bool {
	lines, otherLines := max(link.top-near.line, near.line-link.bottom, 0), max(than.top-near.line, near.line-than.bottom, 0)
	if lines != otherLines {
		return lines < otherLines
	}
	columns, otherColumns := link.ground.columns, than.ground.columns
	return max(columns.start-near.column, near.column-columns.end, 0) < max(otherColumns.start-near.column, near.column-otherColumns.end, 0)
}

// holds is whether a spot lies on the link: on a line from its first to its last, a tile's edges among them, and in its
// ground's columns.
func (link laidLink) holds(at spot) bool {
	return at.line >= link.top && at.line <= link.bottom && link.ground.columns.holds(at.column)
}

// laidGroup is a group of the blog's list whose name stands in the names' column: the lines its tiles take, from its
// first to the one past its last, and its name's lines, which stay in sight while its tiles scroll.
type laidGroup struct {
	first int
	end   int
	name  []line
}

// junction is which lines meet at a point of a grid of tiles: from above, below, the left and the right.
type junction struct {
	above bool
	below bool
	left  bool
	right bool
}

// junctionMarks are the mark drawn at each meeting of a grid's lines, as tmux draws the borders between its panes.
func junctionMarks() map[junction]string {
	return map[junction]string{
		{above: false, below: true, left: false, right: true}:  frameTopLeft,
		{above: false, below: true, left: true, right: true}:   frameTopJoin,
		{above: false, below: true, left: true, right: false}:  frameTopRight,
		{above: true, below: true, left: false, right: true}:   frameRuleLeft,
		{above: true, below: true, left: true, right: true}:    frameCrossing,
		{above: true, below: true, left: true, right: false}:   frameRuleRight,
		{above: true, below: false, left: false, right: true}:  frameBottomLeft,
		{above: true, below: false, left: true, right: true}:   frameBottomJoin,
		{above: true, below: false, left: true, right: false}:  frameBottomRight,
		{above: false, below: false, left: true, right: true}:  frameAcross,
		{above: false, below: false, left: true, right: false}: frameAcross,
		{above: false, below: false, left: false, right: true}: frameAcross,
	}
}

// list lays out the blog's list (DESIGN.md, The TUI): its head, then each group's tiles, the group's name in the
// column at its left where the page is wide enough, over its tiles where not; with the links a key goes to, the head's
// and the tiles', and the groups whose names stay in sight, and the line each series' group starts at. A block that is
// no group, which compose writes none of, stands as any page lays it out, as the page draws it.
func (layout layout) list(head document.Head, blocks []document.Block) laidTab {
	column := nameColumnOf(layout.width)
	headLines := layout.blogHead(head, column)
	links := headLinks(headLines)
	laid := laidTab{lines: headLines, links: links, opening: 0, groups: []laidGroup{}, split: 0, sections: map[string]int{}, outline: noOutline()}
	if column > 0 {
		laid.split = column + nameGap
	}
	for _, block := range blocks {
		laid.lines = append(laid.lines, make([]line, groupRows)...)
		group, isGroup := document.GroupOf(block)
		if group.Name != "" {
			laid.sections[group.Name] = len(laid.lines)
		}
		switch {
		case !isGroup:
			laid.lines = append(laid.lines, layout.blocks([]document.Block{block}, 0)...)
		case column == 0:
			laid = layout.groupOver(laid, group)
		default:
			laid = layout.groupBeside(laid, group, column)
		}
	}
	// the list opens on its first tile, after the head's links, or on its first link where it has no tile
	if len(laid.links) > len(links) {
		laid.opening = len(links)
	}
	return laid
}

// nameColumnOf is how wide the groups' names' column stands on a page of a width, none on a narrow one.
func nameColumnOf(width int) int {
	switch {
	case width >= widestBlogPage:
		return nameColumnWide
	case width >= nameColumnLeast:
		return nameColumnNarrow
	}
	return 0
}

// blogHead is the list's head (DESIGN.md, BlogHead): the blog's name in the titles' font, in the names' column; beside
// it the blog's line, then the way the list is grouped, the grouping shown chosen, and the tags of the posts it lists;
// under the name where the page has no column for it.
func (layout layout) blogHead(head document.Head, column int) []line {
	named, about := layout, layout
	if column > 0 {
		named.width, about.width = column, layout.width-column-nameGap
	}
	measured := about
	measured.width = min(about.width, blogLineMeasure)
	ways := about.waysLines(head)
	right := slices.Concat(measured.fill(about.words(head.Lead, kindText), 0), blankLines(), ways)
	letters := named.titleRows(document.PlainText(head.Title))
	if column == 0 {
		return stacked(letters, right)
	}
	return besideLines(letters, column+nameGap, right)
}

// waysLines are the ways the list is grouped (compose.ListWords.Groupings), after the site's word before them, the
// grouping shown chosen, a filtered list's the one it filters; then its filters, the tags of the posts it lists; the
// filters on lines of their own where both do not fit a line. Each way and each filter opens the list it leads to.
func (layout layout) waysLines(head document.Head) []line {
	listWords := layout.siteWords.ListWords
	grouped, _ := head.Page.Unfiltered()
	grouping := []word{wordOf(listWords.By, kindMuted)}
	for _, way := range listWords.Groupings(layout.siteWords.Translation) {
		shown := kindText
		if way.List == grouped {
			shown = kindChosen
		}
		grouping = append(grouping, chipWord(way.Words, shown, way.List))
	}
	filters := make([]word, 0, len(head.Filters))
	for _, filter := range head.Filters {
		filters = append(filters, filterWord(filter, filter.Page))
	}
	ways, tags := wordRuns(grouping), tagRuns(filters)
	if len(tags) == 0 {
		return []line{ways}
	}
	if runsWidth(ways)+waysGap+runsWidth(tags) <= layout.width {
		return []line{slices.Concat(ways, line{gapRun(waysGap)}, tags)}
	}
	return slices.Concat([]line{ways}, blankLines(), layout.fill(filters, 0))
}

// filterWord is a filter a head offers as the page writes it (DESIGN.md, BlogHead and Tag), opening a page, none where
// the keys never stand on it: its tag's muted word after a faint hash; the filter the list is filtered by chosen whole,
// hash and all, a space either side, so it stands as one chip.
func filterWord(filter document.Filter, opens document.Page) word {
	if !filter.Current {
		return tagWord(filter.Tag, kindMuted, opens)
	}
	chosen := chipWord(tagMark+filter.Tag, kindChosen, opens)
	chosen.spaceBefore = kindChosen
	return chosen
}

// chipWord is a text of a kind standing as a chip, opening a page: a space either side, each a run of its own that
// opens none, so a link's ground and underline keep to the text, as lipgloss draws the spaces of an underlined run
// without their reversal.
func chipWord(text string, shown kind, opens document.Page) word {
	padding := plainRun(space, shown)
	return word{runs: []run{padding, linkRun(text, shown, opens), padding}, highlightBefore: false, spaceBefore: kindText}
}

// tagWords are tags as the page writes them (DESIGN.md, Tag), each tag's word of a kind, opening no page.
func tagWords(tags []string, wordKind kind) []word {
	words := make([]word, 0, len(tags))
	for _, tag := range tags {
		words = append(words, tagWord(tag, wordKind, document.PageOf(document.PageNone)))
	}
	return words
}

// tagWord is a tag as the page writes it (DESIGN.md, Tag), opening a page: its word after a faint hash, the word of a
// kind.
func tagWord(tag string, wordKind kind, opens document.Page) word {
	return word{runs: []run{linkRun(tagMark, kindFaint, opens), linkRun(tag, wordKind, opens)}, highlightBefore: false, spaceBefore: kindText}
}

// tagRuns are filters on one line, each two the page's gap apart.
func tagRuns(filters []word) line {
	runs := line{}
	for index, filter := range filters {
		if index > 0 {
			runs = append(runs, gapRun(tagGap))
		}
		runs = append(runs, filter.runs...)
	}
	return runs
}

// headLinks are the links among the lines of the list's head, in reading order: runs side by side on a line that open
// one page, as a tag's hash and word, are one link.
func headLinks(lines []line) []laidLink {
	links := []laidLink{}
	for row, shown := range lines {
		column := 0
		for _, piece := range shown {
			end := column + textWidth(piece.text)
			if piece.opensPage() {
				links = withLinkRun(links, piece.opens, row, columnSpan{start: column, end: end})
			}
			column = end
		}
	}
	return links
}

// withLinkRun is links with a run that opens a page at columns of a row: the last link widened over it where the run
// goes on from it, opening the same page; else a link of its own after them.
func withLinkRun(links []laidLink, opens document.Page, row int, columns columnSpan) []laidLink {
	last := len(links) - 1
	if last >= 0 && links[last].page == opens && links[last].top == row && links[last].ground.columns.end == columns.start {
		links[last].ground.columns.end = columns.end
		return links
	}
	return append(links, wordsLink(opens, row, columns))
}

// groupOver lays out a group on a page with no names' column: its name in the gradient and how many posts it holds on
// one line over its tiles, a series' line under them.
func (layout layout) groupOver(laid laidTab, group document.Group) laidTab {
	name := append(append(line{gapRun(0)}, wordRuns(gradientWords(document.PlainText(group.Title), kindHeading))...), gapRun(waysGap))
	heading := slices.Concat(layout.continued(name, layout.words(group.Posts.Count, kindMuted), 0), layout.fill(layout.words(group.Posts.Line, kindMuted), 0))
	laid.lines = slices.Concat(laid.lines, heading, blankLines())
	tiles := layout.tiles(group.Posts.List)
	return laid.joined(tiles.lines, tiles.links, 0)
}

// groupBeside lays out a group beside its name, in the names' column of a width: the name in the titles' font, a
// series' line and how many posts it holds under it, standing in sight while the group's tiles scroll.
func (layout layout) groupBeside(laid laidTab, group document.Group, column int) laidTab {
	named, tiled := layout, layout
	named.width, tiled.width = column, layout.width-column-nameGap
	about := slices.Concat(named.fill(named.words(group.Posts.Line, kindMuted), 0), named.fill(named.words(group.Posts.Count, kindMuted), 0))
	name := stacked(named.titleRows(document.PlainText(group.Title)), about)
	first, tiles := len(laid.lines), tiled.tiles(group.Posts.List)
	laid = laid.joined(besideLines(name, column+nameGap, tiles.lines), tiles.links, column+nameGap)
	laid.groups = append(laid.groups, laidGroup{first: first, end: len(laid.lines), name: name})
	return laid
}

// tiles lays out posts as a grid of tiles: each row of tiles as tall as its tallest, the tiles sharing their edges, as
// tmux's panes share theirs. A post out is a tile a key goes to, a link; a part to come is none.
func (layout layout) tiles(posts []document.ListedPost) laidTab {
	grid := gridOf(layout.width)
	inside := layout
	inside.width = grid.inside()
	cells := make([]gridCell, 0, len(posts))
	for _, post := range posts {
		_, isOut := post.Readiness.(document.PostOut)
		cells = append(cells, gridCell{label: layout.edgeLabel(post), words: inside.tileWordsOf(post), leads: isOut})
	}
	framed := grid.framed(cells)
	laid := laidTab{lines: framed.lines, links: []laidLink{}, opening: 0, groups: nil, split: 0, sections: nil, outline: noOutline()}
	for index, post := range posts {
		if out, isOut := post.Readiness.(document.PostOut); isOut {
			row, left := index/grid.columns, index%grid.columns*grid.width+textWidth(frameDown)
			columns := columnSpan{start: left, end: left + grid.width - textWidth(frameDown)}
			laid.links = append(laid.links, tileLink(out.Page, framed.edges[row], framed.edges[row+1], columns))
		}
	}
	return laid
}

// gridCell is a cell of a grid framed as tmux frames its panes, a tile of the blog's list or a pane of a post's foot:
// what is set into its top edge, its words, and whether it leads somewhere, which its frame says.
type gridCell struct {
	label line
	words tileWords
	leads bool
}

// framedGrid is a grid of cells laid out: its lines, and the line of each row's top edge, then of the grid's bottom
// edge, so a row of cells stands between two of them.
type framedGrid struct {
	lines []line
	edges []int
}

// framed lays out cells as a grid, as many in a row as the grid has columns, each row as tall as its tallest cell, the
// cells sharing their edges.
func (grid tileGrid) framed(cells []gridCell) framedGrid {
	laid, above := framedGrid{lines: []line{}, edges: []int{}}, []bool{}
	for row := range slices.Chunk(cells, grid.columns) {
		labels, words, leads := make([]line, 0, len(row)), make([]tileWords, 0, len(row)), make([]bool, 0, len(row))
		for _, cell := range row {
			labels, words, leads = append(labels, cell.label), append(words, cell.words), append(leads, cell.leads)
		}
		laid.edges = append(laid.edges, len(laid.lines))
		laid.lines = append(append(laid.lines, grid.edgeLine(above, leads, labels)), grid.tileRows(words, leads)...)
		above = leads
	}
	laid.edges = append(laid.edges, len(laid.lines))
	laid.lines = append(laid.lines, grid.edgeLine(above, nil, nil))
	return laid
}

// joined is a laid list with lines after its own, among them tiles' links laid out from the lines' top, standing a
// number of columns in from the page's left edge.
func (laid laidTab) joined(lines []line, tiles []laidLink, left int) laidTab {
	first := len(laid.lines)
	for _, tile := range tiles {
		laid.links = append(laid.links, tile.shifted(first, left))
	}
	laid.lines = append(laid.lines, lines...)
	return laid
}

// tileGrid is how the list's tiles, or a post's panes, share a width: how many stand in a row, how wide each is with
// its left edge, and how much wider the last column stands, so a row fills the width whole.
type tileGrid struct {
	columns   int
	width     int
	lastWider int
}

// gridOf is the grid of tiles over a width: as many in a row as are at least the least tile wide, their right edge
// closing the row, the room left over after it.
func gridOf(width int) tileGrid {
	room := width - textWidth(frameDown)
	columns := max(room/leastTile, 1)
	return tileGrid{columns: columns, width: room / columns, lastWider: 0}
}

// filledGrid is a grid of a number of columns sharing a width whole, the last column taking what is left over.
func filledGrid(columns, width int) tileGrid {
	room := width - textWidth(frameDown)
	return tileGrid{columns: columns, width: room / columns, lastWider: room % columns}
}

// widthOf is how wide a column of the grid stands with its left edge.
func (grid tileGrid) widthOf(column int) int {
	if column == grid.columns-1 {
		return grid.width + grid.lastWider
	}
	return grid.width
}

// inside is the room for a tile's words, its padding at each side gone.
func (grid tileGrid) inside() int {
	return grid.width - textWidth(frameDown) - sides*tilePadding
}

// tileWords are a tile's or a pane's words laid out: its title and lead at its top, and at its foot where it stands,
// after a row of room.
type tileWords struct {
	top  []line
	foot []line
}

// tileWordsOf lays out a tile's words (DESIGN.md, Tile): a post out by its title bold and its lead muted, a part to
// come by its title faint and not bold; at its foot its part and its series where its group does not name the series,
// and its tags, faint. A post out's title opens it.
func (layout layout) tileWordsOf(post document.ListedPost) tileWords {
	top := []line{}
	switch typed := post.Readiness.(type) {
	case document.PostOut:
		title := titleLines(layout.fill(textWords(document.PlainText(post.Title), kindStrong), 0), typed.Page)
		top = slices.Concat(title, layout.fill(layout.words(post.Lead, kindMuted), 0))
	case document.ComingPart:
		top = layout.fill(layout.words(post.Title, kindFaint), 0)
	}
	foot := blankLines()
	if len(post.Series) > 0 {
		series := []word{wordOf(document.PlainText(post.Part), kindStrong), wordOf(placeMark, kindFaint)}
		foot = append(foot, layout.fill(append(series, layout.words(post.Series, kindText)...), 0)...)
	}
	return tileWords{top: top, foot: append(foot, layout.fill(tagWords(post.Tags, kindFaint), 0)...)}
}

// titleLines are a tile's title's lines, opening its post's page, each one run after its indent, so the tile the keys
// stand on underlines the spaces between its words too, as the page underlines a link whole.
func titleLines(lines []line, page document.Page) []line {
	whole := make([]line, 0, len(lines))
	for _, shown := range lines {
		var text strings.Builder
		for _, piece := range shown[1:] {
			text.WriteString(piece.text)
		}
		whole = append(whole, line{shown[0], linkRun(text.String(), kindStrong, page)})
	}
	return whole
}

// edgeLabel is what a tile's top edge holds (DESIGN.md, Tile): when its post came out, in the site's words for a day,
// muted, after its part, bold, where it stands in its series' own group; for a part to come, its part and the word
// saying it comes, faint.
func (layout layout) edgeLabel(post document.ListedPost) line {
	partKind, label := kindStrong, line{}
	if _, isComing := post.Readiness.(document.ComingPart); isComing {
		partKind = kindFaint
	}
	if post.InItsSeriesGroup() {
		label = line{plainRun(document.PlainText(post.Part), partKind), plainRun(space+placeMark+space, kindFaint)}
	}
	switch typed := post.Readiness.(type) {
	case document.PostOut:
		return append(label, plainRun(layout.siteWords.Days.Of(typed.Published), kindMuted))
	case document.ComingPart:
		return append(label, plainRun(document.PlainText(typed.Coming), kindFaint))
	}
	return label
}

// edgeLine is the grid's line across between a row of tiles above and one below, by whether each tile leads
// somewhere, each label of those below, where it has one, set into its top edge; each piece of it framed as the tiles
// it bounds have it.
func (grid tileGrid) edgeLine(above, below []bool, labels []line) line {
	marks, widest := junctionMarks(), max(len(above), len(below))
	edge := line{gapRun(0)}
	for point := 0; point <= widest; point++ {
		across := grid.widthOf(point) - textWidth(frameDown)
		meeting := junction{above: point <= len(above) && len(above) > 0, below: point <= len(below) && len(below) > 0, left: point > 0, right: point < widest}
		around := frameKind(leadsAt(above, point-1), leadsAt(above, point), leadsAt(below, point-1), leadsAt(below, point))
		edge = append(edge, plainRun(marks[meeting], around))
		beside := frameKind(leadsAt(above, point), leadsAt(below, point))
		switch {
		case point == widest:
		case point < len(below) && len(labels[point]) > 0:
			label := slices.Concat(line{gapRun(spaceWidth)}, labels[point], line{gapRun(spaceWidth)})
			rest := max(across-tileLabelInset-runsWidth(label), 0)
			edge = slices.Concat(edge, line{plainRun(strings.Repeat(frameAcross, tileLabelInset), beside)}, label, line{plainRun(strings.Repeat(frameAcross, rest), beside)})
		default:
			edge = append(edge, plainRun(strings.Repeat(frameAcross, across), beside))
		}
	}
	return edge
}

// tileRows are a row of tiles' rows, by whether each tile leads somewhere: as many as the tallest tile's words and the
// room above and below them, each tile's foot standing on its last row of room, each edge down framed as the tiles
// beside it have it.
func (grid tileGrid) tileRows(tiles []tileWords, leads []bool) []line {
	topRows, footRows := 0, 0
	for _, tile := range tiles {
		topRows, footRows = max(topRows, len(tile.top)), max(footRows, len(tile.foot))
	}
	height := tileRowsAround + topRows + footRows + tileRowsAround
	footStart := height - tileRowsAround - footRows
	lines := make([]line, 0, height)
	for row := range height {
		shown := line{gapRun(0)}
		for column, tile := range tiles {
			cell := append(line{gapRun(tilePadding)}, tile.wordsOn(row, footStart)...)
			edge := plainRun(frameDown, frameKind(leadsAt(leads, column-1), leadsAt(leads, column)))
			shown = slices.Concat(shown, line{edge}, cell, line{gapRun(grid.widthOf(column) - textWidth(frameDown) - runsWidth(cell))})
		}
		lines = append(lines, append(shown, plainRun(frameDown, frameKind(leadsAt(leads, len(leads)-1)))))
	}
	return lines
}

// leadsAt is whether the tile at a column of a row leads somewhere; none does past the row's ends.
func leadsAt(row []bool, column int) bool {
	return column >= 0 && column < len(row) && row[column]
}

// frameKind is how a piece of a frame reads by the tiles or panes it bounds (DESIGN.md, Colour): muted where one of
// them leads somewhere, off where none does.
func frameKind(leads ...bool) kind {
	if slices.Contains(leads, true) {
		return kindMuted
	}
	return kindOff
}

// wordsOn is a tile's words on a row of it: its title and lead under the row of room at its top, its foot from the row
// it starts at; nothing on any other row.
func (tile tileWords) wordsOn(row, footStart int) line {
	switch {
	case row >= tileRowsAround && row-tileRowsAround < len(tile.top):
		return tile.top[row-tileRowsAround]
	case row >= footStart && row-footStart < len(tile.foot):
		return tile.foot[row-footStart]
	}
	return line{}
}

// paintedGroup is a group's name painted alone, to stand in sight while its tiles scroll: the lines its tiles take,
// from its first to the one past its last; its name's rows and a row of nothing, each as wide as the names' column and
// the gap after it.
type paintedGroup struct {
	first int
	end   int
	rows  []stuckLine
	blank stuckLine
}

// groupNames paints the names of the blog's list's groups alone, once, each row as wide as the names' column and its
// gap, the split of the list's lines.
func (paint paint) groupNames(groups []laidGroup, split int, frame frame) []paintedGroup {
	painted := make([]paintedGroup, 0, len(groups))
	blank := paint.stuckLineOf(line{}, 0, split, frame)
	for _, group := range groups {
		painted = append(painted, paintedGroup{first: group.first, end: group.end, rows: paint.stuckLines(group.name, 0, split, frame), blank: blank})
	}
	return painted
}

// stuckNames are, for each row of a page of a number of rows at a scroll, the name's row stuck in sight over the
// row's left part, none where none is: each group whose top has scrolled past the row a name keeps from the page's top
// keeps its name at that row, its other rows blank; where the group's tiles end above the name's last row, the name
// rises with them.
func stuckNames(groups []paintedGroup, scroll, pageRows int) []stuckLine {
	stuck := make([]stuckLine, pageRows)
	for _, group := range groups {
		nameTop := min(scroll+stuckRow, group.end-len(group.rows))
		if nameTop <= group.first {
			continue
		}
		for row := max(group.first-scroll, 0); row < min(group.end-scroll, pageRows); row++ {
			stuck[row] = group.blank
			if nameRow := scroll + row - nameTop; nameRow >= 0 && nameRow < len(group.rows) {
				stuck[row] = group.rows[nameRow]
			}
		}
	}
	return stuck
}

// linkStoodOn is the place among a list's links of the one the keys stand on, which the tab's position counts from the
// link the list opens on.
func (model Model) linkStoodOn(place int) int {
	return model.laid[place].opening + model.positions[place].link
}

// pointedLink is the link the keys stand on in a tab, the zero link where the tab holds no links.
func (model Model) pointedLink(place int) laidLink {
	links := model.laid[place].links
	if len(links) == 0 {
		noGround := linkGround{lines: lineSpan{first: 0, end: 0}, columns: columnSpan{start: 0, end: 0}}
		return laidLink{page: document.PageOf(document.PageNone), top: 0, bottom: 0, ground: noGround}
	}
	return links[model.linkStoodOn(place)]
}

// onList is whether the tab chosen is the blog's list, whose keys go from link to link.
func (model Model) onList() bool {
	return len(model.laid[model.chosen].links) > 0
}

// downward is j's and the down arrow's: on the blog's list a link on, else a line down.
func (model Model) downward() (Model, tea.Cmd) {
	return model.movedBy(forwardStep), nil
}

// upward is k's and the up arrow's: on the blog's list a link back, else a line up.
func (model Model) upward() (Model, tea.Cmd) {
	return model.movedBy(backwardStep), nil
}

// movedBy is the model a step on or back: on the blog's list a link, else a line.
func (model Model) movedBy(step int) Model {
	if model.onList() {
		return model.toLink(model.linkStoodOn(model.chosen) + step)
	}
	return model.scrolledTo(model.positions[model.chosen].scroll + step)
}

// toLink has the keys stand on a link of the chosen list, by its place in reading order, and the page scrolled to show
// it whole; past the first link or the last the keys stay and the page scrolls by the line the step went past, so the
// list's head and foot come in sight.
func (model Model) toLink(wanted int) Model {
	links, scroll := model.laid[model.chosen].links, model.positions[model.chosen].scroll
	chosen := min(max(wanted, 0), len(links)-1)
	if chosen == model.linkStoodOn(model.chosen) {
		return model.scrolledTo(scroll + wanted - chosen)
	}
	model = model.choseLink(chosen)
	return model.scrolledTo(model.scrollShowing(model.chosen, links[chosen]))
}

// choseLink has the keys stand on a link of the chosen list, by its place, the link left and the link chosen painted
// again and no other line.
func (model Model) choseLink(chosen int) Model {
	laid := model.laid[model.chosen]
	left, link := model.pointedLink(model.chosen), laid.links[chosen]
	marks := model.drawn.tabs[model.chosen].marks
	marks.pointed = link
	model.positions = slices.Clone(model.positions)
	model.positions[model.chosen].link = chosen - laid.opening
	return model.repaintedLines(slices.Concat(linesBetween(left.top, left.bottom), linesBetween(link.top, link.bottom)), marks)
}

// withChosenLinkNear is the model where, on the blog's list chosen, the link the keys stand on has gone out of sight:
// the keys stand on the link nearest a spot the page went to, among those in sight where any is.
func (model Model) withChosenLinkNear(near spot) Model {
	links, chosen := model.laid[model.chosen].links, model.chosen
	if len(links) == 0 || model.inSight(chosen, model.pointedLink(chosen)) {
		return model
	}
	nearest := 0
	for index, link := range links {
		best := links[nearest]
		if model.inSight(chosen, link) != model.inSight(chosen, best) {
			if model.inSight(chosen, link) {
				nearest = index
			}
			continue
		}
		if link.nearer(best, near) {
			nearest = index
		}
	}
	return model.choseLink(nearest)
}

// withFirstLinkFrom is the model, on the blog's list chosen, with the keys on the first link at or after a line, where
// any is.
func (model Model) withFirstLinkFrom(line int) Model {
	first := slices.IndexFunc(model.laid[model.chosen].links, func(link laidLink) bool { return link.top >= line })
	if first < 0 || first == model.linkStoodOn(model.chosen) {
		return model
	}
	return model.choseLink(first)
}

// inSight is whether a line of a link shows on a tab's page at its scroll.
func (model Model) inSight(place int, link laidLink) bool {
	scroll := model.positions[place].scroll
	return link.bottom >= scroll && link.top < scroll+model.frame.rows
}

// withLinksInSight is the model with each list among the tabs whose link the keys stand on is out of sight scrolled
// as little as shows it whole; a list whose link shows stays where it stands, at its top as it opens.
func (model Model) withLinksInSight() Model {
	model.positions = slices.Clone(model.positions)
	for place := range model.tabs {
		if pointed := model.pointedLink(place); len(model.laid[place].links) > 0 && !model.inSight(place, pointed) {
			model.positions[place].scroll = model.scrollShowing(place, pointed)
		}
	}
	return model.clamped()
}

// withMatchLinkTop is the model, on the blog's list, with the page scrolled up where it shows a line of the link the
// keys stand on but not its first, so a tile shows its date, as j shows it.
func (model Model) withMatchLinkTop(line int) Model {
	link := model.pointedLink(model.chosen)
	if !model.onList() || line < link.top || line > link.bottom || link.top >= model.positions[model.chosen].scroll {
		return model
	}
	return model.scrolledTo(link.top)
}

// linesBetween are the lines from a first to a last, both among them.
func linesBetween(first, last int) []int {
	lines := make([]int, 0, last-first+1)
	for line := first; line <= last; line++ {
		lines = append(lines, line)
	}
	return lines
}

// scrollShowing is the scroll of a tab nearest its own that shows a link whole, a tile's edges with it, or its top
// where the page is too short for the whole.
func (model Model) scrollShowing(place int, link laidLink) int {
	return min(max(model.positions[place].scroll, link.bottom+1-model.frame.rows), link.top)
}

// followLink is Enter's: on the blog's list the page the link the keys stand on opens, as the page's link leads: a
// grouping's list or a tag's, shown in the list's window, or a tile's post, beside the list.
func (model Model) followLink() (Model, tea.Cmd) {
	if !model.onList() {
		return model, nil
	}
	return model.showing(model.pointedLink(model.chosen).page), nil
}
