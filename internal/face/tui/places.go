package tui

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// The name of the place that opens the window of keys, as Neovim's :help opens its own
const helpPlace = "help"

// The tab of a place that is no tab
const noTab = -1

// The kinds the menu over the command line names the places by, as Neovim's popup menu names a completion's: a tab of
// the page shown, and a project's page; the window of keys goes by its title
const (
	tabKind     = "tab"
	projectKind = "project"
)

// place is somewhere the command line goes by its name: a tab of the page shown, a project's page, or the window of
// keys; its kind; the tab it is, where it is one; whether Tab offers it before the reader types anything; and what
// going there does.
type place struct {
	name string
	kind string
	tab  int
	// offeredIdle is whether Tab offers the place before anything is typed: the tabs and the window of keys it does,
	// the projects' pages only once the reader types, so Tab goes round the page shown first
	offeredIdle bool
	// act is what going there does, as a key's action does it, with no count
	act action
}

// placesOf are the places the command line goes to from a page: each of its tabs, each project's page, as :e opens
// it, and the window of keys, as the question mark opens it, in that order; a name taken already by one before is
// left to it, so each name goes to one place.
func placesOf(tabs []tab, slugs []string) []place {
	places := make([]place, 0, len(tabs)+len(slugs)+1)
	for index, shown := range tabs {
		places = append(places, place{name: shown.slug, kind: tabKind, tab: index, offeredIdle: true, act: func(model Model, _ int) (Model, tea.Cmd) { return model.choose(index), nil }})
	}
	for _, slug := range slugs {
		places = append(places, place{name: slug, kind: projectKind, tab: noTab, offeredIdle: false, act: func(model Model, _ int) (Model, tea.Cmd) { return model.edit(slug) }})
	}
	places = append(places, place{name: helpPlace, kind: keysTitle, tab: noTab, offeredIdle: true, act: Model.openKeys})
	unique := make([]place, 0, len(places))
	for _, known := range places {
		if !slices.ContainsFunc(unique, func(taken place) bool { return taken.name == known.name }) {
			unique = append(unique, known)
		}
	}
	return unique
}

// placesFor are the places Tab goes round for what is typed while a tab is chosen, as completions: before anything
// is typed, those offered idle; after, every place whose name starts with what is typed; never the tab chosen, where
// the reader already is.
func (model Model) placesFor(chosen int, typed string) []completion {
	listed := []completion{}
	for _, known := range model.places {
		if known.tab == chosen || !strings.HasPrefix(known.name, typed) || typed == "" && !known.offeredIdle {
			continue
		}
		listed = append(listed, completion{written: known.name, name: known.name, kind: known.kind})
	}
	return listed
}

// placeNamed is the place of a name, and whether there is one.
func (model Model) placeNamed(name string) (place, bool) {
	for _, known := range model.places {
		if known.name == name {
			return known, true
		}
	}
	return place{name: "", kind: "", tab: noTab, offeredIdle: false, act: nil}, false
}
