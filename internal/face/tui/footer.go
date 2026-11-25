package tui

import (
	"slices"
	"strconv"
	"strings"

	"github.com/vrppaul/paranin/internal/compose"
)

// The marks of the footer, as tmux's status line writes them: what holds the session's name; what parts a window's
// number from its name; what marks the window shown; and what stands for the windows that give way
const (
	sessionOpening = "["
	sessionClosing = "]"
	numberMark     = ":"
	shownTabMark   = "*"
	givenWayMark   = "…"
)

// The room in the footer: a column of the panel on each side of a tab's number and name, as tmux pads a window's, and
// one before each tab; the least room left before what stands at the right; the room between the key's word and the
// ruler; and the columns the footer keeps clear at each side where it does not stand on the page's column
const (
	tabPadding = 1
	tabGap     = 1
	rightGap   = 1
	keysGap    = 4
	edgeMargin = 1
)

// footerPlace is where the footer stands on a screen: the columns it keeps clear at each side, the page's margin
// where it fits on the page's column and the screen's edge where not; whether the key to the window of keys stands
// in it; and the columns left for the session and the tabs.
type footerPlace struct {
	margin    int
	keysShown bool
	room      int
}

// paintedFooter is what the footer holds on any tab, painted: where it stands, its margin of the panel's ground, and
// the key to the window of keys, or nothing where it gives way.
type paintedFooter struct {
	place  footerPlace
	gutter string
	keys   footerBar
}

// footerTarget is where a click on a part of the footer goes, as the key that part names goes: a window, a session not
// shown, the window of keys, or nowhere.
//
//sumtype:decl
type footerTarget interface {
	footerTarget()
}

// windowTarget is a window of the session shown, by its place among the tabs, which a click chooses as its digit does.
type windowTarget struct {
	place int
}

// sessionTarget is a session not shown, which a click shows as its reader left it.
type sessionTarget struct {
	session compose.Session
}

// keysTarget is the window of keys, which a click opens as ? does.
type keysTarget struct{}

// noTarget is a part of the footer a click does nothing on: a gap, the mark of windows given way, the session shown.
type noTarget struct{}

func (windowTarget) footerTarget()  {}
func (sessionTarget) footerTarget() {}
func (keysTarget) footerTarget()    {}
func (noTarget) footerTarget()      {}

// footerItem is pieces of the footer that a click on goes one way.
type footerItem struct {
	pieces []piece
	leads  footerTarget
}

// footerSpot is the screen's columns an item of the footer stands in, and where a click on them goes.
type footerSpot struct {
	columns columnSpan
	leads   footerTarget
}

// footerBar is items of the footer painted, and the spot each stands in.
type footerBar struct {
	painted paintedBar
	spots   []footerSpot
}

// footerItems paints items of the footer cut at a width, standing a number of columns in from the screen's left edge,
// and says where each stands, in the one walk that paints them, so where a click finds an item is where it shows.
func (paint paint) footerItems(items []footerItem, width, left int) footerBar {
	var builder strings.Builder
	used, spots := 0, make([]footerSpot, 0, len(items))
	for _, item := range items {
		painted := paint.cut(item.pieces, width-used)
		builder.WriteString(painted.text)
		spots = append(spots, footerSpot{columns: columnSpan{start: left + used, end: left + used + painted.width}, leads: item.leads})
		used += painted.width
	}
	return footerBar{painted: paintedBar{text: builder.String(), width: used}, spots: spots}
}

// itemsWidth is the columns items of the footer take, uncut.
func itemsWidth(items []footerItem) int {
	width := 0
	for _, item := range items {
		width += piecesWidth(item.pieces)
	}
	return width
}

// footerPlaceOf is where a footer of a width, the key to the window of keys aside, stands on a frame: on the page's
// column with the key, then without it, then at the screen's edges likewise; where none fits, at the edges without
// the key, the tabs giving way.
func footerPlaceOf(frame frame, width, keysWidth int) footerPlace {
	for _, margin := range []int{max(frame.left, edgeMargin), edgeMargin} {
		for _, keysShown := range []bool{true, false} {
			if place := placeAt(frame, margin, keysShown, keysWidth); width <= place.room {
				return place
			}
		}
	}
	return placeAt(frame, edgeMargin, false, keysWidth)
}

// placeAt is the footer standing a margin in from each side of a frame, the key to the window of keys in it or not:
// the margin no more than half the screen, and the room the ruler and the key leave.
func placeAt(frame frame, margin int, keysShown bool, keysWidth int) footerPlace {
	margin = min(margin, frame.width/sides)
	right := widestRuler + rightGap
	if keysShown {
		right += keysWidth
	}
	return footerPlace{margin: margin, keysShown: keysShown, room: max(frame.width-sides*margin-right, 0)}
}

