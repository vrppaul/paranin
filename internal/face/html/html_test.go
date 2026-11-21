package html

import (
	"html"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/golden"
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
		{page: document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, file: "index.html"},
		{page: document.Page{Kind: document.PageResume, Slug: "", Tab: "experience"}, file: "experience.html"},
		{page: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}, file: "p/tidy-tool.html"},
		{page: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: "why"}, file: "p/tidy-tool/why.html"},
	}
	for _, testCase := range cases {
		test.Run(testCase.file, func(test *testing.T) {
			whole := document.Page{Kind: testCase.page.Kind, Slug: testCase.page.Slug, Tab: ""}
			composed := composedOf(test, snapshot, whole)
			golden.Match(test, testCase.file, lined(fixtureWriter(test).Page(composed.Document, composed.Record, testCase.page.Tab)))
		})
	}
}

// Every phrase of a document as written, all but a fact or a point that is only a gap, reaches one of its pages, its
// opening's or a tab's, as HTML writes text: the facts a browser shows are the facts the written document holds.
func TestEveryPhraseReachesThePage(test *testing.T) {
	for _, composed := range compose.Pages(fixtures.LoadedContent(test)) {
		pages := fixtureWriter(test).Page(composed.Document, composed.Record, "")
		for _, tab := range document.TabsOf(composed.Document).Rest {
			pages += fixtureWriter(test).Page(composed.Document, composed.Record, tab.Slug)
		}
		written := document.Document{Head: composed.Document.Head, Blocks: document.WrittenBlocks(composed.Document.Blocks)}
		for _, phrase := range document.Phrases(written) {
			if !strings.Contains(pages, html.EscapeString(phrase)) {
				test.Errorf("the pages of %+v hold no %q", composed.Document.Head.Page, phrase)
			}
		}
	}
}

