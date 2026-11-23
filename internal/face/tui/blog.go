package tui

import (
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// blogNames are the pages of the blog among documents by the names a terminal reaches them by (D26): each list by its
// address, and each post by its address and by its number, so blog/12 reaches the post blog/<slug> does.
func blogNames(documents []document.Document) map[string]document.Page {
	named := map[string]document.Page{}
	for _, held := range documents {
		page := held.Head.Page
		if !page.InBlog() {
			continue
		}
		named[routes.TerminalName(routes.Path(page, routes.FormatChosen))] = page
		if page.Kind == document.PagePost {
			named[routes.TerminalName(routes.NumberPath(held.Head.Number))] = page
		}
	}
	return named
}
