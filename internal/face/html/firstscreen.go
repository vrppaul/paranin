package html

import (
	"slices"
	"strings"
	"time"

	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// The most facts the first screen briefs the reader with (DESIGN.md, Facts): the first the résumé's first section
// holds, in the order the profile writes them, which is the author's choice; that section keeps the rest where they
// stood
const briefFacts = 4

// The at a mail address parts at, between its mailbox and its host
const addressAt = "@"

// resumeParts are the résumé's main part as the page lays it out (DESIGN.md, Pages): the facts its first screen
// briefs the reader with and the address heading them, beside the head, and the career line under them; then the
// blocks that are no section, and the sections, the first holding the facts the brief leaves.
type resumeParts struct {
	brief    []document.Fact
	address  []document.Span
	career   careerDrawing
	others   []document.Block
	sections []document.Section
}

// mailAddress is an address as the brief heads with it (DESIGN.md, Mail): where it leads, its mailbox and its host.
type mailAddress struct {
	reference string
	box       string
	host      string
}

// mainPart is what a page's main part holds: on the résumé whose first section holds facts, the first screen and
// the résumé's parts under it; on any other page, its head, then each block.
func (writer PageWriter) mainPart(composed document.Document) templ.Component {
	if composed.Head.Page.Kind != document.PageResume {
		return writer.plainMain(composed)
	}
	others, sections := document.SectionsOf(composed.Blocks)
	briefed := document.BriefOf(sections)
	if briefed == nil {
		return writer.plainMain(composed)
	}
	return writer.resumeMain(composed.Head, partsOf(briefed, others, writer.today))
}

// partsOf are the résumé's parts round its brief, on a day: the first facts briefed, the rest kept in their section
// where the brief stood, and the roles on the career line, which runs to the day.
func partsOf(briefed *document.Briefed, others []document.Block, today time.Time) resumeParts {
	list := briefed.Brief.List
	shown := min(briefFacts, len(list))
	kept := []document.Block{}
	if shown < len(list) {
		kept = append(kept, document.Facts{Kind: briefed.Brief.Kind, List: list[shown:]})
	}
	first := document.Section{Title: briefed.Title, Blocks: slices.Concat(briefed.Before, kept, briefed.After)}
	return resumeParts{
		brief: list[:shown], address: briefed.Address, career: careerDrawingOf(document.CareerOf(briefed.Roles, today)),
		others: others, sections: slices.Concat([]document.Section{first}, briefed.Rest),
	}
}

// mailOf is an address as the brief heads with it, parted at its at; not parted where it is not one link with an
// address whose words hold an at.
func mailOf(address []document.Span) (mailAddress, bool) {
	none := mailAddress{reference: "", box: "", host: ""}
	if len(address) != 1 {
		return none, false
	}
	link, isLink := address[0].(document.Link)
	if !isLink {
		return none, false
	}
	reference := routes.Reference(link.Target, routes.FormatChosen)
	box, host, parted := strings.Cut(document.PlainText(link.Spans), addressAt)
	if reference == "" || !parted {
		return none, false
	}
	return mailAddress{reference: reference, box: box, host: host}, true
}
