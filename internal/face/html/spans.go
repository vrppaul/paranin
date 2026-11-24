package html

import (
	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// spanWriter draws text inside a block, the page's and a feed's alike, a link by the address its writer leads by: a
// page's by its path from the root, which a browser reads against the page (routes.Site's Reference), a feed's whole,
// since a reader shows it away from the site (routes.Site's Address).
type spanWriter struct {
	addressOf func(document.Target, routes.Format) string
}

// span is the component of a span; a link whose target has no address reads as its text.
func (writer spanWriter) span(span document.Span) templ.Component {
	switch typed := span.(type) {
	case document.Text:
		return text(typed.Text)
	case document.Code:
		return code(typed.Text)
	case document.Gap:
		return gap(typed.Text)
	case document.Highlight:
		return writer.highlight(typed.Spans)
	case document.Link:
		if address := writer.addressOf(typed.Target, routes.FormatChosen); address != "" {
			return writer.link(address, typed.Spans)
		}
		return writer.spans(typed.Spans)
	case document.Emphasis:
		return writer.emphasis(typed.Spans)
	case document.Strong:
		return writer.strong(typed.Spans)
	case document.Note:
		// a note's words stand beside the line or at the end; where a line cites it, its number
		return citation(typed.Number)
	}
	return templ.NopComponent
}
