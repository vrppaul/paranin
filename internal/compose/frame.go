package compose

import (
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// Frame is what a face lays round the documents it draws, on every page alike, in the site's words: the words of
// the link past it to the page's main part, the titles of the résumé's sections, which it leads to in their order,
// and the name of the theme button, which is pressed while the page is dark.
type Frame struct {
	SkipToContent string
	Sections      []string
	DarkTheme     string
}

// FrameOf is the frame of a site's pages in its words, its sections those of the résumé among the pages composed;
// none where they hold no résumé.
func FrameOf(words content.Words, pages []Composed) Frame {
	sections := []string{}
	for _, page := range pages {
		if page.Document.Head.Page.Kind != document.PageResume {
			continue
		}
		for _, block := range page.Document.Blocks {
			if section, isSection := block.(document.Section); isSection {
				sections = append(sections, document.PlainText(section.Title))
			}
		}
	}
	return Frame{
		SkipToContent: words.Text(content.WordSkipToContent),
		Sections:      sections,
		DarkTheme:     words.Text(content.WordDarkTheme),
	}
}
