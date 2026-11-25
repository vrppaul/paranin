package html

import (
	"html"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/assets"
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
// code, a fact as a term and its description, a skill table's roles, which its fold as a grid would take from it, and
// the words between the parts of a line. Each page of the fixture
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
		{page: document.Page{Kind: document.PageResume, Slug: "", Tab: "skills"}, file: "skills.html"},
		{page: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""}, file: "p/tidy-tool.html"},
		{page: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: "why"}, file: "p/tidy-tool/why.html"},
	}
	for _, testCase := range cases {
		test.Run(testCase.file, func(test *testing.T) {
			whole := document.Page{Kind: testCase.page.Kind, Slug: testCase.page.Slug, Tab: ""}
			composed := composedOf(test, siteOnTheDay(snapshot).Pages, whole)
			golden.Match(test, testCase.file, lined(fixtureWriter(test).Page(composed.Document, composed.Record, testCase.page.Tab)))
		})
	}
}

// Every phrase of a document as written, all but a fact or a point that is only a gap, reaches one of its pages, its
// opening's or a tab's, as HTML writes text: the facts a browser shows are the facts the written document holds.
func TestEveryPhraseReachesThePage(test *testing.T) {
	for _, composed := range siteOnTheDay(fixtures.LoadedContent(test)).Pages {
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
// under its title, which titles the page with its page's after it, the address it is known by and its markdown the
// tab's own.
func TestAPageShowsOneTabOfItsDocument(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	pages := siteOnTheDay(snapshot).Pages
	resume := composedOf(test, pages, document.Page{Kind: document.PageResume, Slug: "", Tab: ""})
	project := composedOf(test, pages, document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""})
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
				experience, `<title>Experience — Mira Example</title>`, `<h1 aria-label="Experience">`,
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
			holds: []string{why, `<title>Why — tidy-tool</title>`, `<h1 aria-label="Why">`, `<link rel="canonical" href="https://example.test/p/tidy-tool/why">`},
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

// What a heading is read from, its tag, whose name the pattern's first group holds and its level's digit the second;
// the level of a page's title, and the one before it, where a page's headings start from
const (
	headingTag      = `<(h([1-6]))[\s>]`
	headingName     = 1
	headingDigit    = 2
	titleLevel      = 1
	levelAboveTitle = titleLevel - 1
	headingsBetween = " "
)

// A page reads as an outline, as a screen reader lists its headings: its title the h1, and each heading after it at
// most one level below the one before it. Every page the face writes for the fixture's site is held to it, and a
// résumé whose opening holds a subsection, which no profile writes, so the opening's blocks are held under its h2.
func TestEveryPageStepsDownItsHeadingsOneLevelAtATime(test *testing.T) {
	about := document.Section{
		Title:  textOf("About"),
		Name:   "",
		Blocks: []document.Block{document.Facts{Kind: document.FactsDetails, List: []document.Fact{{Label: "now", Value: textOf("here")}}}, emptySection("More")},
	}
	subsectioned := writtenPage{name: "a résumé whose opening holds a subsection", markup: fixtureWriter(test).Page(resumeOf(about), blankPerson(), "")}
	heading := regexp.MustCompile(headingTag)
	for _, page := range append(everyWrittenPage(test), subsectioned) {
		test.Run(page.name, func(test *testing.T) {
			run := []string{}
			previous := levelAboveTitle
			for _, match := range heading.FindAllStringSubmatch(page.markup, everyMatch) {
				// the pattern's group is one digit, which always reads as a number
				level, _ := strconv.Atoi(match[headingDigit])
				run = append(run, match[headingName])
				if level > previous+1 {
					test.Errorf("the headings run %s: an h%d after an h%d, want at most an h%d", strings.Join(run, headingsBetween), level, previous, previous+1)
				}
				previous = level
			}
		})
	}
}

// writtenPage is a page the HTML face writes, named by its path.
type writtenPage struct {
	name   string
	markup string
}

// everyWrittenPage is every page the HTML face writes for the fixture's site, through the writers its golden files
// are written by: the résumé's and each project's opening and tabs, and in each language the blog's posts, their older
// versions, its lists, filtered or not, and the page of an address the site has none at.
func everyWrittenPage(test *testing.T) []writtenPage {
	snapshot := fixtures.LoadedContent(test)
	site := siteOnTheDay(snapshot)
	writer := fixtureWriter(test)
	pages := []writtenPage{}
	for _, composed := range site.Pages {
		opening := composed.Document.Head.Page
		pages = append(pages, writtenPage{name: routes.Path(opening, routes.FormatChosen), markup: writer.Page(composed.Document, composed.Record, "")})
		for _, tab := range document.TabsOf(composed.Document).Rest {
			shown := opening
			shown.Tab = tab.Slug
			pages = append(pages, writtenPage{name: routes.Path(shown, routes.FormatChosen), markup: writer.Page(composed.Document, composed.Record, tab.Slug)})
		}
	}
	for _, blog := range slices.Concat([]compose.Blog{site.Own}, site.Translations) {
		pages = append(pages, blogPages(test, snapshot, blog)...)
	}
	return pages
}

// The path asked for where the site has no page, and the status its page names
const (
	missingPath   = "/nothing"
	missingStatus = "404 Not Found"
)

// blogPages are the pages the HTML face writes of the blog in one language, in its frame: its posts, their older
// versions, its lists, filtered or not, and the page of an address the site has none at.
func blogPages(test *testing.T, snapshot *content.Snapshot, blog compose.Blog) []writtenPage {
	build := fixtures.LoadedBuild(test)
	writer := NewPageWriter(fixtures.Site(snapshot), build, fixtures.LoadedCards(test), blog.Frame, fixtures.Day())
	pages := []writtenPage{}
	for _, post := range blog.Posts {
		pages = append(pages, writtenPage{name: routes.Path(post.Document.Head.Page, routes.FormatChosen), markup: writer.Post(post.Document, post.Record)})
	}
	for _, older := range blog.Versions {
		pages = append(pages, writtenPage{name: routes.Path(older.Head.Page, routes.FormatChosen), markup: writer.Post(older, nil)})
	}
	for _, list := range slices.Concat(blog.Lists, blog.Filtered) {
		pages = append(pages, writtenPage{name: routes.Path(list.Head.Page, routes.FormatChosen), markup: writer.List(list)})
	}
	missing := NewNotFound(fixtures.Site(snapshot), missingStatus, build, blog.Frame).For(missingPath)
	return append(pages, writtenPage{name: missingStatus + " in " + blog.Frame.Language, markup: missing})
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

// Every page ends with the foot line, the page of no page among them: the sessions, the site's name leading to where
// it begins and the blog's leading to its list, the one the page belongs to current, the other quiet; the document's
// tabs as windows by their numbers and names, on the blog's pages its list the first window, in either grouping, and a
// post the second, each leading to its page, the one shown the current page; the programs that reach the site and
// read it in a terminal, each opening its hint, a window that holds its command whole to copy and says what it gives,
// curl's reading the page whole, the résumé where there is none; the markdown of the page whole, the résumé's where a
// page has none; the switch that turns the vim keys on, and the theme button. Over it stands the command line those
// keys open, with the places it goes to by name: the document's tabs, the one shown among them, then the page's own
// places, a post's headings and the posts and series of the blog's list, then the page of each project that has one,
// by its slug, shown where its opening is; and the pages its :e opens, the projects' and the blog's lists in every
// language, the one shown among them, the blog's posts on its own pages alone, and the blog's folder in each language,
// where :e opens any other.
func TestEveryPageEndsWithTheFootLine(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	writer := fixtureWriter(test)
	pages := siteOnTheDay(snapshot).Pages
	resume := composedOf(test, pages, document.Page{Kind: document.PageResume, Slug: "", Tab: ""})
	project := composedOf(test, pages, document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: ""})
	post := fixturePost(test)
	missing := NewNotFound(fixtures.Site(snapshot), missingStatus, fixtures.LoadedBuild(test), siteOnTheDay(snapshot).Own.Frame)
	every := []string{
		`popovertarget="_hint-ssh"><span class="way-mark" aria-hidden="true">❯</span>ssh</button>`,
		`popovertarget="_hint-curl"><span class="way-mark" aria-hidden="true">❯</span>curl</button>`,
		`<div id="_hint-ssh" class="hint" popover role="dialog" aria-labelledby="_hint-ssh-title" data-part="hint"><p id="_hint-ssh-title" class="hint-title">ssh</p><div class="command" data-part="command"><span class="command-prompt" aria-hidden="true">❯</span> <code data-part="command-words">ssh `,
		`<p class="hint-gives">The page as a manual in a terminal.</p>`,
		`popovertarget="_hint-ssh" popovertargetaction="hide"`, `popovertarget="_hint-curl" popovertargetaction="hide"`,
		`data-part="vim-keys"`, `data-part="theme"`, `data-part="command-line"`, `role="combobox"`,
		`data-place="list" data-name="blog/series" data-address="/blog/series"`,
		`data-place="list" data-name="ru/blog" data-address="/ru/blog"`,
	}
	// the commands that read the résumé whole, which the page of no page offers too, and the project's page whole
	const (
		resumeReading  = `<code data-part="command-words">curl https://example.test/</code>`
		projectReading = `<code data-part="command-words">curl https://example.test/p/tidy-tool</code>`
	)
	// the place of the fixture's one project with a page; another project has none, so no place
	const (
		projectPlace = `data-place="project" data-name="tidy-tool" data-address="/p/tidy-tool"`
		noPagePlace  = `data-name="other-tool"`
	)
	// the ssh hint of the pages the site's own command opens, the résumé's and the page of no page's, of a post's, which
	// its number opens, and of the blog's lists (D26): each its page's command and what it gives
	const (
		resumeSSH = `<code data-part="command-words">ssh example.test</code>`
		postSSH   = `<code data-part="command-words">ssh blog/1@example.test</code>`
		blogSSH   = `<code data-part="command-words">ssh blog@example.test</code>`
		seriesSSH = `<code data-part="command-words">ssh blog/series@example.test</code>`
	)
	const (
		resumeGives = `<p class="hint-gives">The résumé as an app in a terminal.</p>`
		postGives   = `<p class="hint-gives">This post in a terminal.</p>`
		blogGives   = `<p class="hint-gives">The blog in a terminal.</p>`
	)
	// the résumé's markdown, which the page of no page offers too
	const resumeMarkdown = `<a class="way-chip" href="/index.md">index.md</a>`
	// the two sessions, each leading to where it begins: the site's current on every page but the blog's, the blog's,
	// leading to its list, quiet beside it; on the blog's pages the other way round
	const (
		siteSessions = `<span class="sessions"><a class="session" href="/"><span aria-hidden="true">[</span>example.test<span aria-hidden="true">]</span></a>` +
			`<a class="session session-other" href="/blog"><span aria-hidden="true">[</span>blog<span aria-hidden="true">]</span></a></span>`
		blogSessions = `<span class="sessions"><a class="session session-other" href="/"><span aria-hidden="true">[</span>example.test<span aria-hidden="true">]</span></a>` +
			`<a class="session" href="/blog"><span aria-hidden="true">[</span>blog<span aria-hidden="true">]</span></a></span>`
	)
	lists := siteOnTheDay(snapshot).Own.Lists
	cases := []struct {
		name  string
		page  string
		holds []string
		lacks []string
	}{
		{
			name: "the résumé's opening", page: writer.Page(resume.Document, resume.Record, ""),
			holds: slices.Concat(every, []string{
				siteSessions, resumeSSH, resumeGives, resumeMarkdown, resumeReading, `<a class="window" href="/" aria-current="page"><span class="window-number">1:</span>about</a>`,
				`<a class="window" href="/experience"><span class="window-number">2:</span>experience</a>`,
				`<a class="window" href="/contact"><span class="window-number">5:</span>contact</a>`,
				`data-place="tab" data-name="about" data-address="/" data-current`,
				`data-place="tab" data-name="contact" data-address="/contact"`, projectPlace, projectPlace + ` data-edit`,
				`data-part="folder" data-name="blog/" data-address="/blog/"`, `data-part="folder" data-name="ru/blog/" data-address="/ru/blog/"`,
			}),
			lacks: []string{noPagePlace, `data-name="blog/reading-slowly"`},
		},
		{
			name: "a tab of the résumé", page: writer.Page(resume.Document, resume.Record, "experience"),
			holds: slices.Concat(every, []string{
				siteSessions, resumeSSH, resumeGives, resumeMarkdown, resumeReading, `<a class="window" href="/"><span class="window-number">1:</span>about</a>`,
				`<a class="window" href="/experience" aria-current="page"><span class="window-number">2:</span>experience</a>`,
			}),
			lacks: nil,
		},
		{
			name: "a project's page", page: writer.Page(project.Document, project.Record, ""),
			holds: slices.Concat(every, []string{
				siteSessions, resumeSSH, resumeGives, `<a class="way-chip" href="/p/tidy-tool.md">tidy-tool.md</a>`, projectReading,
				`<a class="window" href="/p/tidy-tool" aria-current="page"><span class="window-number">1:</span>tidy-tool</a>`,
				`<a class="window" href="/p/tidy-tool/why"><span class="window-number">2:</span>why</a>`,
				`data-place="tab" data-name="tidy-tool" data-address="/p/tidy-tool" data-current`,
				`data-place="tab" data-name="why" data-address="/p/tidy-tool/why"`, projectPlace + ` data-current`,
			}),
			lacks: nil,
		},
		{
			name: "a tab of a project's page", page: writer.Page(project.Document, project.Record, "why"),
			holds: slices.Concat(every, []string{siteSessions, projectPlace + ` hidden`, projectReading}), lacks: []string{projectPlace + ` data-current`},
		},
		{
			name: "the page of no page", page: missing.For(missingPath),
			holds: slices.Concat(every, []string{siteSessions, resumeSSH, resumeGives, resumeMarkdown, resumeReading, projectPlace}),
			lacks: []string{`class="window"`, `data-place="tab"`},
		},
		{
			name: "a post's page", page: writer.Post(post.Document, post.Record),
			holds: slices.Concat(every, []string{
				blogSessions, postSSH, postGives, `<a class="way-chip" href="/blog/reading-slowly.md">reading-slowly.md</a>`,
				`<code data-part="command-words">curl https://example.test/blog/reading-slowly</code>`,
				`<a class="window" href="/blog"><span class="window-number">1:</span>posts</a>`,
				`<a class="window" href="/blog/reading-slowly" aria-current="page"><span class="window-number">2:</span>reading-slowly</a>`,
				`data-place="tab" data-name="posts" data-address="/blog" hidden`,
				`data-place="tab" data-name="reading-slowly" data-address="/blog/reading-slowly" data-current`,
				`data-place="heading" data-name="in-a-terminal" data-address="/blog/reading-slowly#in-a-terminal"`, projectPlace,
				`data-place="post" data-name="blog/reading-slowly" data-address="/blog/reading-slowly" data-current data-edit`,
			}),
			lacks: nil,
		},
		{
			name: "the blog's list by year", page: writer.List(lists[0]),
			holds: slices.Concat(every, []string{
				blogSessions, blogSSH, blogGives, `<a class="way-chip" href="/blog.md">blog.md</a>`,
				`<code data-part="command-words">curl https://example.test/blog</code>`,
				`<a class="window" href="/blog" aria-current="page"><span class="window-number">1:</span>posts</a>`,
				`data-place="tab" data-name="posts" data-address="/blog" data-current`,
				`data-place="post" data-name="reading-slowly" data-address="/blog/reading-slowly"`,
				`data-place="series" data-name="reading" data-address="/blog/series#reading"`, projectPlace,
				`data-place="list" data-name="blog" data-address="/blog" data-current data-edit`,
				`data-place="post" data-name="blog/reading-slowly" data-address="/blog/reading-slowly" data-edit`,
			}),
			lacks: []string{`<span class="window-number">2:`},
		},
		{
			name: "the blog's list by series", page: writer.List(lists[1]),
			holds: slices.Concat(every, []string{
				blogSessions, seriesSSH, blogGives, `<a class="way-chip" href="/blog/series.md">series.md</a>`,
				`<code data-part="command-words">curl https://example.test/blog/series</code>`,
				`<a class="window" href="/blog/series" aria-current="page"><span class="window-number">1:</span>posts</a>`,
				`data-place="series" data-name="reading" data-address="/blog/series#reading"`,
			}),
			lacks: []string{`<span class="window-number">2:`},
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
// status in its title and on the page, in English as HTTP writes it; quotes the path asked for as Go quotes a string,
// so no control character in it reaches the page; and leads to where the site begins, by its path, under the site's
// name, its foot line in its language's words. Each language's is held to its golden file whole.
func TestAMissingPageReadsAsItsGoldenMarkup(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	site := siteOnTheDay(snapshot)
	cases := []struct {
		file  string
		frame compose.Frame
		path  string
	}{
		{file: "notfound.html", frame: site.Own.Frame, path: "/nothing\x1b[2J<b>"},
		{file: "ru/notfound.html", frame: site.Translations[0].Frame, path: "/ru/nothing"},
	}
	for _, testCase := range cases {
		test.Run(testCase.file, func(test *testing.T) {
			missing := NewNotFound(fixtures.Site(snapshot), missingStatus, fixtures.LoadedBuild(test), testCase.frame).For(testCase.path)
			golden.Match(test, testCase.file, lined(missing))
		})
	}
}

// A person who works nowhere now is written with no employer, rather than with an employer of no name.
func TestAPersonWorkingNowhereNamesNoEmployer(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	resume := document.Page{Kind: document.PageResume, Slug: "", Tab: ""}
	record := blankPerson()
	record.Name = "Mira Example"
	if page := fixtureWriter(test).Page(composedOf(test, siteOnTheDay(snapshot).Pages, resume).Document, record, ""); strings.Contains(page, "worksFor") {
		test.Errorf("the page %q names an employer", page)
	}
}

// The blog's list is read as a browser, a screen reader and a search engine read it: the blog's name the page's one
// h1, in block letters, its line, the Switch of its groupings and its tags, each a link to the list it filters; each
// group under its name, an h2, its posts in order, each a link to its post, a part to come no link; then the foot line
// of the blog's session. Each grouping is held to its golden file, named after its own address, and the list filtered
// by a tag to its own, named by the filter, since the list's path holds a query.
func TestTheListReadsAsItsGoldenMarkup(test *testing.T) {
	site := siteOnTheDay(fixtures.LoadedContent(test))
	cases := []struct {
		page document.Page
		file string
	}{
		{page: document.PageOf(document.PageBlog), file: "blog.html"},
		{page: document.PageOf(document.PageBlogSeries), file: "blog/series.html"},
		{page: document.TagFilter("reading"), file: "filtered/tag-reading.html"},
	}
	for _, testCase := range cases {
		test.Run(testCase.file, func(test *testing.T) {
			list := fixtures.DocumentOf(test, slices.Concat(site.Own.Lists, site.Own.Filtered), testCase.page)
			golden.Match(test, testCase.file, lined(fixtureWriter(test).List(list)))
		})
	}
}

// A page no search engine may index tells it so and names no address it is known by, no record and no languages: a
// filtered list, a view of the list and no page of its own, called by the blog's name and its filter, with no preview
// either, since a part of the list is no copy of it; an unlisted post, which keeps its preview at its own address. The
// list, the list by series, called by the blog's name and its grouping so no two lists share a title, and a post
// published are indexed and known at their own addresses. A draft, which its author alone reads, shows no preview.
func TestAPageNotIndexedNamesNoAddress(test *testing.T) {
	site := siteOnTheDay(fixtures.LoadedContent(test))
	lists := slices.Concat(site.Own.Lists, site.Own.Filtered)
	writer := fixtureWriter(test)
	unlisted, published := unlistedPost(test), fixturePost(test)
	draftsSite := compose.SiteOf(compose.Prepare(fixtures.LoadedContent(test)), compose.Audience{At: fixtures.Day(), Drafts: true})
	draft := composedOf(test, draftsSite.Own.Posts, document.PostPage("notes-on-tools"))
	cases := []struct {
		name  string
		page  string
		holds []string
		lacks []string
	}{
		{
			name: "filtered by a tag", page: writer.List(fixtures.DocumentOf(test, lists, document.TagFilter("reading"))),
			holds: []string{noIndexTag, `<title>blog — posts tagged reading</title>`}, lacks: []string{`rel="canonical"`, `og:url`},
		},
		{
			name: "the list", page: writer.List(fixtures.DocumentOf(test, lists, document.PageOf(document.PageBlog))),
			holds: []string{`<link rel="canonical" href="https://example.test/blog">`, `<title>blog</title>`}, lacks: []string{noIndexTag},
		},
		{
			name: "the list by series", page: writer.List(fixtures.DocumentOf(test, lists, document.PageOf(document.PageBlogSeries))),
			holds: []string{`<link rel="canonical" href="https://example.test/blog/series">`, "<title>blog" + dot + "series</title>"},
			lacks: []string{noIndexTag},
		},
		{
			name: "an unlisted post", page: writer.Post(unlisted.Document, unlisted.Record),
			holds: []string{noIndexTag, fixturePostAddress},
			lacks: []string{`rel="canonical"`, `<link rel="alternate" hreflang=`, `application/ld+json`},
		},
		{
			name: "a draft", page: writer.Post(draft.Document, draft.Record),
			holds: []string{noIndexTag}, lacks: []string{`rel="canonical"`, `og:url`, `application/ld+json`},
		},
		{
			name: "a post published", page: writer.Post(published.Document, published.Record),
			holds: []string{`<link rel="canonical" href="https://example.test/blog/reading-slowly">`, fixturePostAddress, `<link rel="alternate" hreflang="ru"`},
			lacks: []string{noIndexTag},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			for _, held := range testCase.holds {
				if !strings.Contains(testCase.page, held) {
					test.Errorf("the page %q holds no %q", testCase.page, held)
				}
			}
			for _, lacked := range testCase.lacks {
				if strings.Contains(testCase.page, lacked) {
					test.Errorf("the page %q holds %q", testCase.page, lacked)
				}
			}
		})
	}
}

// A page a preview shows names its card, the cards' manifest's for its page, a tab's its page's: the card's whole
// address, its type, its size and the page's title as its alt, and X's large card where the card is named, the small
// one where it is not. An indexed page lets a search engine show the card large, a page not indexed keeps its noindex
// and says no more, a post's record names the card as its image, and a page no preview shows names none.
func TestAPageThatShowsAPreviewNamesItsCard(test *testing.T) {
	site := siteOnTheDay(fixtures.LoadedContent(test))
	lists := slices.Concat(site.Own.Lists, site.Own.Filtered)
	writer := fixtureWriter(test)
	resume := composedOf(test, site.Pages, document.PageOf(document.PageResume))
	published, unlisted := fixturePost(test), unlistedPost(test)
	undated := document.Document{Head: fixtures.HeadOf(document.PostPage("undated"), textOf("Undated")), Blocks: nil}
	draftsSite := compose.SiteOf(compose.Prepare(fixtures.LoadedContent(test)), compose.Audience{At: fixtures.Day(), Drafts: true})
	draft := composedOf(test, draftsSite.Own.Posts, document.PostPage("notes-on-tools"))
	cases := []struct {
		name  string
		page  string
		holds []string
		lacks []string
	}{
		{
			name: "the résumé", page: writer.Page(resume.Document, resume.Record, ""),
			holds: slices.Concat(cardTags(resumeCardAddress, "Mira Example"), []string{twitterLargeCard, indexedImageRule}),
			lacks: []string{twitterSmallCard, noIndexTag},
		},
		{
			name: "a tab of the résumé shows its page's card", page: writer.Page(resume.Document, resume.Record, "experience"),
			holds: slices.Concat(cardTags(resumeCardAddress, "Experience — Mira Example"), []string{twitterLargeCard, indexedImageRule}),
		},
		{
			name: "the list", page: writer.List(fixtures.DocumentOf(test, lists, document.PageOf(document.PageBlog))),
			holds: slices.Concat(cardTags(pageCardAddress, "blog"), []string{twitterLargeCard, indexedImageRule}), lacks: []string{noIndexTag},
		},
		{
			name: "a post published", page: writer.Post(published.Document, published.Record),
			holds: slices.Concat(cardTags(pageCardAddress, "Reading slowly"), []string{twitterLargeCard, indexedImageRule, `"image":"` + pageCardAddress + `"`}),
		},
		{
			name: "an unlisted post", page: writer.Post(unlisted.Document, unlisted.Record),
			holds: slices.Concat(cardTags(pageCardAddress, "Reading slowly"), []string{twitterLargeCard, noIndexTag}), lacks: []string{"max-image-preview"},
		},
		{
			name: "a post whose card the manifest does not name", page: writer.Post(undated, undatedRecord(undated.Head.Page)),
			holds: []string{twitterSmallCard}, lacks: []string{"og:image", twitterLargeCard, "max-image-preview", `"image"`},
		},
		{
			name: "a filtered list", page: writer.List(fixtures.DocumentOf(test, lists, document.TagFilter("reading"))),
			holds: []string{noIndexTag}, lacks: []string{"og:image", "twitter:card", "max-image-preview"},
		},
		{
			name: "a draft", page: writer.Post(draft.Document, draft.Record),
			holds: []string{noIndexTag}, lacks: []string{"og:image", "twitter:card", "max-image-preview"},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			for _, held := range testCase.holds {
				if !strings.Contains(testCase.page, held) {
					test.Errorf("the page %q holds no %q", testCase.page, held)
				}
			}
			for _, lacked := range testCase.lacks {
				if strings.Contains(testCase.page, lacked) {
					test.Errorf("the page %q holds %q", testCase.page, lacked)
				}
			}
		})
	}
}

// cardTags are the tags a head names a card by: its address, its alt, its type and its size.
func cardTags(address, alt string) []string {
	return []string{
		openGraphTag("og:image", address), openGraphTag("og:image:alt", alt), openGraphTag("og:image:type", assets.CardType),
		openGraphTag("og:image:width", strconv.Itoa(assets.CardWidth)), openGraphTag("og:image:height", strconv.Itoa(assets.CardHeight)),
	}
}

// openGraphTag is the tag of an Open Graph property with its content.
func openGraphTag(property, content string) string {
	return `<meta property="` + property + `" content="` + content + `">`
}

// undatedRecord is the record of a post of a page that has no day.
func undatedRecord(page document.Page) compose.PostRecord {
	return compose.PostRecord{
		Title: "Undated", Description: "", Page: page, Published: time.Time{}, VersionOut: time.Time{},
		Author: compose.Author{Name: "Mira Example", Page: document.PageOf(document.PageResume)}, Language: "en", Tags: nil, Series: nil,
	}
}

// The two cards X shows, and the robots' line of an indexed page, which lets a search engine show the card large
const (
	twitterLargeCard = `<meta name="twitter:card" content="summary_large_image">`
	twitterSmallCard = `<meta name="twitter:card" content="summary">`
	indexedImageRule = `<meta name="robots" content="max-image-preview:large">`
)

// The addresses of the résumé's card and of every other page's in the cards' fixture
const (
	resumeCardAddress = "https://example.test/cards/resume-fixture.png"
	pageCardAddress   = "https://example.test/cards/page-fixture.png"
)

// A head's filters stand in their order, the list's and a post's alike: each a link to where it leads, a tag's word
// after a hash a screen reader skips; the filter shown current, named to a screen reader as the way to the whole list;
// and a filter that leads nowhere, as a tag of a post the lists name nowhere, as words.
func TestAHeadsFiltersLeadWhereTheyLead(test *testing.T) {
	whole, nowhere := document.PageOf(document.PageBlog), document.PageOf(document.PageNone)
	cases := []struct {
		name    string
		page    document.Page
		filters []document.Filter
		want    string
	}{
		{
			name: "a list's tags", page: whole,
			filters: []document.Filter{tagFilter("go", document.TagFilter("go")), tagFilter("rust", document.TagFilter("rust"))},
			want:    `<ul class="tags"><li>` + tagLink("go", "/blog?tag=go") + `</li><li>` + tagLink("rust", "/blog?tag=rust") + `</li></ul>`,
		},
		{
			name: "the tag shown, the way to the whole list", page: document.TagFilter("go"),
			filters: []document.Filter{{Tag: "go", Page: whole, Current: true, WayOut: "posts tagged go, show all posts"}},
			want: `<ul class="tags"><li><a class="tag" href="/blog" aria-current="true" aria-label="posts tagged go, show all posts">` +
				`<span class="tag-mark" aria-hidden="true">#</span>go</a></li></ul>`,
		},
		{
			name: "a tag that leads nowhere", page: document.PostPage("one"),
			filters: []document.Filter{tagFilter("go", nowhere)},
			want:    `<ul class="tags"><li><span class="tag-mark" aria-hidden="true">#</span>go</li></ul>`,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			head := fixtures.HeadOf(testCase.page, textOf("blog"))
			head.Filters = testCase.filters
			if page := fixtureWriter(test).List(document.Document{Head: head, Blocks: nil}); !strings.Contains(page, testCase.want) {
				test.Errorf("the page %q holds no %q", page, testCase.want)
			}
		})
	}
}

// tagFilter is a tag a head offers, leading to a page, not the filter shown.
func tagFilter(tag string, leads document.Page) document.Filter {
	return document.Filter{Tag: tag, Page: leads, Current: false, WayOut: ""}
}

// tagLink is a tag as a head's filters write it, by its word and the address it leads to.
func tagLink(name, address string) string {
	return `<a class="tag" href="` + address + `"><span class="tag-mark" aria-hidden="true">#</span>` + name + `</a>`
}

// Every phrase of the blog's list reaches its page, as HTML writes text, in each grouping and each filter: every post
// it names is there, with where it stands and its tags, and the filter.
func TestEveryPhraseOfTheListReachesItsPage(test *testing.T) {
	site := siteOnTheDay(fixtures.LoadedContent(test))
	for _, list := range slices.Concat(site.Own.Lists, site.Own.Filtered) {
		page := fixtureWriter(test).List(list)
		for _, phrase := range document.Phrases(list) {
			if !strings.Contains(page, html.EscapeString(phrase)) {
				test.Errorf("the page of %+v holds no %q", list.Head.Page, phrase)
			}
		}
	}
}

// The list's head offers both its groupings, each a link to the list grouped so, the one shown the current page; on a
// filtered list the grouping it shows, by year, current, though the list it leads to is another page.
func TestTheSwitchMarksTheGroupingShown(test *testing.T) {
	cases := []struct {
		name string
		page document.Page
		want string
	}{
		{
			name: "by year", page: document.PageOf(document.PageBlog),
			want: `<span class="switch"><a class="switch-way" href="/blog" aria-current="page">years</a><a class="switch-way" href="/blog/series">series</a></span>`,
		},
		{
			name: "by series", page: document.PageOf(document.PageBlogSeries),
			want: `<span class="switch"><a class="switch-way" href="/blog">years</a><a class="switch-way" href="/blog/series" aria-current="page">series</a></span>`,
		},
		{
			name: "filtered", page: document.TagFilter("go"),
			want: `<span class="switch"><a class="switch-way" href="/blog" aria-current="true">years</a><a class="switch-way" href="/blog/series">series</a></span>`,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			list := document.Document{Head: fixtures.HeadOf(testCase.page, textOf("blog")), Blocks: nil}
			if page := fixtureWriter(test).List(list); !strings.Contains(page, testCase.want) {
				test.Errorf("the page %q holds no %q", page, testCase.want)
			}
		})
	}
}

// A tile is a post the list names, the whole of it leading to the post by its title's link, whose words open with what
// its top edge says: in its series' own group its part and its day there; in a group that is no series' own its day
// there, and at its foot its part and its series' title, a link of its own to its series' group on the list by series,
// then its tags, as words; a part still to come its part and the word saying it comes, no link.
func TestATileSaysWhereItsPostStands(test *testing.T) {
	out := document.PostOut{Page: document.PostPage("one"), Published: time.Date(2026, time.November, 10, 9, 0, 0, 0, time.UTC)}
	stintGroup := document.Anchor{Page: document.Page{Kind: document.PageBlogSeries, Slug: "", Tab: ""}, Name: "stint"}
	cases := []struct {
		name string
		post document.ListedPost
		want string
	}{
		{
			name: "a part in its series' group",
			post: document.ListedPost{Readiness: out, Title: textOf("One"), Lead: textOf("About one."), Part: textOf("part 1"), SeriesGroup: stintGroup},
			want: `<div class="pane tile"><a class="tile-title" href="/blog/one"><span class="edge"><span class="edge-name tile-edge"><span class="tile-part">part 1</span>` + dot +
				`<time datetime="2026-11-10">10 Nov 2026</time></span></span> One</a> <span class="tile-lead">About one.</span> </div>`,
		},
		{
			name: "a part in its year's group, tagged",
			post: document.ListedPost{
				Readiness: out, Title: textOf("One"), Part: textOf("part 1"), Series: textOf("Stint"), SeriesGroup: stintGroup, Tags: []string{"payments", "agents"},
			},
			want: `<div class="pane tile"><a class="tile-title" href="/blog/one"><span class="edge"><span class="edge-name tile-edge"><time datetime="2026-11-10">10 Nov 2026</time></span></span> ` +
				`One</a> <div class="tile-foot"><span class="tile-series"><span class="tile-part">part 1</span>` +
				`<span class="tile-dot">` + dot + `</span> <a class="tile-series-link" href="/blog/series#stint">Stint</a></span> ` +
				`<ul class="tags"><li><span class="tag-mark" aria-hidden="true">#</span>payments</li>` +
				`<li><span class="tag-mark" aria-hidden="true">#</span>agents</li></ul></div></div>`,
		},
		{
			name: "a part to come",
			post: document.ListedPost{Readiness: document.ComingPart{Coming: textOf("soon")}, Title: textOf("Three"), Part: textOf("part 3"), SeriesGroup: stintGroup},
			want: `<div class="pane tile tile-to-come"><span class="edge"><span class="edge-name tile-edge"><span class="tile-part">part 3</span>` + dot +
				`soon</span></span><span class="tile-title">Three</span> </div>`,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			group := document.Section{Title: textOf("Stint"), Blocks: []document.Block{document.Posts{List: []document.ListedPost{testCase.post}}}}
			list := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageBlogSeries, Slug: "", Tab: ""}, textOf("blog")), Blocks: []document.Block{group}}
			if page := fixtureWriter(test).List(list); !strings.Contains(page, testCase.want) {
				test.Errorf("the page %q holds no %q", page, testCase.want)
			}
		})
	}
}

