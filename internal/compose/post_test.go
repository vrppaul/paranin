package compose

import (
	"os"
	"path"
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

// A drawing a case writes beside the sample post, by its name, four pixels by two
const (
	sampleDrawingName = "margin.svg"
	sampleDrawing     = `<svg xmlns="http://www.w3.org/2000/svg" width="4" height="2"/>`
)

// A post's body becomes the document's blocks kind by kind: its marks spans, its escapes and references read, its
// headings sections that nest by their level, its lists, quotes, code, tables and pauses blocks, and each citation of a
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
			want: []document.Block{document.CodeBlock{Language: "go", Runs: []document.CodeRun{
				{Kind: document.CodeFunction, Text: "main"}, {Kind: document.CodePunctuation, Text: "()"},
			}}},
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
			name: "headings nest into sections by their level, each named by its words",
			body: "Before.\n\n## One\n\nIn one.\n\n### Inner\n\nIn inner.\n\n## Two\n\nIn two.",
			want: []document.Block{
				paragraphOf(document.Text{Text: "Before."}),
				document.Section{Title: textSpans("One"), Name: "one", Blocks: []document.Block{
					paragraphOf(document.Text{Text: "In one."}),
					document.Section{Title: textSpans("Inner"), Name: "inner", Blocks: []document.Block{paragraphOf(document.Text{Text: "In inner."})}},
				}},
				document.Section{Title: textSpans("Two"), Name: "two", Blocks: []document.Block{paragraphOf(document.Text{Text: "In two."})}},
			},
		},
		{
			name: "a heading of words named before takes its count after them, and one of no words a name of its own",
			body: "## Why\n\nOne.\n\n## Why\n\nTwo.\n\n## Why 2\n\nThree.\n\n## ?!\n\nFour.",
			want: []document.Block{
				document.Section{Title: textSpans("Why"), Name: "why", Blocks: []document.Block{paragraphOf(document.Text{Text: "One."})}},
				document.Section{Title: textSpans("Why"), Name: "why-2", Blocks: []document.Block{paragraphOf(document.Text{Text: "Two."})}},
				document.Section{Title: textSpans("Why 2"), Name: "why-2-2", Blocks: []document.Block{paragraphOf(document.Text{Text: "Three."})}},
				document.Section{Title: textSpans("?!"), Name: "heading", Blocks: []document.Block{paragraphOf(document.Text{Text: "Four."})}},
			},
		},
		{
			name: "a heading of the name of another place of the page, a window, a project's page or the keys', takes its count",
			body: "## Posts\n\nOne.\n\n## Notes on tools\n\nTwo.\n\n## Tidy tool\n\nThree.\n\n## Help\n\nFour.",
			want: []document.Block{
				document.Section{Title: textSpans("Posts"), Name: "posts-2", Blocks: []document.Block{paragraphOf(document.Text{Text: "One."})}},
				document.Section{Title: textSpans("Notes on tools"), Name: "notes-on-tools-2", Blocks: []document.Block{paragraphOf(document.Text{Text: "Two."})}},
				document.Section{Title: textSpans("Tidy tool"), Name: "tidy-tool-2", Blocks: []document.Block{paragraphOf(document.Text{Text: "Three."})}},
				document.Section{Title: textSpans("Help"), Name: "help-2", Blocks: []document.Block{paragraphOf(document.Text{Text: "Four."})}},
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
			name: "a pause between paragraphs",
			body: "Before.\n\n***\n\nAfter.",
			want: []document.Block{paragraphOf(document.Text{Text: "Before."}), document.Pause{}, paragraphOf(document.Text{Text: "After."})},
		},
		{
			name: "a table under its header",
			body: "| Reading | Kept |\n| --- | --- |\n| Slow | **Most** |",
			want: []document.Block{document.Table{
				Header: document.Row{Cells: []document.Cell{
					{Kind: document.CellText, Spans: textSpans("Reading")},
					{Kind: document.CellText, Spans: textSpans("Kept")},
				}},
				Rows: []document.Row{{Cells: []document.Cell{
					{Kind: document.CellText, Spans: textSpans("Slow")},
					{Kind: document.CellText, Spans: []document.Span{document.Strong{Spans: textSpans("Most")}}},
				}}},
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

// A picture standing alone in a post's body is a figure, wherever it stands: at its file beside the post, its alt text
// and its caption read as a reader reads them, its size its file's.
func TestAPictureStandingAloneIsAFigure(test *testing.T) {
	file := fixtureFile(samplePostSlug, sampleDrawingName)
	cases := []struct {
		name string
		body string
		want []document.Block
	}{
		{
			name: "a picture under its caption", body: `![A page, read](margin.svg "The margin")`,
			want: []document.Block{document.Figure{Picture: file, Alt: "A page, read", Caption: "The margin", Width: 4, Height: 2}},
		},
		{
			name: "a picture with no caption, its alt text's lines joined", body: "![A page\nread twice](margin.svg)",
			want: []document.Block{document.Figure{Picture: file, Alt: "A page read twice", Caption: "", Width: 4, Height: 2}},
		},
		{
			name: "escapes and references read", body: `![A &amp; B](margin.svg "C &amp; \"D\"")`,
			want: []document.Block{document.Figure{Picture: file, Alt: "A & B", Caption: `C & "D"`, Width: 4, Height: 2}},
		},
		{
			name: "a picture in a list's item, after its words", body: "- One:\n\n  ![A page](margin.svg)",
			want: []document.Block{document.List{Ordered: false, Start: 0, Items: []document.Item{
				{Blocks: []document.Block{
					paragraphOf(document.Text{Text: "One:"}), document.Figure{Picture: file, Alt: "A page", Caption: "", Width: 4, Height: 2},
				}},
			}}},
		},
		{
			name: "a picture in a quote", body: "> ![A page](margin.svg)",
			want: []document.Block{document.Quote{Blocks: []document.Block{document.Figure{Picture: file, Alt: "A page", Caption: "", Width: 4, Height: 2}}}},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			snapshot := withSamplePost(test, samplePostFields+"status: published\n---\n"+testCase.body+"\n",
				samplePicture{name: sampleDrawingName, text: sampleDrawing})
			if got := postNamed(test, Posts(snapshot), samplePostSlug).Blocks; !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("the body %q composes to\n%#v\nwant\n%#v", testCase.body, got, testCase.want)
			}
		})
	}
}

// The site serves the pictures of the posts it serves, each at its file beside its post, and a draft's never.
func TestThePicturesServedAreThePostsServed(test *testing.T) {
	cases := []struct {
		name   string
		status content.Status
		want   []document.File
	}{
		{name: "a published post's", status: content.StatusPublished, want: []document.File{fixtureFile(fixturePostSlug, "margin.png"), fixtureFile(samplePostSlug, sampleDrawingName)}},
		{name: "a draft's", status: content.StatusDraft, want: []document.File{fixtureFile(fixturePostSlug, "margin.png")}},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			snapshot := withSamplePost(test, samplePostFields+"status: "+string(testCase.status)+"\n---\n![A page](margin.svg)\n",
				samplePicture{name: sampleDrawingName, text: sampleDrawing})
			files := []document.File{}
			for _, served := range Pictures(snapshot) {
				files = append(files, served.File)
			}
			if !sameMembers(files, testCase.want) {
				test.Errorf("the pictures served = %+v, want %+v", files, testCase.want)
			}
		})
	}
}

