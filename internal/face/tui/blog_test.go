package tui

import (
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
	"github.com/vrppaul/paranin/internal/look"
)

// The footer's windows of the blog's session on the fixture: its list shown alone, its list beside the post read, and
// the post read
const (
	listAlone      = "1:posts*"
	listBesidePost = "1:posts"
	postRead       = "2:reading-slowly*"
	resumeOpened   = "1:about*"
)

// The footer's windows of the blog's session in Russian, the fixture's translation: its list shown alone, and its list
// beside the post read
const (
	russianListAlone      = "1:посты*"
	russianListBesidePost = "1:посты"
)

// openedModel is a session on what the fixture draws, opened on the page a name reaches, of a size.
func openedModel(test *testing.T, opening string, width, height int) Model {
	return New(sessionDrawing(fixtures.LoadedContent(test), false), opening, width, height, fixtures.Day())
}

// A session opens on the page of the blog its SSH user's name reaches (D26): the blog's list by blog, a post by its
// slug and by its number, the list beside it in the first window; a name in the blog's folder that reaches no post,
// a draft's among them, opens the list and says it finds no page by the name; any other name opens the résumé. A name
// in a language's folder opens that language's blog, framed in its words, as its address does on the web.
func TestASessionOpensOnTheBlogsPageItsNameReaches(test *testing.T) {
	cases := []struct {
		name  string
		user  string
		holds []string
		lacks []string
	}{
		{name: "the list", user: "blog", holds: []string{"[example.test] [blog]", listAlone}, lacks: []string{"2:"}},
		{name: "a post by its number", user: "blog/1", holds: []string{listBesidePost, postRead}, lacks: nil},
		{name: "a post by its slug", user: "blog/reading-slowly", holds: []string{listBesidePost, postRead}, lacks: nil},
		{name: "an older version by its address", user: "blog/reading-slowly/2026-11-10", holds: []string{listBesidePost, postRead}, lacks: nil},
		{name: "a number no post has", user: "blog/9", holds: []string{listAlone, `E447: Can't find file "blog/9" in path`}, lacks: nil},
		{name: "a draft's number", user: "blog/2", holds: []string{listAlone, `E447: Can't find file "blog/2" in path`}, lacks: nil},
		{name: "the list in Russian", user: "ru/blog", holds: []string{"[example.test] [блог]", russianListAlone}, lacks: []string{"2:"}},
		{name: "a translation by its number", user: "ru/blog/1", holds: []string{russianListBesidePost, postRead}, lacks: nil},
		{name: "a translation by its slug", user: "ru/blog/reading-slowly", holds: []string{russianListBesidePost, postRead}, lacks: nil},
		{
			name: "a number no translation has", user: "ru/blog/9", holds: []string{russianListAlone, `E447: Can't find file "ru/blog/9" in path`},
			lacks: nil,
		},
		{name: "a language the site has no words for", user: "de/blog", holds: []string{resumeOpened}, lacks: []string{"E447"}},
		{name: "a login name", user: "reader", holds: []string{resumeOpened}, lacks: []string{"posts", "E447"}},
		{name: "a project's slug, which :e reads and no address of the blog is", user: "tidy-tool", holds: []string{resumeOpened}, lacks: []string{"E447"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			footer := atTheFoot(screenOf(test, openedModel(test, testCase.user, commonWidth, commonHeight)))
			for _, want := range testCase.holds {
				if !strings.Contains(footer, want) {
					test.Errorf("opened by %q the footer = %q, want it to hold %q", testCase.user, footer, want)
				}
			}
			for _, unwanted := range testCase.lacks {
				if strings.Contains(footer, unwanted) {
					test.Errorf("opened by %q the footer = %q, want it not to hold %q", testCase.user, footer, unwanted)
				}
			}
		})
	}
}

// A name from outside reaches the screen as letters alone: a control character in an SSH user's name, as a terminal's
// title written with OSC 0, never reaches the terminal, and the command line says the name without it.
func TestANameFromOutsideReachesTheScreenAsLetters(test *testing.T) {
	view := openedModel(test, "blog/\x1b]0;owned\x07x", commonWidth, commonHeight).View().Content
	if strings.Contains(view, "\x1b]0;") || !strings.Contains(ansi.Strip(view), `E447: Can't find file "blog/]0;ownedx" in path`) {
		test.Errorf("opened by a name holding OSC 0 the screen = %q, want the name's letters alone", view)
	}
}

// A session handed no list of the blog opens a post by its name with no list beside it, q on it going back to the
// résumé; a name of the blog that reaches no page opens the résumé, saying so; and :e of the list it does not hold
// says it finds none.
func TestASessionWithNoListOpensAPostAlone(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	drawn := sessionDrawing(snapshot, false)
	documents := []document.Document{compose.Resume(snapshot, nil), fixtures.DocumentOf(test, drawn.Documents, document.Page{Kind: document.PagePost, Slug: "reading-slowly", Tab: ""})}
	drawing := Drawing{Documents: documents, Frames: drawn.Frames, Site: fixtures.Site(snapshot)}
	cases := []struct {
		name  string
		model Model
		holds []string
	}{
		{name: "a post, closed", model: pressed(test, New(drawing, "blog/1", commonWidth, commonHeight, fixtures.Day()), "q"), holds: []string{resumeOpened}},
		{name: "a name of no page", model: New(drawing, "blog/9", commonWidth, commonHeight, fixtures.Day()), holds: []string{resumeOpened, `E447: Can't find file "blog/9" in path`}},
		{
			name:  "the list it does not hold",
			model: commanded(test, New(drawing, "", commonWidth, commonHeight, fixtures.Day()), "e blog"),
			holds: []string{resumeOpened, `E447: Can't find file "blog" in path`},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			below := atTheFoot(screenOf(test, testCase.model))
			for _, want := range testCase.holds {
				if !strings.Contains(below, want) {
					test.Errorf("the foot = %q, want it to hold %q", below, want)
				}
			}
		})
	}
}

// :e opens the blog's pages from the résumé by the names a terminal reaches them by: the list by year, a post by its
// number and its slug, and the list by series, each in Russian too by the name in its folder; and says it finds no post
// of a number no post has.
func TestEditOpensTheBlogsPages(test *testing.T) {
	cases := []struct {
		name   string
		text   string
		holds  string
		footer string
	}{
		{name: "the list", text: "e blog", holds: "Reading slowly", footer: listAlone},
		{name: "a post by its number", text: "e blog/1", holds: "What a reader keeps", footer: postRead},
		{name: "a post by its slug", text: "e blog/reading-slowly", holds: "What a reader keeps", footer: postRead},
		{name: "the list by series", text: "e blog/series", holds: "1 of 2 out", footer: listAlone},
		{name: "the list in Russian", text: "e ru/blog", holds: "1 пост", footer: russianListAlone},
		{name: "a translation by its number", text: "e ru/blog/1", holds: "Что остаётся у читателя", footer: postRead},
		{name: "the list by series in Russian", text: "e ru/blog/series", holds: "вышла 1 из 2", footer: russianListAlone},
		{name: "a number no post has", text: "e blog/9", holds: `E447: Can't find file "blog/9" in path`, footer: resumeOpened},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, commanded(test, fixtureModel(test, commonWidth, tallHeight), testCase.text))
			if !strings.Contains(screen, testCase.holds) || !strings.Contains(footerOf(screen), testCase.footer) {
				test.Errorf("after :%s the screen =\n%s\nwant %q on it and %q in the footer", testCase.text, screen, testCase.holds, testCase.footer)
			}
		})
	}
}

