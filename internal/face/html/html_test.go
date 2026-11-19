package html

import (
	"html"
	"strings"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/golden"
	"github.com/vrppaul/paranin/internal/routes"
)

func TestMain(tests *testing.M) {
	golden.Main(tests)
}

// The markup of a page is what a browser, a screen reader and a search engine read: the language it names and the
// title it is called by, a heading's depth, a link's address, a month as a time, a result marked, a command as
// code, a fact as a term and its description, and the words between the parts of a line. Each page of the fixture
// is held to its golden file, named after the page's own address, with a line break between each two tags that
// meet, so the file reads in a diff.
func TestEveryPageReadsAsItsGoldenMarkup(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	cases := []struct {
		page document.Page
		file string
	}{
		{page: document.Page{Kind: document.PageResume, Slug: ""}, file: "index.html"},
		{page: document.Page{Kind: document.PageProject, Slug: "tidy-tool"}, file: "p/tidy-tool.html"},
	}
	for _, testCase := range cases {
		test.Run(testCase.file, func(test *testing.T) {
			composed := composedOf(test, snapshot, testCase.page)
			golden.Match(test, testCase.file, lined(fixtureWriter(test).Page(composed.Document, composed.Record)))
		})
	}
}

// Every phrase of a document as written, all but a fact or a point that is only a gap, reaches its page, as HTML
// writes text: the facts a browser shows are the facts the written document holds.
func TestEveryPhraseReachesThePage(test *testing.T) {
	for _, composed := range compose.Pages(fixtures.LoadedContent(test)) {
		page := fixtureWriter(test).Page(composed.Document, composed.Record)
		for _, phrase := range document.Phrases(document.Written(composed.Document)) {
			if !strings.Contains(page, html.EscapeString(phrase)) {
				test.Errorf("the page of %+v holds no %q", composed.Document.Head.Page, phrase)
			}
		}
	}
}

// A link whose target has no address, as a hidden page has none, reads as its words, never as a link to nowhere.
func TestALinkWithNoAddressReadsAsItsWords(test *testing.T) {
	nowhere := document.Link{Target: document.Page{Kind: document.PageHealth, Slug: ""}, Spans: []document.Span{document.Text{Text: "the report"}}}
	composed := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageNone, Slug: ""}, nil),
		Blocks: []document.Block{document.Paragraph{Spans: []document.Span{nowhere}}},
	}
	if page, want := fixtureWriter(test).Page(composed, blankPerson()), "<p>the report</p>"; !strings.Contains(page, want) {
		test.Errorf("the page %q holds no %q", page, want)
	}
}

// A page's title stands in block letters, a drawing a screen reader skips, the heading named by its words; the
// résumé's in the name's font with the cursor after it, another page's in the titles'; a title with a letter the font
// does not draw stays its words.
func TestATitleInBlockLettersIsNamedByItsWords(test *testing.T) {
	cases := []struct {
		name  string
		page  document.Page
		title string
		want  []string
		lacks []string
	}{
		{
			name: "the résumé's name, with the cursor", page: document.Page{Kind: document.PageResume, Slug: ""}, title: "Ada Lovelace",
			want:  []string{`<h1 aria-label="Ada Lovelace">`, `<svg class="letters letters-name" aria-hidden="true"`, `class="letters-cursor"`},
			lacks: nil,
		},
		{
			name: "a project's name, with no cursor", page: document.Page{Kind: document.PageProject, Slug: "tidy"}, title: "Tidy",
			want: []string{`<h1 aria-label="Tidy">`, `<svg class="letters letters-title" aria-hidden="true"`}, lacks: []string{`class="letters-cursor"`},
		},
		{
			name: "a letter no font draws", page: document.Page{Kind: document.PageProject, Slug: "cafe"}, title: "Café",
			want: []string{"<h1>Café</h1>"}, lacks: []string{`class="letters`},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			composed := document.Document{
				Head:   fixtures.HeadOf(testCase.page, []document.Span{document.Text{Text: testCase.title}}),
				Blocks: nil,
			}
			page := lined(fixtureWriter(test).Page(composed, blankPerson()))
			for _, want := range testCase.want {
				if !strings.Contains(page, want) {
					test.Errorf("the page %q holds no %q", page, want)
				}
			}
			for _, unwanted := range testCase.lacks {
				if strings.Contains(page, unwanted) {
					test.Errorf("the page %q holds %q", page, unwanted)
				}
			}
		})
	}
}

