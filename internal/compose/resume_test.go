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
	phrases := document.Phrases(Resume(fixtures.LoadedContent(test), nil))
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
		{name: "a point", want: "Wrote the first parser"},
		{name: "a role's technology", want: "Docker"},
		{name: "a role's detail", want: "Joined a team of four that ran a "},
		{name: "the experience's line", want: "Ships tools a team "},
		{name: "a skill the experience puts forward", want: "Go"},
		{name: "how long it was used", want: "3y"},
		{name: "a highlighted result", want: "queue"},
		{name: "a project", want: "tidy-tool"},
		{name: "a tagline", want: "Tidies files."},
		{name: "a description", want: "tidies files"},
		{name: "a project's language", want: "Lua"},
		{name: "a repository", want: "https://github.test/mira/tidy"},
		{name: "an install command", want: "go install example.test/tidy@latest"},
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

// The résumé's parts stand in one order, decided once here, which every face follows: three sections, each under the
// site's word for it, about the author, their experience and their projects (DESIGN.md, Pages).
func TestTheResumeKeepsOneOrder(test *testing.T) {
	got := []string{}
	for _, block := range Resume(fixtures.LoadedContent(test), nil).Blocks {
		got = append(got, titleOf(test, as[document.Section](test, block)))
	}
	if want := []string{"About", "Experience", "Projects"}; !slices.Equal(got, want) {
		test.Errorf("sections = %q, want %q", got, want)
	}
}

// The résumé's first section, about the author, holds the summary, the facts, and the ways to reach them, which the
// first screen places, and nothing more.
func TestTheAboutSectionHoldsTheSummaryTheFactsAndTheWays(test *testing.T) {
	about := sectionTitled(test, Resume(fixtures.LoadedContent(test), nil), "About").Blocks
	got := make([]string, 0, len(about))
	for _, block := range about {
		got = append(got, fmt.Sprintf("%T", block))
	}
	if want := []string{"document.Paragraph", "document.Facts", "document.Facts"}; !slices.Equal(got, want) {
		test.Fatalf("the about section holds %q, want %q", got, want)
	}
	if ways := as[document.Facts](test, about[len(about)-1]).Kind; ways != document.FactsWays {
		test.Errorf("the about section's last facts are of kind %d, want the ways' kind, %d", ways, document.FactsWays)
	}
}

// The experience opens with its line, then the skills it puts forward, each by its name and how long it was used, in
// the profile's order, as the skills' list writes a duration; its roles come after.
func TestTheExperienceOpensWithItsLineAndTheSkillsPutForward(test *testing.T) {
	experience := sectionTitled(test, Resume(fixtures.LoadedContent(test), nil), "Experience").Blocks
	got := make([]string, 0, len(experience))
	for _, block := range experience {
		got = append(got, fmt.Sprintf("%T", block))
	}
	if want := []string{"document.Paragraph", "document.Facts", "document.Entry", "document.Entry"}; !slices.Equal(got, want) {
		test.Fatalf("the experience section holds %q, want %q", got, want)
	}
	putForward := as[document.Facts](test, experience[1])
	want := document.Facts{Kind: document.FactsTechnologies, List: []document.Fact{
		{Label: "Go", Value: []document.Span{document.Text{Text: "3y"}}},
		{Label: "Rust", Value: []document.Span{document.Text{Text: "still learning"}}},
	}}
	if !reflect.DeepEqual(putForward, want) {
		test.Errorf("the skills the experience puts forward = %+v, want %+v", putForward, want)
	}
}

// A role's months reach the document typed, so each face writes them its own way, and a role still held stays
// current.
func TestEachRoleKeepsItsPeriod(test *testing.T) {
	_, sections := document.SectionsOf(Resume(fixtures.LoadedContent(test), nil).Blocks)
	got := []document.Period{}
	for _, role := range document.BriefOf(sections).Roles {
		got = append(got, role.Period)
	}
	want := []document.Period{
		{Start: document.KnownMonth{Start: firstMoment(2023, time.April)}, End: document.CurrentMonth{Text: "present"}},
		{Start: document.KnownMonth{Start: firstMoment(2020, time.January)}, End: document.KnownMonth{Start: firstMoment(2023, time.March)}},
	}
	if !reflect.DeepEqual(got, want) {
		test.Errorf("periods = %+v, want %+v", got, want)
	}
}