// The menu names the blog's places by their kinds, those of the tab chosen alone: on a post its headings and its other
// languages; on the list its posts, its tags and its series, a series named as a tag standing beside it, and not the
// headings of the post read beside it, the post read going by its window; and after :e and a space the blog's lists and
// each post they list, in every language, by the names a terminal reaches them by, never a filter of the list, which
// Tab does not offer.
func TestTheMenuNamesTheBlogsPlacesByTheirKinds(test *testing.T) {
	beside := pressed(test, openedModel(test, "blog/1", commonWidth, commonHeight), "h")
	cases := []struct {
		name  string
		model Model
		text  string
		after []string
		holds []string
		lacks []string
	}{
		{name: "a post's headings", model: openedModel(test, "blog/1", commonWidth, commonHeight), text: "a", after: nil, holds: []string{`afterwards +heading`}, lacks: nil},
		{name: "a post's other language", model: openedModel(test, "blog/1", commonWidth, commonHeight), text: "r", after: nil, holds: []string{`ru +language`}, lacks: nil},
		{name: "the list beside the post", model: beside, text: "a", after: nil, holds: nil, lacks: []string{`afterwards`}},
		{
			name: "the list's post read beside it, by its window", model: beside, text: "r", after: nil, holds: []string{`reading-slowly +tab`},
			lacks: []string{`reading-slowly +post`},
		},
		{
			name: "the list's posts, tags and series, a tag and a series named alike each by its kind", model: openedModel(test, "blog", commonWidth, commonHeight),
			text: "re", after: nil, holds: []string{`reading-slowly +post`, `reading +tag`, `reading +series`}, lacks: nil,
		},
		{
			name: "the series Tab chose beside the tag named alike, which Enter goes to", model: openedModel(test, "blog", commonWidth, commonHeight),
			text: "rea", after: []string{"tab", "tab", "tab", "enter"}, holds: []string{`1 of 2 out`}, lacks: nil,
		},
		{name: "the list's other tag", model: openedModel(test, "blog", commonWidth, commonHeight), text: "t", after: nil, holds: []string{`terminals +tag`}, lacks: nil},
		{name: "the list's series", model: handSeriesModel(test, commonHeight), text: "ea", after: nil, holds: []string{`early +series`}, lacks: nil},
		{
			name: "the blog's pages :e opens", model: openedModel(test, "", commonWidth, commonHeight), text: "e b", after: nil,
			holds: []string{`blog +list`, `blog/series +list`, `blog/reading-slowly +post`}, lacks: []string{`blog/1`, `blog\?tag`},
		},
		{
			name: "the blog's pages :e opens in Russian", model: openedModel(test, "", commonWidth, commonHeight), text: "e r", after: nil,
			holds: []string{`ru/blog +list`, `ru/blog/series +list`, `ru/blog/reading-slowly +post`}, lacks: []string{`ru/blog/1`},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			above := aboveTheCommandLine(screenOf(test, pressed(test, typed(test, pressed(test, testCase.model, ":"), testCase.text), testCase.after...)))
			for _, want := range testCase.holds {
				if !regexp.MustCompile(want).MatchString(above) {
					test.Errorf("after :%s the screen above the command line =\n%s\nwant it to hold %q", testCase.text, above, want)
				}
			}
			for _, unwanted := range testCase.lacks {
				if regexp.MustCompile(unwanted).MatchString(above) {
					test.Errorf("after :%s the screen above the command line =\n%s\nwant it not to hold %q", testCase.text, above, unwanted)
				}
			}
		})
	}
}

// A second post the tests write beside the fixture's, of a tag of its own: its file, and its text; and its Russian
// translation's, made from a text it never had, which the load warns of and takes
const (
	secondPostFile = "posts/second/post.en.md"
	secondPostText = "---\nid: 3\ntitle: The second post\ndescription: About the second.\nstatus: published\npublished: 2026-11-12T09:00+01:00\n" +
		"tags: [tools]\n---\nIts text.\n"
	secondTranslationFile = "posts/second/post.ru.md"
	secondTranslationText = "---\ntitle: Второй пост\ndescription: О втором.\nstatus: published\npublished: 2026-11-16T09:00+03:00\n" +
		"source: en@000000000000\n---\nЕго текст.\n"
)

// fixtureFile is a file a test writes into a copy of the fixture's content: its path there, and its text.
type fixtureFile struct {
	path string
	text string
}

// secondPost is the second post's file, in English.
func secondPost() fixtureFile {
	return fixtureFile{path: secondPostFile, text: secondPostText}
}

// contentWith is the fixture's content with files written beside its own.
func contentWith(test *testing.T, files ...fixtureFile) *content.Snapshot {
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	for _, written := range files {
		fixtures.WriteIn(test, folder, written.path, written.text)
	}
	snapshot, err := content.Load(os.DirFS(folder))
	if err != nil {
		test.Fatalf("load the fixture with %d files beside its own: %v", len(files), err)
	}
	return snapshot
}

// A tag's name shows the list filtered by it in the list's window, and on that list, where it is the filter, the list
// whole again; another tag's filters the list anew. :e opens the list filtered by a tag, by the name a terminal reaches
// it by.
func TestATagsPlaceFiltersTheList(test *testing.T) {
	list := New(sessionDrawing(contentWith(test, secondPost()), false), "blog", commonWidth, tallHeight, fixtures.Day())
	cases := []struct {
		name     string
		commands []string
		holds    []string
		lacks    []string
	}{
		{name: "a tag", commands: []string{"tools"}, holds: []string{"The second post"}, lacks: []string{"Reading slowly"}},
		// the chip the list is filtered by, read beside the head's tag before it, since a tile shows #tools too
		{name: "the tag shown in the list's head", commands: []string{"tools"}, holds: []string{"#terminals   #tools"}, lacks: nil},
		{name: "the tag filtering it", commands: []string{"tools", "tools"}, holds: []string{"The second post", "Reading slowly"}, lacks: nil},
		{name: "another tag", commands: []string{"tools", "terminals"}, holds: []string{"Reading slowly"}, lacks: []string{"The second post"}},
		{name: "a tag's filter by :e", commands: []string{"e blog?tag=tools"}, holds: []string{"The second post"}, lacks: []string{"Reading slowly"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := list
			for _, text := range testCase.commands {
				model = commanded(test, model, text)
			}
			screen := screenOf(test, model)
			for _, want := range testCase.holds {
				if !strings.Contains(screen, want) || !strings.Contains(footerOf(screen), listAlone) {
					test.Errorf("after :%s the screen =\n%s\nwant %q on it and %q in the footer", strings.Join(testCase.commands, ", :"), screen, want, listAlone)
				}
			}
			for _, unwanted := range testCase.lacks {
				if strings.Contains(screen, unwanted) {
					test.Errorf("after :%s the screen =\n%s\nwant no %q on it", strings.Join(testCase.commands, ", :"), screen, unwanted)
				}
			}
		})
	}
}

// Enter on a heading's name brings it to the page's top; on a post's or a part's name opens that post beside the
// list; on a version's, by its day, that version, in the post's window, an older one from the current and back; on a
// language's, by its tag, the post in that language, its blog's list beside it.
func TestEnterOnABlogsPlaceGoesThere(test *testing.T) {
	cases := []struct {
		name   string
		model  Model
		text   string
		top    string
		footer string
	}{
		{name: "a heading", model: openedModel(test, "blog/1", commonWidth, commonHeight), text: "in-a-terminal", top: "In a terminal", footer: postRead},
		{name: "a post from the list", model: openedModel(test, "blog", commonWidth, commonHeight), text: "reading-slowly", top: "Reading slowly", footer: postRead},
		{name: "another part", model: pressed(test, commanded(test, handSeriesModel(test, tallHeight), "late"), "enter"), text: "fifth", top: "The fifth post", footer: "2:fifth*"},
		{name: "an older version", model: openedModel(test, "blog/1", commonWidth, commonHeight), text: "2026-11-10", top: "Slow reading", footer: postRead},
		{
			name: "the current version", model: openedModel(test, "blog/reading-slowly/2026-11-10", commonWidth, commonHeight), text: "2026-11-12",
			top: "Reading slowly", footer: postRead,
		},
		{name: "a translation", model: openedModel(test, "blog/1", commonWidth, commonHeight), text: "ru", top: "Медленное чтение", footer: russianListBesidePost},
		{name: "the original", model: openedModel(test, "ru/blog/1", commonWidth, commonHeight), text: "en", top: "Reading slowly", footer: listBesidePost},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, commanded(test, testCase.model, testCase.text))
			if top := firstRowOf(screen); !strings.Contains(top, testCase.top) || !strings.Contains(footerOf(screen), testCase.footer) {
				test.Errorf("after :%s the screen =\n%s\nwant %q on its first row and %q in the footer", testCase.text, screen, testCase.top, testCase.footer)
			}
		})
	}
}

