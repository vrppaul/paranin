package compose

import (
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// Every fact of the profile a person reads reaches the résumé, as a phrase of its own: a face that shows every
// phrase shows every fact.
func TestEveryFactOfTheProfileReachesTheResume(test *testing.T) {
	phrases := document.Phrases(Resume(fixtures.LoadedContent(test)))
	cases := []struct {
		name string
		want string
	}{
		{name: "the name", want: "Mira Example"},
		{name: "the title", want: "Software Engineer"},
		{name: "the place", want: "Brno, Czechia"},
		{name: "the headline", want: "Builds small tools that last."},
		{name: "the summary", want: "Writes software for people who read it."},
		{name: "a fact's label", want: "available"},
		{name: "a fact's value", want: "5 years"},
		{name: "a gap", want: "[notice period]"},
		{name: "the mail address", want: "mira@example.test"},
		{name: "a profile's label", want: "GitHub"},
		{name: "a profile's address as written", want: "github.test/mira"},
		{name: "the ssh command", want: "ssh example.test"},
		{name: "a company", want: "Northwind"},
		{name: "a role", want: "Senior Engineer"},
		{name: "a role's place", want: "Prague"},
		{name: "a point", want: "Wrote the first parser"},
		{name: "a highlighted result", want: "queue"},
		{name: "a project", want: "tidy-tool"},
		{name: "a tagline", want: "Tidies files."},
		{name: "a description", want: "tidies files"},
		{name: "a project's language", want: "Lua"},
		{name: "a repository", want: "https://github.test/mira/tidy"},
		{name: "an install command", want: "go install example.test/tidy@latest"},
		{name: "a group of skills", want: "Tools"},
		{name: "a skill", want: "Docker"},
		{name: "a skill's span", want: "3y"},
		{name: "a skill used on the side, in the site's words", want: "on the side"},
		{name: "a skill's sources", want: "tidy-tool, a hobby"},
		{name: "a school", want: "A University"},
		{name: "a field of study", want: "Computer Science and the Engineering of Systems That Last"},
		{name: "a month still to fill", want: "[20XX]"},
		{name: "a language", want: "English"},
		{name: "a language's level", want: "fluent"},
		{name: "a word of the site", want: "Experience"},
		{name: "a label of the site", want: "repository"},
		{name: "the word for what still goes on", want: "present"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !slices.Contains(phrases, testCase.want) {
				test.Errorf("phrases = %q, want one %q", phrases, testCase.want)
			}
		})
	}
}

// The résumé's parts stand in one order, decided once here, which every face follows: five sections, each under the
// site's word for it, the summary and the facts in the first; and in the skills, after the groups, the languages and
// the education.
func TestTheResumeKeepsOneOrder(test *testing.T) {
	resume := Resume(fixtures.LoadedContent(test))
	skills := sectionTitled(test, resume, "Skills").Blocks
	cases := []struct {
		name   string
		blocks []document.Block
		want   []string
	}{
		{name: "the sections", blocks: resume.Blocks, want: []string{"About", "Experience", "Projects", "Skills", "Contact"}},
		{name: "the skills' last parts", blocks: skills[len(skills)-2:], want: []string{"Languages", "Education"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			got := []string{}
			for _, block := range testCase.blocks {
				got = append(got, titleOf(test, as[document.Section](test, block)))
			}
			if !slices.Equal(got, testCase.want) {
				test.Errorf("sections = %q, want %q", got, testCase.want)
			}
		})
	}
}

// A skill group's table tells a skill's duration from the words beside it, so each face can draw it set back.
func TestASkillsDurationStandsInACellOfItsOwnKind(test *testing.T) {
	group := sectionTitled(test, Resume(fixtures.LoadedContent(test)), "Tools")
	table := as[document.Table](test, group.Blocks[0])
	got := [][]string{}
	for _, row := range table.Rows {
		got = append(got, kindNamesOf(row.Cells))
	}
	want := [][]string{{"text", "duration", "text"}}
	if !reflect.DeepEqual(got, want) {
		test.Errorf("the group's cells are of the kinds %q, want %q", got, want)
	}
}

