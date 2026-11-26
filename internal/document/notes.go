package document

// NotesOf are the notes blocks cite, each once, in the order of their first citation, which their numbers count: what
// a face that gathers the notes at the end lists there.
func NotesOf(blocks []Block) []Note {
	gathering := noteGathering{notes: []Note{}, cited: map[int]bool{}}
	gathering.blocks(blocks)
	return gathering.notes
}

// noteGathering is the notes met so far, and the numbers among them, which a note cited again is known by.
type noteGathering struct {
	notes []Note
	cited map[int]bool
}

func (gathering *noteGathering) blocks(blocks []Block) {
	for _, block := range blocks {
		gathering.block(block)
	}
}

func (gathering *noteGathering) block(block Block) {
	switch typed := block.(type) {
	case Section:
		gathering.spans(typed.Title)
		gathering.blocks(typed.Blocks)
	case Paragraph:
		gathering.spans(typed.Spans)
	case Table:
		for _, row := range append([]Row{typed.Header}, typed.Rows...) {
			for _, cell := range row.Cells {
				gathering.spans(cell.Spans)
			}
		}
	case List:
		for _, item := range typed.Items {
			gathering.blocks(item.Blocks)
		}
	case Quote:
		gathering.blocks(typed.Blocks)
	case CodeBlock, Pause, Figure, Widget, Facts, ExperienceHead, Entry, Points, Side, Card, Contents, Checks, Skills, Posts:
		// a post's body alone cites notes, and it is made of none of these but code, pauses, figures and widgets, which
		// hold no spans
	}
}

func (gathering *noteGathering) spans(spans []Span) {
	for _, span := range spans {
		switch typed := span.(type) {
		case Highlight:
			gathering.spans(typed.Spans)
		case Emphasis:
			gathering.spans(typed.Spans)
		case Strong:
			gathering.spans(typed.Spans)
		case Link:
			gathering.spans(typed.Spans)
		case Note:
			if !gathering.cited[typed.Number] {
				gathering.cited[typed.Number] = true
				gathering.notes = append(gathering.notes, typed)
			}
		case Text, Code, Gap:
		}
	}
}