// A series' name opens the list by series at its group, the keys on its first part, which Enter reads: on a screen
// that shows the whole list, where the tile the keys stood on stays in sight, and on a short one, where the group
// shows.
func TestASeriesNameOpensItsGroup(test *testing.T) {
	// a screen too short for the list by series whole
	const shortScreen = 14
	cases := []struct {
		name   string
		height int
	}{
		{name: "the whole list in sight", height: tallHeight},
		{name: "a short screen", height: shortScreen},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := commanded(test, handSeriesModel(test, testCase.height), "late")
			if seen := screenOf(test, model); !strings.Contains(seen, "4 Nov 2026") {
				test.Errorf("after :late the screen =\n%s\nwant the late series' first tile in sight", seen)
			}
			if footer := footerOf(screenOf(test, pressed(test, model, "enter"))); !strings.Contains(footer, "2:fourth*") {
				test.Errorf("after :late and Enter the footer = %q, want the fourth post read", footer)
			}
		})
	}
}

// handSeriesModel is a session of a height on a résumé and a blog a test builds by hand, opened on its list by year:
// five posts, the first three the early series, the last two the late one, the fourth leading to the fifth, the next
// part; the list by year naming each post's series; and the list by series holding the early series, then the late.
func handSeriesModel(test *testing.T, height int) Model {
	resume := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, textOf("home")), Blocks: nil}
	fourth, fifth := document.Page{Kind: document.PagePost, Slug: "fourth", Tab: ""}, document.Page{Kind: document.PagePost, Slug: "fifth", Tab: ""}
	late := &document.Series{Title: textOf("Late"), Parts: []document.Part{
		{Number: 1, Title: textOf("The fourth post"), State: document.PartReading, Page: fourth},
		{Number: 2, Title: textOf("The fifth post"), State: document.PartNext, Page: fifth},
	}}
	postPlaces := map[string]*document.PostPlace{"fourth": {Series: late, Neighbours: document.Neighbours{Before: nil, After: nil}, Discussions: nil, Listed: true}}
	seriesOf := map[string]string{"first": "early", "second": "early", "third": "early", "fourth": "late", "fifth": "late"}
	documents := []document.Document{resume}
	listed := []document.ListedPost{}
	bySeries := document.Page{Kind: document.PageBlogSeries, Slug: "", Tab: ""}
	for index, slug := range slices.Backward(handPostSlugs()) {
		page := document.Page{Kind: document.PagePost, Slug: slug, Tab: ""}
		head := fixtures.HeadOf(page, textOf("The "+slug+" post"))
		head.Post = postPlaces[slug]
		documents = append(documents, document.Document{Head: head, Blocks: nil})
		out := document.PostOut{Page: page, Published: time.Date(2026, time.November, len(handPostSlugs())-index, 0, 0, 0, 0, time.UTC)}
		listed = append(listed, document.ListedPost{
			Readiness: out, Title: textOf("The " + slug + " post"), Lead: nil, Part: nil, Series: nil,
			SeriesGroup: document.Anchor{Page: bySeries, Name: seriesOf[slug]}, Tags: nil,
		})
	}
	byYear := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageBlog, Slug: "", Tab: ""}, textOf("blog")), Blocks: []document.Block{
		document.Section{Title: textOf(olderYear), Name: "", Blocks: []document.Block{document.Posts{Line: nil, Count: textOf("five"), List: listed}}},
	}}
	series := document.Document{Head: fixtures.HeadOf(bySeries, textOf("blog")), Blocks: []document.Block{
		document.Section{Title: textOf("Early"), Name: "early", Blocks: []document.Block{document.Posts{Line: nil, Count: textOf("three"), List: listed[:3]}}},
		document.Section{Title: textOf("Late"), Name: "late", Blocks: []document.Block{document.Posts{Line: nil, Count: textOf("two"), List: listed[3:]}}},
	}}
	return commanded(test, handModel(test, commonWidth, height, slices.Concat(documents, []document.Document{byYear, series})...), "e blog")
}

// A roomy screen read down to the page's first row, the row under the marks where the frame's lines cross, as the
// page's corners, and on to its end
const pageTopRow = `(?s)^.*?\n *\+ +\+ *\n([^\n]*).*$`

// firstRowOf is the page's first row on a roomy screen, under the marks where the frame's lines cross; the screen whole
// where it has no such marks.
func firstRowOf(screen string) string {
	return regexp.MustCompile(pageTopRow).ReplaceAllString(screen, "$1")
}

// In the blog's session h and l go between the list and the post read; q on the post goes back to the list, its
// window closed, and q on the list back to the résumé, from the list beside a post too, the session going on.
func TestTheBlogsKeysGoBetweenItsWindows(test *testing.T) {
	cases := []struct {
		name  string
		keys  []string
		holds []string
		lacks []string
	}{
		{name: "h to the list", keys: []string{"h"}, holds: []string{"1:posts*", "2:reading-slowly"}, lacks: []string{postRead}},
		{name: "l back to the post", keys: []string{"h", "l"}, holds: []string{postRead}, lacks: []string{"1:posts*"}},
		{name: "q on the post", keys: []string{"q"}, holds: []string{listAlone}, lacks: []string{"2:"}},
		{name: "q on the list", keys: []string{"q", "q"}, holds: []string{resumeOpened}, lacks: []string{"posts"}},
		{name: "q on the list beside the post", keys: []string{"h", "q"}, holds: []string{resumeOpened}, lacks: []string{"posts"}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := openedModel(test, "blog/1", commonWidth, commonHeight)
			for _, key := range testCase.keys {
				next, command := model.Update(keyPress(test, key))
				if command != nil {
					test.Fatalf("after %q a command came, want the session going on", testCase.keys)
				}
				model = as[Model](test, next)
			}
			footer := footerOf(screenOf(test, model))
			for _, want := range testCase.holds {
				if !strings.Contains(footer, want) {
					test.Errorf("on blog/1 after %q the footer = %q, want it to hold %q", testCase.keys, footer, want)
				}
			}
			for _, unwanted := range testCase.lacks {
				if strings.Contains(footer, unwanted) {
					test.Errorf("on blog/1 after %q the footer = %q, want it not to hold %q", testCase.keys, footer, unwanted)
				}
			}
		})
	}
}

// The blog's session owns its windows: :e of a list from a post shows that list in the first window, the post read
// kept beside it, and going to the résumé and back keeps both.
func TestTheBlogKeepsThePostReadBesideTheListShown(test *testing.T) {
	cases := []struct {
		name   string
		opened Model
		holds  string
	}{
		{name: "the list by year", opened: commanded(test, openedModel(test, "blog/1", commonWidth, tallHeight), "e blog"), holds: "1 post"},
		{name: "the list by series", opened: commanded(test, openedModel(test, "blog/1", commonWidth, tallHeight), "e blog/series"), holds: "1 of 2 out"},
		{name: "back from the résumé", opened: commanded(test, commanded(test, openedModel(test, "blog/1", commonWidth, tallHeight), "e tidy-tool"), "e blog"), holds: "1 post"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, testCase.opened)
			if footer := footerOf(screen); !strings.Contains(footer, "1:posts*") || !strings.Contains(footer, "2:reading-slowly") || !strings.Contains(screen, testCase.holds) {
				test.Errorf("the screen =\n%s\nwant the list shown, holding %q, the post's window beside it", screen, testCase.holds)
			}
		})
	}
}

