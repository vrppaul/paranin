package document

// Phrases are every text of a document a person reads, in its order: the leaves of its spans, its labels, tags
// and places, and the gaps of its periods. A face that shows each phrase shows every fact of the document, which
// is what a test of every face holds it to.
func Phrases(document Document) []string {
	head := document.Head
	phrases := spanPhrases(nil, head.Title)
	phrases = spanPhrases(phrases, head.Subtitle)
	phrases = textPhrase(phrases, head.Place)
	phrases = spanPhrases(phrases, head.Lead)
	phrases = factPhrases(phrases, head.Ways)
	return blockPhrases(phrases, document.Blocks)
}

func blockPhrases(phrases []string, blocks []Block) []string {
	for _, block := range blocks {
		switch typed := block.(type) {
		case Section:
			phrases = blockPhrases(spanPhrases(phrases, typed.Title), typed.Blocks)
		case Paragraph:
			phrases = spanPhrases(phrases, typed.Spans)
		case Table:
			phrases = rowPhrases(phrases, typed.Rows)
		case CodeBlock:
			phrases = textPhrase(phrases, typed.Text)
		case Facts:
			phrases = factPhrases(phrases, typed.List)
		case Entry:
			phrases = entryPhrases(phrases, typed)
		case Card:
			phrases = cardPhrases(phrases, typed)
		}
	}
	return phrases
}

func rowPhrases(phrases []string, rows []Row) []string {
	for _, row := range rows {
		for _, cell := range row.Cells {
			phrases = spanPhrases(phrases, cell.Spans)
		}
	}
	return phrases
}

func factPhrases(phrases []string, facts []Fact) []string {
	for _, fact := range facts {
		phrases = spanPhrases(textPhrase(phrases, fact.Label), fact.Value)
	}
	return phrases
}

func entryPhrases(phrases []string, entry Entry) []string {
	phrases = monthPhrase(monthPhrase(phrases, entry.Period.Start), entry.Period.End)
	phrases = textPhrase(spanPhrases(spanPhrases(phrases, entry.Title), entry.Subtitle), entry.Place)
	for _, point := range entry.Points {
		phrases = spanPhrases(phrases, point.Spans)
	}
	return phrases
}

func cardPhrases(phrases []string, card Card) []string {
	phrases = textPhrase(spanPhrases(spanPhrases(phrases, card.Title), card.Subtitle), card.Tag)
	return factPhrases(spanPhrases(phrases, card.Summary), card.Facts)
}

// monthPhrase is a gap's text and the word for what still goes on; a known month is each face's to write.
func monthPhrase(phrases []string, month Month) []string {
	switch typed := month.(type) {
	case GapMonth:
		return textPhrase(phrases, typed.Text)
	case CurrentMonth:
		return textPhrase(phrases, typed.Text)
	case KnownMonth:
	}
	return phrases
}

func spanPhrases(phrases []string, spans []Span) []string {
	for _, span := range spans {
		switch typed := span.(type) {
		case Text:
			phrases = textPhrase(phrases, typed.Text)
		case Code:
			phrases = textPhrase(phrases, typed.Text)
		case Gap:
			phrases = textPhrase(phrases, typed.Text)
		case Highlight:
			phrases = spanPhrases(phrases, typed.Spans)
		case Link:
			phrases = spanPhrases(phrases, typed.Spans)
		}
	}
	return phrases
}

// textPhrase adds a phrase; an empty text is none.
func textPhrase(phrases []string, text string) []string {
	if text == "" {
		return phrases
	}
	return append(phrases, text)
}
