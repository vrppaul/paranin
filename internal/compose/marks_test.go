package compose

import (
	"reflect"
	"testing"

	"github.com/vrppaul/paranin/internal/document"
)

// The profile marks a result with ==…==, a gap still to fill with […] and code with `…`; each becomes a span, a gap or
// code may sit inside a result, as the profile writes "==[N] requests/day==", and inside code no mark is read.
func TestTheProfilesMarksBecomeSpans(test *testing.T) {
	cases := []struct {
		name string
		text string
		want []document.Span
	}{
		{name: "plain text", text: "Wrote it", want: []document.Span{document.Text{Text: "Wrote it"}}},
		{
			name: "a result", text: "Cut it by ==40%== at once",
			want: []document.Span{document.Text{Text: "Cut it by "}, document.Highlight{Spans: []document.Span{document.Text{Text: "40%"}}}, document.Text{Text: " at once"}},
		},
		{name: "a gap", text: "Owned [service]", want: []document.Span{document.Text{Text: "Owned "}, document.Gap{Text: "[service]"}}},
		{
			name: "a gap inside a result", text: "==[N] requests==",
			want: []document.Span{document.Highlight{Spans: []document.Span{document.Gap{Text: "[N]"}, document.Text{Text: " requests"}}}},
		},
		{name: "a bracket that closes nothing", text: "a [b", want: []document.Span{document.Text{Text: "a [b"}}},
		{name: "brackets inside brackets", text: "[a [b] c]", want: []document.Span{document.Text{Text: "[a "}, document.Gap{Text: "[b]"}, document.Text{Text: " c]"}}},
		{name: "empty brackets", text: "a [] b", want: []document.Span{document.Text{Text: "a [] b"}}},
		{name: "a result never closed", text: "==open", want: []document.Span{document.Text{Text: "==open"}}},
		{name: "code", text: "Run `go test` now", want: []document.Span{document.Text{Text: "Run "}, document.Code{Text: "go test"}, document.Text{Text: " now"}}},
		{name: "no mark inside code", text: "`==a== [b]`", want: []document.Span{document.Code{Text: "==a== [b]"}}},
		{name: "a backtick with no partner", text: "a ` b", want: []document.Span{document.Text{Text: "a ` b"}}},
		{
			name: "code inside a result", text: "==`go` fast==",
			want: []document.Span{document.Highlight{Spans: []document.Span{document.Code{Text: "go"}, document.Text{Text: " fast"}}}},
		},
		{
			name: "a result's mark inside code inside a result", text: "==a `b==c` d==",
			want: []document.Span{document.Highlight{Spans: []document.Span{document.Text{Text: "a "}, document.Code{Text: "b==c"}, document.Text{Text: " d"}}}},
		},
	}
	for _, testCase := range cases {
		test.Run(testCase.name, func(test *testing.T) {
			if got := markedSpans(testCase.text); !reflect.DeepEqual(got, testCase.want) {
				test.Errorf("markedSpans(%q) = %+v, want %+v", testCase.text, got, testCase.want)
			}
		})
	}
}