// A section of the résumé stands under its title in block letters, named by its words; a section of another page
// under its title as words.
func TestTheResumesSectionsStandUnderBlockLetters(test *testing.T) {
	section := document.Section{Title: []document.Span{document.Text{Text: "About"}}, Blocks: nil}
	cases := []struct {
		name string
		page document.Page
		want string
	}{
		{name: "the résumé's", page: document.Page{Kind: document.PageResume, Slug: ""}, want: `<h2 aria-label="About">`},
		{name: "a project's", page: document.Page{Kind: document.PageProject, Slug: "tidy"}, want: "<h2>About</h2>"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			composed := document.Document{Head: fixtures.HeadOf(testCase.page, nil), Blocks: []document.Block{section}}
			if page := lined(fixtureWriter(test).Page(composed, blankPerson())); !strings.Contains(page, testCase.want) {
				test.Errorf("the page %q holds no %q", page, testCase.want)
			}
		})
	}
}

// Every page opens with the top bar, the page of no page among them: a link past it to the page's main part, the
// site's name leading to where it begins, each of the résumé's sections leading to its place on the résumé, and the
// theme button. The résumé's sections stand at those places.
func TestEveryPageOpensWithTheTopBar(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	writer := fixtureWriter(test)
	resume := composedOf(test, snapshot, document.Page{Kind: document.PageResume, Slug: ""})
	project := composedOf(test, snapshot, document.Page{Kind: document.PageProject, Slug: "tidy-tool"})
	bar := []string{
		`href="#main">Skip to content</a>`, `<main id="main"`, `href="/">example.test</a>`,
		`<a href="/#about">About</a>`, `<a href="/#experience">Experience</a>`, `<a href="/#projects">Projects</a>`,
		`<a href="/#skills">Skills</a>`, `<a href="/#contact">Contact</a>`, `data-part="theme"`,
	}
	missing := NewNotFound(fixtureSite(snapshot), "404 Not Found", "en", fixtures.LoadedBuild(test), fixtureFrame(snapshot))
	cases := []struct {
		name string
		page string
		want []string
	}{
		{name: "the résumé", page: writer.Page(resume.Document, resume.Record), want: bar},
		{name: "a project's page", page: writer.Page(project.Document, project.Record), want: bar},
		{name: "the page of no page", page: missing.For("/nothing"), want: bar},
		{
			name: "the résumé's sections, where the bar leads", page: writer.Page(resume.Document, resume.Record),
			want: []string{`<section id="about"`, `<section id="experience"`, `<section id="projects"`, `<section id="skills"`, `<section id="contact"`},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			for _, want := range testCase.want {
				if !strings.Contains(testCase.page, want) {
					test.Errorf("the page %q holds no %q", testCase.page, want)
				}
			}
		})
	}
}

// Every page ends with the footer, the page of no page among them: the ways the same résumé is served, the command
// that reaches it over ssh, the one that reads the page in a terminal, the page's markdown where it has one, and the
// index agents read.
func TestEveryPageEndsWithTheFooter(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	writer := fixtureWriter(test)
	resume := composedOf(test, snapshot, document.Page{Kind: document.PageResume, Slug: ""})
	project := composedOf(test, snapshot, document.Page{Kind: document.PageProject, Slug: "tidy-tool"})
	missing := NewNotFound(fixtureSite(snapshot), "404 Not Found", "en", fixtures.LoadedBuild(test), fixtureFrame(snapshot))
	cases := []struct {
		name  string
		page  string
		want  []string
		lacks []string
	}{
		{
			name: "the résumé", page: writer.Page(resume.Document, resume.Record),
			want: []string{
				"<footer", "<code>ssh example.test</code>", "<code>curl https://example.test/</code>",
				`href="/index.md">/index.md</a>`, `href="/llms.txt">/llms.txt</a>`,
			},
			lacks: nil,
		},
		{
			name: "a project's page", page: writer.Page(project.Document, project.Record),
			want: []string{
				"<code>ssh example.test</code>", "<code>curl https://example.test/p/tidy-tool</code>",
				`href="/p/tidy-tool.md">/p/tidy-tool.md</a>`, `href="/llms.txt">/llms.txt</a>`,
			},
			lacks: nil,
		},
		{
			name: "the page of no page, which has no markdown", page: missing.For("/nothing"),
			want:  []string{"<code>ssh example.test</code>", "<code>curl https://example.test/</code>", `href="/llms.txt">/llms.txt</a>`},
			lacks: []string{".md"},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			for _, want := range testCase.want {
				if !strings.Contains(testCase.page, want) {
					test.Errorf("the page %q holds no %q", testCase.page, want)
				}
			}
			for _, unwanted := range testCase.lacks {
				if strings.Contains(testCase.page, unwanted) {
					test.Errorf("the page %q holds %q", testCase.page, unwanted)
				}
			}
		})
	}
}