// The résumé's first section, about the author, holds the summary and then the facts, and nothing more.
func TestTheAboutSectionHoldsTheSummaryAndTheFacts(test *testing.T) {
	about := sectionTitled(test, Resume(fixtures.LoadedContent(test)), "About").Blocks
	got := make([]string, 0, len(about))
	for _, block := range about {
		got = append(got, fmt.Sprintf("%T", block))
	}
	if want := []string{"document.Paragraph", "document.Facts"}; !slices.Equal(got, want) {
		test.Errorf("the about section holds %q, want %q", got, want)
	}
}

// A role's months reach the document typed, so each face writes them its own way, and a role still held stays
// current.
func TestEachRoleKeepsItsPeriod(test *testing.T) {
	experience := sectionTitled(test, Resume(fixtures.LoadedContent(test)), "Experience")
	got := []document.Period{}
	for _, block := range experience.Blocks {
		got = append(got, as[document.Entry](test, block).Period)
	}
	want := []document.Period{
		{Start: document.KnownMonth{Start: firstMoment(2023, time.April)}, End: document.CurrentMonth{Text: "present"}},
		{Start: document.KnownMonth{Start: firstMoment(2020, time.January)}, End: document.KnownMonth{Start: firstMoment(2023, time.March)}},
	}
	if !reflect.DeepEqual(got, want) {
		test.Errorf("periods = %+v, want %+v", got, want)
	}
}

// Each entry of the résumé says what its time was spent on: an entry of the experience is a role, one of the
// education a school, so a face takes either by its kind, not by the section it stands in.
func TestEachEntryTellsARoleFromASchool(test *testing.T) {
	resume := Resume(fixtures.LoadedContent(test))
	kindNames := map[document.EntryKind]string{0: "of no kind", document.EntryRole: "a role", document.EntrySchool: "a school"}
	cases := []struct {
		name    string
		section string
		want    document.EntryKind
	}{
		{name: "a role", section: "Experience", want: document.EntryRole},
		{name: "a school", section: "Education", want: document.EntrySchool},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			for _, block := range sectionTitled(test, resume, testCase.section).Blocks {
				if got := as[document.Entry](test, block).Kind; got != testCase.want {
					test.Errorf("an entry of %q is %s, want %s", testCase.section, kindNames[got], kindNames[testCase.want])
				}
			}
		})
	}
}

// Every link of the résumé and of a project's page leads where the profile says: a face writes the address, but
// never chooses it.
func TestEveryLinkLeadsWhereTheProfileSays(test *testing.T) {
	documents := SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay()).Documents()
	resume := fixtures.DocumentOf(test, documents, document.Page{Kind: document.PageResume, Slug: "", Tab: ""})
	project := fixtures.DocumentOf(test, documents, document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""})
	projects := sectionTitled(test, resume, "Projects")
	pageless, tidy := cardTitled(test, projects, "other-tool"), cardTitled(test, projects, "tidy-tool")
	ways := waysOf(test, resume)
	cases := []struct {
		name string
		got  document.Target
		want document.Target
	}{
		{name: "the mail address", got: linkTarget(test, factLabelled(test, ways, "mail").Value), want: document.Mail{Address: "mira@example.test"}},
		{name: "a public profile", got: linkTarget(test, factLabelled(test, ways, "GitHub").Value), want: document.Address{URL: "https://github.test/mira"}},
		{name: "a project's page", got: tidy.Page, want: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}},
		{name: "a project with no page", got: pageless.Page, want: document.Page{Kind: document.PageNone, Slug: "", Tab: ""}},
		{name: "a project's repository", got: linkTarget(test, factLabelled(test, tidy.Facts, "repository").Value), want: document.Address{URL: "https://github.test/mira/tidy"}},
		{name: "a project's page back to the résumé", got: linkTarget(test, factLabelled(test, detailsOf(test, project), "author").Value), want: document.Page{Kind: document.PageResume, Slug: "", Tab: ""}},
		{name: "a project's page to its code", got: linkTarget(test, factLabelled(test, project.Head.Repository.Facts(), "repository").Value), want: document.Address{URL: "https://github.test/mira/tidy"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if testCase.got != testCase.want {
				test.Errorf("target = %+v, want %+v", testCase.got, testCase.want)
			}
		})
	}
}