// A block of the blog's list that is no group, which compose writes none of, stands as any page draws it, never left
// out: words, and posts outside a section, as tiles with no name over them.
func TestABlockOfTheListThatIsNoGroupStandsAsAnyPageDraws(test *testing.T) {
	post := document.ListedPost{Readiness: document.ComingPart{Coming: textOf("soon")}, Title: textOf("Loose post")}
	blocks := []document.Block{document.Paragraph{Spans: textOf("Loose words.")}, document.Posts{List: []document.ListedPost{post}}}
	list := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageBlog, Slug: "", Tab: ""}, textOf("blog")), Blocks: blocks}
	page := fixtureWriter(test).List(list)
	for _, want := range []string{"Loose words.", `<span class="tile-title">Loose post</span>`} {
		if !strings.Contains(page, want) {
			test.Errorf("the page %q holds no %q", page, want)
		}
	}
}

// A post's page is the post whole, as a browser, a screen reader and a search engine read it: its title the page's
// one h1, its lead and facts beside it, its paragraphs numbered, its headings, lists, quote, table and code, each note
// beside the line that first cites it, then the foot line of the blog's session. It is held to its golden file,
// named after the post's own address.
func TestAPostReadsAsItsGoldenMarkup(test *testing.T) {
	post := fixturePost(test)
	golden.Match(test, "blog/reading-slowly.html", lined(fixtureWriter(test).Post(post.Document, post.Record)))
}

