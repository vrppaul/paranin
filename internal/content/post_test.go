package content

import (
	"bytes"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/yuin/goldmark/ast"
	extensionast "github.com/yuin/goldmark/extension/ast"
)

// A post that loads, which a case writes beside the fixture's own posts, changed: its body's lines are the file's 8th
// to 12th
const samplePost = `---
id: 3
title: A sample post
description: What a sample says.
status: published
published: 2026-11-10T09:00+01:00
---
A paragraph.

## A section

Its text.
`

// A series that loads, which a case writes beside the fixture's own series, changed
const sampleSeries = `title: Tools
line: On the tools a writer keeps.
posts: [notes-on-tools]
`

// Where a case writes a sample unless it names another place, and the fixture's post of every kind of block
const (
	samplePostFile   = "posts/sample-post/post.en.md"
	sampleSeriesFile = "series/tools.yaml"
	samplePostSlug   = "sample-post"
	fixturePostSlug  = "reading-slowly"
)

// postFields are the fields of a loaded post a test reads, its times as RFC 3339 writes them.
type postFields struct {
	ID          int
	Title       string
	Description string
	Status      Status
	Published   string
	Updated     string
	Tags        []string
	Discussions []discussionFields
}

type discussionFields struct {
	Site    string
	Address string
}

// faultPlace is where a fault stands: its file, and its field or its line where it names one.
type faultPlace struct {
	file  string
	field string
	line  int
}

func TestAPostLoadsWithItsFieldsAndItsBodyParsed(test *testing.T) {
	snapshot, err := Load(fixtureFolder(test))
	if err != nil {
		test.Fatalf("Load of the fixture = %v, want no error", err)
	}
	post := postNamed(test, snapshot, fixturePostSlug)
	got := postFields{
		ID: post.ID, Title: post.Title, Description: post.Description, Status: post.Status,
		Published: post.Published.at.Format(time.RFC3339), Updated: post.Updated.start.Format(time.DateOnly),
		Tags: post.Tags, Discussions: nil,
	}
	for _, discussion := range post.Discussions {
		got.Discussions = append(got.Discussions, discussionFields{Site: discussion.Site, Address: discussion.Address.Written()})
	}
	want := postFields{
		ID: 1, Title: "Reading slowly", Description: "Why a page read twice is read once.", Status: StatusPublished,
		Published: "2026-11-10T09:00:00+01:00", Updated: "2026-11-12", Tags: []string{"reading", "terminals"},
		Discussions: []discussionFields{{Site: "Lobsters", Address: "https://lobsters.example.test/s/slow"}},
	}
	if !reflect.DeepEqual(got, want) {
		test.Errorf("%s loads as %+v, want %+v", fixturePostSlug, got, want)
	}
	if !bytes.HasPrefix(post.Body.Source, []byte("A page read twice")) {
		test.Errorf("%s's body = %q, want it to start after the front matter", fixturePostSlug, post.Body.Source)
	}
	kinds := kindsOf(post.Body.Tree)
	for _, kind := range []ast.NodeKind{extensionast.KindTable, extensionast.KindFootnoteList, ast.KindList, ast.KindBlockquote, ast.KindFencedCodeBlock} {
		if !slices.Contains(kinds, kind) {
			test.Errorf("%s's tree holds %v, want a %v among them", fixturePostSlug, kinds, kind)
		}
	}
}

// The snapshot keeps the posts in the order of their slugs, which no edit of a post's fields changes.
func TestThePostsComeInTheOrderOfTheirSlugs(test *testing.T) {
	snapshot, err := Load(writtenFile(fixtureFolder(test), "posts/a-first/post.en.md", samplePost))
	if err != nil {
		test.Fatalf("Load = %v, want no error", err)
	}
	slugs := []string{}
	for _, post := range snapshot.Posts {
		slugs = append(slugs, post.Slug)
	}
	if want := []string{"a-first", "notes-on-tools", fixturePostSlug}; !slices.Equal(slugs, want) {
		test.Errorf("the posts' slugs = %q, want %q", slugs, want)
	}
}

func TestASeriesLoadsWithItsPartsInOrder(test *testing.T) {
	snapshot, err := Load(fixtureFolder(test))
	if err != nil {
		test.Fatalf("Load of the fixture = %v, want no error", err)
	}
	want := []Series{{
		Name: "reading", Title: "Reading", Line: "How a reader keeps what they read.", Posts: []string{fixturePostSlug},
		ToCome: []string{"Reading aloud"},
	}}
	if !reflect.DeepEqual(snapshot.Series, want) {
		test.Errorf("the series = %+v, want %+v", snapshot.Series, want)
	}
}

