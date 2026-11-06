package tui

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// The name of the place that opens the window of keys, as Neovim's :help opens its own, and what joins the words of
// a tab's title in its place's name, as a path joins them
const (
	helpPlace      = "help"
	placeWordJoint = "-"
)

// The tab of a place that is no tab
const noTab = -1

// place is somewhere the prompt goes by its name: a tab of the page shown, a project's page, or the window of keys;
// the tab it is, where it is one; whether it is listed before the reader types anything; and what going there does.
type place struct {
	name string
	tab  int
	// listedIdle is whether the place is listed before anything is typed: the tabs and the window of keys are, the
	// projects' pages only once the reader types, so the places of the page shown fit on the line
	listedIdle bool
	// act is what going there does, as a key's action does it, with no count
	act action
}

// placesOf are the places the prompt goes to from a page: each of its tabs, each project's page, as :e opens it, and
// the window of keys, as the question mark opens it, in that order; a name taken already by one before is left to
// it, so each name goes to one place.
func placesOf(tabs []tab, slugs []string) []place {
	places := make([]place, 0, len(tabs)+len(slugs)+1)
	for index, shown := range tabs {
		places = append(places, place{name: placeName(shown.title), tab: index, listedIdle: true, act: func(model Model, _ int) (Model, tea.Cmd) { return model.choose(index), nil }})
	}
	for _, slug := range slugs {
		places = append(places, place{name: slug, tab: noTab, listedIdle: false, act: func(model Model, _ int) (Model, tea.Cmd) { return model.edit(slug) }})
	}
	places = append(places, place{name: helpPlace, tab: noTab, listedIdle: true, act: Model.openKeys})
	unique := make([]place, 0, len(places))
	for _, known := range places {
		if !slices.ContainsFunc(unique, func(taken place) bool { return taken.name == known.name }) {
			unique = append(unique, known)
		}
	}
	return unique
}

// placeName is the name a tab's place goes by: its title in small letters, its words joined as a path joins them.
func placeName(title string) string {
	return strings.ToLower(strings.Join(strings.Fields(title), placeWordJoint))
}

// placesFor are the places the prompt lists for what is typed while a tab is chosen, as completions: before anything
// is typed, those listed idle; after, every place whose name starts with what is typed; never the tab chosen, where
// the reader already is.
func (model Model) placesFor(chosen int, typed string) []completion {
	listed := []completion{}
	for _, known := range model.places {
		if known.tab == chosen || !strings.HasPrefix(known.name, typed) || typed == "" && !known.listedIdle {
			continue
		}
		listed = append(listed, completion{shown: known.name, written: known.name})
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
	return place{name: "", tab: noTab, listedIdle: false, act: nil}, false
}
