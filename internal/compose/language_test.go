package compose

import (
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The fixture's Russian translation, and the language it is in
const (
	translationFile     = "posts/reading-slowly/post.ru.md"
	russian             = "ru"
	translationStatus   = "status: published"
	translationOutLater = "published: 2026-12-01T10:00+03:00"
	translationOut      = "published: 2026-11-15T10:00+03:00"
	// the sources the translation names the fixture post's text by: as its revision of the 10th keeps it, the
	// fixture's, and as it stands
	fixtureRevisionSource = "en@59b3c768ccba"
	fixtureCurrentSource  = "en@520f97557153"
	// the Russian blog's place among the site's translations, the fixture's one
	russianBlog = 0
)

// fixtureLanguages are the languages the fixture has words for, by their tags, in the order the site's blogs stand.
func fixtureLanguages() []string {
	return []string{"en", russian}
}

// seenBlog is the blog in a language as a test reads it.
type seenBlog struct {
	lists     []document.Page
	listTitle string
	posts     []document.Page
	languages []string
	feeds     []document.Feed
	frameList document.Page
	locale    string
}

// A translation is a blog of its own language: its lists, posts, feeds and frame in its folder and its words alone.
func TestATranslationIsABlogOfItsOwnLanguage(test *testing.T) {
	site := SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay())
	blogLanguages := []string{}
	for _, composed := range site.Blogs() {
		blogLanguages = append(blogLanguages, composed.Lists[0].Head.Language)
	}
	if !slices.Equal(blogLanguages, fixtureLanguages()) {
		test.Fatalf("the site's blogs are in %q, want %q", blogLanguages, fixtureLanguages())
	}
	got := seenBlogOf(site.Translations[russianBlog])
	post := document.PostPage(fixturePostSlug).InLanguage(russian)
	want := seenBlog{
		lists:     []document.Page{document.PageOf(document.PageBlog).InLanguage(russian), document.PageOf(document.PageBlogSeries).InLanguage(russian)},
		listTitle: "блог",
		posts:     []document.Page{post},
		languages: []string{russian},
		feeds:     []document.Feed{{Series: "", Translation: russian}, {Series: fixtureSeriesName, Translation: russian}},
		frameList: document.PageOf(document.PageBlog).InLanguage(russian),
		locale:    "ru_RU",
	}
	if !reflect.DeepEqual(got, want) {
		test.Errorf("the Russian blog = %+v, want %+v", got, want)
	}
}

// seenBlogOf is a blog as a test reads it.
func seenBlogOf(composed Blog) seenBlog {
	seen := seenBlog{
		lists: []document.Page{}, listTitle: document.PlainText(composed.Lists[0].Head.Title), posts: []document.Page{}, languages: []string{},
		feeds: []document.Feed{}, frameList: composed.Frame.ListWords.List, locale: composed.Frame.Locale,
	}
	for _, list := range composed.Lists {
		seen.lists = append(seen.lists, list.Head.Page)
	}
	for _, post := range composed.Posts {
		seen.posts, seen.languages = append(seen.posts, post.Head.Page), append(seen.languages, post.Head.Language)
	}
	for _, feed := range composed.Feeds {
		seen.feeds = append(seen.feeds, feed.Target)
	}
	return seen
}

// A post's Switch offers the language read and each other the lists name the post in, none where it is read in one.
func TestAPostsSwitchOffersEachLanguageTheListsNameThePostIn(test *testing.T) {
	original, translated := document.PostPage(fixturePostSlug), document.PostPage(fixturePostSlug).InLanguage(russian)
	english := document.Language{Tag: "en", Name: "English", Locale: "en_US", Page: original, Read: false}
	russianRead := document.Language{Tag: russian, Name: "Русский", Locale: "ru_RU", Page: translated, Read: true}
	englishRead, russianOffered := english, russianRead
	englishRead.Read, russianOffered.Read = true, false
	bothFromEnglish := []document.Language{englishRead, russianOffered}
	cases := []struct {
		name         string
		status       string
		audience     Audience
		wantOriginal []document.Language
		wantOwn      []document.Language
	}{
		{name: "published and updated", status: translationStatus + "\nupdated: 2026-11-16", audience: publicOnTheDay(), wantOriginal: bothFromEnglish, wantOwn: []document.Language{english, russianRead}},
		{name: "unlisted", status: "status: unlisted", audience: publicOnTheDay(), wantOriginal: nil, wantOwn: []document.Language{english, russianRead}},
		{name: "a preview", status: "status: preview", audience: publicOnTheDay(), wantOriginal: nil, wantOwn: []document.Language{english, russianRead}},
		{name: "a draft its author reads", status: "status: draft", audience: Audience{At: fixtures.Day(), Drafts: true}, wantOriginal: nil, wantOwn: []document.Language{english, russianRead}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			site := SiteOf(Prepare(translationWith(test, translationStatus, testCase.status)), testCase.audience)
			own := translationRead(test, site.Translations[russianBlog], translated)
			if got := postNamed(test, site.Own.Posts, fixturePostSlug).Head.Languages; !reflect.DeepEqual(got, testCase.wantOriginal) {
				test.Errorf("the original's Switch with the translation %s = %+v, want %+v", testCase.name, got, testCase.wantOriginal)
			}
			if got := own.Head.Languages; !reflect.DeepEqual(got, testCase.wantOwn) {
				test.Errorf("the translation's own Switch, %s = %+v, want %+v", testCase.name, got, testCase.wantOwn)
			}
		})
	}
}