// Every fact of a project the profile gives reaches its page, as a phrase of its own: a face that shows every phrase
// shows every fact.
func TestEveryFactOfAProjectReachesItsPage(test *testing.T) {
	project := fixtures.DocumentOf(test, SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay()).Documents(), document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""})
	phrases := document.Phrases(project)
	cases := []struct {
		name string
		want string
	}{
		{name: "its name", want: "tidy-tool"},
		{name: "its tagline", want: "Tidies files."},
		{name: "its language", want: "Lua"},
		{name: "what it is built on", want: "LPeg"},
		{name: "its repository", want: "https://github.test/mira/tidy"},
		{name: "a fact of the profile", want: "done"},
		{name: "its install command", want: "go install example.test/tidy@latest"},
		{name: "its recording", want: "$ tidy .\ntidied 3 files"},
		{name: "a section's paragraph", want: "Files were untidy."},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !slices.Contains(phrases, testCase.want) {
				test.Errorf("phrases = %q, want one %q", phrases, testCase.want)
			}
		})
	}
}

// A project's page opens with what the project is made of, its language and then its stack, as the profile lists
// them, so a face shows them beside its name.
func TestAProjectsHeadNamesWhatTheProjectIsMadeOf(test *testing.T) {
	tidyTool := document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}
	cases := []struct {
		name     string
		snapshot *content.Snapshot
		want     []string
	}{
		{name: "a project written in a language", snapshot: fixtures.LoadedContent(test), want: []string{"Lua", "Penlight", "LPeg"}},
		{name: "a project that names no language", snapshot: withoutLanguages(fixtures.LoadedContent(test)), want: []string{"Penlight", "LPeg"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := fixtures.DocumentOf(test, SiteOf(Prepare(testCase.snapshot), publicOnTheDay()).Documents(), tidyTool).Head.MadeOf; !slices.Equal(got, testCase.want) {
				test.Errorf("the tags of tidy-tool's page = %q, want %q", got, testCase.want)
			}
		})
	}
}

// A project's page names where its code lives once, in its head, and its details hold the rest: who made it, the
// profile's facts, how it is installed.
func TestAProjectsPageNamesItsRepositoryOnce(test *testing.T) {
	project := fixtures.DocumentOf(test, SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay()).Documents(), document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""})
	cases := []struct {
		name  string
		facts []document.Fact
		want  []string
	}{
		{name: "its head", facts: project.Head.Repository.Facts(), want: []string{"repository"}},
		{name: "its details", facts: detailsOf(test, project), want: []string{"author", "status", "install"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			got := []string{}
			for _, fact := range testCase.facts {
				got = append(got, fact.Label)
			}
			if !slices.Equal(got, testCase.want) {
				test.Errorf("labels = %q, want %q", got, testCase.want)
			}
		})
	}
}

// The index lists each page under the site's word for its kind, by the title and the line its own page opens
// with, so an agent finds each page by what it is called there.
func TestTheIndexListsEachPageUnderItsWord(test *testing.T) {
	index := SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay()).Index
	cases := []struct {
		name    string
		section string
		want    []document.Listing
	}{
		{name: "the résumé", section: "Résumé", want: []document.Listing{{
			Page:  document.Page{Kind: document.PageResume, Slug: "", Tab: ""},
			Title: []document.Span{document.Text{Text: "Mira Example"}},
			Note:  []document.Span{document.Text{Text: "Software Engineer"}},
		}}},
		{name: "the projects' pages", section: "Projects", want: []document.Listing{{
			Page:  document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""},
			Title: []document.Span{document.Text{Text: "tidy-tool"}},
			Note:  []document.Span{document.Text{Text: "Tidies files."}},
		}}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := contentsOf(test, sectionTitled(test, index, testCase.section)).Listings; !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("listings under %q = %+v, want %+v", testCase.section, got, testCase.want)
			}
		})
	}
}

