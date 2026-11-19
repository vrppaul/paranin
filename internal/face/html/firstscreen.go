package html

import (
	"slices"
	"time"

	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/document"
)

// The most facts the first screen briefs the reader with (DESIGN.md, Facts): the first the résumé's first section
// holds, in the order the profile writes them, which is the author's choice; that section keeps the rest where they
// stood
const briefFacts = 4

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

// mainPart is what a page's main part holds: on the résumé whose first section holds facts, the first screen and
// the résumé's parts under it; on a résumé with none to brief, its head, then its sections; on any other page, its
// head, then its blocks.
func (writer PageWriter) mainPart(composed document.Document) templ.Component {
	if composed.Head.Page.Kind != document.PageResume {
		return writer.plainMain(composed)
	}
	others, sections := document.SectionsOf(composed.Blocks)
	briefed := document.BriefOf(sections)
	if briefed == nil {
		return writer.unbriefedResumeMain(composed)
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