// fixtureFile is the file of a name beside the post of a slug.
func fixtureFile(slug, name string) document.File {
	return document.File{Page: document.Page{Kind: document.PagePost, Slug: slug, Tab: ""}, Name: name}
}

// A post's code is split into runs of the kinds of word a reader of code tells apart, by the language its author
// names in any case, each run the longest of its kind, so a face colours each kind; code in a language compose lexes
// none of, or in none, stays one plain run, its text whole.
func TestCodeIsSplitIntoItsKindsOfWord(test *testing.T) {
	cases := []struct {
		name string
		body string
		want document.CodeBlock
	}{
		{
			name: "a Go function",
			body: "```go\nfunc twice(page string) int {\n\treturn len(page) * 2 // twice\n}\n```",
			want: document.CodeBlock{Language: "go", Runs: []document.CodeRun{
				{Kind: document.CodeKeyword, Text: "func"},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeFunction, Text: "twice"},
				{Kind: document.CodePunctuation, Text: "("},
				{Kind: document.CodePlain, Text: "page "},
				{Kind: document.CodeType, Text: "string"},
				{Kind: document.CodePunctuation, Text: ")"},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeType, Text: "int"},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodePunctuation, Text: "{"},
				{Kind: document.CodePlain, Text: "\n\t"},
				{Kind: document.CodeKeyword, Text: "return"},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeBuiltin, Text: "len"},
				{Kind: document.CodePunctuation, Text: "("},
				{Kind: document.CodePlain, Text: "page"},
				{Kind: document.CodePunctuation, Text: ")"},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeOperator, Text: "*"},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeNumber, Text: "2"},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeComment, Text: "// twice"},
				{Kind: document.CodePlain, Text: "\n"},
				{Kind: document.CodePunctuation, Text: "}"},
			}},
		},
		{
			name: "a Go call of a string and a constant",
			body: "```go\nfmt.Println(\"twice\", true)\n```",
			want: document.CodeBlock{Language: "go", Runs: []document.CodeRun{
				{Kind: document.CodePlain, Text: "fmt"},
				{Kind: document.CodePunctuation, Text: "."},
				{Kind: document.CodeFunction, Text: "Println"},
				{Kind: document.CodePunctuation, Text: "("},
				{Kind: document.CodeString, Text: `"twice"`},
				{Kind: document.CodePunctuation, Text: ","},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeNumber, Text: "true"},
				{Kind: document.CodePunctuation, Text: ")"},
			}},
		},
		{
			name: "a shell's line",
			body: "```sh\nPAGER=less man \"$HOME/notes\"\n```",
			want: document.CodeBlock{Language: "sh", Runs: []document.CodeRun{
				{Kind: document.CodeVariable, Text: "PAGER"},
				{Kind: document.CodeOperator, Text: "="},
				{Kind: document.CodePlain, Text: "less man "},
				{Kind: document.CodeString, Text: `"`},
				{Kind: document.CodeVariable, Text: "$HOME"},
				{Kind: document.CodeString, Text: `/notes"`},
			}},
		},
		{
			name: "YAML",
			body: "```yaml\ntitle: Slowly\ncount: 2 # read twice\nread: 2026-11-10\n```",
			want: document.CodeBlock{Language: "yaml", Runs: []document.CodeRun{
				{Kind: document.CodeFunction, Text: "title"},
				{Kind: document.CodePunctuation, Text: ":"},
				{Kind: document.CodePlain, Text: " Slowly\n"},
				{Kind: document.CodeFunction, Text: "count"},
				{Kind: document.CodePunctuation, Text: ":"},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeNumber, Text: "2"},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeComment, Text: "# read twice"},
				{Kind: document.CodePlain, Text: "\n"},
				{Kind: document.CodeFunction, Text: "read"},
				{Kind: document.CodePunctuation, Text: ":"},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeString, Text: "2026-11-10"},
			}},
		},
		{
			name: "a Python class, its operators written as words plain",
			body: "```python\nclass Page:\n    read = 2 in pages and None\n```",
			want: document.CodeBlock{Language: "python", Runs: []document.CodeRun{
				{Kind: document.CodeKeyword, Text: "class"},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeType, Text: "Page"},
				{Kind: document.CodePunctuation, Text: ":"},
				{Kind: document.CodePlain, Text: "\n    read "},
				{Kind: document.CodeOperator, Text: "="},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeNumber, Text: "2"},
				{Kind: document.CodePlain, Text: " in pages and "},
				{Kind: document.CodeNumber, Text: "None"},
			}},
		},
		{
			name: "a Rust constant",
			body: "```Rust\nconst MAX: u32 = 2;\n```",
			want: document.CodeBlock{Language: "Rust", Runs: []document.CodeRun{
				{Kind: document.CodeKeyword, Text: "const"},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeNumber, Text: "MAX"},
				{Kind: document.CodePlain, Text: ": "},
				{Kind: document.CodeType, Text: "u32"},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeOperator, Text: "="},
				{Kind: document.CodePlain, Text: " "},
				{Kind: document.CodeNumber, Text: "2"},
				{Kind: document.CodePunctuation, Text: ";"},
			}},
		},
		{
			name: "a language the highlighter does not know",
			body: "```klingon\nqaStaH nuq \"'e'\"\n```",
			want: document.CodeBlock{Language: "klingon", Runs: []document.CodeRun{{Kind: document.CodePlain, Text: `qaStaH nuq "'e'"`}}},
		},
		{
			name: "a language the highlighter knows and compose lexes none of",
			body: "```ruby\nPAGES = 2\n```",
			want: document.CodeBlock{Language: "ruby", Runs: []document.CodeRun{{Kind: document.CodePlain, Text: "PAGES = 2"}}},
		},
		{
			name: "code indented, in no language",
			body: "    ls -la\n    ls",
			want: document.CodeBlock{Language: "", Runs: []document.CodeRun{{Kind: document.CodePlain, Text: "ls -la\nls"}}},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			snapshot := withSamplePost(test, samplePostFields+"status: published\n---\n"+testCase.body+"\n")
			want := []document.Block{testCase.want}
			if got := postNamed(test, Posts(snapshot), samplePostSlug).Blocks; !reflect.DeepEqual(got, want) {
				test.Errorf("the body %q composes to\n%#v\nwant\n%#v", testCase.body, got, want)
			}
		})
	}
}

