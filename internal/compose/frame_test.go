package compose

import (
	"testing"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// A page of the blog is reached over ssh as the user the session opens it for (D26): a post by its number, in its
// language's folder, a list by its path; any other page by the site's own command, which opens the résumé.
func TestAPageIsReachedOverSSHByTheNameItsSessionOpens(test *testing.T) {
	site := SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay())
	cases := []struct {
		name string
		page document.Page
		want string
	}{
		{name: "the résumé", page: document.PageOf(document.PageResume), want: "ssh example.test"},
		{name: "a project", page: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: "", Revision: "", Translation: ""}, want: "ssh example.test"},
		{name: "the list", page: document.PageOf(document.PageBlog), want: "ssh blog@example.test"},
		{name: "the list by series", page: document.PageOf(document.PageBlogSeries), want: "ssh blog/series@example.test"},
		{name: "a post", page: document.PostPage(fixturePostSlug), want: "ssh blog/1@example.test"},
		{name: "a translation", page: document.PostPage(fixturePostSlug).InLanguage("ru"), want: "ssh ru/blog/1@example.test"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			head := fixtures.DocumentOf(test, site.served, testCase.page).Head
			if got := site.Own.Frame.SSHCommandTo(head); got != testCase.want {
				test.Errorf("the ssh command to %+v = %q, want %q", testCase.page, got, testCase.want)
			}
		})
	}
}

// The site is named after its person, as the profile writes the name, its marks read and gone, in every language's
// frame alike.
func TestTheSiteIsNamedAfterItsPerson(test *testing.T) {
	marked := *fixtures.LoadedContent(test)
	marked.Profile.Person.Name = "==Mira== Example"
	for _, composed := range SiteOf(Prepare(&marked), publicOnTheDay()).Blogs() {
		if got, want := composed.Frame.SiteName, "Mira Example"; got != want {
			test.Errorf("the site's name in the frame of %q = %q, want %q", composed.Frame.Language, got, want)
		}
	}
}
