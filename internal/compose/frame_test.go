package compose

import (
	"testing"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// A page of the blog is reached over ssh as the user the session opens it for (D26): a post by its number, in its
// language's folder, a list by its path; any other page by the site's own command, which opens the résumé. The way
// says what it gives in the page's own words, in its language: a post, the blog or the résumé.
func TestAPageIsReachedOverSSHByTheNameItsSessionOpens(test *testing.T) {
	site := SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay())
	cases := []struct {
		name  string
		page  document.Page
		want  string
		gives string
	}{
		{name: "the résumé", page: document.PageOf(document.PageResume), want: "ssh example.test", gives: "The résumé as an app in a terminal."},
		{name: "a project", page: document.Page{Kind: document.PageProject, Slug: "tidy-tool", Tab: "", Revision: "", Translation: ""}, want: "ssh example.test", gives: "The résumé as an app in a terminal."},
		{name: "the list", page: document.PageOf(document.PageBlog), want: "ssh blog@example.test", gives: "The blog in a terminal."},
		{name: "the list by series", page: document.PageOf(document.PageBlogSeries), want: "ssh blog/series@example.test", gives: "The blog in a terminal."},
		{name: "a post", page: document.PostPage(fixturePostSlug), want: "ssh blog/1@example.test", gives: "This post in a terminal."},
		{name: "a translation", page: document.PostPage(fixturePostSlug).InLanguage("ru"), want: "ssh ru/blog/1@example.test", gives: "Этот пост в терминале."},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			head := fixtures.DocumentOf(test, site.served, testCase.page).Head
			frame := site.Frames()[testCase.page.Translation]
			if got := frame.SSHCommandTo(head); got != testCase.want {
				test.Errorf("the ssh command to %+v = %q, want %q", testCase.page, got, testCase.want)
			}
			way := frame.SSHWayTo(head)
			if way.Command.Text != testCase.want || way.Gives != testCase.gives {
				test.Errorf("the ssh way to %+v = %q giving %q, want %q giving %q", testCase.page, way.Command.Text, way.Gives, testCase.want, testCase.gives)
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