// In every language compose lexes, a post's code is read into more than one kind of word and its runs join to its
// text whole, a letter past ASCII among it, no line end added or changed, since a reader copies the text and the text
// faces write it; and a case stands for every language lexedLanguages names, so a language added is held too.
func TestCodeRunsJoinToTheCodesTextWhole(test *testing.T) {
	cases := []struct {
		language string
		code     string
	}{
		{language: "go", code: `x := "é"`},
		{language: "sh", code: `echo "$HOME" é`},
		{language: "bash", code: `echo "$HOME" é`},
		{language: "shell", code: `echo "$HOME" é`},
		{language: "console", code: "$ echo \"$HOME\"\n/home/é"},
		{language: "shell-session", code: "$ echo \"$HOME\"\n/home/é"},
		{language: "yaml", code: "title: é # read"},
		{language: "yml", code: "title: é # read"},
		{language: "json", code: `{"title": "é"}`},
		{language: "toml", code: `title = "é"`},
		{language: "python", code: `def read(): return "é"`},
		{language: "py", code: `def read(): return "é"`},
		{language: "rust", code: `fn main() { let title = "é"; }`},
		{language: "rs", code: `fn main() { let title = "é"; }`},
		{language: "typescript", code: `const title: string = "é"`},
		{language: "ts", code: `const title: string = "é"`},
		{language: "javascript", code: `const title = "é"`},
		{language: "js", code: `const title = "é"`},
		{language: "sql", code: "SELECT 'é' FROM pages"},
		{language: "dockerfile", code: "FROM golang:1.26 # é"},
		{language: "makefile", code: "all:\n\tgo build é"},
		{language: "make", code: "all:\n\tgo build é"},
		{language: "html", code: `<p class="title">é</p>`},
		{language: "css", code: "p { color: red } /* é */"},
		{language: "c", code: `int main(void) { return 0; } // é`},
		{language: "svelte", code: `<script>let title = "é"</script>`},
	}
	languages := []string{}
	for _, testCase := range cases {
		languages = append(languages, testCase.language)
		test.Run(testCase.language, func(test *testing.T) {
			snapshot := withSamplePost(test, samplePostFields+"status: published\n---\n```"+testCase.language+"\n"+testCase.code+"\n```\n")
			code := codeOf(test, postNamed(test, Posts(snapshot), samplePostSlug))
			if got := code.Text(); got != testCase.code || len(kindsOfRuns(code.Runs)) < 2 {
				test.Errorf("the code %q in %s = runs %+v, joined %q; want them joined to the code, of more than one kind",
					testCase.code, testCase.language, code.Runs, got)
			}
		})
	}
	if lexed := lexedNames(); !sameMembers(languages, lexed) {
		test.Errorf("the languages a case holds = %q, want every language compose lexes, %q", languages, lexed)
	}
}

