// Package compose turns content into the documents every face draws (docs/ARCHITECTURE.md, ADR-0003). It
// decides once which facts a page shows and in what order, so no face picks them again; it reads the profile's
// marks into spans and carries its months across. It writes no format of any face: how a span or a period reads
// is each face's own.
package compose

import (
	"strings"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// The scheme of a link that writes a mail
const mailScheme = "mailto:"

// What separates the values of one cell: the places a skill was proven
const listSeparator = ", "

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

func resumeHead(profile content.Profile, words content.Words) document.Head {
	person := profile.Person
	return document.Head{
		Title:    markedSpans(person.Name),
		Subtitle: markedSpans(person.Title),
		Place:    person.Location,
		Lead:     markedSpans(person.Headline),
		Ways:     ways(person, profile.Site, words),
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
			Period:   periodOf(role.Start, role.End),
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
			Period:   periodOf(school.Start, school.End),
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
			Facts:    projectFacts(project, words),
			Page:     projectPage(project),
		})
	}
	return document.Section{Title: textSpans(words.Text(content.WordProjects)), Blocks: cards}
}

// projectPage is a project's own page, or no page where it has none.
func projectPage(project content.Project) document.Page {
	if !project.Page {
		return document.Page{Kind: document.PageNone, Slug: ""}
	}
	return document.Page{Kind: document.PageProject, Slug: project.Slug}
}

// projectFacts are where a project's code lives and how it is installed, each where the project has one.
func projectFacts(project content.Project, words content.Words) []document.Fact {
	facts := []document.Fact{}
	if project.Repo != "" {
		facts = append(facts, document.Fact{Label: words.Text(content.WordRepository), Value: linkSpans(document.Address{URL: project.Repo}, project.Repo)})
	}
	if project.Install != "" {
		facts = append(facts, document.Fact{Label: words.Text(content.WordInstall), Value: []document.Span{document.Code{Text: project.Install}}})
	}
	return facts
}

// skillsSection is a section of each group of skills, each a table of a skill, how long, and where it was proven.
func skillsSection(groups []content.SkillGroup, words content.Words) document.Section {
	sections := make([]document.Block, 0, len(groups))
	for _, group := range groups {
		rows := make([]document.Row, 0, len(group.Skills))
		for _, skill := range group.Skills {
			rows = append(rows, document.Row{Cells: []document.Cell{
				{Spans: markedSpans(skill.Name)},
				{Spans: markedSpans(skill.Duration)},
				{Spans: markedSpans(strings.Join(skill.Sources, listSeparator))},
			}})
		}
		sections = append(sections, document.Section{Title: markedSpans(group.Title), Blocks: []document.Block{document.Table{Rows: rows}}})
	}
	return document.Section{Title: textSpans(words.Text(content.WordSkills)), Blocks: sections}
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