// The index lists the résumé and every project's page compose makes, in the order it makes them: a project's page
// added to the site is a page added to the index.
func TestTheIndexListsEveryPageComposeMakes(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	index := SiteOf(Prepare(snapshot), publicOnTheDay()).Index
	got := []document.Page{}
	for _, title := range []string{"Résumé", "Projects"} {
		for _, listed := range contentsOf(test, sectionTitled(test, index, title)).Listings {
			got = append(got, listed.Page)
		}
	}
	want := []document.Page{}
	for _, page := range SiteOf(Prepare(snapshot), publicOnTheDay()).Documents() {
		want = append(want, page.Head.Page)
	}
	if !slices.Equal(got, want) {
		test.Errorf("listed pages = %+v, want the pages compose makes, %+v", got, want)
	}
}

// seenListing is a page the index lists as a test reads it: the page, its title and its line.
type seenListing struct {
	page  document.Page
	title string
	note  string
}

// The index lists the blog under the site's word for it: the blog's list in each grouping, by its title naming the
// grouping and by the blog's line, then each post the list names, newest first, by its title and its description;
// never a draft, a preview, a post unlisted or one that comes later than its reader.
func TestTheIndexListsTheBlogsListsThenEveryPostListedNewestFirst(test *testing.T) {
	posts := []writtenPost{
		{slug: "first", id: 3, status: "published", published: earlier, tags: ""},
		{slug: "toolbox", id: 4, status: "published", published: latest, tags: ""},
		{slug: "drafted", id: 5, status: "draft", published: later, tags: ""},
		{slug: "previewed", id: 6, status: "preview", published: later, tags: ""},
		{slug: "hidden", id: 7, status: "unlisted", published: later, tags: ""},
		{slug: "coming", id: 8, status: "published", published: afterTheDay, tags: ""},
	}
	index := SiteOf(Prepare(withPostsAndSeries(test, posts, nil)), publicOnTheDay()).Index
	want := []seenListing{
		{page: document.PageOf(document.PageBlog), title: "blog by years", note: "What the writer reads, and how."},
		{page: document.PageOf(document.PageBlogSeries), title: "blog by series", note: "What the writer reads, and how."},
		{page: document.Page{Kind: document.PagePost, Slug: "toolbox", Tab: ""}, title: "toolbox", note: "About toolbox."},
		{page: document.Page{Kind: document.PagePost, Slug: fixturePostSlug, Tab: ""}, title: fixturePostTitle, note: "Why a page read twice is read once."},
		{page: document.Page{Kind: document.PagePost, Slug: "first", Tab: ""}, title: "first", note: "About first."},
	}
	got := []seenListing{}
	for _, listed := range contentsOf(test, sectionTitled(test, index, "blog")).Listings {
		got = append(got, seenListing{page: listed.Page, title: document.PlainText(listed.Title), note: document.PlainText(listed.Note)})
	}
	if !reflect.DeepEqual(got, want) {
		test.Errorf("listings under the blog = %+v, want %+v", got, want)
	}
}

// The projects' word stands in the index only where a project has a page to list under it.
func TestTheIndexHasNoProjectsWhereNoProjectHasAPage(test *testing.T) {
	withoutPages := *fixtures.LoadedContent(test)
	withoutPages.Profile.Projects = slices.Clone(withoutPages.Profile.Projects)
	for index := range withoutPages.Profile.Projects {
		withoutPages.Profile.Projects[index].Page = false
	}
	got := []string{}
	for _, block := range SiteOf(Prepare(&withoutPages), publicOnTheDay()).Index.Blocks {
		got = append(got, titleOf(test, as[document.Section](test, block)))
	}
	if want := []string{"Résumé", "blog"}; !slices.Equal(got, want) {
		test.Errorf("sections = %q, want %q", got, want)
	}
}

