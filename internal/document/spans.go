package document

import "strings"

// Span is one kind of text inside a block. A highlight and a link hold spans; text, code and a gap are leaves.
//
//sumtype:decl
type Span interface {
	span()
}

// Text is text as it reads.
type Text struct {
	Text string
}

// Code is text a reader types or copies as it is: a command.
type Code struct {
	Text string
}

// Gap is a place the author has still to fill, written as the content writes it.
type Gap struct {
	Text string
}

// Highlight is a result worth seeing first.
type Highlight struct {
	Spans []Span
}

// Link is text that leads to its target.
type Link struct {
	Target Target
	Spans  []Span
}

func (Text) span()      {}
func (Code) span()      {}
func (Gap) span()       {}
func (Highlight) span() {}
func (Link) span()      {}

// PlainText is the text of spans, marks gone: a title where a face takes text alone, as a terminal's tab and a
// browser's title bar do.
func PlainText(spans []Span) string {
	var builder strings.Builder
	for _, span := range spans {
		switch typed := span.(type) {
		case Text:
			builder.WriteString(typed.Text)
		case Code:
			builder.WriteString(typed.Text)
		case Gap:
			builder.WriteString(typed.Text)
		case Highlight:
			builder.WriteString(PlainText(typed.Spans))
		case Link:
			builder.WriteString(PlainText(typed.Spans))
		}
	}
	return builder.String()
}
