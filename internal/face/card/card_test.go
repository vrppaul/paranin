package card

import (
	"testing"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The fixture's post, its folder and its file, and where it says it is published
const (
	fixturePostSlug = "reading-slowly"
	fixturePostFile = "posts/reading-slowly/post.en.md"
	publishedStatus = "status: published"
	unlistedStatus  = "status: unlisted"
	fixtureDraft    = "notes-on-tools"
	olderVersion    = "2026-11-10"
	russian         = "ru"
)

// The cards' mark in the fixture, the blog's name in the site's own language, and the commands that reach its pages
// over ssh: the site's own, the list's, the post's by its number and its translation's
const (
	fixtureMark   = "blog"
	fixtureWayIn  = "ssh example.test"
	listWay       = "ssh blog@example.test"
	postWay       = "ssh blog/1@example.test"
	translatedWay = "ssh ru/blog/1@example.test"
)

// Each page a preview shows has the facts of its card, read off its document in its language's words, the command that
// reaches it over ssh its way in: the résumé, a project, the blog's list by year and by series, a post, its series the
// one its page shows, an unlisted post, which a preview shows too, and a translation, under the cards' mark in the
// site's own language. A tab has none of its own: its page's card stands for it.
func TestEachPageAPreviewShowsHasTheFactsOfItsCard(test *testing.T) {
	site := siteOf(fixtures.LoadedContent(test), false)
	unlisted := siteOf(fixtures.LoadedFrom(test, unlistedCopy(test)), false)
	post := Post{
		Mark: fixtureMark, Day: "10 Nov 2026", Series: "Reading", Title: "Reading slowly", Lead: "Why a page read twice is read once.",
		WayIn: postWay,
	}
	list := List{Mark: fixtureMark, Line: "What the writer reads, and how.", WayIn: listWay}
	cases := []struct {
		name string
		site compose.Site
		page document.Page
		want Facts
	}{
		{
			name: "the résumé", site: site, page: document.PageOf(document.PageResume),
			want: Resume{Name: "Mira Example", Line: "Software Engineer", WayIn: fixtureWayIn},
		},
		{
			name: "a project", site: site, page: projectPage("tidy-tool", ""),
			want: Project{Name: "tidy-tool", Line: "Tidies files.", WayIn: fixtureWayIn},
		},
		{name: "the list by year", site: site, page: document.PageOf(document.PageBlog), want: list},
		{
			name: "the list by series", site: site, page: document.PageOf(document.PageBlogSeries),
			want: List{Mark: fixtureMark, Line: list.Line, WayIn: "ssh blog/series@example.test"},
		},
		{name: "a post", site: site, page: document.PostPage(fixturePostSlug), want: post},
		{
			name: "an unlisted post, whose page shows no series", site: unlisted, page: document.PostPage(fixturePostSlug),
			want: Post{Mark: fixtureMark, Day: post.Day, Series: "", Title: post.Title, Lead: post.Lead, WayIn: postWay},
		},
		{
			name: "a translation", site: site, page: document.PostPage(fixturePostSlug).InLanguage(russian),
			want: Post{
				Mark: fixtureMark, Day: "15 нояб. 2026", Series: "Чтение", Title: "Медленное чтение",
				Lead: "Почему страницу, прочитанную дважды, читают один раз.", WayIn: translatedWay,
			},
		},
		{
			name: "the list in a translation", site: site, page: document.PageOf(document.PageBlog).InLanguage(russian),
			want: List{Mark: fixtureMark, Line: "Что читает автор, и как.", WayIn: "ssh ru/blog@example.test"},
		},
		{name: "a tab of the résumé", site: site, page: resumeTab("experience"), want: nil},
		{name: "a tab of a project", site: site, page: projectPage("tidy-tool", "why"), want: nil},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			composed := fixtures.DocumentOf(test, servedOf(testCase.site), testCase.page)
			got, _ := NewWriter(testCase.site).FactsOf(composed)
			if got != testCase.want {
				test.Errorf("the facts of %+v = %+v, want %+v", testCase.page, got, testCase.want)
			}
		})
	}
}

// A page shows its own card, a tab its page's; a page no preview shows shows none: an older version of a post, which a
// newer stands over, a filtered list, a view of the list, and a draft, which its author alone reads.
func TestAPageShowsItsOwnCardOrItsPagesOrNone(test *testing.T) {
	site := siteOf(fixtures.LoadedContent(test), false)
	drafts := siteOf(fixtures.LoadedContent(test), true)
	none := document.PageOf(document.PageNone)
	cases := []struct {
		name string
		site compose.Site
		page document.Page
		want document.Page
	}{
		{name: "the résumé", site: site, page: document.PageOf(document.PageResume), want: document.PageOf(document.PageResume)},
		{name: "a tab of the résumé", site: site, page: resumeTab("experience"), want: document.PageOf(document.PageResume)},
		{name: "a tab of a project", site: site, page: projectPage("tidy-tool", "why"), want: projectPage("tidy-tool", "")},
		{name: "a post", site: site, page: document.PostPage(fixturePostSlug), want: document.PostPage(fixturePostSlug)},
		{name: "an older version", site: site, page: document.PostPage(fixturePostSlug).OlderVersion(olderVersion), want: none},
		{name: "a filtered list", site: site, page: document.TagFilter("reading"), want: none},
		{name: "a draft", site: drafts, page: document.PostPage(fixtureDraft), want: none},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := PageOf(fixtures.DocumentOf(test, servedOf(testCase.site), testCase.page).Head); got != testCase.want {
				test.Errorf("the card %+v shows = %+v, want %+v", testCase.page, got, testCase.want)
			}
		})
	}
}

// siteOf is the site of a snapshot as it is read on the tests' day, by the public, or by its author with the drafts.
func siteOf(snapshot *content.Snapshot, drafts bool) compose.Site {
	return compose.SiteOf(compose.Prepare(snapshot), compose.Audience{At: fixtures.Day(), Drafts: drafts})
}

// unlistedCopy is a copy of the fixture content whose post is unlisted.
func unlistedCopy(test *testing.T) string {
	test.Helper()
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	fixtures.ReplaceIn(test, folder, fixturePostFile, publishedStatus, unlistedStatus)
	return folder
}

// servedOf are the documents of every page a site serves: the index, the résumé and the projects' pages and their tabs,
// and every blog's lists, filtered or not, posts and older versions.
func servedOf(site compose.Site) []document.Document {
	served := []document.Document{site.Index}
	for _, page := range site.Pages {
		served = append(served, page.PageAndTabs()...)
	}
	for _, blog := range site.Blogs() {
		served = append(served, blog.Served()...)
	}
	return served
}

// resumeTab is a tab of the résumé, by its name.
func resumeTab(tab string) document.Page {
	return document.Page{Kind: document.PageResume, Slug: "", Tab: tab, Revision: "", Translation: document.Untranslated}
}

// projectPage is a project's page, or a tab of it by its name.
func projectPage(slug, tab string) document.Page {
	return document.Page{Kind: document.PageProject, Slug: slug, Tab: tab, Revision: "", Translation: document.Untranslated}
}