// A translation's post and its language's list read as their golden files, in the language's frame.
func TestATranslationReadsAsItsGoldenMarkupInItsLanguage(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	// the fixture is translated into Russian alone
	translated := siteOnTheDay(snapshot).Translations[0]
	writer := NewPageWriter(fixtures.Site(snapshot), fixtures.LoadedBuild(test), fixtures.LoadedCards(test), translated.Frame, fixtures.Day())
	post := composedOf(test, translated.Posts, document.PostPage("reading-slowly").InLanguage("ru"))
	list := fixtures.DocumentOf(test, translated.Lists, document.PageOf(document.PageBlog).InLanguage("ru"))
	golden.Match(test, "ru/blog/reading-slowly.html", lined(writer.Post(post.Document, post.Record)))
	golden.Match(test, "ru/blog.html", lined(writer.List(list)))
}

// An older version of a post reads as the post does, with its own title, lead and body, its version's pick and under
// it the way to the newer, held to its golden file, named after its own address; it tells a search engine not to index
// it and names no address it is known by, as a filtered list names none, while the post itself is known at its own.
func TestAnOlderVersionReadsAsItsGoldenMarkupIndexedByNone(test *testing.T) {
	older := fixtures.DocumentOf(test, siteOnTheDay(fixtures.LoadedContent(test)).Own.Versions, document.PostPage("reading-slowly").OlderVersion("2026-11-10"))
	page := fixtureWriter(test).Post(older, nil)
	golden.Match(test, "blog/reading-slowly/2026-11-10.html", lined(page))
	for _, want := range []string{noIndexTag, `<a class="version-note" href="/blog/reading-slowly">`} {
		if !strings.Contains(page, want) {
			test.Errorf("the older version's page %q holds no %q", page, want)
		}
	}
	if strings.Contains(page, `rel="canonical"`) {
		test.Errorf("the older version's page %q names an address it is known by", page)
	}
}