// translationRead is the document of a translation as its reader reads it: among the posts served in its language, or
// among its previews.
func translationRead(test *testing.T, composed Blog, page document.Page) document.Document {
	test.Helper()
	documents := slices.Clone(composed.Posts)
	for _, previewed := range composed.Previews {
		documents = append(documents, previewed.Post)
	}
	return fixtures.DocumentOf(test, documents, page)
}

// translationWith is the fixture content with a piece of its Russian translation written otherwise.
func translationWith(test *testing.T, original, replacement string) *content.Snapshot {
	test.Helper()
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	fixtures.ReplaceIn(test, folder, translationFile, original, replacement)
	snapshot, err := content.Load(os.DirFS(folder))
	if err != nil {
		test.Fatalf("load the fixture with %q: %v", replacement, err)
	}
	return snapshot
}

// A translation older than its original says so under its day, leading to the original as its Switch does; one made
// from the original as it stands says nothing, nor one whose original no list names, which the Switch leads nowhere.
func TestATranslationOlderThanItsOriginalLeadsToTheOriginal(test *testing.T) {
	cases := []struct {
		name     string
		snapshot *content.Snapshot
		want     *document.Outdated
	}{
		{
			name: "made from an older version", snapshot: fixtures.LoadedContent(test),
			want: &document.Outdated{Original: document.PostPage(fixturePostSlug), Language: "en", Words: "старше оригинала"},
		},
		{name: "made from the original as it stands", snapshot: translationWith(test, fixtureRevisionSource, fixtureCurrentSource), want: nil},
		{name: "of an original no list names", snapshot: changedFixture(test, fixturePostUnlisted), want: nil},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			site := SiteOf(Prepare(testCase.snapshot), publicOnTheDay())
			read := fixtures.DocumentOf(test, site.Translations[russianBlog].Posts, document.PostPage(fixturePostSlug).InLanguage(russian))
			version, isOut := read.Head.Version.(document.VersionOut)
			if !isOut || !reflect.DeepEqual(version.Outdated, testCase.want) {
				test.Errorf("the translation's version, %s = %+v, want one out saying %+v", testCase.name, read.Head.Version, testCase.want)
			}
		})
	}
}

// fixturePostUnlisted makes the fixture's post unlisted, which its address serves and no list names.
//
//wrapper:a case of the table names the change it writes
func fixturePostUnlisted(test *testing.T, folder string) {
	fixtures.ReplaceIn(test, folder, fixturePostFile, publishedStatus, "status: unlisted")
}

// A blog in a language writes a day in its language's months, and a series in the words its translation of it writes.
func TestABlogWritesDaysAndSeriesInItsLanguage(test *testing.T) {
	site := SiteOf(Prepare(fixtures.LoadedContent(test)), publicOnTheDay())
	day := fixtures.MomentOf(test, "2026-11-15T10:00+03:00")
	russianPost := fixtures.DocumentOf(test, site.Translations[russianBlog].Posts, document.PostPage(fixturePostSlug).InLanguage(russian))
	got := []string{
		site.Own.Frame.Days.Of(day), site.Translations[russianBlog].Frame.Days.Of(day),
		document.PlainText(postNamed(test, site.Own.Posts, fixturePostSlug).Head.Post.Series.Title), document.PlainText(russianPost.Head.Post.Series.Title),
	}
	if want := []string{"15 Nov 2026", "15 нояб. 2026", "Reading", "Чтение"}; !slices.Equal(got, want) {
		test.Errorf("the day and the series' title in English and in Russian = %q, want %q", got, want)
	}
}

// The site changes as a translation dated later comes out, as it does when a post does.
func TestATranslationDatedLaterChangesTheSiteWhenOut(test *testing.T) {
	site := SiteOf(Prepare(translationWith(test, translationOut, translationOutLater)), publicOnTheDay())
	want := fixtures.MomentOf(test, "2026-12-01T10:00+03:00")
	if russianPosts := len(site.Translations[russianBlog].Posts); !site.Until.Equal(want) || russianPosts != 0 {
		test.Errorf("the site with a translation out later holds until %s with %d Russian posts, want %s and none", site.Until, russianPosts, want)
	}
}