// A page shows one tab of its document, as the TUI does: at its own address its opening, the résumé's first screen and
// first section, another page's head and the blocks before its first section; at a tab's address that tab alone,
// under its title, which titles the page, the address it is known by and its markdown the tab's own.
func TestAPageShowsOneTabOfItsDocument(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	resume := composedOf(test, snapshot, document.Page{Kind: document.PageResume, Slug: "", Tab: ""})
	project := composedOf(test, snapshot, document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""})
	writer := fixtureWriter(test)
	const (
		// a line of the résumé's first screen, of its first section, and of its experience
		greeting   = "Hello, my name is"
		summary    = "Writes software for people who read it."
		experience = "Northwind — Senior Engineer"
		// a line of a project's details, and of its section Why
		details = "go install example.test/tidy@latest"
		why     = "Files were untidy."
	)
	cases := []struct {
		name  string
		page  string
		holds []string
		lacks []string
	}{
		{
			name: "the résumé's opening", page: writer.Page(resume.Document, resume.Record, ""),
			holds: []string{greeting, summary, `<link rel="canonical" href="https://example.test/">`}, lacks: []string{experience},
		},
		{
			name: "a tab of the résumé", page: writer.Page(resume.Document, resume.Record, "experience"),
			holds: []string{
				experience, `<title>Experience</title>`, `<h1 aria-label="Experience">`,
				`<link rel="canonical" href="https://example.test/experience">`, `type="text/markdown" href="/experience.md"`,
			},
			lacks: []string{greeting, summary},
		},
		{
			name: "a project's opening", page: writer.Page(project.Document, project.Record, ""),
			holds: []string{details, `<h1 aria-label="tidy-tool">`}, lacks: []string{why},
		},
		{
			name: "a tab of a project's page", page: writer.Page(project.Document, project.Record, "why"),
			holds: []string{why, `<h1 aria-label="Why">`, `<link rel="canonical" href="https://example.test/p/tidy-tool/why">`},
			lacks: []string{details},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			for _, want := range testCase.holds {
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

// A link whose target has no address, as a hidden page has none, reads as its words, never as a link to nowhere.
func TestALinkWithNoAddressReadsAsItsWords(test *testing.T) {
	nowhere := document.Link{Target: document.Page{Kind: document.PageHealth, Slug: "", Tab: ""}, Spans: []document.Span{document.Text{Text: "the report"}}}
	composed := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageNone, Slug: "", Tab: ""}, nil),
		Blocks: []document.Block{document.Paragraph{Spans: []document.Span{nowhere}}},
	}
	if page, want := fixtureWriter(test).Page(composed, blankPerson(), ""), "<p>the report</p>"; !strings.Contains(page, want) {
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
			name: "the résumé's name, with the cursor", page: document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, title: "Ada Lovelace",
			want:  []string{`<h1 aria-label="Ada Lovelace">`, `<svg class="letters letters-name" aria-hidden="true"`, `class="letters-cursor"`},
			lacks: nil,
		},
		{
			name: "a project's name, with no cursor", page: document.Page{Kind: document.PageProject, Slug: "tidy", Tab: ""}, title: "Tidy",
			want: []string{`<h1 aria-label="Tidy">`, `<svg class="letters letters-title" aria-hidden="true"`}, lacks: []string{`class="letters-cursor"`},
		},
		{
			name: "a letter no font draws", page: document.Page{Kind: document.PageProject, Slug: "cafe", Tab: ""}, title: "Café",
			want: []string{"<h1>Café</h1>"}, lacks: []string{`class="letters`},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			composed := document.Document{
				Head:   fixtures.HeadOf(testCase.page, []document.Span{document.Text{Text: testCase.title}}),
				Blocks: nil,
			}
			page := lined(fixtureWriter(test).Page(composed, blankPerson(), ""))
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

// A tab shown alone stands under its title in block letters, the page's h1, named by its words, on the résumé as on
// any other page; a title with a letter the font does not draw stays its words.
func TestATabStandsUnderItsTitleInBlockLetters(test *testing.T) {
	cases := []struct {
		name  string
		page  document.Page
		title string
		want  string
	}{
		{name: "the résumé's", page: document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, title: "Work", want: `<h1 aria-label="Work">`},
		{name: "a project's", page: document.Page{Kind: document.PageProject, Slug: "tidy", Tab: ""}, title: "Usage", want: `<h1 aria-label="Usage">`},
		{name: "a letter no font draws", page: document.Page{Kind: document.PageProject, Slug: "tidy", Tab: ""}, title: "Café", want: "<h1>Café</h1>"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			section := document.Section{Title: []document.Span{document.Text{Text: testCase.title}}, Blocks: nil}
			composed := document.Document{Head: fixtures.HeadOf(testCase.page, nil), Blocks: []document.Block{section}}
			page := lined(fixtureWriter(test).Page(composed, blankPerson(), document.SlugOf(testCase.title)))
			if !strings.Contains(page, testCase.want) {
				test.Errorf("the page %q holds no %q", page, testCase.want)
			}
		})
	}
}

// Every page ends with the foot line, the page of no page among them: the session, the site's name leading to where
// it begins; the document's tabs as windows by their numbers and names, each leading to its page, the one shown the
// current page; the programs that reach the site and read it in a terminal; the markdown of the page whole,
// the résumé's where a page has none; the switch that turns the vim keys on, and the theme button. Over it stands
// the command line those keys open, with the places it goes to by name: the document's tabs, the one shown among
// them, then the page of each project that has one, by its slug, shown where its opening is.
func TestEveryPageEndsWithTheFootLine(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	writer := fixtureWriter(test)
	resume := composedOf(test, snapshot, document.Page{Kind: document.PageResume, Slug: "", Tab: ""})
	project := composedOf(test, snapshot, document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""})
	missing := NewNotFound(fixtures.Site(snapshot), "404 Not Found", "en", fixtures.LoadedBuild(test), compose.FrameOf(snapshot))
	every := []string{
		`<a class="session" href="/"><span aria-hidden="true">[</span>example.test<span aria-hidden="true">]</span></a>`,
		`<span class="way-mark" aria-hidden="true">❯</span>ssh</button>`, `<span class="way-mark" aria-hidden="true">❯</span>curl</button>`,
		`data-part="vim-keys"`, `data-part="theme"`, `data-part="command-line"`, `role="combobox"`,
	}
	// the place of the fixture's one project with a page; another project has none, so no place
	const (
		projectPlace = `data-place="project" data-name="tidy-tool" data-address="/p/tidy-tool"`
		noPagePlace  = `data-name="other-tool"`
	)
	// the résumé's markdown, which the page of no page offers too
	const resumeMarkdown = `<a class="way-chip" href="/index.md">index.md</a>`
	cases := []struct {
		name  string
		page  string
		holds []string
		lacks []string
	}{
		{
			name: "the résumé's opening", page: writer.Page(resume.Document, resume.Record, ""),
			holds: slices.Concat(every, []string{
				resumeMarkdown, `<a class="window" href="/" aria-current="page"><span class="window-number">1:</span>about</a>`,
				`<a class="window" href="/experience"><span class="window-number">2:</span>experience</a>`,
				`<a class="window" href="/contact"><span class="window-number">5:</span>contact</a>`,
				`data-place="tab" data-name="about" data-address="/" data-current`,
				`data-place="tab" data-name="contact" data-address="/contact"`, projectPlace,
			}),
			lacks: []string{noPagePlace},
		},
		{
			name: "a tab of the résumé", page: writer.Page(resume.Document, resume.Record, "experience"),
			holds: slices.Concat(every, []string{
				resumeMarkdown, `<a class="window" href="/"><span class="window-number">1:</span>about</a>`,
				`<a class="window" href="/experience" aria-current="page"><span class="window-number">2:</span>experience</a>`,
			}),
			lacks: nil,
		},
		{
			name: "a project's page", page: writer.Page(project.Document, project.Record, ""),
			holds: slices.Concat(every, []string{
				`<a class="way-chip" href="/p/tidy-tool.md">tidy-tool.md</a>`,
				`<a class="window" href="/p/tidy-tool" aria-current="page"><span class="window-number">1:</span>tidy-tool</a>`,
				`<a class="window" href="/p/tidy-tool/why"><span class="window-number">2:</span>why</a>`,
				`data-place="tab" data-name="tidy-tool" data-address="/p/tidy-tool" data-current`,
				`data-place="tab" data-name="why" data-address="/p/tidy-tool/why"`, projectPlace + ` data-current`,
			}),
			lacks: nil,
		},
		{
			name: "a tab of a project's page", page: writer.Page(project.Document, project.Record, "why"),
			holds: slices.Concat(every, []string{projectPlace + ` hidden`}), lacks: []string{projectPlace + ` data-current`},
		},
		{
			name: "the page of no page", page: missing.For("/nothing"),
			holds: slices.Concat(every, []string{resumeMarkdown, projectPlace}), lacks: []string{`class="window"`, `data-place="tab"`},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			for _, want := range testCase.holds {
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
// facts after them stay in that section, under the first screen, where the facts stood.
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
	firstScreen, aboutSection, _ := strings.Cut(fixtureWriter(test).Page(resumeOf(about), blankPerson(), ""), `<section class="section"`)
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
			firstScreen, _, _ := strings.Cut(fixtureWriter(test).Page(resumeOf(about, contact), blankPerson(), ""), `<section class="section"`)
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
	firstScreen, _, _ := strings.Cut(fixtureWriter(test).Page(resumeOf(about, work), blankPerson(), ""), `<section class="section"`)
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

// The page shows a document as a reader of the finished work sees it (document.WrittenBlocks): a fact or a point that
// is only a gap still to fill is left out, on the first screen, which briefs the reader with the first four facts
// written, in a tab, and in another page's opening; a gap among written words stays, marked as a gap.
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
	firstScreen, _, _ := strings.Cut(fixtureWriter(test).Page(resumeOf(about, work), blankPerson(), ""), `<section class="section"`)
	workTab := fixtureWriter(test).Page(resumeOf(about, work), blankPerson(), "work")
	project := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageProject, Slug: "tidy", Tab: ""}, nil),
		Blocks: []document.Block{about.Blocks[0]},
	}
	projectOpening := fixtureWriter(test).Page(project, blankPerson(), "")
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
			name: "a point with a gap among its words, and none that is only one", part: workTab,
			holds: `Owned <span class="gap">[service]</span></li>`, lacks: "[What was built]",
		},
		{name: "another page's opening, its facts written", part: projectOpening, holds: "<dt>first</dt><dd>first value</dd>", lacks: "[later]"},
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

