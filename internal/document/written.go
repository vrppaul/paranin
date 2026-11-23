package document

import (
	"slices"
	"strings"
)

// WrittenBlocks are blocks as a reader of the finished work sees them: a fact or a point that holds nothing but gaps
// still to fill is left out until its author fills it, and a list of facts left with none goes with it; a gap among
// written words stays. A face that keeps the author's notes, as the text faces do, draws the blocks whole. The blocks
// it is handed keep everything they had.
func WrittenBlocks(blocks []Block) []Block {
	kept := make([]Block, 0, len(blocks))
	for _, block := range blocks {
		if shown, isShown := writtenBlock(block); isShown {
			kept = append(kept, shown)
		}
	}
	return kept
}

// writtenBlock is a block with its facts and points that are only gaps left out; not shown at all where it is a list
// of facts left with none.
func writtenBlock(block Block) (Block, bool) {
	switch typed := block.(type) {
	case Section:
		typed.Blocks = WrittenBlocks(typed.Blocks)
		return typed, true
	case Facts:
		typed.List = writtenFacts(typed.List)
		return typed, len(typed.List) > 0
	case Entry:
		typed.Points = slices.DeleteFunc(slices.Clone(typed.Points), func(point Point) bool { return !written(point.Spans) })
		return typed, true
	case Card:
		typed.Facts = writtenFacts(typed.Facts)
		return typed, true
	case Paragraph, Table, CodeBlock, List, Quote, Pause, Figure, Contents, Checks, Skills, Posts:
	}
	return block, true
}

// writtenFacts are the facts whose values hold a written word, in their order.
func writtenFacts(facts []Fact) []Fact {
	return slices.DeleteFunc(slices.Clone(facts), func(fact Fact) bool { return !written(fact.Value) })
}

// written is whether spans hold a word the author wrote, not gaps alone.
//
//wrapper:the one test every part of a document makes of its words before it is shown
func written(spans []Span) bool {
	return slices.ContainsFunc(spans, spanWritten)
}

// spanWritten is whether a span holds a word the author wrote: text that is more than spaces, code, or a mark round
// such a span; a gap is still to write.
func spanWritten(span Span) bool {
	switch typed := span.(type) {
	case Text:
		return strings.TrimSpace(typed.Text) != ""
	case Code:
		return true
	case Highlight:
		return written(typed.Spans)
	case Emphasis:
		return written(typed.Spans)
	case Strong:
		return written(typed.Spans)
	case Link:
		return written(typed.Spans)
	case Note:
		return written(typed.Spans)
	case Gap:
	}
	return false
}
