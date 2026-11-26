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
		pages = append(pages, composedPage(projectDocument(project, profile, snapshot.Own.Words), workRecord(project, author)))
	}
	return pages
}

// PageAndTabs are the documents a page is served as: the page whole, then each of its tabs alone.
func (composed Composed) PageAndTabs() []document.Document {
	return slices.Concat([]document.Document{composed.Document}, composed.Tabs)
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
		Lead:        nil,
		Version:     nil,
		Project:     nil,
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

// Resume is the résumé as one document: its head, naming the feeds a reader handed the site's address finds, then three
// sections, each under the site's word for it, as a reader asks of an author: about them and how to reach them, their
// experience, and their projects (DESIGN.md, Pages).
func Resume(snapshot *content.Snapshot, feeds []document.Link) document.Document {
	profile, words := snapshot.Profile, snapshot.Own.Words
	return document.Document{
		Head: resumeHead(profile, words, feeds),
		Blocks: []document.Block{
			aboutSection(profile, words),
			experienceSection(profile.Experience, profile.SkillGroups, words),
			projectsSection(profile.Projects, words),
		},
	}
}

// aboutSection is the section about the author: the summary, the facts, then the ways to reach them, which the first
// screen places as each face chooses.
func aboutSection(profile content.Profile, words content.Words) document.Section {
	return document.Section{
		Title: textSpans(words.Text(content.WordAbout)),
		Name:  "",
		Blocks: []document.Block{
			document.Paragraph{Spans: markedSpans(profile.Summary)},
			document.Facts{Kind: document.FactsDetails, List: factList(profile.Facts)},
			document.Facts{Kind: document.FactsWays, List: ways(profile.Person, profile.Site, words)},
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
		projects = append(projects, listing(projectHead(project, snapshot.Profile, words)))
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
			Lead:        resume.Lead,
			Version:     nil,
			Project:     nil,
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
	return document.Listing{Target: head.Page, Title: head.Title, Note: head.Subtitle}
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
		Lead:        headline,
		Version:     nil,
		Project:     nil,
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
	return append(list, document.Fact{Label: words.Text(content.WordSSH), Value: []document.Span{document.Code{Text: sshCommand(site.SSHHost, anyUser)}}})
}

// experienceSection is the author's work: its head, where the profile writes its line or puts any skill forward, each
// by its name and how long it was used; then each role, with the technologies it worked with.
func experienceSection(experience content.Experience, groups []content.SkillGroup, words content.Words) document.Section {
	blocks := make([]document.Block, 0, len(experience.Roles)+1)
	if experience.Line != "" || len(experience.PutForward) > 0 {
		blocks = append(blocks, document.ExperienceHead{Line: markedSpans(experience.Line), Technologies: putForward(experience.PutForward, groups, words)})
	}
	for _, role := range experience.Roles {
		blocks = append(blocks, document.Entry{
			Period:       periodOf(role.Start, role.End, words),
			Title:        markedSpans(role.Company),
			Subtitle:     markedSpans(role.Title),
			Technologies: technologiesOf(role.Company, groups),
			Points:       points(role.Bullets),
			Description:  paragraphs(role.Description),
		})
	}
	return document.Section{Title: textSpans(words.Text(content.WordExperience)), Name: "", Blocks: blocks}
}

// technologiesOf are the names of the skills whose sources name a company, in the profile's order of groups and
// skills, a skill still being learnt among them: a skill's sources are the one place the profile says where it was
// used, a role or anything else.
func technologiesOf(company string, groups []content.SkillGroup) []string {
	named := plainText(company)
	technologies := []string{}
	for _, group := range groups {
		for _, skill := range group.Skills {
			if slices.ContainsFunc(skill.Sources, func(source string) bool { return plainText(source) == named }) {
				technologies = append(technologies, skill.Name)
			}
		}
	}
	return technologies
}

// putForward are the skills of names the groups list as technologies, each by its name and how long it was used, as
// the skills' list writes it; the check holds every name to a skill of a group.
func putForward(names []string, groups []content.SkillGroup, words content.Words) []document.Technology {
	durations := map[string]content.Duration{}
	for _, group := range groups {
		for _, skill := range group.Skills {
			durations[skill.Name] = skill.Duration
		}
	}
	technologies := make([]document.Technology, 0, len(names))
	for _, name := range names {
		technologies = append(technologies, document.Technology{Name: name, Used: durationSpans(durations[name], words)})
	}
	return technologies
}

// paragraphs are texts of the profile as paragraphs, their marks read.
func paragraphs(texts []string) []document.Paragraph {
	written := make([]document.Paragraph, 0, len(texts))
	for _, text := range texts {
		written = append(written, document.Paragraph{Spans: markedSpans(text)})
	}
	return written
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

// projectDocument is a project's own page, one of the profile's: its head, then its sections.
func projectDocument(project content.Project, profile content.Profile, words content.Words) document.Document {
	blocks, taken := make([]document.Block, 0, len(project.Sections)), anchors{}
	for _, section := range project.Sections {
		blocks = append(blocks, projectSection(section, taken))
	}
	return document.Document{Head: projectHead(project, profile, words), Blocks: blocks}
}

// projectSection is a section of a project's page, named by its title from the page's anchors as a post's headings
// are: its paragraphs, then its points, then its side, its links, where it has any.
func projectSection(section content.Section, taken anchors) document.Section {
	blocks := make([]document.Block, 0, len(section.Paragraphs))
	for _, paragraph := range paragraphs(section.Paragraphs) {
		blocks = append(blocks, paragraph)
	}
	if len(section.Points) > 0 {
		blocks = append(blocks, document.Points{List: points(section.Points)})
	}
	if len(section.Links) > 0 {
		blocks = append(blocks, document.Side{Blocks: []document.Block{document.Contents{Listings: linkListings(section.Links)}}})
	}
	title := markedSpans(section.Title)
	return document.Section{Title: title, Name: taken.headingAnchor(title), Blocks: blocks}
}

// linkListings are a section's links as contents: each leading to its address by its name, as written, with its line.
func linkListings(links []content.Link) []document.Listing {
	listings := make([]document.Listing, 0, len(links))
	for _, link := range links {
		listings = append(listings, document.Listing{Target: document.Address{URL: link.URL.Written()}, Title: textSpans(link.Name), Note: markedSpans(link.Line)})
	}
	return listings
}

// projectHead is the head of a project's page, in the language of the site's words: its name, what it is in a
// line, what it does, its description, what gets it and tells about it, and where its reader goes next.
func projectHead(project content.Project, profile content.Profile, words content.Words) document.Head {
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
		Lead:        description,
		Version:     nil,
		Project:     projectPart(project, profile, words),
		Post:        nil,
		Description: description,
		Feeds:       nil,
		Languages:   nil,
	}
}

// projectPart is what a project's head says of it beside its name and lines, each part under the site's word: where
// its code lives; its install's command and the other ways, each note as prose and each command as written; what it is
// made of, where its releases are read and its licence; and where its reader goes next, the profile's other projects
// with a page, the way to them all on the résumé's tab of them, and the author's mail.
func projectPart(project content.Project, profile content.Profile, words content.Words) *document.Project {
	install, alternatives := projectInstall(project.Install)
	projectsTab := document.TabPage(document.PageOf(document.PageResume), textSpans(words.Text(content.WordProjects)))
	return &document.Project{
		Repository:   projectRepository(project, words),
		Install:      install,
		Alternatives: alternatives,
		Stack:        projectMadeOf(project),
		Release:      projectRelease(project.Release, words),
		Licence:      project.Licence,
		Others:       otherProjects(project, profile.Projects),
		AllProjects:  document.Link{Target: projectsTab, Spans: textSpans(words.Text(content.WordAllProjects))},
		Mail:         document.Mail{Address: profile.Person.Email},
		Labels: document.ProjectLabels{
			Install: words.Text(content.WordInstall),
			Stack:   words.Text(content.WordStack),
			Release: words.Text(content.WordRelease),
			Licence: words.Text(content.WordLicence),
			Others:  words.Text(content.WordOtherProjects),
			Contact: words.Text(content.WordContact),
		},
	}
}

// otherProjects are the projects with a page of their own a project's page leads to, in the profile's order, the
// project itself left out: each by its name, its tagline, what it is made of and its page.
func otherProjects(project content.Project, projects []content.Project) []document.OtherProject {
	others := []document.OtherProject{}
	for _, other := range projectsWithPages(projects) {
		if other.Slug == project.Slug {
			continue
		}
		others = append(others, document.OtherProject{
			Name: markedSpans(other.Name), Tagline: markedSpans(other.Tagline), Stack: projectMadeOf(other), Page: projectPage(other),
		})
	}
	return others
}

// projectInstall is a project's install: its command and the other ways, each note as prose; none where the profile
// gives no install.
func projectInstall(install *content.Install) (string, []document.Alternative) {
	if install == nil {
		return "", nil
	}
	alternatives := make([]document.Alternative, 0, len(install.Alternatives))
	for _, alternative := range install.Alternatives {
		alternatives = append(alternatives, document.Alternative{Note: markedSpans(alternative.Note), Command: alternative.Command})
	}
	return install.Command, alternatives
}

// projectRelease is where a project's releases are read: a link to the newest and one to their history, each by the
// site's word, with no number, which would go stale; none where the profile gives no release.
func projectRelease(release *content.Release, words content.Words) *document.Release {
	if release == nil {
		return nil
	}
	return &document.Release{
		Latest:    document.Link{Target: document.Address{URL: release.Latest.Written()}, Spans: textSpans(words.Text(content.WordLatest))},
		Changelog: document.Link{Target: document.Address{URL: release.Changelog.Written()}, Spans: textSpans(words.Text(content.WordChangelog))},
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

// installFacts are how a project is installed as its card says it, its install's command under the site's word for
// it, where the project gives one.
func installFacts(project content.Project, words content.Words) []document.Fact {
	if project.Install == nil {
		return nil
	}
	return []document.Fact{document.CommandFact(words.Text(content.WordInstall), project.Install.Command)}
}

// listedSkill is a skill as the hidden list of skills holds it: its name, how long in text and in months, the places it
// was proven in one text, and whether the author is still learning it.
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
