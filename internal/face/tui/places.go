package tui

import (
	"iter"
	"maps"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
)

// The tab of a place that is no tab
const noTab = -1

// place is somewhere the command line goes by its name: a tab of the page shown, a place of the tab's page's own, a
// project's page, or the window of keys; the kind the menu over the command line names it by, as Neovim's popup menu
// names a completion's, the window of keys going by its title; the tab it is, where it is one; where it leads; whether
// Tab offers it before the reader types anything; and what going there does.
type place struct {
	name string
	kind string
	tab  int
	// goesTo is where the place leads, which a place of the same name leading there too takes from it
	goesTo destination
	// offeredIdle is whether Tab offers the place before anything is typed: the tabs and the window of keys it does,
	// the page's own places and the projects' pages only once the reader types, so Tab goes round the page shown first
	offeredIdle bool
	// act is what going there does, as a key's action does it
	act action
}

// destination is where a place leads: a page, the tab of its buffer and the section of it, by its anchor; the zero
// destination is the window of keys, which is no page.
type destination struct {
	page    document.Page
	tab     int
	section string
}

// placeKey is a place as the menu tells two apart: by its name and where it leads, so a name two places share stands
// once where both lead to one place, as the post read and the list's own post do, and twice where they lead apart, as
// a tag and a series named alike do, each by its kind.
type placeKey struct {
	name   string
	goesTo destination
}

// key is the place as the menu tells it from another.
func (known place) key() placeKey {
	return placeKey{name: known.name, goesTo: known.goesTo}
}

// shownPlaces are the places the command line goes to from the tabs shown, in the order Tab goes round them, a name
// two share where both lead to one place going to the first, as the page's line has it: the tabs; each tab's page's
// own places, by the tab's place, but those a tab takes; and the site's places, but those named holds for the tab
// chosen, the tabs and its page's own places, worked out as the tabs are shown.
type shownPlaces struct {
	tabs  []place
	own   [][]place
	site  []place
	named []map[placeKey]bool
}

// sitePlaces are the places every page's command line goes to after its tabs and its own: each project's page, as :e
// opens it, and the window of keys, as the question mark opens it.
func sitePlaces(projects []document.Place) []place {
	places := make([]place, 0, len(projects)+1)
	for _, project := range projects {
		slug := project.Name
		places = append(places, place{
			name: slug, kind: string(project.Kind), tab: noTab, goesTo: destination{page: project.Page, tab: 0, section: ""}, offeredIdle: false,
			act: func(model Model) (Model, tea.Cmd) { return model.edit(slug) },
		})
	}
	keys := destination{page: document.PageOf(document.PageNone), tab: 0, section: ""}
	return append(places, place{name: compose.KeysPlace, kind: keysTitle, tab: noTab, goesTo: keys, offeredIdle: true, act: Model.openKeys})
}

// shownPlacesOf are the places of tabs shown, beside the site's, each page's own as pagePlaces holds them.
func shownPlacesOf(tabs []tab, site []place, pagePlaces map[document.Page][]document.Place) shownPlaces {
	tabsTaken := map[placeKey]bool{}
	shown := shownPlaces{tabs: make([]place, 0, len(tabs)), own: make([][]place, 0, len(tabs)), site: site, named: make([]map[placeKey]bool, 0, len(tabs))}
	for index, held := range tabs {
		tabPlace := place{
			name: held.slug, kind: string(document.PlaceTab), tab: index, goesTo: destination{page: held.page, tab: held.place, section: ""}, offeredIdle: true,
			act: func(model Model) (Model, tea.Cmd) { return model.choose(index), nil },
		}
		shown.tabs = append(shown.tabs, firstNamed([]place{tabPlace}, tabsTaken)...)
	}
	for _, held := range tabs {
		named := maps.Clone(tabsTaken)
		shown.own, shown.named = append(shown.own, firstNamed(ownPlaces(pagePlaces[held.page]), named)), append(shown.named, named)
	}
	return shown
}

// firstNamed are the places none before them took, by name and where they lead, which they take in turn.
func firstNamed(places []place, taken map[placeKey]bool) []place {
	kept := make([]place, 0, len(places))
	for _, known := range places {
		if !taken[known.key()] {
			taken[known.key()] = true
			kept = append(kept, known)
		}
	}
	return kept
}

// ownPlaces are a page's own places as the command line goes to them, each by its name.
func ownPlaces(offered []document.Place) []place {
	places := make([]place, 0, len(offered))
	for _, own := range offered {
		places = append(places, place{
			name: own.Name, kind: string(own.Kind), tab: noTab, goesTo: destination{page: own.Page, tab: 0, section: own.Section}, offeredIdle: false,
			act: func(model Model) (Model, tea.Cmd) { return model.wentTo(own), nil },
		})
	}
	return places
}

// from are the places the command line goes to from a tab shown, in the order Tab goes round them, each once.
func (shown shownPlaces) from(chosen int) iter.Seq[place] {
	return func(yield func(place) bool) {
		if handedOn(shown.tabs, kept, yield) && handedOn(shown.own[chosen], kept, yield) {
			handedOn(shown.site, func(known place) bool { return !shown.named[chosen][known.key()] }, yield)
		}
	}
}

// handedOn is whether yield took each of places a test keeps, handed on in turn, and asked for more.
func handedOn(places []place, keeps, yield func(place) bool) bool {
	for _, known := range places {
		if keeps(known) && !yield(known) {
			return false
		}
	}
	return true
}

// kept keeps every place handed on, as the tabs and a page's own are, each name once already (shownPlacesOf).
func kept(place) bool {
	return true
}

// placesFor are the places Tab goes round for what is typed while a tab is chosen, as completions: before anything
// is typed, those offered idle; after, every place whose name starts with what is typed; never the tab chosen, where
// the reader already is.
func (model Model) placesFor(chosen int, typed string) []completion {
	listed := []completion{}
	for known := range model.places.from(chosen) {
		if known.tab == chosen || !strings.HasPrefix(known.name, typed) || typed == "" && !known.offeredIdle {
			continue
		}
		listed = append(listed, completion{written: known.name, name: known.name, kind: known.kind, goes: known.act})
	}
	return listed
}

// placeNamed is the place of a name from the tab chosen, and whether there is one.
func (model Model) placeNamed(name string) (place, bool) {
	for known := range model.places.from(model.chosen) {
		if known.name == name {
			return known, true
		}
	}
	return place{name: "", kind: "", tab: noTab, goesTo: destination{page: document.PageOf(document.PageNone), tab: 0, section: ""}, offeredIdle: false, act: nil}, false
}
