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

// resumeParts are the résumé's opening as the page lays it out (DESIGN.md, Pages): the facts its first screen briefs
// the reader with, the address heading them and the profiles ending them, beside the head, and the career line under
// them; then the blocks that are no section, and the first section, holding the facts the brief leaves.
type resumeParts struct {
	brief    []document.Fact
	address  []document.Span
	profiles []document.Link
	career   careerDrawing
	others   []document.Block
	first    document.Section
}

// openingPart is what the main part holds where a page shows its opening (document.TabsOf), as a reader of the
// finished work sees it: on a project's page its head and its blocks on the project's grid; on the résumé that opens
// with a brief, the first screen of the facts written and the résumé's parts under it; on any other page, and on a
// résumé whose brief would hold nothing but gaps, its head, then the opening's blocks.
func (writer PageWriter) openingPart(composed document.Document, opening document.Opening) templ.Component {
	blocks := document.WrittenBlocks(opening.Blocks)
	if project := composed.Head.Project; project != nil {
		return writer.projectMain(composed.Head, *project, blocks)
	}
	if opening.Brief == nil {
		return writer.plainMain(composed.Head, blocks)
	}
	others, sections := document.SectionsOf(document.WrittenBlocks(composed.Blocks))
	briefed := document.BriefOf(sections)
	if briefed == nil {
		return writer.plainMain(composed.Head, blocks)
	}
	return writer.resumeMain(composed.Head, partsOf(briefed, others, writer.today))
}

// tabPart is what the main part holds where a page shows a tab past its opening, its blocks under its title in block
// letters as the page's h1: on the résumé the tab stands as a section, the title beside its blocks, whose headings
// stand under the h1; on any other page the title heads its blocks, as a page's head does.
func (writer PageWriter) tabPart(kind document.PageKind, title []document.Span, blocks []document.Block) templ.Component {
	heading := titleHeading(document.PlainText(title), writer.spans(title))
	if kind == document.PageResume {
		return writer.resumeSection(blocks, heading, firstHeadingDepth)
	}
	return writer.tabMain(heading, blocks)
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
	return resumeParts{
		brief: list[:shown], address: briefed.Address(), profiles: profilesOf(briefed.Ways),
		career: careerDrawingOf(document.CareerOf(briefed.Roles, today)),
		others: others, first: document.Section{Title: briefed.Title, Name: "", Blocks: slices.Concat(briefed.Before, kept, briefed.After)},
	}
}

// linkLine is links in one line, the brief's ways to the author's profiles or where a post is discussed, each drawn as
// any link is, the line's dot between two. It is joined here, not in a template's loop, where templ writes a space
// before each part.
func (writer PageWriter) linkLine(links []document.Link) templ.Component {
	parts := make([]templ.Component, 0, len(links))
	for index, link := range links {
		if index > 0 {
			parts = append(parts, text(dot))
		}
		parts = append(parts, writer.span(link))
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