// The versions' pick lists every version of a post newest first, each by its day and where it stands: the current one,
// the version read, the current page and no link, and an older one a link to its page; a post of one version alone
// lists that one.
func TestTheVersionsPickListsEveryVersionAndWhereEachStands(test *testing.T) {
	post := fixturePost(test)
	page := fixtureWriter(test).Post(post.Document, post.Record)
	onlyVersion := fixtureWriter(test).Post(oneVersionPost(test), nil)
	const currentRow = `<li class="pick-part version-read" aria-current="page"><span class="pick-row version-row"><time datetime="2026-11-12">12 Nov 2026</time> <span class="version-state">current</span></span></li>`
	cases := []struct {
		name  string
		page  string
		want  string
		count int
	}{
		{name: "the chip, the day read", page: page, want: `popovertarget="_versions"><time datetime="2026-11-12">12 Nov 2026</time> <span class="pick-mark"`, count: 1},
		{name: "the window titled in the site's word", page: page, want: `class="pick-title">versions</p>`, count: 1},
		{name: "the current version, read", page: page, want: currentRow, count: 1},
		{
			name:  "an older version, a link",
			page:  page,
			want:  `<li class="pick-part"><a class="pick-row version-row" href="/blog/reading-slowly/2026-11-10"><time datetime="2026-11-10">10 Nov 2026</time></a></li>`,
			count: 1,
		},
		{name: "a post of one version, its one row", page: onlyVersion, want: currentRow, count: 1},
		{name: "a post of one version, its one link", page: onlyVersion, want: `class="pick-row version-row" href=`, count: 0},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := strings.Count(testCase.page, testCase.want); got != testCase.count {
				test.Errorf("the page %q holds %q %d times, want %d", testCase.page, testCase.want, got, testCase.count)
			}
		})
	}
}