// A post is CommonMark with GFM's tables and footnotes, and nothing else of GFM: the rest reads as CommonMark does.
func TestAPostsBodyReadsAsCommonMarkWithTablesAndFootnotes(test *testing.T) {
	cases := []struct {
		name    string
		body    string
		kind    ast.NodeKind
		present bool
	}{
		{name: "a table", body: "| Reading | Kept |\n| --- | --- |\n| Slow | Most |", kind: extensionast.KindTable, present: true},
		{name: "a footnote", body: "Its text.[^note]\n\n[^note]: A note.", kind: extensionast.KindFootnote, present: true},
		{name: "an address in angle brackets", body: "See <https://example.test>.", kind: ast.KindAutoLink, present: true},
		{name: "a bare address, as text", body: "See https://example.test now.", kind: ast.KindAutoLink, present: false},
		{name: "a strikethrough, as text", body: "~~Its~~ text.", kind: extensionast.KindStrikethrough, present: false},
		{name: "a task, as an item", body: "- [ ] a task", kind: extensionast.KindTaskCheckBox, present: false},
		{name: "dashes and quotes, as written", body: `Its -- "text".`, kind: ast.KindString, present: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			written := sampleWith(test, samplePost, "Its text.", testCase.body)
			snapshot, err := Load(writtenFile(fixtureFolder(test), samplePostFile, written))
			if err != nil {
				test.Fatalf("Load with the body %q = %v, want no error", testCase.body, err)
			}
			kinds := kindsOf(postNamed(test, snapshot, samplePostSlug).Body.Tree)
			if slices.Contains(kinds, testCase.kind) != testCase.present {
				test.Errorf("the body %q reads as %v, want a %v among them: %t", testCase.body, kinds, testCase.kind, testCase.present)
			}
		})
	}
}

// The format lets a post through that holds a tab inside code, which gofmt writes, a reference to a footnote written
// as code or escaped, and front matter whose fences end as a file written on Windows ends its lines.
func TestAPostTheFormatAllowsLoads(test *testing.T) {
	cases := []struct {
		name        string
		original    string
		replacement string
	}{
		{name: "a tab in fenced code", original: "Its text.", replacement: "```go\nfunc main() {\n\tprintln()\n}\n```"},
		{name: "a tab in indented code", original: "Its text.", replacement: "\tprintln()"},
		{name: "a tab in a code span", original: "Its text.", replacement: "Its `a\tb` text."},
		{name: "a reference written as code", original: "Its text.", replacement: "Its `[^note]` text."},
		{name: "a reference escaped", original: "Its text.", replacement: `Its \[^note] text.`},
		{name: "an opening fence with a carriage return", original: "---\nid: 3", replacement: "---\r\nid: 3"},
		{name: "a closing fence with a carriage return", original: "---\nA paragraph.", replacement: "---\r\nA paragraph."},
		{name: "a link to a mailbox", original: "Its text.", replacement: "Its [mail](mailto:mira@example.test)."},
		{name: "a mailbox in angle brackets", original: "Its text.", replacement: "Its <mira@example.test>."},
		{name: "a link over plain HTTP", original: "Its text.", replacement: "Its [text](http://example.test)."},
		{name: "a reference to a letter", original: "Its text.", replacement: "Its &amp; &#233; text."},
		{name: "a reference to a control character as code", original: "Its text.", replacement: "Its `&#27;` text."},
		{name: "a reference to a control character escaped", original: "Its text.", replacement: `Its \&#27; text.`},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			written := sampleWith(test, samplePost, testCase.original, testCase.replacement)
			if _, err := Load(writtenFile(fixtureFolder(test), samplePostFile, written)); err != nil {
				test.Errorf("Load with %q = %v, want no error", testCase.replacement, err)
			}
		})
	}
}

// A heading of the title's level is said once: the level it stands for is the author's to say, so the heading after
// it is no fault of a skipped level.
func TestAHeadingOfTheTitlesLevelIsTheOneFaultOfItsHeadings(test *testing.T) {
	written := sampleWith(test, samplePost, "## A section\n\nIts text.", "# A section\n\n### Its part")
	faults := faultsOf(test, writtenFile(fixtureFolder(test), samplePostFile, written))
	places := []faultPlace{}
	for _, fault := range faults {
		places = append(places, placeOf(fault))
	}
	if want := []faultPlace{{file: samplePostFile, field: "", line: 10}}; !slices.Equal(places, want) {
		test.Errorf("faults with a heading of level 1, then one of 3 = %q, want one at %+v alone", faults, want)
	}
}

