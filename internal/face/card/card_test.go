package card

import (
	"slices"
	"testing"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/golden"
)

func TestMain(tests *testing.M) {
	golden.Main(tests)
}

// The fixture's post, its folder and its file, and where it says it is published
const (
	fixturePostSlug = "reading-slowly"
	fixturePostFile = "posts/reading-slowly/post.en.md"
	publishedStatus = "status: published"
	unlistedStatus  = "status: unlisted"
	fixtureDraft    = "notes-on-tools"
	olderVersion    = "2026-11-10"
	russian         = "ru"
	fixtureWayIn    = "ssh example.test"
)

// Each page a preview shows has the facts of its card, read off its document in its language's words: the résumé, a
// project, the blog's list by year and by series alike, a post, its series the one its page shows, an unlisted post,
// which a preview shows too, and a translation. A tab has none of its own: its page's card stands for it.
func TestEachPageAPreviewShowsHasTheFactsOfItsCard(test *testing.T) {
	site := siteOf(fixtures.LoadedContent(test), false)
	unlisted := siteOf(fixtures.LoadedFrom(test, unlistedCopy(test)), false)
	post := Post{
		Blog: "blog", Day: "10 Nov 2026", Series: "Reading", Title: "Reading slowly", Lead: "Why a page read twice is read once.",
		WayIn: fixtureWayIn,
	}
	list := List{Title: "blog", Line: "What the writer reads, and how.", WayIn: fixtureWayIn}
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
		{name: "the list by series", site: site, page: document.PageOf(document.PageBlogSeries), want: list},
		{name: "a post", site: site, page: document.PostPage(fixturePostSlug), want: post},
		{
			name: "an unlisted post, whose page shows no series", site: unlisted, page: document.PostPage(fixturePostSlug),
			want: Post{Blog: post.Blog, Day: post.Day, Series: "", Title: post.Title, Lead: post.Lead, WayIn: fixtureWayIn},
		},
		{
			name: "a translation", site: site, page: document.PostPage(fixturePostSlug).InLanguage(russian),
			want: Post{
				Blog: "блог", Day: "15 нояб. 2026", Series: "Чтение", Title: "Медленное чтение",
				Lead: "Почему страницу, прочитанную дважды, читают один раз.", WayIn: fixtureWayIn,
			},
		},
		{
			name: "the list in a translation", site: site, page: document.PageOf(document.PageBlog).InLanguage(russian),
			want: List{Title: "блог", Line: "Что читает автор, и как.", WayIn: fixtureWayIn},
		},
		{name: "a tab of the résumé", site: site, page: resumeTab("experience"), want: nil},
		{name: "a tab of a project", site: site, page: projectPage("tidy-tool", "why"), want: nil},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			composed := fixtures.DocumentOf(test, servedOf(testCase.site), testCase.page)
			got, _ := NewWriter(frameOf(testCase.site, testCase.page)).FactsOf(composed)
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
			if got, _ := PageOf(fixtures.DocumentOf(test, servedOf(testCase.site), testCase.page).Head); got != testCase.want {
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

// frameOf is the frame of a page's language: its blog's, the site's own for a page of no blog.
func frameOf(site compose.Site, page document.Page) compose.Frame {
	blogs := site.Blogs()
	found := slices.IndexFunc(blogs, func(blog compose.Blog) bool { return blog.Frame.Translation == page.Translation })
	return blogs[found].Frame
}

// resumeTab is a tab of the résumé, by its name.
func resumeTab(tab string) document.Page {
	return document.Page{Kind: document.PageResume, Slug: "", Tab: tab, Revision: "", Translation: document.Untranslated}
}

// projectPage is a project's page, or a tab of it by its name.
func projectPage(slug, tab string) document.Page {
	return document.Page{Kind: document.PageProject, Slug: slug, Tab: tab, Revision: "", Translation: document.Untranslated}
}