// The blog's session reads one language at a time, as a reader of the web stays in a language's folder: a page of
// another language shows in its window, and the other window follows it there, the post read in its translation where
// it has one, an older version's current one, and closed where not, its own language's list opening it no more; the
// list in the same grouping and filter where the language's list has that filter, and by year where not.
func TestTheBlogsSessionReadsOneLanguage(test *testing.T) {
	translated := []fixtureFile{secondPost(), {path: secondTranslationFile, text: secondTranslationText}}
	cases := []struct {
		name     string
		files    []fixtureFile
		user     string
		commands []string
		keys     []string
		holds    []string
		lacks    []string
	}{
		{
			name: "the post read, translated", files: nil, user: "blog/1", commands: []string{"e ru/blog"}, keys: []string{"l"},
			holds: []string{russianListBesidePost, postRead, "Медленное чтение"}, lacks: nil,
		},
		{
			name: "the post read, with no translation", files: []fixtureFile{secondPost()}, user: "blog/second", commands: []string{"e ru/blog"}, keys: nil,
			holds: []string{russianListAlone}, lacks: []string{"2:"},
		},
		{
			name: "an older version read, at its translation's current version", files: nil, user: "blog/reading-slowly/2026-11-10",
			commands: []string{"e ru/blog"}, keys: []string{"l"}, holds: []string{russianListBesidePost, postRead, "Медленное чтение"}, lacks: nil,
		},
		{
			name: "the post closed, which its own language's list opens no more", files: []fixtureFile{secondPost()}, user: "blog/second",
			commands: []string{"e ru/blog", "e blog"}, keys: nil, holds: []string{listAlone}, lacks: []string{"2:"},
		},
		{
			name: "the original read", files: nil, user: "ru/blog/1", commands: []string{"e blog/series"}, keys: []string{"l"},
			holds: []string{listBesidePost, postRead, "Reading slowly"}, lacks: nil,
		},
		{
			name: "the list's filter, which the language has", files: translated, user: "blog?tag=reading", commands: []string{"e ru/blog/1"}, keys: []string{"h"},
			holds: []string{russianListAlone, "Медленное чтение"}, lacks: []string{"Второй пост"},
		},
		{
			name: "the list's filter, which the language lacks", files: []fixtureFile{secondPost()}, user: "blog?tag=tools", commands: []string{"e ru/blog/1"},
			keys: []string{"h"}, holds: []string{russianListAlone, "Медленное чтение"}, lacks: nil,
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			model := New(sessionDrawing(contentWith(test, testCase.files...), false), testCase.user, commonWidth, tallHeight, fixtures.Day())
			for _, text := range testCase.commands {
				model = commanded(test, model, text)
			}
			screen := screenOf(test, pressed(test, model, testCase.keys...))
			for _, want := range testCase.holds {
				if !strings.Contains(screen, want) {
					test.Errorf("opened by %q, after :%s and %q, the screen =\n%s\nwant %q on it", testCase.user, testCase.commands, testCase.keys, screen, want)
				}
			}
			for _, unwanted := range testCase.lacks {
				if strings.Contains(screen, unwanted) {
					test.Errorf("opened by %q, after :%s and %q, the screen =\n%s\nwant no %q on it", testCase.user, testCase.commands, testCase.keys, screen, unwanted)
				}
			}
		})
	}
}

// The slugs of the posts of a list a test builds by hand, in its order, newest first: the first of one year, the rest
// of the year before
func handPostSlugs() []string {
	return []string{"fifth", "fourth", "third", "second", "first"}
}

// The years of the list a test builds by hand, newest first
const (
	newerYear = "2027"
	olderYear = "2026"
)

// handBlogModel is a session of a size on a résumé and a blog's list a test builds by hand, of five posts in two
// years, each a post of its own, opened on the list; each year's count of its posts names the year.
func handBlogModel(test *testing.T, width, height int) Model {
	resume := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, textOf("home")), Blocks: nil}
	list := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageBlog, Slug: "", Tab: ""}, textOf("blog")), Blocks: nil}
	documents := []document.Document{resume}
	years := map[string][]document.ListedPost{}
	for index, slug := range handPostSlugs() {
		page := document.Page{Kind: document.PagePost, Slug: slug, Tab: ""}
		documents = append(documents, document.Document{Head: fixtures.HeadOf(page, textOf("The "+slug+" post")), Blocks: nil})
		year := olderYear
		if index == 0 {
			year = newerYear
		}
		out := document.PostOut{Page: page, Published: time.Date(2026, time.November, index+1, 0, 0, 0, 0, time.UTC)}
		years[year] = append(years[year], document.ListedPost{Readiness: out, Title: textOf("The " + slug + " post"), Lead: nil, Part: nil, Series: nil, Tags: nil})
	}
	for _, year := range []string{newerYear, olderYear} {
		posts := document.Posts{Line: nil, Count: textOf("posts of " + year), List: years[year]}
		list.Blocks = append(list.Blocks, document.Section{Title: textOf(year), Blocks: []document.Block{posts}})
	}
	// the list by series holds the same posts in one series, the oldest first
	inSeries := slices.Concat(years[newerYear], years[olderYear])
	slices.Reverse(inSeries)
	series := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageBlogSeries, Slug: "", Tab: ""}, textOf("blog")),
		Blocks: []document.Block{document.Section{Title: textOf("A series"), Blocks: []document.Block{document.Posts{Line: nil, Count: textOf("five"), List: inSeries}}}},
	}
	return commanded(test, handModel(test, width, height, slices.Concat(documents, []document.Document{list, series})...), "e blog")
}

// Each of the blog's lists keeps the tile its keys stood on, as it keeps its scroll: back from the résumé, and back
// from the other list, the keys of each stand where they stood on it.
func TestEachListKeepsTheTileItsKeysStoodOn(test *testing.T) {
	model := handBlogModel(test, commonWidth, tallHeight)
	byYear := commanded(test, pressed(test, model, "j", "q"), "e blog")
	if footer := footerOf(screenOf(test, pressed(test, byYear, "enter"))); !strings.Contains(footer, "2:fourth*") {
		test.Errorf("after j, q, :e blog and Enter the footer = %q, want %q", footer, "2:fourth*")
	}
	bySeries := commanded(test, pressed(test, commanded(test, pressed(test, model, "j"), "e blog/series"), "j", "j"), "e blog")
	if footer := footerOf(screenOf(test, pressed(test, bySeries, "enter"))); !strings.Contains(footer, "2:fourth*") {
		test.Errorf("after j, :e blog/series, j j, :e blog and Enter the footer = %q, want %q", footer, "2:fourth*")
	}
	if footer := footerOf(screenOf(test, pressed(test, commanded(test, bySeries, "e blog/series"), "enter"))); !strings.Contains(footer, "2:third*") {
		test.Errorf("back on the list by series, Enter gives the footer %q, want %q", footer, "2:third*")
	}
}