// A post may be updated on the day it was published, as its own zone counts the day, or later.
func TestAPostsUpdateComesNoEarlierThanTheDayOfItsPublishing(test *testing.T) {
	cases := []struct {
		name      string
		published string
		updated   string
		loads     bool
	}{
		{name: "the next day", published: "2026-11-10T09:00+01:00", updated: "2026-11-11", loads: true},
		{name: "the same day, before midnight in UTC", published: "2026-11-10T00:30+01:00", updated: "2026-11-10", loads: true},
		{name: "the same day, after midnight in UTC", published: "2026-11-10T23:30-02:00", updated: "2026-11-10", loads: true},
		{name: "the day before", published: "2026-11-10T09:00+01:00", updated: "2026-11-09", loads: false},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			written := sampleWith(test, samplePost, "published: 2026-11-10T09:00+01:00",
				"published: "+testCase.published+"\nupdated: "+testCase.updated)
			_, err := Load(writtenFile(fixtureFolder(test), samplePostFile, written))
			if (err == nil) != testCase.loads {
				test.Errorf("Load of a post published %s, updated %s = %v, want it to load: %t", testCase.published,
					testCase.updated, err, testCase.loads)
			}
		})
	}
}

// Every fault of a post or a series stands at its file and at its field, or at its line in the post's body, so the
// author finds it from the fault alone.
func TestEveryFaultOfAPostOrASeriesStandsAtItsPlace(test *testing.T) {
	cases := []struct {
		name        string
		sample      string
		file        string
		original    string
		replacement string
		want        faultPlace
	}{
		{name: "raw HTML of a block", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "<div>Its text.</div>", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "raw HTML in a line", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its <b>text</b>.", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "raw HTML in a table", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "| Kept |\n| --- |\n| <b>Most</b> |", want: faultPlace{file: samplePostFile, field: "", line: 14}},
		{name: "a heading of the title's level", sample: samplePost, file: samplePostFile, original: "## A section", replacement: "# A section", want: faultPlace{file: samplePostFile, field: "", line: 10}},
		{name: "a first heading that skips a level", sample: samplePost, file: samplePostFile, original: "## A section", replacement: "### A section", want: faultPlace{file: samplePostFile, field: "", line: 10}},
		{name: "a heading that skips a level after another", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its text.\n\n#### Deeper", want: faultPlace{file: samplePostFile, field: "", line: 14}},
		{name: "a widget", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "```widget Cache\nalt: A cache.\n```", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "an escape in the body", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its \x1b[2J text.", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "a mark of direction in the body", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its ‮ text.", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "a tab in a paragraph", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its\ttext.", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "a tab after a code span", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its `text`\there.", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "a footnote no reference cites", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its text.\n\n[^note]: A note.", want: faultPlace{file: samplePostFile, field: "", line: 14}},
		{name: "a reference to no footnote", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its text[^note].", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "a reference to no footnote on a later line", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its text,\nand [^note].", want: faultPlace{file: samplePostFile, field: "", line: 13}},
		{name: "a table's row of more cells than its header", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "| A |\n| --- |\n| 1 | 2 |", want: faultPlace{file: samplePostFile, field: "", line: 14}},
		{name: "a footnote of two paragraphs", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its text.[^note]\n\n[^note]: A note.\n\n    More of it.", want: faultPlace{file: samplePostFile, field: "", line: 14}},
		{name: "a footnote of a list", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its text.[^note]\n\n[^note]: - a point", want: faultPlace{file: samplePostFile, field: "", line: 14}},
		{name: "a footnote citing a footnote", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its text.[^one]\n\n[^one]: See [^two].\n[^two]: A source.", want: faultPlace{file: samplePostFile, field: "", line: 14}},
		{name: "a picture", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its ![a cache](cache.svg) text.", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "a thematic break", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its text.\n\n***\n\nMore.", want: faultPlace{file: samplePostFile, field: "", line: 14}},
		{name: "a hard line break by a backslash", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its text,\\\nand more.", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "a hard line break by two spaces", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its text,  \nand more.", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "a link inside the site", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its [text](/p/tidy-tool).", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "an address of another scheme in angle brackets", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its <ftp://example.test/a>.", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "a link to a mailbox of no bare address", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its [mail](mailto:Mira).", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "a link with a title", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: `Its [text](https://example.test "a title").`, want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "a reference to a control character", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its &#27;[2J text.", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "a reference to a control character in a link's address", sample: samplePost, file: samplePostFile, original: "Its text.", replacement: "Its [text](https://example.test/&#27;).", want: faultPlace{file: samplePostFile, field: "", line: 12}},
		{name: "an escape in a field", sample: samplePost, file: samplePostFile, original: "title: A sample post", replacement: `title: "A \e[2J post"`, want: faultPlace{file: samplePostFile, field: "title", line: 0}},
		{name: "no front matter", sample: samplePost, file: samplePostFile, original: "---\nid: 3\n", replacement: "id: 3\n", want: faultPlace{file: samplePostFile, field: "", line: 1}},
		{name: "front matter never closed", sample: samplePost, file: samplePostFile, original: "---\nA paragraph.", replacement: "A paragraph.", want: faultPlace{file: samplePostFile, field: "", line: 1}},
		{name: "a field no post has", sample: samplePost, file: samplePostFile, original: "id: 3", replacement: "id: 3\nauthor: Mira", want: faultPlace{file: samplePostFile, field: "", line: 0}},
		{name: "a number another post has", sample: samplePost, file: samplePostFile, original: "id: 3", replacement: "id: 1", want: faultPlace{file: samplePostFile, field: "id", line: 0}},
		{name: "no number", sample: samplePost, file: samplePostFile, original: "id: 3\n", replacement: "", want: faultPlace{file: samplePostFile, field: "id", line: 0}},
		{name: "an empty title", sample: samplePost, file: samplePostFile, original: "title: A sample post", replacement: `title: ""`, want: faultPlace{file: samplePostFile, field: "title", line: 0}},
		{name: "a title of two lines", sample: samplePost, file: samplePostFile, original: "title: A sample post", replacement: `title: "A sample\npost"`, want: faultPlace{file: samplePostFile, field: "title", line: 0}},
		{name: "no description", sample: samplePost, file: samplePostFile, original: "description: What a sample says.\n", replacement: "", want: faultPlace{file: samplePostFile, field: "description", line: 0}},
		{name: "a status of no kind", sample: samplePost, file: samplePostFile, original: "status: published", replacement: "status: posted", want: faultPlace{file: samplePostFile, field: "status", line: 0}},
		{name: "no status", sample: samplePost, file: samplePostFile, original: "status: published\n", replacement: "", want: faultPlace{file: samplePostFile, field: "status", line: 0}},
		{name: "a preview with no date", sample: samplePost, file: samplePostFile, original: "status: published\npublished: 2026-11-10T09:00+01:00\n", replacement: "status: preview\n", want: faultPlace{file: samplePostFile, field: "published", line: 0}},
		{name: "a date with no zone", sample: samplePost, file: samplePostFile, original: "published: 2026-11-10T09:00+01:00", replacement: "published: 2026-11-10T09:00", want: faultPlace{file: samplePostFile, field: "published", line: 0}},
		{name: "an update before the post", sample: samplePost, file: samplePostFile, original: "published: 2026-11-10T09:00+01:00", replacement: "published: 2026-11-10T09:00+01:00\nupdated: 2026-11-09", want: faultPlace{file: samplePostFile, field: "updated", line: 0}},
		{name: "an update of no date", sample: samplePost, file: samplePostFile, original: "published: 2026-11-10T09:00+01:00", replacement: "published: 2026-11-10T09:00+01:00\nupdated: soon", want: faultPlace{file: samplePostFile, field: "updated", line: 0}},
		{name: "a tag in capitals", sample: samplePost, file: samplePostFile, original: "status: published", replacement: "status: published\ntags: [Reading]", want: faultPlace{file: samplePostFile, field: "tags[0]", line: 0}},
		{name: "a tag twice", sample: samplePost, file: samplePostFile, original: "status: published", replacement: "status: published\ntags: [reading, reading]", want: faultPlace{file: samplePostFile, field: "tags[1]", line: 0}},
		{name: "a discussion over plain HTTP", sample: samplePost, file: samplePostFile, original: "status: published", replacement: "status: published\ndiscussions: [{site: Lobsters, address: \"http://lobsters.example.test/s/x\"}]", want: faultPlace{file: samplePostFile, field: "discussions[0].address", line: 0}},
		{name: "a discussion at no site", sample: samplePost, file: samplePostFile, original: "status: published", replacement: "status: published\ndiscussions: [{site: \"\", address: \"https://lobsters.example.test/s/x\"}]", want: faultPlace{file: samplePostFile, field: "discussions[0].site", line: 0}},
		{name: "a slug in capitals", sample: samplePost, file: "posts/Sample-Post/post.en.md", original: "id: 3", replacement: "id: 3", want: faultPlace{file: "posts/Sample-Post", field: "", line: 0}},
		{name: "a slug of digits alone", sample: samplePost, file: "posts/12/post.en.md", original: "id: 3", replacement: "id: 3", want: faultPlace{file: "posts/12", field: "", line: 0}},
		{name: "a translation", sample: samplePost, file: "posts/reading-slowly/post.ru.md", original: "id: 3", replacement: "id: 3", want: faultPlace{file: "posts/reading-slowly/post.ru.md", field: "", line: 0}},
		{name: "a revision", sample: samplePost, file: "posts/reading-slowly/revisions/2026-11-10.en.md", original: "id: 3", replacement: "id: 3", want: faultPlace{file: "posts/reading-slowly/revisions", field: "", line: 0}},
		{name: "a widget's file", sample: samplePost, file: "posts/reading-slowly/widgets/Cache.svelte", original: "id: 3", replacement: "id: 3", want: faultPlace{file: "posts/reading-slowly/widgets", field: "", line: 0}},
		{name: "a file of no part of a post", sample: samplePost, file: "posts/reading-slowly/notes.txt", original: "id: 3", replacement: "id: 3", want: faultPlace{file: "posts/reading-slowly/notes.txt", field: "", line: 0}},
		{name: "a post's folder with no post", sample: samplePost, file: "posts/empty-post/notes.txt", original: "id: 3", replacement: "id: 3", want: faultPlace{file: "posts/empty-post/post.en.md", field: "", line: 0}},
		{name: "a file among the posts' folders", sample: samplePost, file: "posts/README.md", original: "id: 3", replacement: "id: 3", want: faultPlace{file: "posts/README.md", field: "", line: 0}},
		{name: "a series naming no post", sample: sampleSeries, file: sampleSeriesFile, original: "posts: [notes-on-tools]", replacement: "posts: [notes-on-tools, no-such-post]", want: faultPlace{file: sampleSeriesFile, field: "posts[1]", line: 0}},
		{name: "a post twice in a series", sample: sampleSeries, file: sampleSeriesFile, original: "posts: [notes-on-tools]", replacement: "posts: [notes-on-tools, notes-on-tools]", want: faultPlace{file: sampleSeriesFile, field: "posts[1]", line: 0}},
		{name: "a post in two series", sample: sampleSeries, file: sampleSeriesFile, original: "posts: [notes-on-tools]", replacement: "posts: [reading-slowly]", want: faultPlace{file: sampleSeriesFile, field: "posts[0]", line: 0}},
		{name: "a series of no posts", sample: sampleSeries, file: sampleSeriesFile, original: "posts: [notes-on-tools]", replacement: "posts: []", want: faultPlace{file: sampleSeriesFile, field: "posts", line: 0}},
		{name: "a series with no title", sample: sampleSeries, file: sampleSeriesFile, original: "title: Tools", replacement: `title: ""`, want: faultPlace{file: sampleSeriesFile, field: "title", line: 0}},
		{name: "a series with no line", sample: sampleSeries, file: sampleSeriesFile, original: "line: On the tools a writer keeps.\n", replacement: "", want: faultPlace{file: sampleSeriesFile, field: "line", line: 0}},
		{name: "a part to come with no title", sample: sampleSeries, file: sampleSeriesFile, original: "posts: [notes-on-tools]", replacement: "posts: [notes-on-tools]\nto_come: [\"\"]", want: faultPlace{file: sampleSeriesFile, field: "to_come[0]", line: 0}},
		{name: "an escape in a series", sample: sampleSeries, file: sampleSeriesFile, original: "title: Tools", replacement: `title: "To\aols"`, want: faultPlace{file: sampleSeriesFile, field: "title", line: 0}},
		{name: "a field no series has", sample: sampleSeries, file: sampleSeriesFile, original: "title: Tools", replacement: "title: Tools\nauthor: Mira", want: faultPlace{file: sampleSeriesFile, field: "", line: 0}},
		{name: "a series name in capitals", sample: sampleSeries, file: "series/Tools.yaml", original: "title: Tools", replacement: "title: Tools", want: faultPlace{file: "series/Tools.yaml", field: "", line: 0}},
		{name: "a series name of digits alone", sample: sampleSeries, file: "series/12.yaml", original: "title: Tools", replacement: "title: Tools", want: faultPlace{file: "series/12.yaml", field: "", line: 0}},
		{name: "a file of no series", sample: sampleSeries, file: "series/notes.txt", original: "title: Tools", replacement: "title: Tools", want: faultPlace{file: "series/notes.txt", field: "", line: 0}},
		{name: "a folder among the series", sample: sampleSeries, file: "series/old/tools.yaml", original: "title: Tools", replacement: "title: Tools", want: faultPlace{file: "series/old", field: "", line: 0}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			written := sampleWith(test, testCase.sample, testCase.original, testCase.replacement)
			faults := faultsOf(test, writtenFile(fixtureFolder(test), testCase.file, written))
			if !slices.ContainsFunc(faults, func(fault Fault) bool { return placeOf(fault) == testCase.want }) {
				test.Errorf("faults with %s at %s = %q, want one at %+v", testCase.replacement, testCase.file, faults, testCase.want)
			}
		})
	}
}

// A character no value may hold, written in a post's text, is one fault at its line: the check of the source's
// lines names it, and the check of the text's references does not name it again.
func TestACharacterWrittenInAPostIsOneFault(test *testing.T) {
	written := sampleWith(test, samplePost, "Its text.", "Its \x1b[2J text.")
	faults := faultsOf(test, writtenFile(fixtureFolder(test), samplePostFile, written))
	if count := len(faults); count != 1 {
		test.Errorf("faults with an escape written in the text = %q, want one", faults)
	}
}

// A file a folder holds besides the posts and the series keeps none of them from being read, so it holds back no
// fault of what they say: the author fixes both in one pass.
func TestAFileBesideThePostsHoldsBackNoFaultOfWhatTheySay(test *testing.T) {
	cases := []struct {
		name  string
		file  string
		stray faultPlace
	}{
		{name: "a translation", file: "posts/sample-post/post.ru.md", stray: faultPlace{file: "posts/sample-post/post.ru.md", field: "", line: 0}},
		{name: "a file among the posts' folders", file: "posts/README.md", stray: faultPlace{file: "posts/README.md", field: "", line: 0}},
		{name: "a file of no series", file: "series/notes.txt", stray: faultPlace{file: "series/notes.txt", field: "", line: 0}},
	}
	heading := faultPlace{file: samplePostFile, field: "", line: 10}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			folder := writtenFile(fixtureFolder(test), samplePostFile, sampleWith(test, samplePost, "## A section", "# A section"))
			faults := faultsOf(test, writtenFile(folder, testCase.file, samplePost))
			places := []faultPlace{}
			for _, fault := range faults {
				places = append(places, placeOf(fault))
			}
			if !slices.Contains(places, testCase.stray) || !slices.Contains(places, heading) {
				test.Errorf("faults with %s and a heading of level 1 = %q, want one at %+v and one at %+v", testCase.file, faults,
					testCase.stray, heading)
			}
		})
	}
}

func placeOf(fault Fault) faultPlace {
	return faultPlace{file: fault.File, field: fault.Field, line: fault.Line}
}

// sampleWith is a sample with one piece of it written otherwise; the piece must be there once.
func sampleWith(test *testing.T, sample, original, replacement string) string {
	test.Helper()
	if strings.Count(sample, original) != 1 {
		test.Fatalf("the sample holds %q %d times, want once", original, strings.Count(sample, original))
	}
	return strings.Replace(sample, original, replacement, 1)
}

// postNamed is the post of a slug in a snapshot.
func postNamed(test *testing.T, snapshot *Snapshot, slug string) Post {
	test.Helper()
	index := slices.IndexFunc(snapshot.Posts, func(post Post) bool { return post.Slug == slug })
	if index < 0 {
		test.Fatalf("the snapshot holds no post %q", slug)
	}
	return snapshot.Posts[index]
}

// kindsOf are the kinds of a tree's nodes, the root's first and then its children's, each in turn.
func kindsOf(node ast.Node) []ast.NodeKind {
	kinds := []ast.NodeKind{node.Kind()}
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		kinds = append(kinds, kindsOf(child)...)
	}
	return kinds
}