// The résumé's first screen briefs the reader with the first four facts its first section holds, in their order; the
// facts after them stay in that section, at the place the top bar leads to, where the facts stood.
func TestTheFirstScreenBriefsTheFirstFourFacts(test *testing.T) {
	facts := []document.Fact{}
	for _, label := range []string{"first", "second", "third", "fourth", "fifth"} {
		facts = append(facts, document.Fact{Label: label, Value: []document.Span{document.Text{Text: label + " value"}}})
	}
	about := document.Section{
		Title: []document.Span{document.Text{Text: "About"}},
		Blocks: []document.Block{
			document.Paragraph{Spans: []document.Span{document.Text{Text: "before"}}},
			document.Facts{Kind: document.FactsDetails, List: facts},
			document.Paragraph{Spans: []document.Span{document.Text{Text: "after"}}},
		},
	}
	firstScreen, aboutSection, _ := strings.Cut(fixtureWriter(test).Page(resumeOf(about), blankPerson()), `<section id="about"`)
	cases := []struct {
		name  string
		part  string
		holds string
		lacks string
	}{
		{
			name: "the first screen", part: firstScreen,
			holds: "<dt>first</dt><dd>first value</dd><dt>second</dt><dd>second value</dd><dt>third</dt><dd>third value</dd><dt>fourth</dt><dd>fourth value</dd></dl>",
			lacks: "<dt>fifth</dt>",
		},
		{
			name: "the section", part: aboutSection,
			holds: `<p>before</p><dl class="facts"><dt>fifth</dt><dd>fifth value</dd></dl><p>after</p>`,
			lacks: "<dt>first</dt>",
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(testCase.part, testCase.holds) {
				test.Errorf("%s %q holds no %q", testCase.name, testCase.part, testCase.holds)
			}
			if strings.Contains(testCase.part, testCase.lacks) {
				test.Errorf("%s %q holds %q", testCase.name, testCase.part, testCase.lacks)
			}
		})
	}
}

// The brief is headed by the first way to reach the author: a link that writes to its mailbox, parted at its at into
// the mailbox and the host; an address that is no such link reads as its words.
func TestTheBriefIsHeadedByTheMail(test *testing.T) {
	cases := []struct {
		name    string
		address []document.Span
		want    string
	}{
		{
			name:    "a link to a mailbox",
			address: []document.Span{document.Link{Target: document.Mail{Address: "mira@example.test"}, Spans: []document.Span{document.Text{Text: "mira@example.test"}}}},
			want:    `href="mailto:mira@example.test"><span class="mail-box">mira</span><span class="mail-at">@</span><span class="mail-host">example.test</span></a>`,
		},
		{
			name:    "words with no link",
			address: []document.Span{document.Text{Text: "write to mira"}},
			want:    "<p>write to mira</p>",
		},
		{
			name:    "a link to no mailbox",
			address: []document.Span{document.Link{Target: document.Address{URL: "https://example.test/write"}, Spans: []document.Span{document.Text{Text: "the form"}}}},
			want:    `<p><a href="https://example.test/write">the form</a></p>`,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			about := document.Section{
				Title:  []document.Span{document.Text{Text: "About"}},
				Blocks: []document.Block{document.Facts{Kind: document.FactsDetails, List: []document.Fact{{Label: "now", Value: []document.Span{document.Text{Text: "here"}}}}}},
			}
			contact := document.Section{
				Title:  []document.Span{document.Text{Text: "Contact"}},
				Blocks: []document.Block{document.Facts{Kind: document.FactsWays, List: []document.Fact{{Label: "mail", Value: testCase.address}}}},
			}
			firstScreen, _, _ := strings.Cut(fixtureWriter(test).Page(resumeOf(about, contact), blankPerson()), `<section id="about"`)
			if !strings.Contains(firstScreen, testCase.want) {
				test.Errorf("the first screen %q holds no %q", firstScreen, testCase.want)
			}
		})
	}
}