// A role's technologies are the skills whose sources name its company, in the profile's order of groups and skills,
// a skill still being learnt among them.
func TestARolesTechnologiesAreTheSkillsItsCompanyIsASourceOf(test *testing.T) {
	_, sections := document.SectionsOf(Resume(fixtures.LoadedContent(test), nil).Blocks)
	got := map[string][]string{}
	for _, role := range document.BriefOf(sections).Roles {
		got[document.PlainText(role.Title)] = role.Technologies
	}
	want := map[string][]string{
		"Northwind": {"Shell scripts for the build machines", "Docker"},
		"Southwind": {"Shell scripts for the build machines"},
	}
	if !reflect.DeepEqual(got, want) {
		test.Errorf("technologies by company = %q, want %q", got, want)
	}
}

// Every link of the résumé and of a project's page leads where the profile says: a face writes the address, but
// never chooses it.
func TestEveryLinkLeadsWhereTheProfileSays(test *testing.T) {
	documents := DocumentsOf(SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay()).Pages)
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
		{name: "a project's page to its code", got: project.Head.Project.Repository.Address, want: document.Address{URL: "https://github.test/mira/tidy"}},
		{name: "a project's page to its newest release", got: project.Head.Project.Release.Latest.Target, want: document.Address{URL: "https://github.test/mira/tidy/releases/latest"}},
		{name: "a project's page to its releases' history", got: project.Head.Project.Release.Changelog.Target, want: document.Address{URL: "https://github.test/mira/tidy/blob/HEAD/CHANGELOG.md"}},
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
	project := fixtures.DocumentOf(test, DocumentsOf(SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay()).Pages), document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""})
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
		{name: "its install command", want: "go install example.test/tidy@latest"},
		{name: "a later install's note", want: "On a machine with no Go"},
		{name: "a later install's command", want: "curl -fsSL https://example.test/tidy.sh | sh"},
		{name: "the link to its newest release", want: "latest"},
		{name: "the link to its releases' history", want: "changelog"},
		{name: "its licence", want: "MIT"},
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

// What tells about a project in its head: what it is made of, where its releases are read, its licence
type toldAbout struct {
	stack   []string
	release *document.Release
	licence string
}

// A project's head tells about the project as DESIGN.md's ProjectHead does: what it is made of, each name whole, its
// language first; its newest release and their history, each a link by the site's word; its licence. What the profile
// does not give is left out.
func TestAProjectsHeadTellsAboutTheProject(test *testing.T) {
	tidyTool := document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}
	release := &document.Release{
		Latest:    document.Link{Target: document.Address{URL: "https://github.test/mira/tidy/releases/latest"}, Spans: []document.Span{document.Text{Text: "latest"}}},
		Changelog: document.Link{Target: document.Address{URL: "https://github.test/mira/tidy/blob/HEAD/CHANGELOG.md"}, Spans: []document.Span{document.Text{Text: "changelog"}}},
	}
	cases := []struct {
		name     string
		snapshot *content.Snapshot
		want     toldAbout
	}{
		{name: "a project the profile tells all of", snapshot: fixtures.LoadedContent(test), want: toldAbout{stack: []string{"Lua", "Penlight", "LPeg"}, release: release, licence: "MIT"}},
		{name: "a project that names no language", snapshot: withoutLanguages(fixtures.LoadedContent(test)), want: toldAbout{stack: []string{"Penlight", "LPeg"}, release: release, licence: "MIT"}},
		{name: "a project the profile tells nothing of", snapshot: withoutAbout(fixtures.LoadedContent(test)), want: toldAbout{stack: nil, release: nil, licence: ""}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			project := fixtures.DocumentOf(test, DocumentsOf(SiteOf(Prepare(testCase.snapshot), publicOnTheDay()).Pages), tidyTool).Head.Project
			if got := (toldAbout{stack: project.Stack, release: project.Release, licence: project.Licence}); !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("what tidy-tool's head tells of it = %+v, want %+v", got, testCase.want)
			}
		})
	}
}

