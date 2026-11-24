// Package compose turns content into the documents every face draws (docs/ARCHITECTURE.md, ADR-0003). It
// decides once which facts a page shows and in what order, so no face picks them again; it reads the profile's
// marks into spans and carries its months across. Beside each page a person reads it makes the page's record, the
// facts a search engine reads as data, typed and plain. Beside the documents it hands a face the words of the frame
// it lays round them, the page's theme button among them (frame.go), since a face reads no content. It writes
// no format of any face: how a span or a period reads, and the vocabulary a record is written in, are each face's
// own.
package compose

import (
	"slices"
	"strings"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// What separates the places a skill was proven, written in one text
const listSeparator = ", "

// Composed is a page a person reads, as compose makes it: its document, its tabs past the opening, each a document
// alone, and its record, made together so none is ever without the others.
type Composed struct {
	Document document.Document
	Tabs     []document.Document
	Record   Record
}

// pagesOf are the pages a snapshot makes for a person to read, each with its record: the résumé, naming the blog's feed
// as the blog reads it, so a feed reader handed the site's address finds it; then the page of each project that has
// one. The site's index lists them.
func pagesOf(snapshot *content.Snapshot, read blog) []Composed {
	profile := snapshot.Profile
	pages := []Composed{composedPage(Resume(snapshot, read.feedLinks(noSeries)), personRecord(profile))}
	author := authorOf(profile.Person)
	for _, project := range projectsWithPages(profile.Projects) {
		pages = append(pages, composedPage(projectDocument(project, profile.Person, snapshot.Own.Words), workRecord(project, author)))
	}
	return pages
}

// composedPage is a page of a document and its record, with the document's tabs.
func composedPage(composed document.Document, record Record) Composed {
	return Composed{Document: composed, Tabs: tabDocuments(composed), Record: record}
}

// tabDocuments are a document's tabs past its opening (document.TabsOf), in their order, each a document of the tab
// alone, for a face that draws a document whole.
func tabDocuments(composed document.Document) []document.Document {
	rest := document.TabsOf(composed).Rest
	documents := make([]document.Document, 0, len(rest))
	for _, tab := range rest {
		documents = append(documents, document.Document{Head: tabHead(composed.Head, tab), Blocks: tab.Blocks})
	}
	return documents
}

// tabHead is the head of a tab shown alone: its page's, naming the tab, under the tab's title, with the language,
// description and feeds of its page's head.
func tabHead(head document.Head, tab document.Tab) document.Head {
	return document.Head{
		Page:        document.Page{Kind: head.Page.Kind, Slug: head.Page.Slug, Tab: tab.Slug, Revision: "", Translation: head.Page.Translation},
		Number:      0,
		Language:    head.Language,
		Greeting:    "",
		Title:       tab.Title,
		Subtitle:    nil,
		Place:       "",
		Filters:     nil,
		MadeOf:      nil,
		Lead:        nil,
		Version:     nil,
		Repository:  nil,
		Post:        nil,
		Description: head.Description,
		Feeds:       head.Feeds,
		Languages:   nil,
	}
}

// projectsWithPages are the projects that have a page of their own, in their order.
func projectsWithPages(projects []content.Project) []content.Project {
	return slices.DeleteFunc(slices.Clone(projects), func(project content.Project) bool { return !project.Page })
}

// Resume is the résumé as one document: its head, naming the feeds a reader handed the site's address finds, then five
// sections, each under the site's word for it, as a reader asks of an author: about them, their experience, their
// projects, their skills, and how to reach them.
func Resume(snapshot *content.Snapshot, feeds []document.Link) document.Document {
	profile, words := snapshot.Profile, snapshot.Own.Words
	return document.Document{
		Head: resumeHead(profile, words, feeds),
		Blocks: []document.Block{
			aboutSection(profile, words),
			experienceSection(profile.Experience, words),
			projectsSection(profile.Projects, words),
			skillsSection(profile.SkillGroups, words, languagesSection(profile.Languages, words), educationSection(profile.Education, words)),
			section(words.Text(content.WordContact), document.Facts{Kind: document.FactsWays, List: ways(profile.Person, profile.Site, words)}),
		},
	}
}

// aboutSection is the section about the author: the summary, then the facts.
func aboutSection(profile content.Profile, words content.Words) document.Section {
	return document.Section{
		Title: textSpans(words.Text(content.WordAbout)),
		Name:  "",
		Blocks: []document.Block{
			document.Paragraph{Spans: markedSpans(profile.Summary)},
			document.Facts{Kind: document.FactsDetails, List: factList(profile.Facts)},
		},
	}
}

// indexOf is the site's index, the list of its pages an agent reads first: whose site it is and its headline, then
// the résumé, the page of each project that has one, and the blog's lists and the posts they name, each under the
// site's word for it, the projects' word only where a project has a page. The résumé and a project's page are
// listed by the title and subtitle their own heads open with; the blog's, as blogListings has them.
func indexOf(snapshot *content.Snapshot, read blog) document.Document {
	words := snapshot.Own.Words
	resume := resumeHead(snapshot.Profile, words, nil)
	blocks := []document.Block{section(words.Text(content.WordResume), document.Contents{Listings: []document.Listing{listing(resume)}})}
	projects := []document.Listing{}
	for _, project := range projectsWithPages(snapshot.Profile.Projects) {
		projects = append(projects, listing(projectHead(project, words)))
	}
	if len(projects) > 0 {
		blocks = append(blocks, section(words.Text(content.WordProjects), document.Contents{Listings: projects}))
	}
	blocks = append(blocks, section(words.Text(content.WordBlog), document.Contents{Listings: blogListings(read, words)}))
	return document.Document{
		Head: document.Head{
			Page:        document.PageOf(document.PageIndex),
			Number:      0,
			Language:    words.Language(),
			Greeting:    "",
			Title:       resume.Title,
			Subtitle:    nil,
			Place:       "",
			Filters:     nil,
			MadeOf:      nil,
			Lead:        resume.Lead,
			Version:     nil,
			Repository:  nil,
			Post:        nil,
			Description: nil,
			Feeds:       nil,
			Languages:   nil,
		},
		Blocks: blocks,
	}
}

// listing is a page as a list of contents names it: by its head's title and subtitle.
func listing(head document.Head) document.Listing {
	return document.Listing{Page: head.Page, Title: head.Title, Note: head.Subtitle}
}

// resumeHead is the résumé's head, its headline its description: it greets the reader before the author's name,
// names its opening, in the site's words, and the feeds handed it.
func resumeHead(profile content.Profile, words content.Words, feeds []document.Link) document.Head {
	person := profile.Person
	headline := markedSpans(person.Headline)
	return document.Head{
		Page:        document.PageOf(document.PageResume),
		Number:      0,
		Language:    words.Language(),
		Greeting:    words.Text(content.WordGreeting),
		Title:       markedSpans(person.Name),
		Subtitle:    markedSpans(person.Title),
		Place:       person.Location,
		Filters:     nil,
		MadeOf:      nil,
		Lead:        headline,
		Version:     nil,
		Repository:  nil,
		Post:        nil,
		Description: headline,
		Feeds:       feeds,
		Languages:   nil,
	}
}

// ways are the ways to reach the author, in the order a reader tries them: mail, the public profiles, ssh.
func ways(person content.Person, site content.Site, words content.Words) []document.Fact {
	mail := document.Mail{Address: person.Email}
	list := []document.Fact{{Label: words.Text(content.WordMail), Value: linkSpans(mail, person.Email)}}
	for _, link := range person.Links {
		list = append(list, document.Fact{Label: link.Label, Value: linkSpans(document.Address{URL: link.URL}, link.Text)})
	}
	return append(list, document.Fact{Label: words.Text(content.WordSSH), Value: []document.Span{document.Code{Text: site.SSH}}})
}

func experienceSection(roles []content.Role, words content.Words) document.Section {
	entries := make([]document.Block, 0, len(roles))
	for _, role := range roles {
		entries = append(entries, document.Entry{
			Kind:     document.EntryRole,
			Period:   periodOf(role.Start, role.End, words),
			Title:    markedSpans(role.Company),
			Subtitle: markedSpans(role.Title),
			Place:    role.Location,
			Points:   points(role.Bullets),
		})
	}
	return document.Section{Title: textSpans(words.Text(content.WordExperience)), Name: "", Blocks: entries}
}

func educationSection(schools []content.School, words content.Words) document.Section {
	entries := make([]document.Block, 0, len(schools))
	for _, school := range schools {
		entries = append(entries, document.Entry{
			Kind:     document.EntrySchool,
			Period:   periodOf(school.Start, school.End, words),
			Title:    markedSpans(school.Name),
			Subtitle: markedSpans(school.Field),
			Place:    "",
			Points:   nil,
		})
	}
	return document.Section{Title: textSpans(words.Text(content.WordEducation)), Name: "", Blocks: entries}
}

func projectsSection(projects []content.Project, words content.Words) document.Section {
	cards := make([]document.Block, 0, len(projects))
	for _, project := range projects {
		cards = append(cards, document.Card{
			Title:    markedSpans(project.Name),
			Subtitle: markedSpans(project.Tagline),
			Tag:      project.Language,
			Summary:  markedSpans(project.Description),
			Facts:    append(projectRepository(project, words).Facts(), installFacts(project, words)...),
			Page:     projectPage(project),
		})
	}
	return document.Section{Title: textSpans(words.Text(content.WordProjects)), Name: "", Blocks: cards}
}

// projectDocument is a project's own page: its head; its details, its author first, a link back to the résumé, and
// how it is installed last; its recording, where it has one; then its sections.
func projectDocument(project content.Project, author content.Person, words content.Words) document.Document {
	resume := document.PageOf(document.PageResume)
	details := []document.Fact{{Label: words.Text(content.WordAuthor), Value: linkSpans(resume, author.Name)}}
	details = append(details, factList(project.Facts)...)
	blocks := []document.Block{document.Facts{Kind: document.FactsDetails, List: append(details, installFacts(project, words)...)}}
	if project.Recording != "" {
		blocks = append(blocks, codeBlock(project.Recording, noLanguage))
	}
	for _, section := range project.Sections {
		paragraphs := make([]document.Block, 0, len(section.Paragraphs))
		for _, paragraph := range section.Paragraphs {
			paragraphs = append(paragraphs, document.Paragraph{Spans: markedSpans(paragraph)})
		}
		blocks = append(blocks, document.Section{Title: markedSpans(section.Title), Name: "", Blocks: paragraphs})
	}
	return document.Document{Head: projectHead(project, words), Blocks: blocks}
}

// projectHead is the head of a project's page, in the language of the site's words: its name, what it is in a
// line, what it is made of, what it does, its description, and where its code lives.
func projectHead(project content.Project, words content.Words) document.Head {
	description := markedSpans(project.Description)
	return document.Head{
		Page:        projectPage(project),
		Number:      0,
		Language:    words.Language(),
		Greeting:    "",
		Title:       markedSpans(project.Name),
		Subtitle:    markedSpans(project.Tagline),
		Place:       "",
		Filters:     nil,
		MadeOf:      projectMadeOf(project),
		Lead:        description,
		Version:     nil,
		Repository:  projectRepository(project, words),
		Post:        nil,
		Description: description,
		Feeds:       nil,
		Languages:   nil,
	}
}

// projectMadeOf is what a project is made of: the language it is written in, where the profile names one, then what
// it is built on.
func projectMadeOf(project content.Project) []string {
	if project.Language == "" {
		return slices.Clone(project.Stack)
	}
	return append([]string{project.Language}, project.Stack...)
}

// projectPage is a project's own page, or no page where it has none.
func projectPage(project content.Project) document.Page {
	if !project.Page {
		return document.PageOf(document.PageNone)
	}
	return document.Page{Kind: document.PageProject, Slug: project.Slug, Tab: "", Revision: "", Translation: ""}
}

// projectRepository is where a project's code lives, under the site's word for it; none where the project names no
// place.
func projectRepository(project content.Project, words content.Words) *document.Repository {
	if project.Repo == "" {
		return nil
	}
	return &document.Repository{Label: words.Text(content.WordRepository), Address: document.Address{URL: project.Repo}}
}

// installFacts are how a project is installed, under the site's word for it, where the project says.
func installFacts(project content.Project, words content.Words) []document.Fact {
	if project.Install == "" {
		return nil
	}
	return []document.Fact{{Label: words.Text(content.WordInstall), Value: []document.Span{document.Code{Text: project.Install}}}}
}

// skillsSection is what the author can do: a section of each group of skills, each a table of a skill, how long, and
// where it was proven; then the sections handed after them.
func skillsSection(groups []content.SkillGroup, words content.Words, after ...document.Block) document.Section {
	sections := make([]document.Block, 0, len(groups)+len(after))
	for _, group := range groups {
		rows := make([]document.Row, 0, len(group.Skills))
		for _, skill := range group.Skills {
			listed := listedSkill(skill, words)
			rows = append(rows, document.Row{Cells: []document.Cell{
				{Kind: document.CellText, Spans: listed.Name},
				{Kind: document.CellDuration, Spans: listed.Duration},
				{Kind: document.CellText, Spans: listed.Sources},
			}})
		}
		table := document.Table{Header: document.Row{Cells: nil}, Rows: rows}
		sections = append(sections, document.Section{Title: markedSpans(group.Title), Name: "", Blocks: []document.Block{table}})
	}
	return document.Section{Title: textSpans(words.Text(content.WordSkills)), Name: "", Blocks: append(sections, after...)}
}

// languagesSection is the languages the author speaks, each with how well.
func languagesSection(languages []content.Language, words content.Words) document.Section {
	return section(words.Text(content.WordLanguages), document.Facts{Kind: document.FactsDetails, List: languageList(languages)})
}

// listedSkill is a skill as the documents hold it, in the résumé's table and in the list of skills alike: its name,
// how long in text and in months, the places it was proven in one text, and whether the author is still learning it.
func listedSkill(skill content.Skill, words content.Words) document.Skill {
	return document.Skill{
		Name:     markedSpans(skill.Name),
		Duration: durationSpans(skill.Duration, words),
		Months:   skill.Duration.Months(),
		Sources:  markedSpans(strings.Join(skill.Sources, listSeparator)),
		Learning: skill.Learning,
	}
}

// durationSpans are how long a skill was used, by every kind a snapshot holds: a span as written, a gap as a gap
// still to fill, and a skill used on the side or still being learnt in the site's words.
func durationSpans(duration content.Duration, words content.Words) []document.Span {
	switch duration.Kind() {
	case content.DurationMeasured:
		return textSpans(duration.Written())
	case content.DurationGap:
		return markedSpans(duration.Written())
	case content.DurationSide:
		return textSpans(words.Text(content.WordSide))
	case content.DurationLearning:
		return textSpans(words.Text(content.WordLearning))
	}
	// A snapshot holds no other kind: a load refuses a duration it cannot read and one left out alike. Shown as
	// written, such a duration would say what the content says
	return textSpans(duration.Written())
}

func factList(facts []content.Fact) []document.Fact {
	list := make([]document.Fact, 0, len(facts))
	for _, fact := range facts {
		list = append(list, document.Fact{Label: fact.Label, Value: markedSpans(fact.Value)})
	}
	return list
}

func languageList(languages []content.Language) []document.Fact {
	list := make([]document.Fact, 0, len(languages))
	for _, language := range languages {
		list = append(list, document.Fact{Label: language.Name, Value: markedSpans(language.Level)})
	}
	return list
}

func points(bullets []string) []document.Point {
	list := make([]document.Point, 0, len(bullets))
	for _, bullet := range bullets {
		list = append(list, document.Point{Spans: markedSpans(bullet)})
	}
	return list
}

// section is a section under a word of the site, holding one block.
func section(title string, block document.Block) document.Section {
	return document.Section{Title: textSpans(title), Name: "", Blocks: []document.Block{block}}
}

// textSpans is a text with no marks of its own: a word of the site, or the text of a link the profile writes.
func textSpans(text string) []document.Span {
	return []document.Span{document.Text{Text: text}}
}

func linkSpans(target document.Target, text string) []document.Span {
	return []document.Span{document.Link{Target: target, Spans: textSpans(text)}}
}
