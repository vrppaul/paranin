package document

import "strings"

// Phrases are every text of a document a face must show, in its order: the leaves of its spans, its labels, tags and
// places, the word for what still goes on, and the words an older version of a post says a newer stands by; not the
// head's greeting, which a face may leave out. A face that shows each phrase of the document it draws, whole or as
// WrittenBlocks leaves it, shows every fact of it, which is what a test of every face holds it to.
func Phrases(document Document) []string {
	head := document.Head
	phrases := spanPhrases(nil, head.Title)
	phrases = spanPhrases(phrases, head.Subtitle)
	phrases = textPhrase(phrases, head.Place)
	for _, tag := range head.Tags() {
		phrases = textPhrase(phrases, tag)
	}
	phrases = spanPhrases(phrases, head.Lead)
	phrases = projectPhrases(phrases, head.Project)
	if older, isOlder := head.Version.(VersionOlder); isOlder {
		phrases = textPhrase(textPhrase(phrases, strings.TrimSpace(older.NewerWords.Before)), strings.TrimSpace(older.NewerWords.After))
	}
	return blockPhrases(phrases, document.Blocks)
}

func blockPhrases(phrases []string, blocks []Block) []string {
	for _, block := range blocks {
		phrases = oneBlockPhrases(phrases, block)
	}
	return phrases
}

func oneBlockPhrases(phrases []string, block Block) []string {
	switch typed := block.(type) {
	case Section:
		return blockPhrases(spanPhrases(phrases, typed.Title), typed.Blocks)
	case Paragraph:
		return spanPhrases(phrases, typed.Spans)
	case Table:
		return rowPhrases(phrases, append([]Row{typed.Header}, typed.Rows...))
	case CodeBlock:
		return textPhrase(phrases, typed.Text())
	case List:
		for _, item := range typed.Items {
			phrases = blockPhrases(phrases, item.Blocks)
		}
		return phrases
	case Quote:
		return blockPhrases(phrases, typed.Blocks)
	case Pause:
		// a pause says nothing: each face draws its stars
	case Figure:
		// a figure's address is each face's to write, or to show the picture at
		return textPhrase(textPhrase(phrases, typed.Alt), typed.Caption)
	case Widget:
		// a widget's name and its link are each face's to write, as a figure's address is
		return textPhrase(textPhrase(phrases, typed.Alt), typed.Caption)
	case Facts:
		return factPhrases(phrases, typed.List)
	case ExperienceHead:
		phrases = spanPhrases(phrases, typed.Line)
		for _, technology := range typed.Technologies {
			phrases = spanPhrases(textPhrase(phrases, technology.Name), technology.Used)
		}
		return phrases
	case Entry:
		return entryPhrases(phrases, typed)
	case Points:
		return pointPhrases(phrases, typed.List)
	case Side:
		return blockPhrases(phrases, typed.Blocks)
	case Card:
		return cardPhrases(phrases, typed)
	case Contents:
		return listingPhrases(phrases, typed.Listings)
	case Checks:
		return checkPhrases(spanPhrases(phrases, typed.Title), typed.List)
	case Skills:
		return skillPhrases(phrases, typed.List)
	case Posts:
		return listedPhrases(spanPhrases(spanPhrases(phrases, typed.Line), typed.Count), typed.List)
	}
	return phrases
}

