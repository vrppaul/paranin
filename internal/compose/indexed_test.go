package compose

import (
	"slices"
	"testing"
	"time"

	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/routes"
)

// The day a page has none of, as time.DateOnly writes the zero time
const noDay = "0001-01-01"

// The site names each page a search engine may index once, by the day its text last changed: a post its current
// version's, a list the latest current version's day among the posts it names, the résumé, a project's page and their
// tabs none; never a page no browser reads at an address of its own, as the index, an unlisted post, a draft, an older
// version or a filtered list.
func TestTheSiteNamesEachPageIndexedByItsDay(test *testing.T) {
	posts := []writtenPost{
		{slug: "fresh", id: 3, status: "published", published: "2026-11-11T09:00+01:00", tags: ""},
		{slug: "hidden", id: 4, status: "unlisted", published: later, tags: ""},
	}
	site := SiteOf(Prepare(withPostsAndSeries(test, posts, nil)), Audience{At: fixtures.Day(), Drafts: true})
	seen := map[string][]string{}
	for _, indexed := range site.IndexedPages() {
		path := routes.Path(indexed.Page, routes.FormatChosen)
		seen[path] = append(seen[path], indexed.Changed.Format(time.DateOnly))
	}
	cases := []struct {
		name string
		path string
		want []string
	}{
		{name: "the résumé", path: "/", want: []string{noDay}},
		{name: "a tab of the résumé", path: "/experience", want: []string{noDay}},
		{name: "a project's page", path: "/p/tidy-tool", want: []string{noDay}},
		{name: "a post updated, by its current version's day", path: "/blog/reading-slowly", want: []string{"2026-11-12"}},
		{name: "a post never updated, by the day it came out", path: "/blog/fresh", want: []string{"2026-11-11"}},
		{name: "a translation, by its own version's day", path: "/ru/blog/reading-slowly", want: []string{"2026-11-15"}},
		{name: "the list by year, by an older post updated since the newest came out", path: "/blog", want: []string{"2026-11-12"}},
		{name: "the list by series, by an older post updated since the newest came out", path: "/blog/series", want: []string{"2026-11-12"}},
		{name: "the list in a language, by its posts' latest version", path: "/ru/blog", want: []string{"2026-11-15"}},
		{name: "a page no browser reads at an address of its own, as the index", path: "", want: nil},
		{name: "an unlisted post", path: "/blog/hidden", want: nil},
		{name: "a draft its author reads", path: "/blog/notes-on-tools", want: nil},
		{name: "an older version", path: "/blog/reading-slowly/2026-11-10", want: nil},
		{name: "a filtered list", path: "/blog?tag=reading", want: nil},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := seen[testCase.path]; !slices.Equal(got, testCase.want) {
				test.Errorf("the site names %q by the days %q, want %q", testCase.path, got, testCase.want)
			}
		})
	}
}