// oneVersionPost is a post a test builds by hand, out on the fixture's current day with no older version.
func oneVersionPost(test *testing.T) document.Document {
	day, err := time.Parse(time.DateOnly, "2026-11-12")
	if err != nil {
		test.Fatalf("read the day: %v", err)
	}
	head := fixtures.HeadOf(document.PostPage("alone"), textOf("Alone"))
	head.Version = document.VersionOut{Day: day, Versions: []document.PostVersion{{Page: document.PostPage("alone"), Day: day, Name: "2026-11-12", Current: true, Read: true}}}
	return document.Document{Head: head, Blocks: nil}
}

// A draft not yet dated, whose record knows no moment it came out and no day of its version, names neither, in its
// linked data or its Open Graph, rather than a day that never was; its author and its kind it still names.
func TestAnUndatedDraftsPageNamesNoDay(test *testing.T) {
	page := document.PostPage("undated")
	record := undatedRecord(page)
	written := fixtureWriter(test).Post(document.Document{Head: fixtures.HeadOf(page, textOf("Undated")), Blocks: nil}, record)
	for _, want := range []string{`"@type":"BlogPosting"`, `<meta property="article:author" content="https://example.test/">`} {
		if !strings.Contains(written, want) {
			test.Errorf("the undated draft's page %q holds no %q", written, want)
		}
	}
	for _, unwanted := range []string{"datePublished", "dateModified", "article:published_time", "article:modified_time"} {
		if strings.Contains(written, unwanted) {
			test.Errorf("the undated draft's page %q names %q", written, unwanted)
		}
	}
}

