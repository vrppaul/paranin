// Package compose turns content into the documents every face draws (docs/ARCHITECTURE.md, ADR-0003). It
// decides once which facts a page shows and in what order, so no face picks them again; it reads the profile's
// marks into spans and carries its months across. Beside each page a person reads it makes the page's record, the
// facts a search engine reads as data, typed and plain. Beside the documents it hands a face the words of the
// controls it lays round them, the page's choice of theme (controls.go), since a face reads no content. It writes
// no format of any face: how a span or a period reads, and the vocabulary a record is written in, are each face's
// own.
package compose

import (
	"slices"
	"strings"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// The scheme of a link that writes a mail
const mailScheme = "mailto:"

// What separates the places a skill was proven, written in one text
const listSeparator = ", "

// Composed is a page a person reads, as compose makes it: its document, and its record, made together so neither
// is ever without the other.
type Composed struct {
	Document document.Document
	Record   Record
}

// Pages are the pages a snapshot makes for a person to read, each with its record: the résumé, then the page of
// each project that has one. Index lists them.
func Pages(snapshot *content.Snapshot) []Composed {
	profile := snapshot.Profile
	pages := []Composed{{Document: Resume(snapshot), Record: personRecord(profile)}}
	for _, project := range projectsWithPages(profile.Projects) {
		page := Composed{Document: projectDocument(project, profile.Person, snapshot.Words), Record: workRecord(project, profile.Person)}
		pages = append(pages, page)
	}
	return pages
}

// Documents are the documents of the pages a person reads, in their order, for a face that draws no record.
func Documents(snapshot *content.Snapshot) []document.Document {
	pages := Pages(snapshot)
	documents := make([]document.Document, 0, len(pages))
	for _, page := range pages {
		documents = append(documents, page.Document)
	}
	return documents
}

// resumePage is the résumé's page, the site's first.
func resumePage() document.Page {
	return document.Page{Kind: document.PageResume, Slug: ""}
}

// projectsWithPages are the projects that have a page of their own, in their order.
func projectsWithPages(projects []content.Project) []content.Project {
	return slices.DeleteFunc(slices.Clone(projects), func(project content.Project) bool { return !project.Page })
}

// Resume is the résumé as one document: its head, the summary, then its facts, experience, projects, skills,
// education and languages, each a section under the site's word for it.
func Resume(snapshot *content.Snapshot) document.Document {
	profile, words := snapshot.Profile, snapshot.Words
	return document.Document{
		Head: resumeHead(profile, words),
		Blocks: []document.Block{
			document.Paragraph{Spans: markedSpans(profile.Summary)},
			section(words.Text(content.WordFacts), document.Facts{List: factList(profile.Facts)}),
			experienceSection(profile.Experience, words),
			projectsSection(profile.Projects, words),
			skillsSection(profile.SkillGroups, words),
			educationSection(profile.Education, words),
			section(words.Text(content.WordLanguages), document.Facts{List: languageList(profile.Languages)}),
		},
	}
}

// Index is the site's index, the list of its pages an agent reads first: whose site it is and its headline, then
// the résumé and the page of each project that has one, each under the site's word for it, the projects' word
// only where a project has a page. A page is listed by the title and subtitle its own head opens with.
func Index(snapshot *content.Snapshot) document.Document {
	words := snapshot.Words
	resume := resumeHead(snapshot.Profile, words)
	blocks := []document.Block{section(words.Text(content.WordResume), document.Contents{Listings: []document.Listing{listing(resume)}})}
	projects := []document.Listing{}
	for _, project := range projectsWithPages(snapshot.Profile.Projects) {
		projects = append(projects, listing(projectHead(project, words)))
	}
	if len(projects) > 0 {
		blocks = append(blocks, section(words.Text(content.WordProjects), document.Contents{Listings: projects}))
	}
	return document.Document{
		Head: document.Head{
			Page:         document.Page{Kind: document.PageIndex, Slug: ""},
			Language:     words.Language(),
			Greeting:     "",
			Title:        resume.Title,
			Subtitle:     nil,
			Place:        "",
			Tags:         nil,
			Lead:         resume.Lead,
			Ways:         nil,
			Repository:   nil,
			Description:  nil,
			OpeningTitle: "",
		},
		Blocks: blocks,
	}
}

// listing is a page as a list of contents names it: by its head's title and subtitle.
func listing(head document.Head) document.Listing {
	return document.Listing{Page: head.Page, Title: head.Title, Note: head.Subtitle}
}

// resumeHead is the résumé's head, its headline its description: it greets the reader before the author's name,
// and names its opening, in the site's words.
func resumeHead(profile content.Profile, words content.Words) document.Head {
	person := profile.Person
	headline := markedSpans(person.Headline)
	return document.Head{
		Page:         resumePage(),
		Language:     words.Language(),
		Greeting:     words.Text(content.WordGreeting),
		Title:        markedSpans(person.Name),
		Subtitle:     markedSpans(person.Title),
		Place:        person.Location,
		Tags:         nil,
		Lead:         headline,
		Ways:         ways(person, profile.Site, words),
		Repository:   nil,
		Description:  headline,
		OpeningTitle: words.Text(content.WordAbout),
	}
}

// ways are the ways to reach the author, in the order a reader tries them: mail, the public profiles, ssh.
func ways(person content.Person, site content.Site, words content.Words) []document.Fact {
	mail := document.Address{URL: mailScheme + person.Email}
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
			Period:   periodOf(role.Start, role.End, words),
			Title:    markedSpans(role.Company),
			Subtitle: markedSpans(role.Title),
			Place:    role.Location,
			Points:   points(role.Bullets),
		})
	}
	return document.Section{Title: textSpans(words.Text(content.WordExperience)), Blocks: entries}
}

