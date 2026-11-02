package tui

import (
	"maps"
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/vrppaul/paranin/internal/document"
)

// What the command line says where :e finds no page by the name typed, as Neovim says it of a file
const (
	pageNotFoundOpening = `E447: Can't find file "`
	pageNotFoundClosing = `" in path`
)

// resumePage is the page a session opens on, and the one q closes a project's page back to.
func resumePage() document.Page {
	return document.Page{Kind: document.PageResume, Slug: ""}
}

// readerPlace is where a reader stood in a buffer: the tab chosen, and how far down each tab they scrolled.
type readerPlace struct {
	chosen  int
	scrolls []int
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

// projectSlugs are the slugs of the projects' pages among the buffers, in the order of their letters.
func projectSlugs(buffers map[document.Page][]tab) []string {
	slugs := []string{}
	for page := range buffers {
		if page.Kind == document.PageProject {
			slugs = append(slugs, page.Slug)
		}
	}
	slices.Sort(slugs)
	return slugs
}

// placesLeft are where a reader stands in each buffer but the one shown before they first read it: on its first
// tab, scrolled nowhere.
func placesLeft(buffers map[document.Page][]tab, shown document.Page) map[document.Page]readerPlace {
	left := make(map[document.Page]readerPlace, len(buffers))
	for page, tabs := range buffers {
		if page != shown {
			left[page] = readerPlace{chosen: 0, scrolls: make([]int, len(tabs))}
		}
	}
	return left
}

// showing is the model showing the buffer of a page instead of the one it showed: where the reader stood in the new
// one comes back, and where they stood in the old one is kept for it. The buffer is laid out and painted for the
// terminal as it is.
func (model Model) showing(page document.Page) Model {
	left := maps.Clone(model.left)
	left[model.shown] = readerPlace{chosen: model.chosen, scrolls: model.scrolls}
	place := left[page]
	delete(left, page)
	model.left, model.shown, model.tabs, model.commands = left, page, model.buffers[page], commands(closingHint(page))
	model.chosen, model.scrolls = place.chosen, place.scrolls
	return model.relaid()
}

// edit shows the page of a project by its slug, as Neovim's :e opens a file, or says it finds no such page.
func (model Model) edit(slug string) (Model, tea.Cmd) {
	wanted := document.Page{Kind: document.PageProject, Slug: slug}
	if _, held := model.buffers[wanted]; !held {
		model.said = commandMessage{text: pageNotFoundOpening + slug + pageNotFoundClosing, failed: true}
		return model, nil
	}
	return model.showing(wanted), nil
}

// closed is the model after q, as Neovim's :q closes a window: on a project's page the résumé shows again, where
// its reader left it, as the first version closed a report; on the résumé the session ends.
func (model Model) closed(int) (Model, tea.Cmd) {
	if model.shown != resumePage() {
		return model.showing(resumePage()), nil
	}
	return model, tea.Quit
}

// closingHint is what the frame says q does on a page: it closes a project's page, and quits on the résumé.
func closingHint(shown document.Page) string {
	if shown == resumePage() {
		return quitHint
	}
	return closeHint
}