// A draft's page, as its author reads it, says it is a draft where a post out says the day of the version read.
func TestADraftsPageIsMarkedAsADraft(test *testing.T) {
	site := compose.SiteOf(compose.Prepare(fixtures.LoadedContent(test)), compose.Audience{At: fixtures.Day(), Drafts: true})
	draft := composedOf(test, site.Own.Posts, document.PostPage("notes-on-tools"))
	const mark = "<dt>version</dt><dd>draft</dd>"
	if page := fixtureWriter(test).Post(draft.Document, draft.Record); !strings.Contains(page, mark) {
		test.Errorf("the draft's page %q holds no %q", page, mark)
	}
}

// The robots' tag that keeps a search engine from indexing a page
const noIndexTag = `<meta name="robots" content="noindex">`

// The fixture post's own address, as a link's preview shows it
const fixturePostAddress = `<meta property="og:url" content="https://example.test/blog/reading-slowly">`

// A pre whole, and a tag of markup, which a browser shows none of
const (
	preformatted = `(?s)<pre[^>]*>.*?</pre>`
	markupTag    = `<[^>]*>`
)

// Every phrase of a post reaches the text its page shows, as HTML writes text, the tags round its code's kinds of
// word gone: the facts a browser shows are the post's whole.
func TestEveryPhraseOfAPostReachesItsPage(test *testing.T) {
	post := fixturePost(test)
	shown := codeAsShown(fixtureWriter(test).Post(post.Document, post.Record))
	for _, phrase := range document.Phrases(post.Document) {
		if !strings.Contains(shown, html.EscapeString(phrase)) {
			test.Errorf("the page of %+v holds no %q", post.Document.Head.Page, phrase)
		}
	}
}

// codeAsShown is a page with every tag inside a pre gone, the spans its code's kinds of word stand in among them, so
// the code reads as a browser shows it and every other part as it is written.
func codeAsShown(page string) string {
	tag := regexp.MustCompile(markupTag)
	return regexp.MustCompile(preformatted).ReplaceAllStringFunc(page, func(pre string) string { return tag.ReplaceAllString(pre, "") })
}

// A post's page is read as a reader moves through it: its title the one h1; a heading and a paragraph each a place a
// link names, every paragraph at the top of the body or a section numbered by a link to it; a note's number raised
// where a line cites it, the note beside the line that first cites it, and every note once at the text's end.
func TestAPostsPlacesAndNotesStandWhereTheyAreRead(test *testing.T) {
	post := fixturePost(test)
	page := fixtureWriter(test).Post(post.Document, post.Record)
	// the fixture's two notes as the page writes each
	const (
		firstNote  = `<p class="note"><span class="note-number">1</span> Mira Example, on her own reading.</p>`
		secondNote = `<p class="note"><span class="note-number">2</span> A soft one, which rubs out.</p>`
	)
	cases := []struct {
		name  string
		want  string
		count int
	}{
		{name: "one h1", want: "<h1", count: 1},
		{name: "the h1, the post's title", want: "<h1 class=\"post-title\">Reading slowly</h1>", count: 1},
		{name: "a heading a link names", want: `id="what-a-reader-keeps">What a reader keeps</h2>`, count: 1},
		{name: "a heading under it", want: `<h3 id="in-a-terminal">In a terminal</h3>`, count: 1},
		{name: "the first paragraph's number, leading to it", want: `<a class="paragraph-number" href="#_p1">1</a><p id="_p1">`, count: 1},
		{name: "a paragraph's number after the lists", want: `<a class="paragraph-number" href="#_p2">2</a><p id="_p2">`, count: 1},
		{name: "a paragraph's number in the next section, counting on", want: `<a class="paragraph-number" href="#_p3">3</a><p id="_p3">`, count: 1},
		{name: "each citation of the note cited twice, its number raised", want: `<sup class="citation">1</sup>`, count: 2},
		{name: "the note cited twice, beside the line first citing it", want: `</sup></p><aside class="notes">` + firstNote, count: 1},
		{name: "a note cited inside a list, beside the list", want: `</ul><aside class="notes">` + secondNote, count: 1},
		{name: "every note once at the text's end", want: `<div class="sources-end">` + firstNote + secondNote + `</div>`, count: 1},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := strings.Count(page, testCase.want); got != testCase.count {
				test.Errorf("the page %q holds %q %d times, want %d", page, testCase.want, got, testCase.count)
			}
		})
	}
}

// A heading's anchor is its name, whatever its words, and no other part of its page takes it: a heading named as a
// paragraph, the command line's menu, one of its places, a hint, the series' pick or the panes' heading keeps its name,
// once on the page, since every id of the page's own stands in a form no anchor takes.
func TestAHeadingsAnchorIsItsName(test *testing.T) {
	post := seriesPost(emptySection("P1"), emptySection("Places"), emptySection("Place 1"), emptySection("Hint ssh"),
		emptySection("Parts"), emptySection("More"), document.Paragraph{Spans: textOf("First.")})
	page := fixtureWriter(test).Post(post, nil)
	cases := []struct {
		name string
		want string
	}{
		{name: "a heading of a paragraph's name", want: `id="p1"`},
		{name: "a heading of the command line's menu's name", want: `id="places"`},
		{name: "a heading of a place's name on the menu", want: `id="place-1"`},
		{name: "a heading of a hint's name", want: `id="hint-ssh"`},
		{name: "a heading of the series' pick's name", want: `id="parts"`},
		{name: "a heading of the panes' heading's name", want: `id="more"`},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := strings.Count(page, testCase.want); got != 1 {
				test.Errorf("the page %q holds %q %d times, want once", page, testCase.want, got)
			}
		})
	}
}

// A post of a series, which a test builds by hand, reads as its golden markup: the page the fixture's post, the one
// part of its series out, cannot show, with its other parts among its places.
func TestAPostOfASeriesReadsAsItsGoldenMarkup(test *testing.T) {
	golden.Match(test, "blog/second.html", lined(fixtureWriter(test).Post(seriesPost(emptySection("More")), nil)))
}

// A page's own places lead where the page draws them: a heading to its anchor, another part of a post's series to its
// page, and a series to its group on the list by series, which that list names by the series' name.
func TestAPagesOwnPlacesLeadWhereThePageDrawsThem(test *testing.T) {
	post := fixtureWriter(test).Post(seriesPost(emptySection("More")), nil)
	lists := siteOnTheDay(fixtures.LoadedContent(test)).Own.Lists
	cases := []struct {
		name string
		page string
		want string
	}{
		{name: "a heading, at its anchor", page: post, want: `data-place="heading" data-name="more" data-address="/blog/second#more"`},
		{name: "another part, at its page", page: post, want: `data-place="part" data-name="first" data-address="/blog/first"`},
		{name: "a series, at its group", page: fixtureWriter(test).List(lists[0]), want: `data-place="series" data-name="reading" data-address="/blog/series#reading"`},
		{name: "the group a series leads to", page: fixtureWriter(test).List(lists[1]), want: `<section id="reading" class="group">`},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(testCase.page, testCase.want) {
				test.Errorf("the page %q holds no %q", testCase.page, testCase.want)
			}
		})
	}
}