// lexedNames are the names of every language compose lexes, as an author writes one.
func lexedNames() []string {
	names := []string{}
	for _, lexed := range lexedLanguages() {
		names = append(names, lexed.names...)
	}
	return names
}

// codeOf is the one block of a post whose body is a block of code.
func codeOf(test *testing.T, post document.Document) document.CodeBlock {
	test.Helper()
	code, isCode := post.Blocks[0].(document.CodeBlock)
	if len(post.Blocks) != 1 || !isCode {
		test.Fatalf("the post's blocks = %+v, want one block of code", post.Blocks)
	}
	return code
}

// kindsOfRuns are the kinds runs of code hold, each once.
func kindsOfRuns(runs []document.CodeRun) map[document.CodeKind]bool {
	kinds := map[document.CodeKind]bool{}
	for _, run := range runs {
		kinds[run.Kind] = true
	}
	return kinds
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
		{name: "code", want: `less "$HOME/notes"`},
		{name: "a link's words", want: "the reader's notes"},
		{name: "a note's words", want: "A soft one, which rubs out."},
		{name: "a picture's alt text", want: "A noted margin"},
		{name: "a picture's caption", want: "The margin, where a slow reader writes back."},
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
		kind     ast.NodeKind
		body     string
		pictures []samplePicture
		phrase   string
	}{
		{kind: ast.KindDocument, body: "A word.", pictures: nil, phrase: "A word."},
		{kind: ast.KindParagraph, body: "A word.", pictures: nil, phrase: "A word."},
		{kind: ast.KindText, body: "A word.", pictures: nil, phrase: "A word."},
		{kind: ast.KindTextBlock, body: "- an item", pictures: nil, phrase: "an item"},
		{kind: ast.KindList, body: "- an item", pictures: nil, phrase: "an item"},
		{kind: ast.KindListItem, body: "- an item", pictures: nil, phrase: "an item"},
		{kind: ast.KindHeading, body: "## A heading", pictures: nil, phrase: "A heading"},
		{kind: ast.KindFencedCodeBlock, body: "```go\nfenced()\n```", pictures: nil, phrase: "fenced()"},
		{kind: ast.KindCodeBlock, body: "    indented()", pictures: nil, phrase: "indented()"},
		{kind: ast.KindBlockquote, body: "> Quoted.", pictures: nil, phrase: "Quoted."},
		// a break has no phrase of its own: its row keeps the kinds whole, and TestAPostsBodyBecomesItsBlocks holds its block
		{kind: ast.KindThematicBreak, body: "Before.\n\n***\n\nAfter.", pictures: nil, phrase: "After."},
		{kind: ast.KindLinkReferenceDefinition, body: "[defined][x]\n\n[x]: https://example.test", pictures: nil, phrase: "defined"},
		{kind: ast.KindCodeSpan, body: "A `span`.", pictures: nil, phrase: "span"},
		{kind: ast.KindEmphasis, body: "*stressed*", pictures: nil, phrase: "stressed"},
		{kind: ast.KindLink, body: "[linked](https://example.test)", pictures: nil, phrase: "linked"},
		{kind: ast.KindAutoLink, body: "<https://example.test>", pictures: nil, phrase: "https://example.test"},
		{
			kind: ast.KindImage, body: "![A page](margin.svg)", pictures: []samplePicture{{name: sampleDrawingName, text: sampleDrawing}},
			phrase: "A page",
		},
		{kind: extensionast.KindTable, body: "| Head |\n| --- |\n| Cell |", pictures: nil, phrase: "Cell"},
		{kind: extensionast.KindTableHeader, body: "| Head |\n| --- |\n| Cell |", pictures: nil, phrase: "Head"},
		{kind: extensionast.KindTableRow, body: "| Head |\n| --- |\n| Cell |", pictures: nil, phrase: "Cell"},
		{kind: extensionast.KindTableCell, body: "| Head |\n| --- |\n| Cell |", pictures: nil, phrase: "Cell"},
		{kind: extensionast.KindFootnoteList, body: "Cited.[^a]\n\n[^a]: A note.", pictures: nil, phrase: "A note."},
		{kind: extensionast.KindFootnote, body: "Cited.[^a]\n\n[^a]: A note.", pictures: nil, phrase: "A note."},
		{kind: extensionast.KindFootnoteLink, body: "Cited.[^a]\n\n[^a]: A note.", pictures: nil, phrase: "A note."},
	}
	kinds := []ast.NodeKind{}
	for _, testCase := range cases {
		kinds = append(kinds, testCase.kind)
		test.Run(testCase.kind.String(), func(test *testing.T) {
			snapshot := withSamplePost(test, samplePostFields+"status: published\n---\n"+testCase.body+"\n", testCase.pictures...)
			post := snapshot.Posts[slices.IndexFunc(snapshot.Posts, func(post content.Post) bool { return post.Slug == samplePostSlug })]
			phrases := document.Phrases(postNamed(test, Posts(snapshot), samplePostSlug))
			if !slices.Contains(kindsOf(post.Body.Tree), testCase.kind) || !slices.Contains(phrases, testCase.phrase) {
				test.Errorf("the body %q = kinds %v, phrases %q; want a %v among the kinds and %q among the phrases",
					testCase.body, kindsOf(post.Body.Tree), phrases, testCase.kind, testCase.phrase)
			}
		})
	}
	if !sameMembers(kinds, content.PostKinds()) {
		test.Errorf("the kinds a case holds = %v, want every kind a post may hold, %v", kinds, content.PostKinds())
	}
}

