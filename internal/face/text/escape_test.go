package text

import (
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/document"
)

// Text reads as the text it is, never as markup: not as emphasis, a link, HTML or a block opening a line. A gap
// keeps its brackets, code every backtick and space it holds, and an address its spaces and brackets.
func TestTextNeverReadsAsMarkup(test *testing.T) {
	cases := []struct {
		name string
		span document.Span
		want string
	}{
		{name: "emphasis marks in text", span: document.Text{Text: "a *b* _c_"}, want: `a \*b\* \_c\_`},
		{name: "a backtick and a backslash in text", span: document.Text{Text: "`d` \\e"}, want: "\\`d\\` \\\\e"},
		{name: "code with a backtick", span: document.Code{Text: "a`b"}, want: "``a`b``"},
		{name: "code that opens and closes with a backtick", span: document.Code{Text: "`x`"}, want: "`` `x` ``"},
		{name: "code between spaces", span: document.Code{Text: " a "}, want: "`  a  `"},
		{name: "a link's markup in text", span: document.Text{Text: "[a](https://evil.test)"}, want: `\[a\](https://evil.test)`},
		{name: "HTML in text", span: document.Text{Text: "<b> &copy;"}, want: `\<b> \&copy;`},
		{name: "a heading's mark opening a line", span: document.Text{Text: "# not a heading"}, want: "\n\\# not a heading"},
		{name: "a number opening a line", span: document.Text{Text: "1. not a list"}, want: "\n1\\. not a list"},
		{name: "a quote's mark opening a line", span: document.Text{Text: "> not a quote"}, want: "\n\\> not a quote"},
		{name: "a list's mark opening a line", span: document.Text{Text: "- not a list"}, want: "\n\\- not a list"},
		{name: "a heading's mark after a line break", span: document.Text{Text: "a\n# b"}, want: "a\n\\# b"},
		{name: "spaces opening a line", span: document.Text{Text: "    not code"}, want: "\n\nnot code\n"},
		{name: "a heading's mark closing a title", span: document.Text{Text: "Learn C #"}, want: "Learn C \\#"},
		{name: "an address with a parenthesis", span: document.Link{Target: document.Address{URL: "https://a.test/b(c)"}, Spans: []document.Span{document.Text{Text: "x"}}}, want: "[x](<https://a.test/b(c)>)"},
		{name: "a gap keeps its brackets", span: document.Gap{Text: "[a *gap*]"}, want: `[a \*gap\*]`},
		{name: "an address with a space", span: document.Link{Target: document.Address{URL: "https://a.test/b c"}, Spans: []document.Span{document.Text{Text: "x"}}}, want: "[x](<https://a.test/b c>)"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			composed := document.Document{
				Head:   document.Head{Page: document.Page{Kind: document.PageNone, Slug: ""}, Title: nil, Subtitle: nil, Place: "", Lead: nil, Ways: nil},
				Blocks: []document.Block{document.Section{Title: []document.Span{testCase.span}, Blocks: nil}, document.Paragraph{Spans: []document.Span{testCase.span}}},
			}
			if got := Markdown(composed, fixtureSite(fixtureSnapshot(test))); !strings.Contains(got, testCase.want) {
				test.Errorf("Markdown = %q, want it to hold %q", got, testCase.want)
			}
		})
	}
}

// A list item reads as the text it is too: a fact's value never opens a list or a quote inside its item.
func TestAListItemNeverOpensABlock(test *testing.T) {
	composed := document.Document{
		Head: document.Head{Page: document.Page{Kind: document.PageNone, Slug: ""}, Title: nil, Subtitle: nil, Place: "", Lead: nil, Ways: nil},
		Blocks: []document.Block{document.Table{Rows: []document.Row{
			{Cells: []document.Cell{{Spans: []document.Span{document.Text{Text: "- a"}}}}},
			{Cells: []document.Cell{{Spans: []document.Span{document.Text{Text: "1. b"}}}}},
		}}},
	}
	want := "- \\- a\n- 1\\. b"
	if got := Markdown(composed, fixtureSite(fixtureSnapshot(test))); !strings.Contains(got, want) {
		test.Errorf("Markdown = %q, want it to hold %q", got, want)
	}
}

// A document's lead reads as one quote of the text it is, every paragraph of it inside, never opening a block
// inside the quote; a document with no lead has no quote.
func TestALeadReadsAsAQuoteOfItsText(test *testing.T) {
	cases := []struct {
		name string
		lead []document.Span
		want string
	}{
		{name: "no lead", lead: nil, want: "# tool\n"},
		{name: "a lead", lead: []document.Span{document.Text{Text: "Tidies files."}}, want: "# tool\n\n> Tidies files.\n"},
		{name: "a lead opening with a heading's mark", lead: []document.Span{document.Text{Text: "# not a heading"}}, want: "# tool\n\n> \\# not a heading\n"},
		{name: "a lead of two paragraphs", lead: []document.Span{document.Text{Text: "first\n\nsecond"}}, want: "# tool\n\n> first\n>\n> second\n"},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			composed := document.Document{
				Head: document.Head{
					Page: document.Page{Kind: document.PageNone, Slug: ""}, Title: []document.Span{document.Text{Text: "tool"}},
					Subtitle: nil, Place: "", Lead: testCase.lead, Ways: nil,
				},
				Blocks: nil,
			}
			if got := Markdown(composed, fixtureSite(fixtureSnapshot(test))); got != testCase.want {
				test.Errorf("Markdown = %q, want %q", got, testCase.want)
			}
		})
	}
}
