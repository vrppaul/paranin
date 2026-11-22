package compose

import (
	"os"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/yuin/goldmark/ast"
	extensionast "github.com/yuin/goldmark/extension/ast"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// The fixture's post of every kind of block, and its draft, whose file a case writes over with a post of its own
const (
	fixturePostSlug  = "reading-slowly"
	samplePostSlug   = "notes-on-tools"
	samplePostFile   = "posts/notes-on-tools/post.en.md"
	samplePostFields = "---\nid: 2\ntitle: Notes on tools\ndescription: The tools a writer keeps.\npublished: 2026-11-10T09:00+01:00\n"
)

// A post's body becomes the document's blocks kind by kind: its marks spans, its escapes and references read, its
// headings sections that nest by their level, its lists, quotes, code and tables blocks, and each citation of a
// footnote the note itself, numbered by the order of the first citations.
func TestAPostsBodyBecomesItsBlocks(test *testing.T) {
	cases := []struct {
		name string
		body string
		want []document.Block
	}{
		{
			name: "marks in a line",
			body: "Its *stressed*, **weighty** and `coded` words.",
			want: []document.Block{paragraphOf(
				document.Text{Text: "Its "}, document.Emphasis{Spans: textSpans("stressed")}, document.Text{Text: ", "},
				document.Strong{Spans: textSpans("weighty")}, document.Text{Text: " and "}, document.Code{Text: "coded"},
				document.Text{Text: " words."},
			)},
		},
		{name: "escapes and references read", body: `A &amp; \*star\*, &#233;.`, want: []document.Block{paragraphOf(document.Text{Text: "A & *star*, é."})}},
		{name: "a reference escaped, read as written", body: `\&amp; and \&ouml;`, want: []document.Block{paragraphOf(document.Text{Text: "&amp; and &ouml;"})}},
		{
			name: "a link's address read as goldmark reads one, its escapes before its references",
			body: `[query](https://example.test/?a=1\&amp;b=2)`,
			want: []document.Block{paragraphOf(document.Link{Target: document.Address{URL: "https://example.test/?a=1&b=2"}, Spans: textSpans("query")})},
		},
		{
			name: "code fenced with words after its language",
			body: "```go title=\"main.go\"\nmain()\n```",
			want: []document.Block{document.CodeBlock{Text: "main()", Language: "go"}},
		},
		{name: "a paragraph's lines joined", body: "Its text,\nand more.", want: []document.Block{paragraphOf(document.Text{Text: "Its text, and more."})}},
		{
			name: "links outside the site and to a mailbox",
			body: "[notes](https://notes.example.test) [mail](mailto:mira@example.test) <https://example.test> <mira@example.test>",
			want: []document.Block{paragraphOf(
				document.Link{Target: document.Address{URL: "https://notes.example.test"}, Spans: textSpans("notes")}, document.Text{Text: " "},
				document.Link{Target: document.Mail{Address: "mira@example.test"}, Spans: textSpans("mail")}, document.Text{Text: " "},
				document.Link{Target: document.Address{URL: "https://example.test"}, Spans: textSpans("https://example.test")}, document.Text{Text: " "},
				document.Link{Target: document.Mail{Address: "mira@example.test"}, Spans: textSpans("mira@example.test")},
			)},
		},
		{
			name: "headings nest into sections by their level",
			body: "Before.\n\n## One\n\nIn one.\n\n### Inner\n\nIn inner.\n\n## Two\n\nIn two.",
			want: []document.Block{
				paragraphOf(document.Text{Text: "Before."}),
				document.Section{Title: textSpans("One"), Blocks: []document.Block{
					paragraphOf(document.Text{Text: "In one."}),
					document.Section{Title: textSpans("Inner"), Blocks: []document.Block{paragraphOf(document.Text{Text: "In inner."})}},
				}},
				document.Section{Title: textSpans("Two"), Blocks: []document.Block{paragraphOf(document.Text{Text: "In two."})}},
			},
		},
		{
			name: "a list nested in an item",
			body: "- one\n- two\n  - inner",
			want: []document.Block{document.List{Ordered: false, Start: 0, Items: []document.Item{
				{Blocks: []document.Block{paragraphOf(document.Text{Text: "one"})}},
				{Blocks: []document.Block{
					paragraphOf(document.Text{Text: "two"}),
					document.List{Ordered: false, Start: 0, Items: []document.Item{{Blocks: []document.Block{paragraphOf(document.Text{Text: "inner"})}}}},
				}},
			}}},
		},
		{
			name: "a list in order from its first number",
			body: "3. three\n4. four",
			want: []document.Block{document.List{Ordered: true, Start: 3, Items: []document.Item{
				{Blocks: []document.Block{paragraphOf(document.Text{Text: "three"})}},
				{Blocks: []document.Block{paragraphOf(document.Text{Text: "four"})}},
			}}},
		},
		{
			name: "a quote of two paragraphs",
			body: "> One.\n>\n> Two.",
			want: []document.Block{document.Quote{Blocks: []document.Block{paragraphOf(document.Text{Text: "One."}), paragraphOf(document.Text{Text: "Two."})}}},
		},
		{
			name: "code fenced with its language",
			body: "```go\nfunc main() {\n\tprintln()\n}\n```",
			want: []document.Block{document.CodeBlock{Text: "func main() {\n\tprintln()\n}", Language: "go"}},
		},
		{name: "code indented", body: "    ls -la\n    ls", want: []document.Block{document.CodeBlock{Text: "ls -la\nls", Language: ""}}},
		{
			name: "a table under its header",
			body: "| Reading | Kept |\n| --- | --- |\n| Slow | **Most** |",
			want: []document.Block{document.Table{
				Header: document.Row{Cells: []document.Cell{{Spans: textSpans("Reading")}, {Spans: textSpans("Kept")}}},
				Rows:   []document.Row{{Cells: []document.Cell{{Spans: textSpans("Slow")}, {Spans: []document.Span{document.Strong{Spans: textSpans("Most")}}}}}},
			}},
		},
		{
			name: "a note at each citation, numbered by the first citations",
			body: "One[^b] two[^a] three[^b].\n\n[^a]: Of *a*.\n[^b]: Of b.",
			want: []document.Block{paragraphOf(
				document.Text{Text: "One"}, document.Note{Number: 1, Spans: textSpans("Of b.")},
				document.Text{Text: " two"}, document.Note{Number: 2, Spans: []document.Span{document.Text{Text: "Of "}, document.Emphasis{Spans: textSpans("a")}, document.Text{Text: "."}}},
				document.Text{Text: " three"}, document.Note{Number: 1, Spans: textSpans("Of b.")}, document.Text{Text: "."},
			)},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			snapshot := withSamplePost(test, samplePostFields+"status: published\n---\n"+testCase.body+"\n")
			if got := postNamed(test, Posts(snapshot), samplePostSlug).Blocks; !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("the body %q composes to\n%#v\nwant\n%#v", testCase.body, got, testCase.want)
			}
		})
	}
}