// Whose site it is, the site's words and every listing's title and line reach the index as phrases of their own:
// a face that shows every phrase lists every page.
func TestEveryPageReachesTheIndex(test *testing.T) {
	phrases := document.Phrases(SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay()).Index)
	cases := []struct {
		name string
		want string
	}{
		{name: "the headline", want: "Builds small tools that last."},
		{name: "a word of the site", want: "Résumé"},
		{name: "a listing's title", want: "tidy-tool"},
		{name: "a listing's line", want: "Tidies files."},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !slices.Contains(phrases, testCase.want) {
				test.Errorf("phrases = %q, want one %q", phrases, testCase.want)
			}
		})
	}
}

// Every document compose makes is in the language its words are written in, the fixture's English, so a face can
// name the language its reader reads.
func TestEveryDocumentIsInTheLanguageOfItsWords(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	site := SiteOf(Prepare(snapshot), publicOnTheDay())
	documents := slices.Concat(site.Documents(), tabDocumentsOf(snapshot), []document.Document{site.Index}, Extras(snapshot))
	for _, composed := range documents {
		if got, want := composed.Head.Language, "en"; got != want {
			test.Errorf("the language of %+v = %q, want %q", composed.Head.Page, got, want)
		}
	}
}

// Each page a person reads has a description, for a search engine's snippet and a link's preview where it is
// shared: the résumé's is its headline, a project's page's what the project does, a tab's its page's.
func TestEachPageHasADescription(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	documents := slices.Concat(SiteOf(Prepare(snapshot), publicOnTheDay()).Documents(), tabDocumentsOf(snapshot))
	cases := []struct {
		name string
		page document.Page
		want string
	}{
		{name: "the résumé", page: document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, want: "Builds small tools that last."},
		{name: "a project's page", page: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}, want: "A tool that tidies files."},
		{name: "a tab of the résumé", page: document.Page{Kind: document.PageResume, Slug: "", Tab: "experience"}, want: "Builds small tools that last."},
		{name: "a tab of a project's page", page: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: "why"}, want: "A tool that tidies files."},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := document.PlainText(fixtures.DocumentOf(test, documents, testCase.page).Head.Description); got != testCase.want {
				test.Errorf("the description of %+v = %q, want %q", testCase.page, got, testCase.want)
			}
		})
	}
}

// The résumé's head greets the reader before the author's name, in the site's words; a project's page does not.
func TestTheResumeHeadGreetsTheReader(test *testing.T) {
	documents := SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay()).Documents()
	cases := []struct {
		name string
		page document.Page
		want string
	}{
		{name: "the résumé", page: document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, want: "Hello, my name is"},
		{name: "a project's page", page: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}, want: ""},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := fixtures.DocumentOf(test, documents, testCase.page).Head.Greeting; got != testCase.want {
				test.Errorf("the greeting of %+v = %q, want %q", testCase.page, got, testCase.want)
			}
		})
	}
}

// tabDocumentsOf are the documents of every tab past an opening of the documents a snapshot makes.
func tabDocumentsOf(snapshot *content.Snapshot) []document.Document {
	documents := []document.Document{}
	for _, composed := range SiteOf(Prepare(snapshot), publicOnTheDay()).Pages {
		documents = append(documents, composed.Tabs...)
	}
	return documents
}

// contentsOf are a section's contents, wherever compose put them.
func contentsOf(test *testing.T, section document.Section) document.Contents {
	for _, block := range section.Blocks {
		if contents, isContents := block.(document.Contents); isContents {
			return contents
		}
	}
	test.Fatalf("the section %+v lists no pages", section.Title)
	return document.Contents{Listings: nil}
}

// withoutLanguages is a snapshot whose projects name no language, the rest as it was.
func withoutLanguages(snapshot *content.Snapshot) *content.Snapshot {
	changed := *snapshot
	changed.Profile.Projects = slices.Clone(snapshot.Profile.Projects)
	for index := range changed.Profile.Projects {
		changed.Profile.Projects[index].Language = ""
	}
	return &changed
}