// listedPhrases are what each post a list names says of it: the word saying a part comes, its title, its lead, where
// it stands, its series and its tags; when a post came out is each face's to write.
func listedPhrases(phrases []string, posts []ListedPost) []string {
	for _, post := range posts {
		phrases = spanPhrases(readinessPhrases(phrases, post.Readiness), post.Title)
		phrases = spanPhrases(spanPhrases(spanPhrases(phrases, post.Lead), post.Part), post.Series)
		for _, tag := range post.Tags {
			phrases = textPhrase(phrases, tag)
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

// projectPhrases are what a project's head says of it: where its code lives; its install's command, and each other way
// with its note; what tells about it, each under its label, each part; then its end (endPhrases). Not the site's
// words over the repository and the install, which a face may leave out where the place they stand in says it.
func projectPhrases(phrases []string, project *Project) []string {
	if project == nil {
		return phrases
	}
	if project.Repository != nil {
		phrases = textPhrase(phrases, project.Repository.Address.URL)
	}
	phrases = textPhrase(phrases, project.Install)
	for _, alternative := range project.Alternatives {
		phrases = textPhrase(spanPhrases(phrases, alternative.Note), alternative.Command)
	}
	for _, told := range project.About() {
		phrases = spanPhrases(textPhrase(phrases, told.Label), told.Parts)
	}
	return endPhrases(phrases, *project)
}

// endPhrases are what a project's end says: where it has others, their heading, the way to them all and each one's name
// and tagline, not what it is made of, which its own page says; and the author's mail. Not the site's word for contact
// over the mail, which a face leaves out where the mail's place says it.
func endPhrases(phrases []string, project Project) []string {
	if len(project.Others) > 0 {
		phrases = spanPhrases(textPhrase(phrases, project.Labels.Others), []Span{project.AllProjects})
		phrases = listingPhrases(phrases, project.OtherListings())
	}
	return textPhrase(phrases, project.Mail.Address)
}

func factPhrases(phrases []string, facts []Fact) []string {
	for _, fact := range facts {
		phrases = spanPhrases(textPhrase(phrases, fact.Label), fact.Value)
	}
	return phrases
}

func entryPhrases(phrases []string, entry Entry) []string {
	// a period's start is a known month, each face's to write
	phrases = monthPhrase(phrases, entry.Period.End)
	phrases = spanPhrases(spanPhrases(phrases, entry.Title), entry.Subtitle)
	for _, technology := range entry.Technologies {
		phrases = textPhrase(phrases, technology)
	}
	phrases = pointPhrases(phrases, entry.Points)
	for _, paragraph := range entry.Description {
		phrases = spanPhrases(phrases, paragraph.Spans)
	}
	return phrases
}

// pointPhrases are what points say, an entry's or a project's section's.
func pointPhrases(phrases []string, points []Point) []string {
	for _, point := range points {
		phrases = spanPhrases(phrases, point.Spans)
	}
	return phrases
}

func cardPhrases(phrases []string, card Card) []string {
	phrases = textPhrase(spanPhrases(spanPhrases(phrases, card.Title), card.Subtitle), card.Tag)
	return factPhrases(spanPhrases(phrases, card.Summary), card.Facts)
}

func listingPhrases(phrases []string, listings []Listing) []string {
	for _, listing := range listings {
		phrases = spanPhrases(spanPhrases(phrases, listing.Title), listing.Note)
	}
	return phrases
}

// checkPhrases are what each check says and its advice; a level is each face's to write.
func checkPhrases(phrases []string, checks []Check) []string {
	for _, check := range checks {
		phrases = spanPhrases(spanPhrases(phrases, check.Spans), check.Advice)
	}
	return phrases
}

// skillPhrases are each skill's name, how long, and where it was proven; whether the author is still learning it is
// each face's to show.
func skillPhrases(phrases []string, skills []Skill) []string {
	for _, skill := range skills {
		phrases = spanPhrases(spanPhrases(spanPhrases(phrases, skill.Name), skill.Duration), skill.Sources)
	}
	return phrases
}

// readinessPhrases are the word saying a part comes; a post's day is each face's to write.
func readinessPhrases(phrases []string, readiness Readiness) []string {
	switch typed := readiness.(type) {
	case ComingPart:
		return spanPhrases(phrases, typed.Coming)
	case PostOut:
	}
	return phrases
}

// monthPhrase is the word for what still goes on; a known month is each face's to write.
func monthPhrase(phrases []string, month Month) []string {
	switch typed := month.(type) {
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
		case Emphasis:
			phrases = spanPhrases(phrases, typed.Spans)
		case Strong:
			phrases = spanPhrases(phrases, typed.Spans)
		case Link:
			phrases = spanPhrases(phrases, typed.Spans)
		case Note:
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