// paintedFooter paints what the footer holds on any tab, where the session and the tabs leave it room.
func (model Model) paintedFooter() paintedFooter {
	keys := []footerItem{
		{pieces: []piece{{text: keyWindowKey, part: partPanelKey}, {text: space + keysTitle, part: partPanelNote}}, leads: keysTarget{}},
		{pieces: []piece{{text: strings.Repeat(space, keysGap), part: partPanelNote}}, leads: noTarget{}},
	}
	whole := slices.Concat(model.sessionItems(), slices.Concat(model.tabItems(anyTab)...))
	keysWidth := itemsWidth(keys)
	place := footerPlaceOf(model.frame, itemsWidth(whole), keysWidth)
	shownKeys := footerBar{painted: paintedBar{text: "", width: 0}, spots: nil}
	if place.keysShown {
		// the key stands just before the ruler's cell (footerRow)
		shownKeys = model.paint.footerItems(keys, keysWidth, model.frame.width-place.margin-widestRuler-keysWidth)
	}
	return paintedFooter{place: place, gutter: model.paint.parts[partPanel].Render(strings.Repeat(space, place.margin)), keys: shownKeys}
}

// A tab whose items the footer's width is worked out from: the tabs take as many columns whichever is shown
const anyTab = 0

// sessionItems are the sessions the footer opens with, as tmux names its sessions (compose.Frame.Sessions), in the
// language the blog's session reads, a space between each two, the one shown standing out and leading nowhere.
func (model Model) sessionItems() []footerItem {
	shown := model.tabs[model.chosen].page
	items := []footerItem{}
	for _, session := range model.sessionWords().Sessions {
		if len(items) > 0 {
			items = append(items, footerItem{pieces: []piece{{text: space, part: partPanel}}, leads: noTarget{}})
		}
		part, leads := partSession, footerTarget(sessionTarget{session: session})
		if session.Holds(shown) {
			part, leads = partHost, noTarget{}
		}
		items = append(items, footerItem{pieces: []piece{{text: sessionOpening + session.Name + sessionClosing, part: part}}, leads: leads})
	}
	return items
}

// tabItems are each tab as the footer shows it while one is shown, as tmux shows each window: a gap, then its number
// and its name, the one shown marked, on a plate that leads to the window.
func (model Model) tabItems(shown int) [][]footerItem {
	gap := footerItem{pieces: []piece{{text: strings.Repeat(space, tabGap), part: partPanel}}, leads: noTarget{}}
	padding := strings.Repeat(space, tabPadding)
	items := make([][]footerItem, 0, len(model.tabs))
	for place, named := range model.tabs {
		number := strconv.Itoa(place+firstTabNumber) + numberMark
		plate := []piece{{text: padding + number, part: partTabNumber}, {text: named.slug + padding, part: partTabName}}
		if place == shown {
			plate = []piece{{text: padding + number + named.slug + shownTabMark + padding, part: partTabShown}}
		}
		items = append(items, []footerItem{gap, {pieces: plate, leads: windowTarget{place: place}}})
	}
	return items
}

// footerTabs paints the footer's left part while a tab is shown: the session, then the tabs that fit the room a place
// of the footer has; where the session leaves the tab shown no room, the tabs alone, their marks kept over the
// session, since they say more tabs lie beside the one shown.
func (model Model) footerTabs(shown int, place footerPlace) footerBar {
	session, items := model.sessionItems(), model.tabItems(shown)
	opening := slices.Concat(session, fittedTabs(items, shown, place.room-itemsWidth(session)))
	if itemsWidth(opening) > place.room {
		opening = fittedTabs(items, shown, place.room)
	}
	return model.paint.footerItems(opening, place.room, place.margin)
}

// fittedTabs are the items of the tabs that fit a room, the one shown always among them: where all do not, those
// farthest from it give way, the later first where two stand as far, a mark standing for those given way on each side.
func fittedTabs(items [][]footerItem, shown, room int) []footerItem {
	first, last := 0, len(items)-1
	for first < last && itemsWidth(tabsBetween(items, first, last)) > room {
		if last-shown >= shown-first {
			last--
		} else {
			first++
		}
	}
	return tabsBetween(items, first, last)
}

// tabsBetween are the items of the tabs from a first to a last, a mark on each side where tabs there give way.
func tabsBetween(items [][]footerItem, first, last int) []footerItem {
	mark := footerItem{pieces: []piece{{text: strings.Repeat(space, tabGap), part: partPanel}, {text: givenWayMark, part: partTabNumber}}, leads: noTarget{}}
	row := []footerItem{}
	if first > 0 {
		row = append(row, mark)
	}
	row = append(row, slices.Concat(items[first:last+1]...)...)
	if last < len(items)-1 {
		row = append(row, mark)
	}
	return row
}

// footerRow is the footer as the chosen tab and its scroll leave it: its session and tabs, painted already, the key
// to the window of keys, and the ruler at the right.
func (model Model) footerRow(drawn drawnTab) string {
	footer := model.drawn.footer
	inside := model.frame.width - sides*footer.place.margin
	keys, panel := footer.keys.painted, model.paint.parts[partPanel]
	// the ruler takes a cell as wide as its widest word, its word at the right, so the key before it stands where
	// paintedFooter placed it however far the page has scrolled
	rulerCell := max(min(widestRuler, inside-drawn.footerTabs.width-keys.width), 0)
	ruler := model.paint.cut([]piece{{text: model.position(), part: partRuler}}, rulerCell)
	fill := strings.Repeat(space, max(inside-drawn.footerTabs.width-keys.width-rulerCell, 0))
	rulerPadding := strings.Repeat(space, rulerCell-ruler.width)
	return footer.gutter + drawn.footerTabs.text + panel.Render(fill) + keys.text + panel.Render(rulerPadding) + ruler.text + footer.gutter
}