func educationSection(schools []content.School, words content.Words) document.Section {
	entries := make([]document.Block, 0, len(schools))
	for _, school := range schools {
		entries = append(entries, document.Entry{
			Period:   periodOf(school.Start, school.End, words),
			Title:    markedSpans(school.Name),
			Subtitle: markedSpans(school.Field),
			Place:    "",
			Points:   nil,
		})
	}
	return document.Section{Title: textSpans(words.Text(content.WordEducation)), Blocks: entries}
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
	return document.Section{Title: textSpans(words.Text(content.WordProjects)), Blocks: cards}
}

// projectDocument is a project's own page: its head; its details, its author first, a link back to the résumé, and
// how it is installed last; its recording, where it has one; then its sections.
func projectDocument(project content.Project, author content.Person, words content.Words) document.Document {
	resume := resumePage()
	details := []document.Fact{{Label: words.Text(content.WordAuthor), Value: linkSpans(resume, author.Name)}}
	details = append(details, factList(project.Facts)...)
	blocks := []document.Block{document.Facts{List: append(details, installFacts(project, words)...)}}
	if project.Recording != "" {
		blocks = append(blocks, document.CodeBlock{Text: project.Recording})
	}
	for _, section := range project.Sections {
		paragraphs := make([]document.Block, 0, len(section.Paragraphs))
		for _, paragraph := range section.Paragraphs {
			paragraphs = append(paragraphs, document.Paragraph{Spans: markedSpans(paragraph)})
		}
		blocks = append(blocks, document.Section{Title: markedSpans(section.Title), Blocks: paragraphs})
	}
	return document.Document{Head: projectHead(project, words), Blocks: blocks}
}

// projectHead is the head of a project's page, in the language of the site's words: its name, what it is in a
// line, what it is made of, what it does, its description, and where its code lives.
func projectHead(project content.Project, words content.Words) document.Head {
	description := markedSpans(project.Description)
	return document.Head{
		Page:         projectPage(project),
		Language:     words.Language(),
		Greeting:     "",
		Title:        markedSpans(project.Name),
		Subtitle:     markedSpans(project.Tagline),
		Place:        "",
		Tags:         projectTags(project),
		Lead:         description,
		Ways:         nil,
		Repository:   projectRepository(project, words),
		Description:  description,
		OpeningTitle: "",
	}
}

// projectTags are what a project is made of: the language it is written in, where the profile names one, then what
// it is built on.
func projectTags(project content.Project) []string {
	if project.Language == "" {
		return slices.Clone(project.Stack)
	}
	return append([]string{project.Language}, project.Stack...)
}

// projectPage is a project's own page, or no page where it has none.
func projectPage(project content.Project) document.Page {
	if !project.Page {
		return document.Page{Kind: document.PageNone, Slug: ""}
	}
	return document.Page{Kind: document.PageProject, Slug: project.Slug}
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

// skillsSection is a section of each group of skills, each a table of a skill, how long, and where it was proven.
func skillsSection(groups []content.SkillGroup, words content.Words) document.Section {
	sections := make([]document.Block, 0, len(groups))
	for _, group := range groups {
		rows := make([]document.Row, 0, len(group.Skills))
		for _, skill := range group.Skills {
			listed := listedSkill(skill)
			rows = append(rows, document.Row{Cells: []document.Cell{{Spans: listed.Name}, {Spans: listed.Duration}, {Spans: listed.Sources}}})
		}
		sections = append(sections, document.Section{Title: markedSpans(group.Title), Blocks: []document.Block{document.Table{Rows: rows}}})
	}
	return document.Section{Title: textSpans(words.Text(content.WordSkills)), Blocks: sections}
}

// listedSkill is a skill as the documents hold it, in the résumé's table and in the list of skills alike: its name,
// how long, the places it was proven in one text, and whether the author is still learning it.
func listedSkill(skill content.Skill) document.Skill {
	return document.Skill{
		Name:     markedSpans(skill.Name),
		Duration: markedSpans(skill.Duration),
		Sources:  markedSpans(strings.Join(skill.Sources, listSeparator)),
		Learning: skill.Learning,
	}
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
	return document.Section{Title: textSpans(title), Blocks: []document.Block{block}}
}

// textSpans is a text with no marks of its own: a word of the site, or the text of a link the profile writes.
func textSpans(text string) []document.Span {
	return []document.Span{document.Text{Text: text}}
}

func linkSpans(target document.Target, text string) []document.Span {
	return []document.Span{document.Link{Target: target, Spans: textSpans(text)}}
}