// Every word a post writes reaches its document as a phrase of its own, wherever it stands: a face that shows every
// phrase shows the whole post.
func TestEveryWordOfAPostReachesItsDocument(test *testing.T) {
	phrases := document.Phrases(postNamed(test, Posts(fixtures.LoadedContent(test)), fixturePostSlug))
	cases := []struct {
		name string
		want string
	}{
		{name: "the title", want: "Reading slowly"},
		{name: "a tag", want: "terminals"},
		{name: "stressed words", want: "slow"},
		{name: "a heading", want: "What a reader keeps"},
		{name: "a table's header", want: "Reading"},
		{name: "strong words in a cell", want: "Most"},
		{name: "a list nested in an item", want: "in ink"},
		{name: "a quote's second paragraph", want: "Twice, with a pen."},
		{name: "code", want: `fmt.Println("slowly")`},
		{name: "a link's words", want: "the reader's notes"},
		{name: "a note's words", want: "A soft one, which rubs out."},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if !slices.Contains(phrases, testCase.want) {
				test.Errorf("phrases = %q, want one %q", phrases, testCase.want)
			}
		})
	}
}

// Every kind of node a post may hold (content.PostKinds) has its case in the walk, so its words reach the document:
// a case of each kind, and no kind the load lets through without one, so a kind an extension brings is drawn here
// before any post holds it.
func TestEveryKindAPostHoldsReachesItsDocument(test *testing.T) {
	cases := []struct {
		kind   ast.NodeKind
		body   string
		phrase string
	}{
		{kind: ast.KindDocument, body: "A word.", phrase: "A word."},
		{kind: ast.KindParagraph, body: "A word.", phrase: "A word."},
		{kind: ast.KindText, body: "A word.", phrase: "A word."},
		{kind: ast.KindTextBlock, body: "- an item", phrase: "an item"},
		{kind: ast.KindList, body: "- an item", phrase: "an item"},
		{kind: ast.KindListItem, body: "- an item", phrase: "an item"},
		{kind: ast.KindHeading, body: "## A heading", phrase: "A heading"},
		{kind: ast.KindFencedCodeBlock, body: "```go\nfenced()\n```", phrase: "fenced()"},
		{kind: ast.KindCodeBlock, body: "    indented()", phrase: "indented()"},
		{kind: ast.KindBlockquote, body: "> Quoted.", phrase: "Quoted."},
		{kind: ast.KindLinkReferenceDefinition, body: "[defined][x]\n\n[x]: https://example.test", phrase: "defined"},
		{kind: ast.KindCodeSpan, body: "A `span`.", phrase: "span"},
		{kind: ast.KindEmphasis, body: "*stressed*", phrase: "stressed"},
		{kind: ast.KindLink, body: "[linked](https://example.test)", phrase: "linked"},
		{kind: ast.KindAutoLink, body: "<https://example.test>", phrase: "https://example.test"},
		{kind: extensionast.KindTable, body: "| Head |\n| --- |\n| Cell |", phrase: "Cell"},
		{kind: extensionast.KindTableHeader, body: "| Head |\n| --- |\n| Cell |", phrase: "Head"},
		{kind: extensionast.KindTableRow, body: "| Head |\n| --- |\n| Cell |", phrase: "Cell"},
		{kind: extensionast.KindTableCell, body: "| Head |\n| --- |\n| Cell |", phrase: "Cell"},
		{kind: extensionast.KindFootnoteList, body: "Cited.[^a]\n\n[^a]: A note.", phrase: "A note."},
		{kind: extensionast.KindFootnote, body: "Cited.[^a]\n\n[^a]: A note.", phrase: "A note."},
		{kind: extensionast.KindFootnoteLink, body: "Cited.[^a]\n\n[^a]: A note.", phrase: "A note."},
	}
	kinds := []ast.NodeKind{}
	for _, testCase := range cases {
		kinds = append(kinds, testCase.kind)
		test.Run(testCase.kind.String(), func(test *testing.T) {
			snapshot := withSamplePost(test, samplePostFields+"status: published\n---\n"+testCase.body+"\n")
			post := snapshot.Posts[slices.IndexFunc(snapshot.Posts, func(post content.Post) bool { return post.Slug == samplePostSlug })]
			phrases := document.Phrases(postNamed(test, Posts(snapshot), samplePostSlug))
			if !slices.Contains(kindsOf(post.Body.Tree), testCase.kind) || !slices.Contains(phrases, testCase.phrase) {
				test.Errorf("the body %q = kinds %v, phrases %q; want a %v among the kinds and %q among the phrases",
					testCase.body, kindsOf(post.Body.Tree), phrases, testCase.kind, testCase.phrase)
			}
		})
	}
	if !sameKinds(kinds, content.PostKinds()) {
		test.Errorf("the kinds a case holds = %v, want every kind a post may hold, %v", kinds, content.PostKinds())
	}
}