// A résumé whose brief would hold nothing but gaps keeps its tabs as every face parts them: its first section stays
// the opening, its words written at the page's own address, and the foot line names the tabs the site serves.
func TestAnOpeningOfGapsStaysTheOpening(test *testing.T) {
	about := document.Section{
		Title: []document.Span{document.Text{Text: "About"}},
		Blocks: []document.Block{
			document.Paragraph{Spans: []document.Span{document.Text{Text: "Writes software."}}},
			document.Facts{Kind: document.FactsDetails, List: []document.Fact{{Label: "later", Value: []document.Span{document.Gap{Text: "[later]"}}}}},
		},
	}
	work := document.Section{Title: []document.Span{document.Text{Text: "Work"}}, Blocks: nil}
	page := fixtureWriter(test).Page(resumeOf(about, work), blankPerson(), "")
	for _, want := range []string{
		"<p>Writes software.</p>",
		`<a class="window" href="/" aria-current="page"><span class="window-number">1:</span>about`,
		`<a class="window" href="/work"><span class="window-number">2:</span>work</a>`,
	} {
		if !strings.Contains(page, want) {
			test.Errorf("the page %q holds no %q", page, want)
		}
	}
	if strings.Contains(page, "[later]") {
		test.Errorf("the page %q holds the gap %q", page, "[later]")
	}
}

