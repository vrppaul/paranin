package tui

import (
	"slices"
	"strconv"
	"strings"
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
	keys   paintedBar
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
	keys := []piece{
		{text: keyWindowKey, part: partPanelKey},
		{text: space + keysTitle + strings.Repeat(space, keysGap), part: partPanelNote},
	}
	whole := slices.Concat(model.sessionPieces(), slices.Concat(model.tabItems(anyTab)...))
	place := footerPlaceOf(model.frame, piecesWidth(whole), piecesWidth(keys))
	shownKeys := paintedBar{text: "", width: 0}
	if place.keysShown {
		shownKeys = model.paint.cut(keys, piecesWidth(keys))
	}
	return paintedFooter{place: place, gutter: model.paint.parts[partPanel].Render(strings.Repeat(space, place.margin)), keys: shownKeys}
}

// A tab whose items the footer's width is worked out from: the tabs take as many columns whichever is shown
const anyTab = 0

// sessionPieces are the sessions the footer opens with, as tmux names its sessions (compose.Frame.Sessions), in the
// language the blog's session reads, a space between each two, the one shown standing out.
func (model Model) sessionPieces() []piece {
	shown := model.tabs[model.chosen].page
	pieces := []piece{}
	for _, session := range model.sessionWords().Sessions {
		if len(pieces) > 0 {
			pieces = append(pieces, piece{text: space, part: partPanel})
		}
		named := piece{text: sessionOpening + session.Name + sessionClosing, part: partSession}
		if session.Holds(shown) {
			named.part = partHost
		}
		pieces = append(pieces, named)
	}
	return pieces
}

// tabItems are each tab as the footer shows it while one is shown, as tmux shows each window: a gap, then its number
// and its name, the one shown marked, on a plate.
func (model Model) tabItems(shown int) [][]piece {
	gap := piece{text: strings.Repeat(space, tabGap), part: partPanel}
	padding := strings.Repeat(space, tabPadding)
	items := make([][]piece, 0, len(model.tabs))
	for place, named := range model.tabs {
		number := strconv.Itoa(place+firstTabNumber) + numberMark
		if place == shown {
			items = append(items, []piece{gap, {text: padding + number + named.slug + shownTabMark + padding, part: partTabShown}})
			continue
		}
		items = append(items, []piece{gap, {text: padding + number, part: partTabNumber}, {text: named.slug + padding, part: partTabName}})
	}
	return items
}

// footerTabs paints the footer's left part while a tab is shown: the session, then the tabs that fit the room a place
// of the footer has; where the session leaves the tab shown no room, the tabs alone, their marks kept over the
// session, since they say more tabs lie beside the one shown.
func (model Model) footerTabs(shown int, place footerPlace) paintedBar {
	session, items := model.sessionPieces(), model.tabItems(shown)
	opening := slices.Concat(session, fittedTabs(items, shown, place.room-piecesWidth(session)))
	if piecesWidth(opening) > place.room {
		opening = fittedTabs(items, shown, place.room)
	}
	return model.paint.cut(opening, place.room)
}

// fittedTabs are the items of the tabs that fit a room, the one shown always among them: where all do not, those
// farthest from it give way, the later first where two stand as far, a mark standing for those given way on each side.
func fittedTabs(items [][]piece, shown, room int) []piece {
	first, last := 0, len(items)-1
	for first < last && piecesWidth(tabsBetween(items, first, last)) > room {
		if last-shown >= shown-first {
			last--
		} else {
			first++
		}
	}
	return tabsBetween(items, first, last)
}

// tabsBetween are the items of the tabs from a first to a last, a mark on each side where tabs there give way.
func tabsBetween(items [][]piece, first, last int) []piece {
	mark := []piece{{text: strings.Repeat(space, tabGap), part: partPanel}, {text: givenWayMark, part: partTabNumber}}
	row := []piece{}
	if first > 0 {
		row = append(row, mark...)
	}
	row = append(row, slices.Concat(items[first:last+1]...)...)
	if last < len(items)-1 {
		row = append(row, mark...)
	}
	return row
}

// footerRow is the footer as the chosen tab and its scroll leave it: its session and tabs, painted already, the key
// to the window of keys, and the ruler at the right.
func (model Model) footerRow(drawn drawnTab) string {
	footer := model.drawn.footer
	inside := model.frame.width - sides*footer.place.margin
	ruler := model.paint.cut([]piece{{text: model.position(), part: partRuler}}, inside-drawn.footerTabs.width-footer.keys.width)
	fill := strings.Repeat(space, max(inside-drawn.footerTabs.width-footer.keys.width-ruler.width, 0))
	return footer.gutter + drawn.footerTabs.text + model.paint.parts[partPanel].Render(fill) + footer.keys.text + ruler.text + footer.gutter
}