// The first screen lays the résumé's roles along the career line, in the order they started, and lists them for a
// phone with their years: a role with a month still to fill is left out, the one still held is marked and runs to the
// site's word for the present, and the line's ends are its first year and that word.
func TestTheCareerLineHoldsTheRolesOfKnownMonths(test *testing.T) {
	about := document.Section{
		Title:  []document.Span{document.Text{Text: "About"}},
		Blocks: []document.Block{document.Facts{Kind: document.FactsDetails, List: []document.Fact{{Label: "now", Value: []document.Span{document.Text{Text: "here"}}}}}},
	}
	work := document.Section{Title: []document.Span{document.Text{Text: "Work"}}, Blocks: []document.Block{
		roleOf("Northwind", monthOf(2023, time.April), document.CurrentMonth{Text: "present"}),
		roleOf("Gapwind", document.GapMonth{Text: "[2019]"}, monthOf(2019, time.December)),
		roleOf("Southwind", monthOf(2020, time.January), monthOf(2022, time.December)),
	}}
	firstScreen, _, _ := strings.Cut(fixtureWriter(test).Page(resumeOf(about, work), blankPerson()), `<section id="about"`)
	cases := []struct {
		name  string
		holds string
	}{
		{name: "a role that ended, to the year of its last month", holds: `<li><span class="career-years">2020 – 2022</span><span class="career-company">Southwind</span>`},
		{
			name:  "the one still held, marked, to the word for the present",
			holds: `<li class="career-now"><span class="career-years">2023 – present</span><span class="career-company">Northwind</span>`,
		},
		{name: "the line's ends", holds: `<span>2020</span><span>present</span>`},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(firstScreen, testCase.holds) {
				test.Errorf("the first screen %q holds no %q", firstScreen, testCase.holds)
			}
		})
	}
	test.Run("the roles in the order they started", func(test *testing.T) {
		if earlier, later := strings.Index(firstScreen, "Southwind"), strings.Index(firstScreen, "Northwind"); earlier > later {
			test.Errorf("the first screen %q names Northwind at %d, before Southwind at %d", firstScreen, later, earlier)
		}
	})
	test.Run("a role with a month still to fill left out", func(test *testing.T) {
		if strings.Contains(firstScreen, "Gapwind") {
			test.Errorf("the first screen %q holds %q", firstScreen, "Gapwind")
		}
	})
}

// The page shows a document as a reader of the finished work sees it (document.Written): a fact or a point that is
// only a gap still to fill is left out, so the first screen briefs the reader with the first four facts written; a gap
// among written words stays, marked as a gap.
func TestAGapAloneIsLeftOutOfThePage(test *testing.T) {
	about := document.Section{
		Title: []document.Span{document.Text{Text: "About"}},
		Blocks: []document.Block{document.Facts{Kind: document.FactsDetails, List: []document.Fact{
			{Label: "first", Value: []document.Span{document.Text{Text: "first value"}}},
			{Label: "later", Value: []document.Span{document.Gap{Text: "[later]"}}},
			{Label: "where", Value: []document.Span{document.Text{Text: "Prague · "}, document.Gap{Text: "[remote?]"}}},
			{Label: "third", Value: []document.Span{document.Text{Text: "third value"}}},
			{Label: "fourth", Value: []document.Span{document.Text{Text: "fourth value"}}},
			{Label: "fifth", Value: []document.Span{document.Text{Text: "fifth value"}}},
		}}},
	}
	role := roleOf("Northwind", monthOf(2023, time.April), document.CurrentMonth{Text: "present"})
	role.Points = []document.Point{
		{Spans: []document.Span{document.Gap{Text: "[What was built]"}}},
		{Spans: []document.Span{document.Text{Text: "Owned "}, document.Gap{Text: "[service]"}}},
	}
	work := document.Section{Title: []document.Span{document.Text{Text: "Work"}}, Blocks: []document.Block{role}}
	page := fixtureWriter(test).Page(resumeOf(about, work), blankPerson())
	firstScreen, sections, _ := strings.Cut(page, `<section id="about"`)
	cases := []struct {
		name  string
		part  string
		holds string
		lacks string
	}{
		{
			name: "the brief, of the first four facts written", part: firstScreen,
			holds: `<dt>first</dt><dd>first value</dd><dt>where</dt><dd>Prague · <span class="gap">[remote?]</span></dd>` +
				"<dt>third</dt><dd>third value</dd><dt>fourth</dt><dd>fourth value</dd></dl>",
			lacks: "[later]",
		},
		{
			name: "a point with a gap among its words, and none that is only one", part: sections,
			holds: `Owned <span class="gap">[service]</span></li>`, lacks: "[What was built]",
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(testCase.part, testCase.holds) {
				test.Errorf("%s %q holds no %q", testCase.name, testCase.part, testCase.holds)
			}
			if strings.Contains(testCase.part, testCase.lacks) {
				test.Errorf("%s %q holds %q", testCase.name, testCase.part, testCase.lacks)
			}
		})
	}
}