// cardTitled is a section's card under a title, wherever compose put it.
func cardTitled(test *testing.T, section document.Section, title string) document.Card {
	for _, block := range section.Blocks {
		if card, isCard := block.(document.Card); isCard && as[document.Text](test, card.Title[0]).Text == title {
			return card
		}
	}
	test.Fatalf("the section has no card %q", title)
	return document.Card{Title: nil, Subtitle: nil, Tag: "", Summary: nil, Facts: nil, Page: document.Page{Kind: document.PageNone, Slug: "", Tab: ""}}
}

// detailsOf are a page's facts, wherever compose put them.
func detailsOf(test *testing.T, page document.Document) []document.Fact {
	for _, block := range page.Blocks {
		if facts, isFacts := block.(document.Facts); isFacts {
			return facts.List
		}
	}
	test.Fatalf("the page %+v has no facts", page.Head.Page)
	return nil
}

// waysOf are the ways to reach the author a page holds, the facts of their kind in any of its sections, found as a
// face finds them.
func waysOf(test *testing.T, page document.Document) []document.Fact {
	for _, block := range page.Blocks {
		// a block that is no section holds no block of its own
		section, _ := block.(document.Section)
		for _, inner := range section.Blocks {
			if facts, isFacts := inner.(document.Facts); isFacts && facts.Kind == document.FactsWays {
				return facts.List
			}
		}
	}
	test.Fatalf("the page %+v has no ways to reach its author", page.Head.Page)
	return nil
}

// factLabelled is the fact under a label.
func factLabelled(test *testing.T, facts []document.Fact, label string) document.Fact {
	for _, fact := range facts {
		if fact.Label == label {
			return fact
		}
	}
	test.Fatalf("no fact is labelled %q among %+v", label, facts)
	return document.Fact{Label: label, Value: nil}
}

// sectionTitled is a document's section under a title, wherever compose put it, nested in another or not.
func sectionTitled(test *testing.T, composed document.Document, title string) document.Section {
	if section, found := sectionIn(test, composed.Blocks, title); found {
		return section
	}
	test.Fatalf("the page %+v has no section %q", composed.Head.Page, title)
	return document.Section{Title: nil, Blocks: nil}
}

// sectionIn is the section under a title among blocks or the sections nested in them, and whether there is one.
func sectionIn(test *testing.T, blocks []document.Block, title string) (document.Section, bool) {
	for _, block := range blocks {
		section, isSection := block.(document.Section)
		if !isSection {
			continue
		}
		if titleOf(test, section) == title {
			return section, true
		}
		if nested, found := sectionIn(test, section.Blocks, title); found {
			return nested, true
		}
	}
	return document.Section{Title: nil, Blocks: nil}, false
}

// kindNamesOf are the names of a row's cells' kinds, in its order; a kind with no name reads as "".
func kindNamesOf(cells []document.Cell) []string {
	names := make([]string, 0, len(cells))
	for _, cell := range cells {
		names = append(names, cellKindNames()[cell.Kind])
	}
	return names
}

// cellKindNames name each kind of a table's cell, so a failure says which kinds came.
func cellKindNames() map[document.CellKind]string {
	return map[document.CellKind]string{document.CellText: "text", document.CellDuration: "duration"}
}

func titleOf(test *testing.T, section document.Section) string {
	return as[document.Text](test, section.Title[0]).Text
}

// linkTarget is where a value that is one link leads.
func linkTarget(test *testing.T, spans []document.Span) document.Target {
	return as[document.Link](test, spans[0]).Target
}

// as is a value as the kind a test expects of it, or the test's end, saying what it was.
func as[Kind any](test *testing.T, value any) Kind {
	typed, isKind := value.(Kind)
	if !isKind {
		test.Fatalf("%+v is no %T", value, typed)
	}
	return typed
}

// firstMoment is a month's first moment, as a known month holds it.
//
//wrapper:names the moment a known month starts at, which every expected period writes
func firstMoment(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
}
