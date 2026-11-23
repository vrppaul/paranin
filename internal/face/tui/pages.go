package tui

import (
	"maps"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// What the command line says where :e finds no page by the name typed, as Neovim says it of a file
const (
	pageNotFoundOpening = `E447: Can't find file "`
	pageNotFoundClosing = `" in path`
)

// The place among the tabs shown of a tab none of them is
const noPlace = -1

// readerPlace is where a reader stood in a buffer: the tab chosen, and how far down each tab they scrolled and the
// tile the keys stood on in each, on the blog's list, by the tab's place.
type readerPlace struct {
	chosen  int
	scrolls []int
	tiles   []int
}

// blogSession is what the blog's session shows, as tmux holds a session's windows: the list its first window shows,
// by year or by series; the post read in its second, PageNone where none is; and whether the reader stands at the
// post's window.
type blogSession struct {
	list   document.Page
	post   document.Page
	atPost bool
}

// buffersOf are the tabs of each document by its page: the buffers a session holds, as Neovim holds the files it
// opened. They never change; where the reader stands in each is the model's.
func buffersOf(documents []document.Document) map[document.Page][]tab {
	buffers := make(map[document.Page][]tab, len(documents))
	for _, held := range documents {
		buffers[held.Head.Page] = tabsOf(held)
	}
	return buffers
}

// pagePlacesOf are each document's own places by its page, read off it once, as it never changes.
func pagePlacesOf(documents []document.Document) map[document.Page][]document.Place {
	places := make(map[document.Page][]document.Place, len(documents))
	for _, held := range documents {
		places[held.Head.Page] = document.PlacesOf(held)
	}
	return places
}

// placesLeft are where a reader stands in each buffer before they first read it: on its first tab, scrolled nowhere,
// on its first tile.
func placesLeft(buffers map[document.Page][]tab) map[document.Page]readerPlace {
	left := make(map[document.Page]readerPlace, len(buffers))
	for page, tabs := range buffers {
		left[page] = readerPlace{chosen: 0, scrolls: make([]int, len(tabs)), tiles: make([]int, len(tabs))}
	}
	return left
}

// showing is the model showing a page in its session: a post in the blog's second window, after its list; a list in
// the blog's first window, the post read kept beside it; any other page in the site's session, in place of the page
// it showed.
func (model Model) showing(page document.Page) Model {
	model = model.kept()
	switch {
	case page.Kind == document.PagePost:
		model.blog.post, model.blog.atPost = page, true
	case page.InBlog():
		model.blog.list, model.blog.atPost = page, false
	default:
		model.sitePage = page
	}
	model.inBlog = page.InBlog()
	return model.windowed()
}

// kept is the model with where the reader stands in the tabs shown kept for their pages, and, in the blog's session,
// whether they stand at the post's window.
func (model Model) kept() Model {
	left := maps.Clone(model.left)
	for index, shown := range model.tabs {
		place := left[shown.page]
		place.scrolls, place.tiles = slices.Clone(place.scrolls), slices.Clone(place.tiles)
		place.scrolls[shown.place], place.tiles[shown.place] = model.scrolls[index], model.tiles[index]
		if index == model.chosen {
			place.chosen = shown.place
		}
		left[shown.page] = place
	}
	model.left = left
	if model.inBlog {
		model.blog.atPost = model.tabs[model.chosen].page.Kind == document.PagePost
	}
	return model
}

// windows are the tabs of the session shown, as tmux's windows, and the place of the one the reader stands at: the
// blog's list and the post read, each named as compose names the blog's windows; or the tabs of the site's page.
func (model Model) windows() ([]tab, int) {
	if !model.inBlog {
		return model.buffers[model.sitePage], model.left[model.sitePage].chosen
	}
	tabs, chosen := []tab{}, 0
	for _, window := range model.siteWords.ListWords.BlogWindows(model.blog.list, model.blog.post) {
		held, isHeld := model.buffers[window.Page]
		if !isHeld {
			continue
		}
		if window.Page.Kind == document.PagePost && model.blog.atPost {
			chosen = len(tabs)
		}
		named := held[0]
		named.slug = window.Name
		tabs = append(tabs, named)
	}
	return tabs, chosen
}

// windowed is the model showing the windows of the session shown, where the reader stood in each coming back. A tab
// shown already keeps its layout and its paint, so only a tab newly shown is laid out and painted; the frame round
// them is painted again, and all of it is laid out again where the session shown takes a page of another width.
func (model Model) windowed() Model {
	tabs, chosen := model.windows()
	before := model
	model.tabs, model.chosen = tabs, chosen
	model.scrolls, model.tiles = make([]int, 0, len(tabs)), make([]int, 0, len(tabs))
	for _, shown := range tabs {
		place := model.left[shown.page]
		model.scrolls, model.tiles = append(model.scrolls, place.scrolls[shown.place]), append(model.tiles, place.tiles[shown.place])
	}
	model.places = shownPlacesOf(tabs, model.sitePlaces, model.pagePlaces)
	if frameOf(model.terminal.width, model.terminal.height, widestPageOf(model.inBlog)) != model.frame {
		return model.relaid()
	}
	model.laid, model.search.shown = make([]laidTab, 0, len(tabs)), make([]shownMatch, 0, len(tabs))
	reused, layout := map[int]paintedPage{}, model.layout()
	for place, shown := range tabs {
		if old := before.placeOf(shown); old != noPlace {
			model.laid, model.search.shown = append(model.laid, before.laid[old]), append(model.search.shown, before.search.shown[old])
			reused[place] = before.drawn.tabs[old].paintedPage
			continue
		}
		model.laid, model.search.shown = append(model.laid, layout.tab(shown)), append(model.search.shown, shownMatch{place: 0, scroll: 0, shown: false})
	}
	pages := make([]paintedPage, 0, len(tabs))
	for place := range tabs {
		painted, isReused := reused[place]
		if !isReused {
			painted = model.paintedPage(place)
		}
		pages = append(pages, painted)
	}
	model.drawn = model.drawnScreen(pages)
	return model.clamped().windowOfKeysSettled()
}

// placeOf is the place among the tabs shown of a tab of a page, noPlace where none of them is.
func (model Model) placeOf(wanted tab) int {
	return slices.IndexFunc(model.tabs, func(shown tab) bool { return shown.page == wanted.page && shown.place == wanted.place })
}

// reached is the model opened on the page a name from outside reaches (New), its letters a terminal does not show as
// they are left out: a page of the blog; where the name is the blog's but names no page, the blog's list where the
// session holds it, the résumé where not, saying it finds no page by the name; the résumé where it names nothing of
// the blog.
func (model Model) reached(given string) Model {
	name := strings.Map(shownLetter, given)
	if page, named := model.opened[name]; named && page.InBlog() && model.holds(page) {
		return model.showing(page)
	}
	opened, blogs := document.PageOf(document.PageResume), routes.IsBlogName(name)
	if list := document.PageOf(document.PageBlog); blogs && model.holds(list) {
		opened = list
	}
	shown := model.showing(opened)
	if blogs {
		shown.said = notFound(name)
	}
	return shown
}

// holds is whether the session holds a page.
func (model Model) holds(page document.Page) bool {
	_, held := model.buffers[page]
	return held
}

// edit shows a page by its name, as Neovim's :e opens a file: by any name compose says :e reads, a project's slug or
// the name a terminal reaches a page of the blog by, where the session holds the page; or says it finds no such page.
func (model Model) edit(name string) (Model, tea.Cmd) {
	page, named := model.opened[name]
	if !named || !model.holds(page) {
		model.said = notFound(name)
		return model, nil
	}
	return model.showing(page), nil
}

// openedPages are the pages :e opens, by every name it reads.
func openedPages(pages []compose.OpenedPage) map[string]document.Page {
	opened := make(map[string]document.Page, len(pages))
	for _, page := range pages {
		opened[page.Place.Name] = page.Place.Page
	}
	return opened
}

// wentTo is the model at a place of a page's own: its page shown, where the tab chosen shows another, then scrolled to
// the place's section, where it names one the page lays out; on the blog's list the keys on the section's first tile.
func (model Model) wentTo(place document.Place) Model {
	if model.tabs[model.chosen].page != place.Page {
		model = model.showing(place.Page)
	}
	line, laid := model.laid[model.chosen].sections[place.Section]
	if !laid {
		return model
	}
	return model.scrolledTo(line).withFirstTileFrom(line)
}

// notFound is what the command line says of a name no page goes by, as Neovim says it of a file.
func notFound(name string) commandMessage {
	return commandMessage{text: pageNotFoundOpening + name + pageNotFoundClosing, failed: true}
}

// closed is the model after q, as Neovim's :q closes a window: on a post's tab its window closes and the blog's list
// shows, where the session holds it; on any other page but the résumé, the blog's list among them, and on a post with
// no list beside it, the résumé shows, where its reader left it; on the résumé the session ends.
func (model Model) closed() (Model, tea.Cmd) {
	resume := document.PageOf(document.PageResume)
	switch shown := model.tabs[model.chosen].page; {
	case shown.Kind == document.PagePost && model.holds(model.blog.list):
		model = model.kept()
		model.blog.post, model.blog.atPost = document.PageOf(document.PageNone), false
		return model.windowed(), nil
	case shown != resume:
		return model.showing(resume), nil
	}
	return model, tea.Quit
}

// closingHint is what the frame says q does on a tab of a page: it closes any page but the résumé, and quits on the
// résumé.
func closingHint(page document.Page) string {
	if page == document.PageOf(document.PageResume) {
		return quitHint
	}
	return closeHint
}