// A fact whose value is a command alone stands as a chip that copies it, named by the command and saying in the toast
// what it copied, on any page that holds one; code among a fact's words stays code.
func TestACommandAloneStandsAsAChipThatCopiesTheCommand(test *testing.T) {
	tool := document.Card{
		Title: []document.Span{document.Text{Text: "tidy"}}, Subtitle: nil, Tag: "", Summary: nil, Page: document.Page{Kind: document.PageNone, Slug: ""},
		Facts: []document.Fact{
			{Label: "install", Value: []document.Span{document.Code{Text: "go install example.test/tidy@latest"}}},
			{Label: "use", Value: []document.Span{document.Code{Text: "tidy ."}, document.Text{Text: " daily"}}},
		},
	}
	composed := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageProject, Slug: "tidy"}, []document.Span{document.Text{Text: "tidy"}}),
		Blocks: []document.Block{tool},
	}
	page := fixtureWriter(test).Page(composed, blankPerson())
	for _, want := range []string{
		`<dt>install</dt><dd><div class="command" data-part="command">`,
		`<code data-part="command-words">go install example.test/tidy@latest</code>`,
		`aria-label="copy go install example.test/tidy@latest" data-copied="copied" data-toast="Copied go install example.test/tidy@latest"`,
		`<dt>use</dt><dd><code>tidy .</code> daily</dd>`,
		`data-part="toast"`,
	} {
		if !strings.Contains(page, want) {
			test.Errorf("the page %q holds no %q", page, want)
		}
	}
}

// The brief ends with the ways to the author's profiles, each a link by its label, a dot between two; the mail that
// heads the brief, a link with words after it and a command to run are none of them.
func TestTheBriefEndsWithTheProfiles(test *testing.T) {
	about := document.Section{
		Title:  []document.Span{document.Text{Text: "About"}},
		Blocks: []document.Block{document.Facts{Kind: document.FactsDetails, List: []document.Fact{{Label: "now", Value: []document.Span{document.Text{Text: "here"}}}}}},
	}
	contact := document.Section{Title: []document.Span{document.Text{Text: "Contact"}}, Blocks: []document.Block{document.Facts{Kind: document.FactsWays, List: []document.Fact{
		{Label: "mail", Value: []document.Span{document.Link{Target: document.Mail{Address: "mira@example.test"}, Spans: []document.Span{document.Text{Text: "mira@example.test"}}}}},
		{Label: "GitHub", Value: []document.Span{document.Link{Target: document.Address{URL: "https://github.test/mira"}, Spans: []document.Span{document.Text{Text: "github.test/mira"}}}}},
		{Label: "LinkedIn", Value: []document.Span{document.Link{Target: document.Address{URL: "https://linkedin.test/mira"}, Spans: []document.Span{document.Text{Text: "linkedin.test/mira"}}}}},
		{Label: "blog", Value: []document.Span{document.Link{Target: document.Address{URL: "https://blog.test"}, Spans: []document.Span{document.Text{Text: "blog.test"}}}, document.Text{Text: ", in Czech"}}},
		{Label: "ssh", Value: []document.Span{document.Code{Text: "ssh example.test"}}},
	}}}}
	firstScreen, _, _ := strings.Cut(fixtureWriter(test).Page(resumeOf(about, contact), blankPerson()), `<section id="about"`)
	want := `<p class="profiles"><a href="https://github.test/mira">GitHub</a>` + dot + `<a href="https://linkedin.test/mira">LinkedIn</a></p>`
	if !strings.Contains(firstScreen, want) {
		test.Errorf("the first screen %q holds no %q", firstScreen, want)
	}
}