// A project's head says how to get the project: its install's command, then each other way with its note, which says
// when to take it instead, in the profile's order.
func TestAProjectsHeadSaysEachWayToInstallTheProject(test *testing.T) {
	project := fixtures.DocumentOf(test, DocumentsOf(SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay()).Pages), document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}).Head.Project
	const wantInstall = "go install example.test/tidy@latest"
	wantAlternatives := []document.Alternative{
		{Note: []document.Span{document.Text{Text: "On a machine with no Go"}}, Command: "curl -fsSL https://example.test/tidy.sh | sh"},
	}
	if project.Install != wantInstall || !reflect.DeepEqual(project.Alternatives, wantAlternatives) {
		test.Errorf("the ways to install tidy-tool = %q and %+v, want %q and %+v", project.Install, project.Alternatives, wantInstall, wantAlternatives)
	}
}

// Where a project's page leads its reader at its end: the other projects, the way to them all, the author's mail
type pageEnd struct {
	others      []document.OtherProject
	allProjects document.Link
	mail        document.Mail
}

// A project's page ends where its reader goes next (DESIGN.md, ProjectEnd): the other project with a page of its own,
// the project itself left out, by its name, its tagline, its stack and its page; the way to the résumé's projects tab;
// and the author's mail.
func TestAProjectsPageEndsWhereItsReaderGoesNext(test *testing.T) {
	pages := DocumentsOf(SiteOf(Prepare(fixtures.WithTwoProjectPages(test)), publicOnTheDay()).Pages)
	project := fixtures.DocumentOf(test, pages, document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}).Head.Project
	want := pageEnd{
		others: []document.OtherProject{{
			Name:    []document.Span{document.Text{Text: "tame-tool"}},
			Tagline: []document.Span{document.Text{Text: "Tames files."}},
			Stack:   []string{"Go"},
			Page:    document.Page{Kind: document.PageProject, Slug: "tame-tool", Tab: ""},
		}},
		allProjects: document.Link{Target: document.Page{Kind: document.PageResume, Slug: "", Tab: "projects"}, Spans: []document.Span{document.Text{Text: "all projects"}}},
		mail:        document.Mail{Address: "mira@example.test"},
	}
	if got := (pageEnd{others: project.Others, allProjects: project.AllProjects, mail: project.Mail}); !reflect.DeepEqual(got, want) {
		test.Errorf("tidy-tool's page ends with %+v, want %+v", got, want)
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
	for _, page := range DocumentsOf(SiteOf(Prepare(snapshot), publicOnTheDay()).Pages) {
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
// never a draft, a post unlisted or one that comes later than its reader.
func TestTheIndexListsTheBlogsListsThenEveryPostListedNewestFirst(test *testing.T) {
	posts := []writtenPost{
		{slug: "first", id: 3, status: "published", published: earlier, tags: ""},
		{slug: "toolbox", id: 4, status: "published", published: latest, tags: ""},
		{slug: "drafted", id: 5, status: "draft", published: later, tags: ""},
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
	documents := slices.Concat(DocumentsOf(site.Pages), tabDocumentsOf(snapshot), []document.Document{site.Index}, Extras(snapshot))
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
	documents := slices.Concat(DocumentsOf(SiteOf(Prepare(snapshot), publicOnTheDay()).Pages), tabDocumentsOf(snapshot))
	cases := []struct {
		name string
		page document.Page
		want string
	}{
		{name: "the résumé", page: document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, want: "Builds small tools that last."},
		{name: "a project's page", page: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}, want: "A tool that tidies files."},
		{name: "a tab of the résumé", page: document.Page{Kind: document.PageResume, Slug: "", Tab: "experience"}, want: "Builds small tools that last."},
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
	documents := DocumentsOf(SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay()).Pages)
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

// withoutAbout is a snapshot whose projects give nothing that tells about them: no language, stack, release or
// licence.
func withoutAbout(snapshot *content.Snapshot) *content.Snapshot {
	changed := *snapshot
	changed.Profile.Projects = slices.Clone(snapshot.Profile.Projects)
	for index := range changed.Profile.Projects {
		changed.Profile.Projects[index].Language = ""
		changed.Profile.Projects[index].Stack = nil
		changed.Profile.Projects[index].Release = nil
		changed.Profile.Projects[index].Licence = ""
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