// sameMembers is whether two lists name the same members, whatever their order.
func sameMembers[Member comparable](first, second []Member) bool {
	return len(first) == len(second) && !slices.ContainsFunc(first, func(member Member) bool { return !slices.Contains(second, member) })
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
// language of the site's words, its title, its description as its lead and its page's description, its tags, when
// it came out, in the zone it was written in, and where it is discussed, each a link by its site's name; and its
// series and the posts beside it.
func TestAPostsHeadHoldsItsFrontMatter(test *testing.T) {
	head := postNamed(test, Posts(fixtures.LoadedContent(test)), fixturePostSlug).Head
	published := head.Published
	head.Published = time.Time{}
	description := textSpans("Why a page read twice is read once.")
	page := document.Page{Kind: document.PagePost, Slug: fixturePostSlug, Tab: ""}
	toCome := document.Page{Kind: document.PageNone, Slug: "", Tab: ""}
	want := document.Head{
		Page: page, Number: 1, Language: "en",
		Greeting: "", Title: textSpans("Reading slowly"), Subtitle: nil, Place: "", Tags: []string{"reading", "terminals"},
		Lead: description, Published: time.Time{}, Repository: nil,
		Post: &document.PostPlace{
			Series: &document.Series{Title: textSpans("Reading"), Parts: []document.Part{
				{Number: 1, Title: textSpans("Reading slowly"), State: document.PartReading, Page: page},
				{Number: 2, Title: textSpans("Reading aloud"), State: document.PartToCome, Page: toCome},
			}},
			Neighbours: document.Neighbours{
				Before: nil,
				After:  &document.Neighbour{Part: 2, Title: textSpans("Reading aloud"), Lead: nil, Page: toCome, Published: time.Time{}},
			},
			Discussions: []document.Link{{Target: document.Address{URL: "https://lobsters.example.test/s/slow"}, Spans: textSpans("Lobsters")}},
		},
		Description: description,
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

// samplePicture is a picture a test writes into the sample post's folder: its file's name, and what it holds.
type samplePicture struct {
	name string
	text string
}

// withSamplePost is the fixture content with the draft's file written over by a post's text, and the pictures beside
// it, loaded.
func withSamplePost(test *testing.T, text string, pictures ...samplePicture) *content.Snapshot {
	test.Helper()
	folder := fixtures.Copy(test, fixtures.ContentFolder(test))
	fixtures.WriteIn(test, folder, samplePostFile, text)
	for _, picture := range pictures {
		fixtures.WriteIn(test, folder, path.Join(path.Dir(samplePostFile), picture.name), picture.text)
	}
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