// The keys never stand on a link out of sight: a move that takes the page from the link they stand on has them stand
// on the link in sight nearest where the page went, G on the last tile, gg on the first grouping, a search's match on
// its own tile, the wheel on the later tile of the row it brings in; and a screen made smaller keeps the link they
// stand on in sight. Enter then follows a link the reader sees.
func TestTheLinkTheKeysStandOnStaysInSight(test *testing.T) {
	// a screen too short for the list whole, and one shorter still; and the steps of the wheel that take it past the
	// first tile
	const (
		shortScreen                = 14
		shorterStill               = 10
		wheelStepsPastTheFirstTile = 5
	)
	list := handBlogModel(test, commonWidth, shortScreen)
	cases := []struct {
		name    string
		moved   Model
		resized int
		// seen is what shows the link in sight: the day set into a tile's top edge, or a grouping's word
		seen string
		// opened is what the screen shows after Enter: the post's window in the footer, or the day of the first tile of
		// the list by year, which the list by series shows out of sight
		opened string
	}{
		{name: "G", moved: pressed(test, list, "G"), resized: shortScreen, seen: "5 Nov 2026", opened: "2:first*"},
		{name: "gg after G", moved: pressed(test, list, "G", "g", "g"), resized: shortScreen, seen: "years", opened: "1 Nov 2026"},
		{name: "a search", moved: pressed(test, typed(test, pressed(test, list, "/"), "third"), "enter"), resized: shortScreen, seen: "3 Nov 2026", opened: "2:third*"},
		{name: "the wheel", moved: updated(test, list, slices.Repeat([]tea.Msg{wheelStep(tea.MouseWheelDown)}, wheelStepsPastTheFirstTile)...), resized: shortScreen, seen: "3 Nov 2026", opened: "2:third*"},
		{name: "the screen made smaller", moved: pressed(test, list, "j", "j", "j"), resized: shorterStill, seen: "4 Nov 2026", opened: "2:second*"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			next, _ := testCase.moved.Update(tea.WindowSizeMsg{Width: commonWidth, Height: testCase.resized})
			model := as[Model](test, next)
			if screen := screenOf(test, model); !strings.Contains(screen, testCase.seen) {
				test.Errorf("the screen =\n%s\nwant %q in sight", screen, testCase.seen)
			}
			if screen := screenOf(test, pressed(test, model, "enter")); !strings.Contains(screen, testCase.opened) {
				test.Errorf("after Enter the screen =\n%s\nwant %q on it", screen, testCase.opened)
			}
		})
	}
}

// The list opens at its top, as gg leaves it, its first tile in sight though not whole, the blog's name whole above
// it; and a search that finds a tile's words shows that tile from its top edge, its date with it, as j does.
func TestTheListOpensAtItsTopAndASearchShowsItsTilesDate(test *testing.T) {
	// a terminal short enough to show the fixture's first tile in part
	const tileInPart = 20
	if footer := footerOf(screenOf(test, openedModel(test, "blog", narrowWidth*sides, tileInPart))); !strings.Contains(footer, startShown) {
		test.Errorf("opened by blog the footer = %q, want the list at its top, %q", footer, startShown)
	}
	found := pressed(test, typed(test, pressed(test, handBlogModel(test, commonWidth, 14), "/"), "third post"), "enter")
	if screen := screenOf(test, found); !strings.Contains(screen, "3 Nov 2026") {
		test.Errorf("after /third post the screen =\n%s\nwant the third post's tile from its top edge, %q", screen, "3 Nov 2026")
	}
}

// The terminals the keys walk the blog's list on: the least a terminal opens at, and a roomy one, wide enough for the
// groups' names' column
const (
	leastWidth  = 80
	leastHeight = 24
	roomyWidth  = 200
	roomyHeight = 50
)

// On the blog's list j and k go from link to link in reading order: the ways the list is grouped, its tags, then its
// tiles, newest first, staying on the first and the last; the list opens on its first tile. Enter follows the link the
// keys stand on, as the page's links lead: a grouping to the list grouped so, a tag to the list it filters, the tag the
// list is filtered by to the whole list, a tile to its post beside the list. So on the list by year, by series and
// filtered, on the least terminal and a roomy one.
func TestTheKeysGoFromLinkToLink(test *testing.T) {
	// what tells the lists apart on their first screen: how many posts the list by year holds, the series' line on
	// the list by series, and the tile each tag keeps
	const (
		byYear       = "2 posts"
		bySeries     = "1 of 2 out"
		taggedTools  = "The second post"
		taggedReader = "Reading slowly"
	)
	cases := []struct {
		name    string
		opening string
		keys    []string
		footer  string
		holds   []string
		lacks   []string
	}{
		{name: "Enter on the first tile, where the list opens", opening: "e blog", keys: []string{"enter"}, footer: "2:second*", holds: nil, lacks: nil},
		{name: "j to the next tile", opening: "e blog", keys: []string{"j", "enter"}, footer: postRead, holds: nil, lacks: nil},
		{name: "j past the last", opening: "e blog", keys: []string{"j", "j", "enter"}, footer: postRead, holds: nil, lacks: nil},
		{name: "the down arrow", opening: "e blog", keys: []string{"down", "enter"}, footer: postRead, holds: nil, lacks: nil},
		{name: "k to the last tag", opening: "e blog", keys: []string{"k", "enter"}, footer: listAlone, holds: []string{taggedTools}, lacks: []string{taggedReader}},
		{name: "k to the first tag", opening: "e blog", keys: []string{"k", "k", "k", "enter"}, footer: listAlone, holds: []string{taggedReader}, lacks: []string{taggedTools}},
		{name: "k to the second grouping", opening: "e blog", keys: []string{"k", "k", "k", "k", "enter"}, footer: listAlone, holds: []string{bySeries}, lacks: nil},
		{name: "k past the first", opening: "e blog", keys: []string{"k", "k", "k", "k", "k", "k", "enter"}, footer: listAlone, holds: []string{byYear}, lacks: nil},
		{name: "the up arrow", opening: "e blog", keys: []string{"up", "enter"}, footer: listAlone, holds: []string{taggedTools}, lacks: []string{taggedReader}},
		{name: "j from the head to the first tile", opening: "e blog", keys: []string{"k", "j", "enter"}, footer: "2:second*", holds: nil, lacks: nil},
		{name: "by series, Enter on its first tile", opening: "e blog/series", keys: []string{"enter"}, footer: postRead, holds: nil, lacks: nil},
		{name: "by series, k to the first grouping", opening: "e blog/series", keys: []string{"k", "k", "k", "k", "k", "enter"}, footer: listAlone, holds: []string{byYear}, lacks: nil},
		{name: "filtered, Enter on its tile", opening: "e blog?tag=tools", keys: []string{"enter"}, footer: "2:second*", holds: nil, lacks: nil},
		{name: "filtered, k to its filter", opening: "e blog?tag=tools", keys: []string{"k", "enter"}, footer: listAlone, holds: []string{byYear}, lacks: nil},
		{name: "filtered, k to another tag", opening: "e blog?tag=tools", keys: []string{"k", "k", "enter"}, footer: listAlone, holds: []string{taggedReader}, lacks: []string{taggedTools}},
	}
	terminals := []struct {
		name   string
		width  int
		height int
	}{
		{name: "the least", width: leastWidth, height: leastHeight},
		{name: "a roomy one", width: roomyWidth, height: roomyHeight},
	}
	drawing := sessionDrawing(contentWith(test, secondPost()), false)
	for _, terminal := range terminals {
		for _, testCase := range cases {
			test.Run(terminal.name+"/"+testCase.name, func(test *testing.T) {
				list := commanded(test, New(drawing, "blog", terminal.width, terminal.height, fixtures.Day()), testCase.opening)
				screen := screenOf(test, pressed(test, list, testCase.keys...))
				if footer := footerOf(screen); !strings.Contains(footer, testCase.footer) {
					test.Errorf("after :%s and %q the footer = %q, want %q", testCase.opening, testCase.keys, footer, testCase.footer)
				}
				for _, want := range testCase.holds {
					if !strings.Contains(screen, want) {
						test.Errorf("after :%s and %q the screen =\n%s\nwant %q on it", testCase.opening, testCase.keys, screen, want)
					}
				}
				for _, unwanted := range testCase.lacks {
					if strings.Contains(screen, unwanted) {
						test.Errorf("after :%s and %q the screen =\n%s\nwant no %q on it", testCase.opening, testCase.keys, screen, unwanted)
					}
				}
			})
		}
	}
}