// A post of a series picks any part of it, from its head and from its foot: the head's chip names the series and the
// post's place in it, the foot's the count of parts, and each opens a window listing every part by its number and
// title, where it stands for the reader beside it: the part read the current page and no link, a part out a link to
// it, the next named so, a part to come no link and named so.
func TestTheSeriesPickListsEveryPartAndWhereEachStands(test *testing.T) {
	head := fixtures.HeadOf(document.PostPage("two"), textOf("Two"))
	series := &document.Series{Title: textOf("Stint"), Parts: []document.Part{
		{Number: 1, Title: textOf("One"), State: document.PartOut, Page: document.PostPage("one")},
		{Number: 2, Title: textOf("Two"), State: document.PartReading, Page: document.PostPage("two")},
		{Number: 3, Title: textOf("Three"), State: document.PartNext, Page: document.PostPage("three")},
		{Number: 4, Title: textOf("Four"), State: document.PartToCome, Page: document.Page{Kind: document.PageNone, Slug: "", Tab: ""}},
	}}
	head.Post = &document.PostPlace{Series: series, Neighbours: document.Neighbours{Before: nil, After: nil}, Discussions: nil, Listed: true}
	page := fixtureWriter(test).Post(document.Document{Head: head, Blocks: nil}, nil)
	cases := []struct {
		name  string
		want  string
		count int
	}{
		{name: "the head's chip, the series and the post's place in it", want: `popovertarget="_parts"><span class="pick-words">Stint</span> <span class="series-count">2 of 4</span>`, count: 1},
		{name: "the foot's chip, the count of parts", want: `popovertarget="_parts-foot"><span class="pick-words">all 4 parts</span> <span class="pick-mark"`, count: 1},
		{name: "each window titled by the series", want: `class="pick-title">Stint</p>`, count: 2},
		{
			name:  "a part out, a link",
			want:  `<li class="pick-part"><a class="pick-row" href="/blog/one"><span class="part-number">1</span> <span class="part-title">One</span> <span class="part-state"></span></a></li>`,
			count: 2,
		},
		{
			name:  "the part read, the current page and no link",
			want:  `<li class="pick-part part-reading" aria-current="page"><span class="pick-row"><span class="part-number">2</span> <span class="part-title">Two</span> <span class="part-state">reading</span></span></li>`,
			count: 2,
		},
		{
			name:  "the part out after it, the next",
			want:  `<li class="pick-part part-next"><a class="pick-row" href="/blog/three"><span class="part-number">3</span> <span class="part-title">Three</span> <span class="part-state">next</span></a></li>`,
			count: 2,
		},
		{
			name:  "a part to come, no link",
			want:  `<li class="pick-part part-to-come"><span class="pick-row"><span class="part-number">4</span> <span class="part-title">Four</span> <span class="part-state">soon</span></span></li>`,
			count: 2,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := strings.Count(page, testCase.want); got != testCase.count {
				test.Errorf("the page %q holds %q %d times, want %d", page, testCase.want, got, testCase.count)
			}
		})
	}
}

// A post's outline lists the headings of its top level, each a link to its heading, which the site's script finds to
// mark the one read; a heading under them is no place of the outline, and a post of no headings has no outline.
func TestTheOutlineListsThePostsHeadings(test *testing.T) {
	headed := document.Document{
		Head: fixtures.HeadOf(document.PostPage("headed"), nil),
		Blocks: []document.Block{
			document.Section{Title: textOf("One"), Name: "one", Blocks: []document.Block{emptySection("Inner")}},
			emptySection("Two"),
		},
	}
	unheaded := document.Document{Head: fixtures.HeadOf(document.PostPage("unheaded"), nil), Blocks: []document.Block{document.Paragraph{Spans: textOf("Words.")}}}
	headedPage, unheadedPage := fixtureWriter(test).Post(headed, nil), fixtureWriter(test).Post(unheaded, nil)
	cases := []struct {
		name  string
		page  string
		want  string
		count int
	}{
		{name: "a heading of the top level, linked", page: headedPage, want: `<li><a href="#one" data-part="outline-link">One</a></li>`, count: 1},
		{name: "the next, linked", page: headedPage, want: `<li><a href="#two" data-part="outline-link">Two</a></li>`, count: 1},
		{name: "a heading the script finds", page: headedPage, want: `<h2 data-part="outline-heading" id="one">One</h2>`, count: 1},
		{name: "a heading under them, no place of the outline", page: headedPage, want: `href="#inner" data-part="outline-link"`, count: 0},
		{name: "no outline where there are no headings", page: unheadedPage, want: `data-part="outline"`, count: 0},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := strings.Count(testCase.page, testCase.want); got != testCase.count {
				test.Errorf("the page %q holds %q %d times, want %d", testCase.page, testCase.want, got, testCase.count)
			}
		})
	}
}

// An item of a post's list that opens with a block, code here, says so, so the block stands on its mark's line, which
// stays in the line a point is copied with; an item of words says nothing of it. Each number of a list in order is
// read to a screen reader, whose own count of the items starts at one.
func TestAnItemOpeningWithABlockIsMarkedAsOne(test *testing.T) {
	code := document.CodeBlock{Language: "go", Runs: []document.CodeRun{{Kind: document.CodePlain, Text: "main()"}}}
	list := document.List{Ordered: true, Start: 1, Items: []document.Item{
		{Blocks: []document.Block{code}}, {Blocks: []document.Block{document.Paragraph{Spans: textOf("Words.")}}},
	}}
	page := fixtureWriter(test).Post(document.Document{Head: fixtures.HeadOf(document.PostPage("listed"), nil), Blocks: []document.Block{list}}, nil)
	for _, want := range []string{
		`<li class="post-item post-item-block"><span class="point-mark">1.</span> <div class="code"`,
		`<li class="post-item"><span class="point-mark">2.</span> Words.</li>`,
	} {
		if !strings.Contains(page, want) {
			test.Errorf("the page %q holds no %q", page, want)
		}
	}
}

// A widget's name set into its window's edge may break between its words alone, each opening at a capital after a
// small letter or a digit, or at the last capital of a run before a small letter.
func TestAWidgetsNameBreaksBetweenItsWords(test *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{name: "ReadingPace", want: "Reading<wbr>Pace"},
		{name: "HTTPRetry", want: "HTTP<wbr>Retry"},
		{name: "LRUCache", want: "LRU<wbr>Cache"},
		{name: "Page2Cache", want: "Page2<wbr>Cache"},
		{name: "CacheLRU", want: "Cache<wbr>LRU"},
		{name: "Cache", want: "Cache"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			live := document.Anchor{Page: document.PostPage("listed"), Name: testCase.name}
			widget := document.Widget{Name: testCase.name, Alt: "A widget.", Caption: "", Live: live}
			page := fixtureWriter(test).Post(document.Document{Head: fixtures.HeadOf(document.PostPage("listed"), nil), Blocks: []document.Block{widget}}, nil)
			if want := `<span class="edge-name">` + testCase.want + `</span>`; !strings.Contains(page, want) {
				test.Errorf("the page of a widget named %q holds no %q", testCase.name, want)
			}
		})
	}
}