// sameKinds is whether two lists of kinds name the same kinds, whatever their order.
func sameKinds(first, second []ast.NodeKind) bool {
	return len(first) == len(second) && !slices.ContainsFunc(first, func(kind ast.NodeKind) bool { return !slices.Contains(second, kind) })
}

// kindsOf are the kinds of a tree's nodes, the root's first and then its children's, each in turn.
func kindsOf(node ast.Node) []ast.NodeKind {
	kinds := []ast.NodeKind{node.Kind()}
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		kinds = append(kinds, kindsOf(child)...)
	}
	return kinds
}

// A post is served while it is published, and while it is unlisted, at its address though no list names it; a draft
// and a preview never.
func TestAPostIsServedWhilePublishedOrUnlisted(test *testing.T) {
	cases := []struct {
		status content.Status
		served bool
	}{
		{status: content.StatusPublished, served: true},
		{status: content.StatusUnlisted, served: true},
		{status: content.StatusDraft, served: false},
		{status: content.StatusPreview, served: false},
	}
	for _, testCase := range cases {
		test.Run(string(testCase.status), func(test *testing.T) {
			snapshot := withSamplePost(test, samplePostFields+"status: "+string(testCase.status)+"\n---\nIts text.\n")
			slugs := []string{}
			for _, post := range Posts(snapshot) {
				slugs = append(slugs, post.Head.Page.Slug)
			}
			if slices.Contains(slugs, samplePostSlug) != testCase.served {
				test.Errorf("the posts served = %q, want %q among them: %t", slugs, samplePostSlug, testCase.served)
			}
		})
	}
}