// A list with no post out, as the blog before its first, opens on its first link, the grouping by year, as gg goes to
// it: Enter there opens the list by year, and after j the list by series.
func TestAListWithNoPostOpensOnItsFirstLink(test *testing.T) {
	// the lead of each list, which tells the two apart
	const (
		byYearLead   = "Every post by its year."
		bySeriesLead = "Every post by its series."
	)
	resume := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, textOf("home")), Blocks: nil}
	byYear := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageBlog, Slug: "", Tab: ""}, textOf("blog")), Blocks: nil}
	byYear.Head.Lead = textOf(byYearLead)
	bySeries := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageBlogSeries, Slug: "", Tab: ""}, textOf("blog")), Blocks: nil}
	bySeries.Head.Lead = textOf(bySeriesLead)
	list := commanded(test, handModel(test, leastWidth, leastHeight, resume, byYear, bySeries), "e blog")
	cases := []struct {
		name  string
		keys  []string
		holds string
		lacks string
	}{
		{name: "Enter where it opens", keys: []string{"enter"}, holds: byYearLead, lacks: bySeriesLead},
		{name: "j to the second grouping", keys: []string{"j", "enter"}, holds: bySeriesLead, lacks: byYearLead},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			screen := screenOf(test, pressed(test, list, testCase.keys...))
			if !strings.Contains(screen, testCase.holds) || strings.Contains(screen, testCase.lacks) {
				test.Errorf("after %q the screen =\n%s\nwant %q on it and no %q", testCase.keys, screen, testCase.holds, testCase.lacks)
			}
		})
	}
}

// Where the list has a column for the groups' names, a group's name stays in sight while its tiles scroll under it:
// at the list's end, the top of the last group's tiles scrolled away, its count still shows.
func TestAGroupsNameStaysInSightWhileItsTilesScroll(test *testing.T) {
	// a terminal wide enough for the groups' names' column, and short enough for a page of five rows
	const (
		columnWidth  = 120
		fiveRowsHigh = 9
	)
	model := pressed(test, handBlogModel(test, columnWidth, fiveRowsHigh), "G")
	if want := "posts of " + olderYear; !strings.Contains(screenOf(test, model), want) {
		test.Errorf("at the list's end the screen =\n%s\nwant %q in sight", screenOf(test, model), want)
	}
}

// A search typed and confirmed paints the lines again and leaves what stays in sight over them: a post's Outline,
// its heading read marked, and on the blog's list the name of the group whose tiles it scrolled to.
func TestASearchConfirmedLeavesWhatStaysInSight(test *testing.T) {
	// a terminal wide enough for the groups' names' column, and short enough for a page of five rows
	const (
		columnWidth  = 120
		fiveRowsHigh = 9
	)
	cases := []struct {
		name  string
		model Model
		query string
		want  string
	}{
		{name: "a post's Outline", model: openedModel(test, "blog/1", outlinedWidth, tallHeight), query: "Afterwards", want: readMark + " Afterwards"},
		{name: "a group's name on the list", model: handBlogModel(test, columnWidth, fiveRowsHigh), query: "The first post", want: "posts of " + olderYear},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			searched := pressed(test, typed(test, pressed(test, testCase.model, "/"), testCase.query), "enter")
			if screen := screenOf(test, searched); !strings.Contains(screen, testCase.want) {
				test.Errorf("after the search %q the screen =\n%s\nwant %q on it", testCase.query, screen, testCase.want)
			}
		})
	}
}

// The blog's list writes its head and each tile as the page does (DESIGN.md, BlogHead and Tile): the blog's line, the
// way it is grouped and the tags of its posts; how many posts a group holds; and on a tile when its post came out, its
// title and lead, its part in its series and the series, and its tags. It writes the same where the page is too narrow
// for the groups' names' column.
func TestTheListWritesItsHeadAndTiles(test *testing.T) {
	// a terminal wide enough for the groups' names' column beside the tiles
	const columnWidth = 120
	want := []string{
		"What the writer reads, and how.", "by years series", "#reading #terminals", "1 post", "10 Nov 2026",
		"Reading slowly", "Why a page read twice", "part 1 · Reading",
	}
	for _, width := range []int{commonWidth, columnWidth, narrowWidth} {
		read := readRows(aboveTheFooter(screenOf(test, openedModel(test, "blog", width, tallHeight))))
		for _, phrase := range want {
			if !strings.Contains(read, phrase) {
				test.Errorf("at %d columns the list =\n%s\nwant %q on it", width, read, phrase)
			}
		}
	}
}

// A post writes its head, its body and its foot as the page does (DESIGN.md, PostHead, Article and Panes): its facts
// under its lead; its paragraphs numbered, a note cited by its number raised, a list's points and numbers, a quote
// after its bar, its code framed under its language, and its pause as three stars; its sources at
// its end, each by its number; the panes' heading and the way to all the parts, each pane by its name and what it
// holds, the next first; and where the post is discussed.
func TestAPostWritesItsHeadBodyAndFoot(test *testing.T) {
	read := readRows(aboveTheFooter(screenOf(test, openedModel(test, "blog/1", commonWidth, tallHeight))))
	cases := []struct {
		name string
		want string
	}{
		{name: "the day", want: "version 12 Nov 2026"},
		{name: "the series", want: "series Reading 1 of 2"},
		{name: "the languages it is read in", want: "language en · ru"},
		{name: "the tags", want: "tags #reading #terminals"},
		{name: "a numbered paragraph citing a note", want: "1 A page read twice is read once, if the first reading was slow.¹"},
		{name: "a paragraph of a section, numbered on", want: "2 See the reader's notes"},
		{name: "a point", want: "- the margins"},
		{name: "a point nested", want: "- in pencil²"},
		{name: "a list in order", want: "3. open the file"},
		{name: "a quote", want: "▎ A book is read with a pencil."},
		{
			name: "a picture, its alt text, its address and its caption a line each",
			want: "\nA noted margin\nhttps://example.test/blog/reading-slowly/margin.png\nThe margin, where a slow reader writes back.\n",
		},
		{name: "code under its language", want: "┌── go ─"},
		{name: "a pause", want: "* * *"},
		{name: "a source", want: "1 Mira Example, on her own reading."},
		{name: "the panes' heading", want: "Other parts of Reading all 2 parts"},
		{name: "the next part, to come", want: "part 2"},
		{name: "the next part's title", want: "Reading aloud soon"},
		{name: "the side before, with nothing there", want: "none: this is the first part"},
		{name: "where it is discussed", want: "discuss Lobsters"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !strings.Contains(read, testCase.want) {
				test.Errorf("the post =\n%s\nwant %q on it", read, testCase.want)
			}
		})
	}
}

// An older version of a post writes its own title and lead, the day it came out, and under it that a newer stands, the
// day of the current version after the site's word for it.
func TestAnOlderVersionSaysANewerStands(test *testing.T) {
	read := readRows(aboveTheFooter(screenOf(test, openedModel(test, "blog/reading-slowly/2026-11-10", commonWidth, tallHeight))))
	for _, want := range []string{"Slow reading", "Why a page read twice is read once, first said.", "version 10 Nov 2026", look.VersionNoteMark + " newer: 12 Nov 2026"} {
		if !strings.Contains(read, want) {
			test.Errorf("the older version =\n%s\nwant %q on it", read, want)
		}
	}
}

// The blog in another language is written in that language's words, as its page is: on its list a count in the
// language's plural and a day in its months; on a translation its facts by the language's labels, and under its day
// that it is older than its original, where the original's text changed since it was made.
func TestATranslationIsWrittenInItsLanguagesWords(test *testing.T) {
	cases := []struct {
		name string
		user string
		want []string
	}{
		{name: "the list", user: "ru/blog", want: []string{"1 пост", "15 нояб. 2026", "часть 1 · Чтение"}},
		{
			name: "a translation", user: "ru/blog/reading-slowly",
			want: []string{"версия 15 нояб. 2026", look.VersionNoteMark + " старше оригинала", "язык en · ru", "теги #reading #terminals"},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			read := readRows(aboveTheFooter(screenOf(test, openedModel(test, testCase.user, commonWidth, tallHeight))))
			for _, want := range testCase.want {
				if !strings.Contains(read, want) {
					test.Errorf("opened by %q the screen =\n%s\nwant %q on it", testCase.user, read, want)
				}
			}
		})
	}
}