// roleOf is a role a test builds by hand: its company and its months, nothing else.
func roleOf(company string, start, end document.Month) document.Entry {
	return document.Entry{
		Kind: document.EntryRole, Period: document.Period{Start: start, End: end}, Title: []document.Span{document.Text{Text: company}},
		Subtitle: nil, Place: "", Points: nil,
	}
}

// monthOf is a month of a year as a role's period holds it.
func monthOf(year int, month time.Month) document.KnownMonth {
	return document.KnownMonth{Start: time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)}
}

// resumeOf is a résumé a test builds by hand: its sections under a head with nothing but the résumé's page.
func resumeOf(sections ...document.Section) document.Document {
	blocks := []document.Block{}
	for _, section := range sections {
		blocks = append(blocks, section)
	}
	return document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: ""}, nil), Blocks: blocks}
}

// A page whose content gives no description names none, so a search engine or a preview writes its own rather
// than an empty one.
func TestAPageWithNoDescriptionNamesNone(test *testing.T) {
	composed := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: ""}, nil),
		Blocks: nil,
	}
	page := fixtureWriter(test).Page(composed, blankPerson())
	for _, unwanted := range []string{`name="description"`, `property="og:description"`} {
		if strings.Contains(page, unwanted) {
			test.Errorf("the page %q holds %q", page, unwanted)
		}
	}
}

// The page of an address the site has no page at opens its head as every page does, names its language, and its
// status in its title and on the page; quotes the path asked for as Go quotes a string, so no control character
// in it reaches the page; and leads to where the site begins, by its path, under the site's name. It is held to
// its golden file whole.
func TestAMissingPageReadsAsItsGoldenMarkup(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	missing := NewNotFound(fixtureSite(snapshot), "404 Not Found", "en", fixtures.LoadedBuild(test), fixtureFrame(snapshot)).For("/nothing\x1b[2J<b>")
	golden.Match(test, "notfound.html", lined(missing))
}

// A person who works nowhere now is written with no employer, rather than with an employer of no name.
func TestAPersonWorkingNowhereNamesNoEmployer(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	resume := document.Page{Kind: document.PageResume, Slug: ""}
	record := blankPerson()
	record.Name = "Mira Example"
	if page := fixtureWriter(test).Page(composedOf(test, snapshot, resume).Document, record); strings.Contains(page, "worksFor") {
		test.Errorf("the page %q names an employer", page)
	}
}

// lined is a page with a line break between each two tags that meet, so its golden file reads in a diff; a space
// or a word between two tags stays where it is.
func lined(page string) string {
	return strings.ReplaceAll(page, "><", ">\n<") + "\n"
}

// fixtureWriter writes the pages of the fixture's site, with the build's fixture and the fixture's words.
func fixtureWriter(test *testing.T) PageWriter {
	snapshot := fixtures.LoadedContent(test)
	return NewPageWriter(fixtureSite(snapshot), fixtures.LoadedBuild(test), fixtureFrame(snapshot), fixtures.Day())
}

// fixtureFrame is the frame of the fixture's pages, in its words.
func fixtureFrame(snapshot *content.Snapshot) compose.Frame {
	return compose.FrameOf(snapshot, compose.Pages(snapshot))
}

// fixtureSite is the site of the fixture, at the address its profile gives.
func fixtureSite(snapshot *content.Snapshot) routes.Site {
	return routes.New(snapshot.Profile.Site.BaseURL)
}

// composedOf is a page of the fixture as compose makes it, its document and its record.
func composedOf(test *testing.T, snapshot *content.Snapshot, page document.Page) compose.Composed {
	for _, composed := range compose.Pages(snapshot) {
		if composed.Document.Head.Page == page {
			return composed
		}
	}
	test.Fatalf("compose made no page %+v", page)
	return compose.Composed{Document: document.Document{Head: fixtures.HeadOf(page, nil), Blocks: nil}, Record: nil}
}

// blankPerson is a record of a person with no facts, for a page a test builds by hand: every page a person reads
// has a record, and a page standing for the résumé has a person's.
func blankPerson() compose.PersonRecord {
	return compose.PersonRecord{
		Name: "", AlternateNames: nil, Title: "", Email: "", Employer: "", Schools: nil, Skills: nil, Languages: nil,
		Profiles: nil,
	}
}
