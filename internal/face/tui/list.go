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

// laidTile is a post's tile on the blog's list, as a key goes to it: the post's page; the lines of its top edge and
// its bottom edge; and the columns between its left edge and its right one. The zero tile is none.
type laidTile struct {
	page   document.Page
	top    int
	bottom int
	left   int
	right  int
}

// groundOn is the columns the tile takes on a line, the rows between its edges: none on any other line.
func (tile laidTile) groundOn(line int) columnSpan {
	if line <= tile.top || line >= tile.bottom {
		return columnSpan{start: 0, end: 0}
	}
	return columnSpan{start: tile.left, end: tile.right}
}

// spot is a place on a tab's lines, its line and its column, which a move went to.
type spot struct {
	line   int
	column int
}

// nearer is whether the tile lies nearer a spot than another: fewer lines between them, or as few and fewer columns.
func (tile laidTile) nearer(than laidTile, near spot) bool {
	lines, otherLines := max(tile.top-near.line, near.line-tile.bottom, 0), max(than.top-near.line, near.line-than.bottom, 0)
	if lines != otherLines {
		return lines < otherLines
	}
	return max(tile.left-near.column, near.column-tile.right, 0) < max(than.left-near.column, near.column-than.right, 0)
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
// column at its left where the page is wide enough, over its tiles where not; with the tiles a key goes to and the
// groups whose names stay in sight, and the line each series' group starts at. A block that is no group, which compose
// writes none of, stands as any page lays it out, as the page draws it.
func (layout layout) list(head document.Head, blocks []document.Block) laidTab {
	column := nameColumnOf(layout.width)
	laid := laidTab{lines: layout.blogHead(head, column), tiles: []laidTile{}, groups: []laidGroup{}, split: 0, sections: map[string]int{}, outline: noOutline()}
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
// grouping shown chosen, a filtered list's the one it filters; then its filters, the series it is filtered by and the
// tags of the posts it lists; the filters on lines of their own where both do not fit a line.
func (layout layout) waysLines(head document.Head) []line {
	listWords := layout.siteWords.ListWords
	grouped, _ := head.Page.Unfiltered()
	grouping := []word{wordOf(listWords.By, kindMuted)}
	for _, way := range listWords.Groupings() {
		shown := kindText
		if way.List == grouped {
			shown = kindChosen
		}
		grouping = append(grouping, wordOf(space+way.Words+space, shown))
	}
	filters := make([][]word, 0, len(head.Filters))
	for _, filter := range head.Filters {
		filters = append(filters, filterWords(filter))
	}
	ways, tags := wordRuns(grouping), tagRuns(filters)
	if len(tags) == 0 {
		return []line{ways}
	}
	if runsWidth(ways)+waysGap+runsWidth(tags) <= layout.width {
		return []line{slices.Concat(ways, line{gapRun(waysGap)}, tags)}
	}
	return slices.Concat([]line{ways}, blankLines(), layout.fill(slices.Concat(filters...), 0))
}

// filterWords are a filter the list's head offers as the page writes it (DESIGN.md, BlogHead and Tag): a tag's muted
// word after a faint hash, a series by its words; the filter the list is filtered by chosen whole, a space either side
// and between its words chosen too, so it stands as one chip and breaks only at a space.
func filterWords(filter document.Filter) []word {
	mark := ""
	if filter.MarkedAsTag() {
		mark = tagMark
	}
	if !filter.Current {
		words := textWords(document.PlainText(filter.Words), kindMuted)
		if len(words) > 0 && mark != "" {
			words[0].runs = append([]run{plainRun(mark, kindFaint)}, words[0].runs...)
		}
		return words
	}
	written := strings.Fields(mark + document.PlainText(filter.Words))
	words := make([]word, 0, len(written))
	for index, piece := range written {
		if index == 0 {
			piece = space + piece
		}
		if index == len(written)-1 {
			piece += space
		}
		chosen := wordOf(piece, kindChosen)
		chosen.spaceBefore = kindChosen
		words = append(words, chosen)
	}
	return words
}

// tagWords are tags as the page writes them (DESIGN.md, Tag): each its word after a faint hash, the word of a kind.
func tagWords(tags []string, wordKind kind) []word {
	words := make([]word, 0, len(tags))
	for _, tag := range tags {
		words = append(words, word{runs: []run{plainRun(tagMark, kindFaint), plainRun(tag, wordKind)}, highlightBefore: false, spaceBefore: kindText})
	}
	return words
}

// tagRuns are filters on one line, each two the page's gap apart, the words of each a space apart.
func tagRuns(filters [][]word) line {
	runs := line{}
	for index, filter := range filters {
		if index > 0 {
			runs = append(runs, gapRun(tagGap))
		}
		runs = append(runs, wordRuns(filter)...)
	}
	return runs
}

// plainRun is a text of a kind, as one run with no mark of its own.
func plainRun(text string, shown kind) run {
	return run{text: text, kind: shown, highlighted: false, address: "", copied: "", tone: 0, cites: noCitation}
}

// groupOver lays out a group on a page with no names' column: its name in the gradient and how many posts it holds on
// one line over its tiles, a series' line under them.
func (layout layout) groupOver(laid laidTab, group document.Group) laidTab {
	name := append(append(line{gapRun(0)}, wordRuns(gradientWords(document.PlainText(group.Title), kindHeading))...), gapRun(waysGap))
	heading := slices.Concat(layout.continued(name, layout.words(group.Posts.Count, kindMuted), 0), layout.fill(layout.words(group.Posts.Line, kindMuted), 0))
	laid.lines = slices.Concat(laid.lines, heading, blankLines())
	tiles := layout.tiles(group.Posts.List)
	return laid.joined(tiles.lines, tiles.tiles, 0)
}

// groupBeside lays out a group beside its name, in the names' column of a width: the name in the titles' font, a
// series' line and how many posts it holds under it, standing in sight while the group's tiles scroll.
func (layout layout) groupBeside(laid laidTab, group document.Group, column int) laidTab {
	named, tiled := layout, layout
	named.width, tiled.width = column, layout.width-column-nameGap
	about := slices.Concat(named.fill(named.words(group.Posts.Line, kindMuted), 0), named.fill(named.words(group.Posts.Count, kindMuted), 0))
	name := stacked(named.titleRows(document.PlainText(group.Title)), about)
	first, tiles := len(laid.lines), tiled.tiles(group.Posts.List)
	laid = laid.joined(besideLines(name, column+nameGap, tiles.lines), tiles.tiles, column+nameGap)
	laid.groups = append(laid.groups, laidGroup{first: first, end: len(laid.lines), name: name})
	return laid
}

// tiles lays out posts as a grid of tiles: each row of tiles as tall as its tallest, the tiles sharing their edges, as
// tmux's panes share theirs. A post out is a tile a key goes to; a part to come is none.
func (layout layout) tiles(posts []document.ListedPost) laidTab {
	grid := gridOf(layout.width)
	inside := layout
	inside.width = grid.inside()
	rows := [][]gridCell{}
	for first := 0; first < len(posts); first += grid.columns {
		row := []gridCell{}
		for _, post := range posts[first:min(first+grid.columns, len(posts))] {
			_, isOut := post.Readiness.(document.PostOut)
			row = append(row, gridCell{label: layout.edgeLabel(post), words: inside.tileWordsOf(post), leads: isOut})
		}
		rows = append(rows, row)
	}
	framed := grid.framed(rows)
	laid := laidTab{lines: framed.lines, tiles: []laidTile{}, groups: nil, split: 0, sections: nil, outline: noOutline()}
	for index, post := range posts {
		if out, isOut := post.Readiness.(document.PostOut); isOut {
			row, left := index/grid.columns, index%grid.columns*grid.width+textWidth(frameDown)
			laid.tiles = append(laid.tiles, laidTile{page: out.Page, top: framed.edges[row], bottom: framed.edges[row+1], left: left, right: left + grid.width - textWidth(frameDown)})
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

// framed lays out rows of cells as a grid, each row as tall as its tallest cell, the cells sharing their edges.
func (grid tileGrid) framed(rows [][]gridCell) framedGrid {
	laid, above := framedGrid{lines: []line{}, edges: []int{}}, []bool{}
	for _, row := range rows {
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

// joined is a laid list with lines after its own, among them tiles laid out from the lines' top, standing a number of
// columns in from the page's left edge.
func (laid laidTab) joined(lines []line, tiles []laidTile, left int) laidTab {
	first := len(laid.lines)
	for _, tile := range tiles {
		tile.top, tile.bottom, tile.left, tile.right = tile.top+first, tile.bottom+first, tile.left+left, tile.right+left
		laid.tiles = append(laid.tiles, tile)
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
// and its tags, faint.
func (layout layout) tileWordsOf(post document.ListedPost) tileWords {
	top := []line{}
	switch post.Readiness.(type) {
	case document.PostOut:
		top = slices.Concat(titleLines(layout.fill(textWords(document.PlainText(post.Title), kindTileTitle), 0)), layout.fill(layout.words(post.Lead, kindMuted), 0))
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

// titleLines are a tile's title's lines, each one run after its indent, so the tile the keys stand on underlines
// the spaces between its words too, as the page underlines a link whole.
func titleLines(lines []line) []line {
	whole := make([]line, 0, len(lines))
	for _, shown := range lines {
		var text strings.Builder
		for _, piece := range shown[1:] {
			text.WriteString(piece.text)
		}
		whole = append(whole, line{shown[0], plainRun(text.String(), kindTileTitle)})
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
// somewhere, each label of those below set into its top edge; each piece of it framed as the tiles it bounds have it.
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
		case point < len(below):
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

// pointedTile is the tile the keys stand on in a tab, the zero tile where the tab holds no tiles.
func (model Model) pointedTile(place int) laidTile {
	tiles := model.laid[place].tiles
	if len(tiles) == 0 {
		return laidTile{page: document.PageOf(document.PageNone), top: 0, bottom: 0, left: 0, right: 0}
	}
	return tiles[model.positions[place].tile]
}

// onList is whether the tab chosen is the blog's list, whose keys go from tile to tile.
func (model Model) onList() bool {
	return len(model.laid[model.chosen].tiles) > 0
}

// downward is j's and the down arrow's: on the blog's list a tile on, else a line down.
func (model Model) downward() (Model, tea.Cmd) {
	return model.movedBy(forwardStep), nil
}

// upward is k's and the up arrow's: on the blog's list a tile back, else a line up.
func (model Model) upward() (Model, tea.Cmd) {
	return model.movedBy(backwardStep), nil
}

// movedBy is the model a step on or back: on the blog's list a tile, else a line.
func (model Model) movedBy(step int) Model {
	if model.onList() {
		return model.toTile(model.positions[model.chosen].tile + step)
	}
	return model.scrolledTo(model.positions[model.chosen].scroll + step)
}

// toTile has the keys stand on a tile of the chosen list, in reading order, and the page scrolled to show it whole;
// past the first tile or the last the keys stay and the page scrolls by the line the step went past, so the list's
// head and foot come in sight.
func (model Model) toTile(wanted int) Model {
	tiles, scroll := model.laid[model.chosen].tiles, model.positions[model.chosen].scroll
	chosen := min(max(wanted, 0), len(tiles)-1)
	if chosen == model.positions[model.chosen].tile {
		return model.scrolledTo(scroll + wanted - chosen)
	}
	model = model.choseTile(chosen)
	return model.scrolledTo(model.scrollShowing(model.chosen, tiles[chosen]))
}

// choseTile has the keys stand on a tile of the chosen list, the tile left and the tile chosen painted again and no
// other line.
func (model Model) choseTile(chosen int) Model {
	tiles := model.laid[model.chosen].tiles
	left, tile := tiles[model.positions[model.chosen].tile], tiles[chosen]
	marks := model.drawn.tabs[model.chosen].marks
	marks.pointed = tile
	model.positions = slices.Clone(model.positions)
	model.positions[model.chosen].tile = chosen
	return model.repaintedLines(slices.Concat(linesBetween(left.top, left.bottom), linesBetween(tile.top, tile.bottom)), marks)
}

// withChosenTileNear is the model where, on the blog's list chosen, the tile the keys stand on has gone out of sight:
// the keys stand on the tile nearest a spot the page went to, among those in sight where any is.
func (model Model) withChosenTileNear(near spot) Model {
	tiles, chosen := model.laid[model.chosen].tiles, model.chosen
	if len(tiles) == 0 || model.inSight(chosen, tiles[model.positions[chosen].tile]) {
		return model
	}
	nearest := 0
	for index, tile := range tiles {
		best := tiles[nearest]
		if model.inSight(chosen, tile) != model.inSight(chosen, best) {
			if model.inSight(chosen, tile) {
				nearest = index
			}
			continue
		}
		if tile.nearer(best, near) {
			nearest = index
		}
	}
	return model.choseTile(nearest)
}

// withFirstTileFrom is the model, on the blog's list chosen, with the keys on the first tile at or after a line, where
// any is.
func (model Model) withFirstTileFrom(line int) Model {
	first := slices.IndexFunc(model.laid[model.chosen].tiles, func(tile laidTile) bool { return tile.top >= line })
	if first < 0 || first == model.positions[model.chosen].tile {
		return model
	}
	return model.choseTile(first)
}

// inSight is whether a line of a tile shows on a tab's page at its scroll.
func (model Model) inSight(place int, tile laidTile) bool {
	scroll := model.positions[place].scroll
	return tile.bottom >= scroll && tile.top < scroll+model.frame.rows
}

// withTilesInSight is the model with each list among the tabs whose tile the keys stand on is out of sight scrolled
// as little as shows it whole; a list whose tile shows stays where it stands, at its top as it opens.
func (model Model) withTilesInSight() Model {
	model.positions = slices.Clone(model.positions)
	for place := range model.tabs {
		if tiles := model.laid[place].tiles; len(tiles) > 0 && !model.inSight(place, tiles[model.positions[place].tile]) {
			model.positions[place].scroll = model.scrollShowing(place, tiles[model.positions[place].tile])
		}
	}
	return model.clamped()
}

// withMatchTileTop is the model, on the blog's list, with the page scrolled up where it shows a line of the tile the
// keys stand on but not its top edge, so the tile shows its date, as j shows it.
func (model Model) withMatchTileTop(line int) Model {
	tile := model.pointedTile(model.chosen)
	if !model.onList() || line < tile.top || line > tile.bottom || tile.top >= model.positions[model.chosen].scroll {
		return model
	}
	return model.scrolledTo(tile.top)
}

// linesBetween are the lines from a first to a last, both among them.
func linesBetween(first, last int) []int {
	lines := make([]int, 0, last-first+1)
	for line := first; line <= last; line++ {
		lines = append(lines, line)
	}
	return lines
}

// scrollShowing is the scroll of a tab nearest its own that shows a tile whole, its edges with it, or its top where the
// page is too short for the whole.
func (model Model) scrollShowing(place int, tile laidTile) int {
	return min(max(model.positions[place].scroll, tile.bottom+1-model.frame.rows), tile.top)
}

// openTile is Enter's: on the blog's list the post of the tile the keys stand on, beside the list.
func (model Model) openTile() (Model, tea.Cmd) {
	if !model.onList() {
		return model, nil
	}
	return model.showing(model.pointedTile(model.chosen).page), nil
}