// A draft, as its author reads it, says it is a draft where a post out says the day of the version read.
func TestADraftsPageIsMarkedAsADraft(test *testing.T) {
	model := New(sessionDrawing(fixtures.LoadedContent(test), true), "blog/2", commonWidth, tallHeight, fixtures.Day())
	if read, want := readRows(aboveTheFooter(screenOf(test, model))), "version draft"; !strings.Contains(read, want) {
		test.Errorf("the draft =\n%s\nwant %q on it", read, want)
	}
}

// A post's panes stand one under the other, the next part's first (DESIGN.md, The TUI).
func TestAPostsPanesStandTheNextFirst(test *testing.T) {
	read := readRows(aboveTheFooter(screenOf(test, openedModel(test, "blog/1", commonWidth, tallHeight))))
	if next, before := strings.Index(read, "── part 2 ─"), strings.Index(read, "── previous ─"); next < 0 || next > before {
		test.Errorf("the post =\n%s\nwant the pane of part 2 over the pane before it", read)
	}
}

// The narrowest terminals whose page, eight cells narrower, holds a post's Outline and its sources beside its text,
// 132 cells, the blog's widest; and its sources alone, 106 cells
const (
	outlinedWidth = 140
	sourcedWidth  = 114
)

// A post stands in the columns its page is wide enough for (DESIGN.md, The TUI): from the width of its sources beside
// its text, its head's facts beside its title, a source beside the line that first cites it, and the panes side by
// side, the side before at the left; from the blog's widest, its Outline beside its body as well, its heading read
// marked, at the post's end its last. Narrower, none of them, each standing under what it stood beside.
func TestAPostStandsInTheColumnsItsPageHasRoomFor(test *testing.T) {
	besides := []struct {
		name    string
		pattern string
		from    int
	}{
		{name: "the facts beside the title", pattern: `(?m)^Reading slowly version 12 Nov 2026$`, from: sourcedWidth},
		{name: "a source beside its line", pattern: `slow\.¹ 1 Mira Example`, from: sourcedWidth},
		{name: "a source beside its line, under the one before", pattern: `in pencil² 2 A soft one`, from: sourcedWidth},
		{name: "the panes side by side", pattern: `── previous ─+┬── part 2 ─`, from: sourcedWidth},
		{name: "the Outline beside its body", pattern: `(?m)^What a reader keeps 1 A page read twice`, from: outlinedWidth},
		{name: "the Outline's heading read", pattern: readMark + ` Afterwards`, from: outlinedWidth},
	}
	for _, width := range []int{outlinedWidth, outlinedWidth - 1, sourcedWidth, sourcedWidth - 1} {
		read := readRows(aboveTheFooter(screenOf(test, openedModel(test, "blog/1", width, tallHeight))))
		for _, beside := range besides {
			test.Run(beside.name+" at "+strconv.Itoa(width), func(test *testing.T) {
				if shown := width >= beside.from; regexp.MustCompile(beside.pattern).MatchString(read) != shown {
					test.Errorf("at %d columns the post =\n%s\nwant %q on it: %t", width, read, beside.pattern, shown)
				}
			})
		}
	}
}

// As a post scrolls, its Outline marks the heading read as the page's does (DESIGN.md, Outline): none while the first
// heading stands on or under the page's reading line, a third of the way down; the first once it has passed the line;
// and at the post's end its last. Where the post's foot scrolls into sight, the Outline rises with its body's end.
func TestScrollingAPostMovesTheHeadingItsOutlineMarksRead(test *testing.T) {
	// a page of twelve rows, its reading line three rows down, and one tall enough for the first heading to stand
	// above its reading line at the post's top; and how far the fixture's post scrolls on the short page until its
	// first heading, seven lines down it, stands on the reading line
	const (
		shortPage          = 16
		tallPage           = 40
		toTheReadingLine   = 5
		pastTheReadingLine = toTheReadingLine + 1
	)
	// no heading marked read anywhere on the screen
	noneRead := `\A[^` + readMark + `]*\z`
	cases := []struct {
		name    string
		height  int
		keys    []string
		pattern string
	}{
		{name: "the top of a short page", height: shortPage, keys: nil, pattern: noneRead},
		{name: "the first heading on the reading line", height: shortPage, keys: slices.Repeat([]string{"j"}, toTheReadingLine), pattern: noneRead},
		{name: "the first heading past the reading line", height: shortPage, keys: slices.Repeat([]string{"j"}, pastTheReadingLine), pattern: readMark + " What a reader keeps"},
		{name: "the end of a short page, beside the body's last line", height: shortPage, keys: []string{"G"}, pattern: readMark + " Afterwards 4 What stays"},
		{name: "the top of a tall page", height: tallPage, keys: nil, pattern: readMark + " What a reader keeps"},
		{name: "the end of a tall page", height: tallPage, keys: []string{"G"}, pattern: readMark + " Afterwards"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			read := readRows(aboveTheFooter(screenOf(test, pressed(test, openedModel(test, "blog/1", outlinedWidth, testCase.height), testCase.keys...))))
			if !regexp.MustCompile(testCase.pattern).MatchString(read) {
				test.Errorf("after %q the post =\n%s\nwant %q on it", testCase.keys, read, testCase.pattern)
			}
		})
	}
}

// A source stands beside the line that first cites it, though another's number is raised beside its own there: the
// second note of two cited side by side stands under the first, not at a later line that cites it alone.
func TestASourceStandsBesideTheFirstLineCitingItsNote(test *testing.T) {
	first, second := document.Note{Number: 1, Spans: textOf("First source.")}, document.Note{Number: 2, Spans: textOf("Second source.")}
	post := handPost(test, outlinedWidth, tallHeight,
		document.Paragraph{Spans: []document.Span{document.Text{Text: "Both cited"}, first, second}},
		document.Pause{}, document.Pause{},
		document.Paragraph{Spans: []document.Span{document.Text{Text: "Cited again"}, second}},
	)
	read := readRows(aboveTheFooter(screenOf(test, post)))
	if beside, later := strings.Index(read, "2 Second source."), strings.Index(read, "Cited again"); beside < 0 || beside > later {
		test.Errorf("the post =\n%s\nwant the second source over the line citing it again", read)
	}
}