// A post's foot leads to the posts beside it, under a heading saying where: in a series its other parts, the series'
// pick beside the heading and no way to the blog's list, each side named by its part, a part out a link to it with its
// lead, a part to come its title and the word saying it comes, no link; outside a series the blog's other posts, the
// way to its list beside the heading, older and newer, each with the day it came out; a side with nothing there named
// by its side and saying why; and under them where the post is discussed, each place a link.
func TestThePanesLeadToThePostsBesideThePost(test *testing.T) {
	series := &document.Series{Title: textOf("Stint"), Parts: nil}
	toCome := document.Page{Kind: document.PageNone, Slug: "", Tab: ""}
	newer := time.Date(2026, time.November, 20, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name        string
		series      *document.Series
		neighbours  document.Neighbours
		discussions []document.Link
		want        []string
		lacks       []string
	}{
		{
			name:   "a part of a series between a part out and a part to come",
			series: series,
			lacks:  []string{`class="panes-all"`},
			neighbours: document.Neighbours{
				Before: &document.Neighbour{Part: 1, Title: textOf("One"), Lead: textOf("About one."), Page: document.PostPage("one"), Published: time.Time{}},
				After:  &document.Neighbour{Part: 3, Title: textOf("Three"), Lead: nil, Page: toCome, Published: time.Time{}},
			},
			want: []string{
				`<h2 id="_more" class="panes-title">Other parts of Stint</h2>`,
				`<a class="pane" href="/blog/one"><span class="edge"><span class="edge-name">part 1</span></span> <span class="pane-title">One</span> <span class="pane-lead">About one.</span> </a>`,
				`<div class="pane pane-after"><span class="edge"><span class="edge-name">part 3</span></span> <span class="pane-title pane-to-come">Three <span class="pane-soon">soon</span></span></div>`,
			},
		},
		{
			name: "the last part of a series", series: series,
			neighbours: document.Neighbours{Before: nil, After: nil},
			want: []string{
				`<div class="pane pane-none"><span class="edge"><span class="edge-name">previous</span></span> <span class="pane-words">none: this is the first part</span></div>`,
				`<div class="pane pane-none pane-after"><span class="edge"><span class="edge-name">next</span></span> <span class="pane-words">none: this is the last part</span></div>`,
			},
		},
		{
			name: "the first post outside a series, a newer one beside it", series: nil,
			neighbours: document.Neighbours{
				Before: nil,
				After:  &document.Neighbour{Part: 0, Title: textOf("Newer"), Lead: textOf("About newer."), Page: document.PostPage("newer"), Published: newer},
			},
			want: []string{
				`<h2 id="_more" class="panes-title">Other posts</h2><a class="panes-all" href="/blog">all posts</a></div>`,
				`<div class="pane pane-none"><span class="edge"><span class="edge-name">older</span></span> <span class="pane-words">none: this is the first post</span></div>`,
				`<a class="pane pane-after" href="/blog/newer"><span class="edge"><span class="edge-name">newer</span></span> <span class="pane-title">Newer</span> <span class="pane-lead">About newer.</span> <span class="pane-date"><time datetime="2026-11-20">20 Nov 2026</time></span></a>`,
			},
		},
		{
			name: "the newest post, discussed in two places", series: nil,
			neighbours: document.Neighbours{Before: nil, After: nil},
			discussions: []document.Link{
				{Target: document.Address{URL: "https://news.example.test/1"}, Spans: textOf("Hacker News")},
				{Target: document.Address{URL: "https://lobsters.example.test/1"}, Spans: textOf("Lobsters")},
			},
			want: []string{
				`<div class="pane pane-none pane-after"><span class="edge"><span class="edge-name">newer</span></span> <span class="pane-words">none: this is the newest post</span></div>`,
				`<p class="discussions"><span class="discuss">discuss</span> <a href="https://news.example.test/1">Hacker News</a>` + dot + `<a href="https://lobsters.example.test/1">Lobsters</a></p>`,
			},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			head := fixtures.HeadOf(document.PostPage("two"), textOf("Two"))
			head.Post = &document.PostPlace{Series: testCase.series, Neighbours: testCase.neighbours, Discussions: testCase.discussions, Listed: true}
			page := fixtureWriter(test).Post(document.Document{Head: head, Blocks: nil}, nil)
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

// seriesPost is a post a test builds by hand, of blocks: the second part of a series of four, the first out before it,
// the third out after it and the fourth to come.
func seriesPost(blocks ...document.Block) document.Document {
	head := fixtures.HeadOf(document.PostPage("second"), textOf("Second"))
	none := document.Page{Kind: document.PageNone, Slug: "", Tab: ""}
	series := &document.Series{Title: textOf("Stint"), Parts: []document.Part{
		{Number: 1, Title: textOf("First"), State: document.PartOut, Page: document.PostPage("first")},
		{Number: 2, Title: textOf("Second"), State: document.PartReading, Page: document.PostPage("second")},
		{Number: 3, Title: textOf("Third"), State: document.PartNext, Page: document.PostPage("third")},
		{Number: 4, Title: textOf("Fourth"), State: document.PartToCome, Page: none},
	}}
	head.Post = &document.PostPlace{Series: series, Neighbours: document.Neighbours{Before: nil, After: nil}, Discussions: nil, Listed: true}
	return document.Document{Head: head, Blocks: blocks}
}

// textOf is words a test builds by hand, with no marks.
func textOf(words string) []document.Span {
	return []document.Span{document.Text{Text: words}}
}

// emptySection is a section a test builds by hand: its title, named by its words, and no blocks.
func emptySection(title string) document.Block {
	return document.Section{Title: []document.Span{document.Text{Text: title}}, Name: document.SlugOf(title), Blocks: nil}
}

// fixturePost is the fixture's post as compose makes it, its document and its record.
func fixturePost(test *testing.T) compose.Composed {
	posts := siteOnTheDay(fixtures.LoadedContent(test)).Own.Posts
	return composedOf(test, posts, document.Page{Kind: document.PagePost, Slug: "reading-slowly", Tab: ""})
}

// unlistedPost is the fixture's post as compose makes it, unlisted, with its record.
func unlistedPost(test *testing.T) compose.Composed {
	post := fixturePost(test)
	place := *post.Document.Head.Post
	place.Listed = false
	post.Document.Head.Post = &place
	return post
}

// siteOnTheDay is the site of a snapshot as the public reads it on the tests' day.
//
//wrapper:names the one site every test of the page draws
func siteOnTheDay(snapshot *content.Snapshot) compose.Site {
	return compose.SiteOf(compose.Prepare(snapshot), compose.Audience{At: fixtures.Day(), Drafts: false})
}

// What lined breaks a line between, two tags that meet, and what it keeps whole, a pre, whose every letter shows
const (
	tagsMeeting   = "><"
	tagsBroken    = ">\n<"
	preOpening    = "<pre"
	preClosing    = "</pre>"
	goldenFileEnd = "\n"
)

// lined is a page with a line break between each two tags that meet, so its golden file reads in a diff; a space
// or a word between two tags stays where it is, and so does all inside a pre, where a break would show, and the
// page's tests, which run over the golden files, would copy it.
func lined(page string) string {
	var builder strings.Builder
	rest := page
	for {
		outside, opened, found := strings.Cut(rest, preOpening)
		if !found {
			return builder.String() + strings.ReplaceAll(outside, tagsMeeting, tagsBroken) + goldenFileEnd
		}
		builder.WriteString(strings.ReplaceAll(outside+preOpening, tagsMeeting, tagsBroken))
		inside, after, _ := strings.Cut(opened, preClosing)
		builder.WriteString(inside)
		rest = preClosing + after
	}
}

// fixtureWriter writes the pages of the fixture's site, with the build's fixture and the fixture's words.
func fixtureWriter(test *testing.T) PageWriter {
	snapshot := fixtures.LoadedContent(test)
	return NewPageWriter(fixtures.Site(snapshot), fixtures.LoadedBuild(test), fixtures.LoadedCards(test), siteOnTheDay(snapshot).Own.Frame, fixtures.Day())
}

// composedOf is a page of the fixture among those compose made, its document and its record.
func composedOf(test *testing.T, pages []compose.Composed, page document.Page) compose.Composed {
	for _, composed := range pages {
		if composed.Document.Head.Page == page {
			return composed
		}
	}
	test.Fatalf("compose made no page %+v", page)
	return compose.Composed{Document: document.Document{Head: fixtures.HeadOf(page, nil), Blocks: nil}, Tabs: nil, Record: nil}
}

// blankPerson is a record of a person with no facts, for a page a test builds by hand: every page a person reads
// has a record, and a page standing for the résumé has a person's.
func blankPerson() compose.PersonRecord {
	return compose.PersonRecord{
		Name: "", AlternateNames: nil, Title: "", Email: "", Employer: "", Schools: nil, Skills: nil, Languages: nil,
		Profiles: nil, Page: document.Page{Kind: document.PageResume, Slug: "", Tab: ""},
	}
}