// A fact whose value is a command alone stands as a chip that copies it, named by the command and saying in the toast
// what it copied, on any page that holds one; code among a fact's words stays code.
func TestACommandAloneStandsAsAChipThatCopiesTheCommand(test *testing.T) {
	tool := document.Card{
		Title: []document.Span{document.Text{Text: "tidy"}}, Subtitle: nil, Tag: "", Summary: nil, Page: document.Page{Kind: document.PageNone, Slug: "", Tab: ""},
		Facts: []document.Fact{
			{Label: "install", Value: []document.Span{document.Code{Text: "go install example.test/tidy@latest"}}},
			{Label: "use", Value: []document.Span{document.Code{Text: "tidy ."}, document.Text{Text: " daily"}}},
		},
	}
	composed := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageProject, Slug: "tidy", Tab: ""}, []document.Span{document.Text{Text: "tidy"}}),
		Blocks: []document.Block{tool},
	}
	page := fixtureWriter(test).Page(composed, blankPerson(), "")
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
	firstScreen, _, _ := strings.Cut(fixtureWriter(test).Page(resumeOf(about, contact), blankPerson(), ""), `<section class="section"`)
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
	return document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, nil), Blocks: blocks}
}

// A page whose content gives no description names none, so a search engine or a preview writes its own rather
// than an empty one.
func TestAPageWithNoDescriptionNamesNone(test *testing.T) {
	composed := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, nil),
		Blocks: nil,
	}
	page := fixtureWriter(test).Page(composed, blankPerson(), "")
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
	missing := NewNotFound(fixtures.Site(snapshot), "404 Not Found", "en", fixtures.LoadedBuild(test), compose.FrameOf(snapshot)).For("/nothing\x1b[2J<b>")
	golden.Match(test, "notfound.html", lined(missing))
}

// A person who works nowhere now is written with no employer, rather than with an employer of no name.
func TestAPersonWorkingNowhereNamesNoEmployer(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	resume := document.Page{Kind: document.PageResume, Slug: "", Tab: ""}
	record := blankPerson()
	record.Name = "Mira Example"
	if page := fixtureWriter(test).Page(composedOf(test, snapshot, resume).Document, record, ""); strings.Contains(page, "worksFor") {
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
	return NewPageWriter(fixtures.Site(snapshot), fixtures.LoadedBuild(test), compose.FrameOf(snapshot), fixtures.Day())
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
		Profiles: nil, Page: document.Page{Kind: document.PageResume, Slug: "", Tab: ""},
	}
}
