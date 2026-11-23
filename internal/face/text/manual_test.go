package text

import (
	"strings"
	"testing"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/fixtures"
)

// A command too long for a line stays on one, so a reader who copies it pastes the whole command; prose around
// it wraps as prose does.
func TestACommandNeverBreaksAcrossLines(test *testing.T) {
	command := "toolkit add --scope user example -- runner --index example-index=https://packages.example.test example-tool"
	composed := document.Document{
		Head: fixtures.HeadOf(document.Page{Kind: document.PageProject, Slug: "tool", Tab: ""}, nil),
		Blocks: []document.Block{document.Facts{List: []document.Fact{
			{Label: "install", Value: []document.Span{document.Code{Text: command}}},
		}}},
	}
	if got := Manual(composed, fixtures.Site(fixtures.LoadedContent(test))); !strings.Contains(got, command) {
		test.Errorf("Manual = %q, want the command %q on one line", got, command)
	}
}

// DESCRIPTION opens with what it holds, the blocks that are no section where the head has no lead, as a tab shown
// alone has none; and a page with nothing to describe has no DESCRIPTION.
func TestTheDescriptionHoldsWhatDescribesThePage(test *testing.T) {
	run := document.Paragraph{Spans: []document.Span{document.Text{Text: "Run it."}}}
	cases := []struct {
		name   string
		blocks []document.Block
		holds  string
		lacks  string
	}{
		{name: "a block with no lead", blocks: []document.Block{run}, holds: "DESCRIPTION\n" + strings.Repeat(" ", sectionIndent) + "Run it.", lacks: "DESCRIPTION\n\n"},
		{
			name:   "sections alone",
			blocks: []document.Block{document.Section{Title: []document.Span{document.Text{Text: "Usage"}}, Blocks: []document.Block{run}}},
			holds:  "USAGE\n", lacks: "DESCRIPTION",
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			composed := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PageProject, Slug: "tool", Tab: ""}, nil), Blocks: testCase.blocks}
			got := Manual(composed, fixtures.Site(fixtures.LoadedContent(test)))
			if !strings.Contains(got, testCase.holds) || strings.Contains(got, testCase.lacks) {
				test.Errorf("Manual = %q, want it to hold %q and not %q", got, testCase.holds, testCase.lacks)
			}
		})
	}
}

// A title and its period never touch: where fewer than two spaces would part them, the period takes a line of its
// own.
func TestATitleAndItsPeriodNeverTouch(test *testing.T) {
	period := document.Period{Start: document.GapMonth{Text: "[2020]"}, End: document.GapMonth{Text: "[2021]"}}
	title := strings.Repeat("a", pageWidth-sectionIndent-textWidth("[2020] – [2021]")-1)
	composed := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageProject, Slug: "tool", Tab: ""}, nil),
		Blocks: []document.Block{document.Entry{Kind: document.EntryRole, Period: period, Title: []document.Span{document.Text{Text: title}}, Subtitle: nil, Place: "", Points: nil}},
	}
	want := strings.Repeat(" ", sectionIndent) + "[2020] – [2021]\n"
	if got := Manual(composed, fixtures.Site(fixtures.LoadedContent(test))); !strings.Contains(got, "\n"+want) {
		test.Errorf("Manual = %q, want the period on a line of its own, %q", got, want)
	}
}

// A tab in code reaches no terminal: it stands as the spaces to its next stop, four columns apart from the code's
// own start, as on the page, so the code lines up as its author's editor showed it, whatever the page's indent.
func TestATabInCodeStandsAsTheSpacesToItsStop(test *testing.T) {
	composed := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PagePost, Slug: "post", Tab: ""}, nil),
		Blocks: []document.Block{document.CodeBlock{Language: "go", Runs: []document.CodeRun{{Kind: document.CodePlain, Text: "if x {\n\treturn\n}\nab\tc"}}}},
	}
	got := Manual(composed, fixtures.Site(fixtures.LoadedContent(test)))
	indent := strings.Repeat(" ", sectionIndent)
	for _, want := range []string{indent + "    return\n", indent + "ab  c\n"} {
		if !strings.Contains(got, want) || strings.Contains(got, "\t") {
			test.Errorf("Manual = %q, want %q and no tab", got, want)
		}
	}
}

// A table under a header too wide for the page in columns stands a row a bullet, its header's first, each cell's
// words whole, so no column runs past the page's edge.
func TestATableTooWideForColumnsStandsARowABullet(test *testing.T) {
	wide := strings.Repeat("word ", pageWidth/len("word "))
	table := document.Table{
		Header: document.Row{Cells: []document.Cell{cellOf("Name"), cellOf("Note")}},
		Rows:   []document.Row{{Cells: []document.Cell{cellOf("Slow"), cellOf(wide)}}},
	}
	composed := document.Document{Head: fixtures.HeadOf(document.Page{Kind: document.PagePost, Slug: "post", Tab: ""}, nil), Blocks: []document.Block{table}}
	got := Manual(composed, fixtures.Site(fixtures.LoadedContent(test)))
	indent := strings.Repeat(" ", sectionIndent)
	for _, want := range []string{indent + "- Name — Note\n", indent + "- Slow — word"} {
		if !strings.Contains(got, want) {
			test.Errorf("Manual = %q, want %q", got, want)
		}
	}
}

// Text that touches a command stays with it, as it is written: a full stop after a command is no word of its own.
func TestTextTouchingACommandStaysBesideTheCommand(test *testing.T) {
	composed := document.Document{
		Head: fixtures.HeadOf(document.Page{Kind: document.PageProject, Slug: "tool", Tab: ""}, nil),
		Blocks: []document.Block{document.Paragraph{Spans: []document.Span{
			document.Text{Text: "Run "}, document.Code{Text: "tool --all"}, document.Text{Text: "."},
		}}},
	}
	if got := Manual(composed, fixtures.Site(fixtures.LoadedContent(test))); !strings.Contains(got, "Run tool --all.") {
		test.Errorf("Manual = %q, want %q", got, "Run tool --all.")
	}
}
