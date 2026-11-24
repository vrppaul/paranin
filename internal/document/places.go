package document

import "slices"

// PlaceKind is what a place of a page's command line is, the word its menu names it by (DESIGN.md, CommandLine), as
// every face writes it: a tab of the page shown or a project's page, which a face offers on every page; a list of the
// blog or a post, as :e opens them; or a place of the page's own (PlacesOf).
type PlaceKind string

const (
	// PlaceTab is a tab of the page shown, a window of the blog's session among them
	PlaceTab PlaceKind = "tab"
	// PlaceProject is a project's page, by its slug, which :e opens too
	PlaceProject PlaceKind = "project"
	// PlaceList is the blog's list in one of its groupings, as :e opens it
	PlaceList PlaceKind = "list"
	// PlaceHeading is a heading of a post, inside its page
	PlaceHeading PlaceKind = "heading"
	// PlacePart is another part out of a post's series
	PlacePart PlaceKind = "part"
	// PlaceVersion is another version of a post, by the day it came out
	PlaceVersion PlaceKind = "version"
	// PlacePost is a post the blog's list names, as the list offers it and :e opens it
	PlacePost PlaceKind = "post"
	// PlaceSeries is a series' group on the blog's list by series
	PlaceSeries PlaceKind = "series"
	// PlaceTag is a tag of the posts the blog's list names, which leads where the list's head leads by it
	PlaceTag PlaceKind = "tag"
)

// Place is somewhere a page's command line goes by its name: its kind, the page it is or lies in, and the section of
// that page it is, by the section's anchor (Section.Name); none where it is the page whole.
type Place struct {
	Name    string
	Kind    PlaceKind
	Page    Page
	Section string
}

// PlacesOf are a page's own places, which its command line offers after its tabs and before the projects' pages, read
// off the document as its tabs and its notes are, so none is kept in step with the blocks: on a post every heading in
// reading order, then its series' other parts out, then its other versions, newest first; on the blog's list, filtered
// or not, each post out it names, then each tag its head names, then each series a post out is a part of, in the order
// they first stand in; none on any other page. exhaustive holds the switch to every kind.
func PlacesOf(composed Document) []Place {
	head := composed.Head
	switch head.Page.Kind {
	case PagePost:
		return slices.Concat(headingPlaces(head.Page, composed.Blocks), partPlaces(head.Post), versionPlaces(head.Version))
	case PageBlog, PageBlogSeries, PageTagFilter, PageSeriesFilter:
		return listPlaces(head, composed.Blocks)
	case PageNone, PageResume, PageProject, PageIndex, PageHealth, PageSkills:
	}
	return nil
}

// headingPlaces are the headings of a post's body and of its sections, in reading order, each by its anchor.
func headingPlaces(page Page, blocks []Block) []Place {
	places := []Place{}
	for _, block := range blocks {
		if section, isSection := block.(Section); isSection {
			places = append(places, Place{Name: section.Name, Kind: PlaceHeading, Page: page, Section: section.Name})
			places = append(places, headingPlaces(page, section.Blocks)...)
		}
	}
	return places
}

// partPlaces are the parts of a post's series a reader can go to, each by its slug: every part out but the one read,
// which is the page shown, and none to come, which has no page; none outside a series. exhaustive holds the switch to
// every state.
func partPlaces(place *PostPlace) []Place {
	places := []Place{}
	if place == nil || place.Series == nil {
		return places
	}
	for _, part := range place.Series.Parts {
		switch part.State {
		case PartOut, PartNext:
			places = append(places, Place{Name: part.Page.Slug, Kind: PlacePart, Page: part.Page, Section: ""})
		case PartReading, PartToCome:
		}
	}
	return places
}

// versionPlaces are the versions of a post a reader can go to, each by its name, the day it came out: every version but
// the one read, which is the page shown; none for a post of one version alone or a draft.
func versionPlaces(version Version) []Place {
	places := []Place{}
	for _, other := range VersionsOf(version) {
		if !other.Read {
			places = append(places, Place{Name: other.Name, Kind: PlaceVersion, Page: other.Page, Section: ""})
		}
	}
	return places
}

// listPlaces are the places of the blog's list of a head: each post out its groups name, by its slug, then each tag
// its head's filters name, leading where the filter leads, then each series one of the posts is a part of, by its
// name, at its group on the list by series.
func listPlaces(head Head, blocks []Block) []Place {
	posts, tags, series := []Place{}, []Place{}, []Place{}
	for _, out := range postsOut(blocks) {
		posts = append(posts, Place{Name: out.page.Slug, Kind: PlacePost, Page: out.page, Section: ""})
		if out.seriesName != "" && !slices.ContainsFunc(series, func(known Place) bool { return known.Name == out.seriesName }) {
			series = append(series, Place{Name: out.seriesName, Kind: PlaceSeries, Page: PageOf(PageBlogSeries), Section: out.seriesName})
		}
	}
	for _, filter := range head.Filters {
		tags = append(tags, tagPlaces(filter)...)
	}
	return slices.Concat(posts, tags, series)
}

// tagPlaces are the places of a filter of a list's head: a tag, by its name, where the filter leads; none for a
// series, which its group stands for as a place of its own. exhaustive holds the switch to every kind.
func tagPlaces(filter Filter) []Place {
	switch filter.Kind {
	case FilterTag:
		return []Place{{Name: filter.Name, Kind: PlaceTag, Page: filter.Page, Section: ""}}
	case FilterSeries:
	}
	return nil
}

// postOut is a post out the blog's list names, as its places read it: its page, and its series' name, none outside one.
type postOut struct {
	page       Page
	seriesName string
}

// postsOut are the posts out the groups of the blog's list name, in their order; a block that is no group holds none.
func postsOut(blocks []Block) []postOut {
	posts := []postOut{}
	for _, block := range blocks {
		group, _ := GroupOf(block)
		for _, listed := range group.Posts.List {
			if out, isOut := listed.Readiness.(PostOut); isOut {
				posts = append(posts, postOut{page: out.Page, seriesName: listed.SeriesName})
			}
		}
	}
	return posts
}
