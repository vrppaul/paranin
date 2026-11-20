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
		Head: fixtures.HeadOf(document.Page{Kind: document.PageProject, Slug: "tool"}, nil),
		Blocks: []document.Block{document.Facts{List: []document.Fact{
			{Label: "install", Value: []document.Span{document.Code{Text: command}}},
		}}},
	}
	if got := Manual(composed, fixtures.Site(fixtures.LoadedContent(test))); !strings.Contains(got, command) {
		test.Errorf("Manual = %q, want the command %q on one line", got, command)
	}
}

// A title and its period never touch: where fewer than two spaces would part them, the period takes a line of its
// own.
func TestATitleAndItsPeriodNeverTouch(test *testing.T) {
	period := document.Period{Start: document.GapMonth{Text: "[2020]"}, End: document.GapMonth{Text: "[2021]"}}
	title := strings.Repeat("a", pageWidth-sectionIndent-textWidth("[2020] – [2021]")-1)
	composed := document.Document{
		Head:   fixtures.HeadOf(document.Page{Kind: document.PageProject, Slug: "tool"}, nil),
		Blocks: []document.Block{document.Entry{Kind: document.EntryRole, Period: period, Title: []document.Span{document.Text{Text: title}}, Subtitle: nil, Place: "", Points: nil}},
	}
	want := strings.Repeat(" ", sectionIndent) + "[2020] – [2021]\n"
	if got := Manual(composed, fixtures.Site(fixtures.LoadedContent(test))); !strings.Contains(got, "\n"+want) {
		test.Errorf("Manual = %q, want the period on a line of its own, %q", got, want)
	}
}

// Text that touches a command stays with it, as it is written: a full stop after a command is no word of its own.
func TestTextTouchingACommandStaysBesideTheCommand(test *testing.T) {
	composed := document.Document{
		Head: fixtures.HeadOf(document.Page{Kind: document.PageProject, Slug: "tool"}, nil),
		Blocks: []document.Block{document.Paragraph{Spans: []document.Span{
			document.Text{Text: "Run "}, document.Code{Text: "tool --all"}, document.Text{Text: "."},
		}}},
	}
	if got := Manual(composed, fixtures.Site(fixtures.LoadedContent(test))); !strings.Contains(got, "Run tool --all.") {
		test.Errorf("Manual = %q, want %q", got, "Run tool --all.")
	}
}