// A heading gone to by the command line is the one the Outline marks read, as the page marks the heading a link goes
// to (DESIGN.md, Outline), though the reading line has passed the next too, its section being short; and at the post's
// end anywhere on the page, though the last is under it. Once a scroll takes it above the page, the heading the
// reading line passed last is read again.
func TestTheOutlineMarksTheHeadingGoneTo(test *testing.T) {
	// a page of thirty-six rows, its reading line eleven rows down; pauses enough to keep the third heading under the
	// line while the second passes it, and to make the post longer than the page
	const (
		pageHeight  = 40
		shortPauses = 2
		longPauses  = 20
	)
	second := append([]document.Block{document.Paragraph{Spans: textOf("Two.")}}, slices.Repeat([]document.Block{document.Pause{}}, shortPauses)...)
	sections := []document.Block{
		document.Section{Title: textOf("First"), Name: "first", Blocks: []document.Block{document.Paragraph{Spans: textOf("One.")}}},
		document.Section{Title: textOf("Second"), Name: "second", Blocks: second},
		document.Section{Title: textOf("Third"), Name: "third", Blocks: slices.Repeat([]document.Block{document.Pause{}}, longPauses)},
		document.Section{Title: textOf("Fourth"), Name: "fourth", Blocks: []document.Block{document.Paragraph{Spans: textOf("Four.")}}},
		document.Section{Title: textOf("Fifth"), Name: "fifth", Blocks: []document.Block{document.Paragraph{Spans: textOf("Five.")}}},
	}
	cases := []struct {
		name   string
		goneTo string
		then   func(test *testing.T, model Model) Model
		marked string
	}{
		{name: "gone to", goneTo: "first", then: func(_ *testing.T, model Model) Model { return model }, marked: readMark + " First"},
		{
			name: "gone to, the post closed and opened again", goneTo: "first",
			then: func(test *testing.T, model Model) Model {
				return commanded(test, pressed(test, model, "q"), "e blog/reading-slowly")
			},
			marked: readMark + " First",
		},
		{name: "scrolled past it", goneTo: "first", then: func(test *testing.T, model Model) Model { return pressed(test, model, "j") }, marked: readMark + " Second"},
		{name: "gone to at the post's end", goneTo: "fourth", then: func(_ *testing.T, model Model) Model { return model }, marked: readMark + " Fourth"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			gone := commanded(test, handPost(test, outlinedWidth, pageHeight, sections...), testCase.goneTo)
			if read := readRows(aboveTheFooter(screenOf(test, testCase.then(test, gone)))); !strings.Contains(read, testCase.marked) {
				test.Errorf("%s, the post =\n%s\nwant %q on it", testCase.name, read, testCase.marked)
			}
		})
	}
}

// A heading whose title wraps in the Outline's column, once read, takes one bar, before its first row, its words going
// on under themselves.
func TestAWrappedHeadingReadTakesOneBarBeforeItsFirstRow(test *testing.T) {
	sections := []document.Block{
		document.Section{Title: textOf("Before"), Name: "before", Blocks: []document.Block{document.Paragraph{Spans: textOf("One.")}}},
		document.Section{Title: textOf("A heading whose title wraps in the rail"), Name: "wrapped", Blocks: []document.Block{document.Paragraph{Spans: textOf("Two.")}}},
	}
	// at the post's end, its last heading read
	screen := screenOf(test, handPost(test, outlinedWidth, tallHeight, sections...))
	read := readRows(aboveTheFooter(screen))
	if marks := strings.Count(screen, readMark); marks != 1 || !regexp.MustCompile(`(?m)^`+readMark+` A heading whose title.*\nwraps in the rail`).MatchString(read) {
		test.Errorf("the post =\n%s\nwant one %q, before the wrapped heading's first row and over its second", read, readMark)
	}
}

// handPost is a session on a résumé and a post a test builds by hand of blocks, of a size, opened on the post, at the
// fixture's post's address, which the frame's :e reads.
func handPost(test *testing.T, width, height int, blocks ...document.Block) Model {
	resume := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageResume, Slug: "", Tab: ""}, textOf("home")), Blocks: nil}
	post := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PagePost, Slug: "reading-slowly", Tab: ""}, textOf("Coded")), Blocks: blocks}
	return commanded(test, handModel(test, width, height, resume, post), "e blog/reading-slowly")
}

// codeModel is a post a test builds by hand, of one block of code.
func codeModel(test *testing.T, code string) Model {
	return handPost(test, commonWidth, tallHeight, document.CodeBlock{Language: "", Runs: []document.CodeRun{{Kind: document.CodePlain, Text: code}}})
}

// A tab in a post's code stands as the spaces to the next stop of its line, look.TabStop apart, never as a tab,
// which a terminal would widen to its own eight: in the middle of a line as at its start.
func TestATabInCodeStandsToItsStop(test *testing.T) {
	cases := []struct {
		name string
		code string
		want string
	}{
		{name: "in the middle of a line", code: "ab\tc", want: "ab  c"},
		{name: "at a line's start", code: "x\n\ty", want: "    y"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if screen := screenOf(test, codeModel(test, testCase.code)); !strings.Contains(screen, testCase.want) || strings.Contains(screen, tabCharacter) {
				test.Errorf("the post of %q =\n%s\nwant %q on it, and no tab", testCase.code, screen, testCase.want)
			}
		})
	}
}

// A line of code wider than its frame goes on under itself, each row it goes on in opening with a mark that says so,
// as Neovim's showbreak does, and no letter of it lost.
func TestALongLineOfCodeGoesOnUnderItsMark(test *testing.T) {
	code := strings.Repeat("word ", 30) + "end"
	screen := screenOf(test, codeModel(test, code))
	read := strings.Join(strings.Fields(strings.NewReplacer(frameDown, "", codeGoesOnMark, "").Replace(aboveTheFooter(screen))), "")
	if !strings.Contains(screen, frameDown+"  "+codeGoesOnMark) || !strings.Contains(read, strings.Join(strings.Fields(code), "")) {
		test.Errorf("the post =\n%s\nwant the line going on under %q, whole", screen, codeGoesOnMark)
	}
}

// A widget, which a terminal runs none of, is its frame, its name and the word saying it runs set into the frame's top
// edge, inside it its alt text and the address of the live one, and its caption under it.
func TestAWidgetIsItsFrameHoldingItsAltText(test *testing.T) {
	live := document.Anchor{Page: document.PostPage("reading-slowly"), Name: "PageCache"}
	widget := document.Widget{Name: "PageCache", Alt: "A page read twice", Caption: "The cache", Live: live}
	read := readRows(aboveTheFooter(screenOf(test, handPost(test, commonWidth, tallHeight, widget))))
	for _, want := range []string{"┌── PageCache ─", "─ ● live ──┐", "│ A page read twice │", "│ https://example.test/blog/reading-slowly#PageCache │", "└─", "\nThe cache"} {
		if !strings.Contains(read, want) {
			test.Errorf("the post =\n%s\nwant %q on it", read, want)
		}
	}
}

// An item of a list that opens with a block, a picture, code, a quote, a table or a list, stands that block's first
// line on its mark's, as an item of words stands its first words there, its point's dash or its number.
func TestAnItemOpeningWithABlockStandsTheBlockOnItsMark(test *testing.T) {
	picture := document.Figure{Picture: document.File{Page: document.PostPage("reading-slowly"), Name: "shelf.svg"}, Alt: "A shelf", Caption: "", Width: 1, Height: 1}
	code := document.CodeBlock{Language: "go", Runs: []document.CodeRun{{Kind: document.CodePlain, Text: "main()"}}}
	quote := document.Quote{Blocks: []document.Block{document.Paragraph{Spans: textOf("A quote")}}}
	table := document.Table{Header: document.Row{Cells: nil}, Rows: []document.Row{{Cells: []document.Cell{
		{Kind: document.CellText, Spans: textOf("Reading")}, {Kind: document.CellText, Spans: textOf("Kept")},
	}}}}
	inner := document.List{Ordered: false, Start: 0, Items: []document.Item{{Blocks: []document.Block{document.Paragraph{Spans: textOf("inside")}}}}}
	items := []document.Item{{Blocks: []document.Block{picture}}, {Blocks: []document.Block{code}}, {Blocks: []document.Block{quote}}, {Blocks: []document.Block{table}}, {Blocks: []document.Block{inner}}}
	// a pause, which content refuses in an item, opens with an empty line, which the item's mark stands on
	paused := document.Item{Blocks: []document.Block{document.Pause{}}}
	points := document.List{Ordered: false, Start: 0, Items: append(items, paused)}
	numbered := document.List{Ordered: true, Start: 1, Items: []document.Item{{Blocks: []document.Block{code}}}}
	read := readRows(aboveTheFooter(screenOf(test, handPost(test, commonWidth, tallHeight, points, numbered))))
	for _, want := range []string{"- A shelf", "- ┌── go ─", "- ▎ A quote", "- Reading Kept", "- - inside", "-\n* * *", "1. ┌── go ─"} {
		if !strings.Contains(read, want) {
			test.Errorf("the post =\n%s\nwant %q on it", read, want)
		}
	}
}