// A post's head holds what its front matter says a face shows first: its page by its slug, its number, the
// language of the site's words, its title, its description as its lead and its page's description, its tags, and
// when it came out, in the zone it was written in.
func TestAPostsHeadHoldsItsFrontMatter(test *testing.T) {
	head := postNamed(test, Posts(fixtures.LoadedContent(test)), fixturePostSlug).Head
	published := head.Published
	head.Published = time.Time{}
	description := textSpans("Why a page read twice is read once.")
	want := document.Head{
		Page: document.Page{Kind: document.PagePost, Slug: fixturePostSlug, Tab: ""}, Number: 1, Language: "en",
		Greeting: "", Title: textSpans("Reading slowly"), Subtitle: nil, Place: "", Tags: []string{"reading", "terminals"},
		Lead: description, Published: time.Time{}, Repository: nil, Description: description,
	}
	if !reflect.DeepEqual(head, want) {
		test.Errorf("the head = %+v, want %+v", head, want)
	}
	if got, wantPublished := published.Format(time.RFC3339), "2026-11-10T09:00:00+01:00"; got != wantPublished {
		test.Errorf("the head's day of publishing = %q, want %q", got, wantPublished)
	}
}

// Two goroutines may compose one snapshot's posts at once (release.Derived), reading one parsed tree: they make the
// same documents, and under -race, as its hook runs it, neither writes what the other reads.
func TestTwoComposesOfOnePostAgree(test *testing.T) {
	snapshot := fixtures.LoadedContent(test)
	composed := make([][]document.Document, 2)
	var group sync.WaitGroup
	for index := range composed {
		group.Go(func() { composed[index] = Posts(snapshot) })
	}
	group.Wait()
	if !reflect.DeepEqual(composed[0], composed[1]) {
		test.Errorf("two composes at once = %+v and %+v, want them alike", composed[0], composed[1])
	}
}

// withSamplePost is the fixture content with the draft's file written over by a post's text, loaded.
func withSamplePost(test *testing.T, text string) *content.Snapshot {
	test.Helper()
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	fixtures.WriteIn(test, folder, samplePostFile, text)
	snapshot, err := content.Load(os.DirFS(folder))
	if err != nil {
		test.Fatalf("load the fixture with the post %q: %v", text, err)
	}
	return snapshot
}

// postNamed is the document of the post of a slug among documents.
func postNamed(test *testing.T, posts []document.Document, slug string) document.Document {
	test.Helper()
	return fixtures.DocumentOf(test, posts, document.Page{Kind: document.PagePost, Slug: slug, Tab: ""})
}

func paragraphOf(spans ...document.Span) document.Paragraph {
	return document.Paragraph{Spans: spans}
}
