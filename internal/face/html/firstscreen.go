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
// briefs the reader with, the address heading them and the profiles ending them, beside the head, and the career line
// under them; then the blocks that are no section, and the sections, the first holding the facts the brief leaves.
type resumeParts struct {
	brief    []document.Fact
	address  []document.Span
	profiles []document.Link
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
		brief: list[:shown], address: briefed.Address(), profiles: profilesOf(briefed.Ways),
		career: careerDrawingOf(document.CareerOf(briefed.Roles, today)),
		others: others, sections: slices.Concat([]document.Section{first}, briefed.Rest),
	}
}

// profileLine is the brief's ways to the author's profiles in one line, each drawn as any link is, the line's dot
// between two. It is joined here, not in a template's loop, where templ writes a space before each part.
func (writer PageWriter) profileLine(profiles []document.Link) templ.Component {
	parts := make([]templ.Component, 0, len(profiles))
	for index, profile := range profiles {
		if index > 0 {
			parts = append(parts, text(dot))
		}
		parts = append(parts, writer.span(profile))
	}
	return templ.Join(parts...)
}

// profilesOf are the ways to the author's profiles, the ways whose value is one link to an address outside the site,
// in their order, each a link by the way's label: not the mail, which heads the brief, nor a command to run.
func profilesOf(ways []document.Fact) []document.Link {
	profiles := []document.Link{}
	for _, way := range ways {
		link, isLink := document.LinkOf(way.Value)
		if _, outside := link.Target.(document.Address); isLink && outside {
			profiles = append(profiles, document.Link{Target: link.Target, Spans: []document.Span{document.Text{Text: way.Label}}})
		}
	}
	return profiles
}
