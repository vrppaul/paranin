package compose

import (
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// Record is what compose makes of a page for whoever reads it as data rather than as text, a search engine
// reading JSON-LD among them (ADR-0003): the facts a vocabulary of data names, each as plain text. A face writes
// the vocabulary; compose picks the facts, as it does for a document. The résumé, each project's page and each post
// have one; the blog's lists, filtered or not, and an older version of a post, which a newer stands over, have none.
//
//sumtype:decl
type Record interface {
	record()
}

// PersonRecord is the résumé's record: who it is about, the names they also go by, their title and mail, where
// they work now, none where no role goes on, the skills they use, not those still learning, the languages they
// speak, the addresses of their public profiles, and the page that is theirs.
type PersonRecord struct {
	Name           string
	AlternateNames []string
	Title          string
	Email          string
	Employer       string
	Skills         []string
	Languages      []string
	Profiles       []string
	Page           document.Page
}

// WorkRecord is the record of a project's page: the project, what it does, what it is written in, its repository,
// its author with the page that is theirs, and its own page, whole even where a tab of it is shown.
type WorkRecord struct {
	Name        string
	Description string
	Language    string
	Repository  string
	Author      Author
	Page        document.Page
}

// PostRecord is the record of a post's current version: its title and lead, its page, when the post came out, the zero
// time for a draft not yet dated, when its current version did, the zero time where neither the post nor an update is
// dated, its author, the language it is written in, its tags, and its series, none where its page shows none.
type PostRecord struct {
	Title       string
	Description string
	Page        document.Page
	Published   time.Time
	VersionOut  time.Time
	Author      Author
	Language    string
	Tags        []string
	Series      *PostSeries
}

// Author is who wrote a work a record names, a project or a post: their name, and the page that is theirs.
type Author struct {
	Name string
	Page document.Page
}

// PostSeries is a series as a post's record names it: its title, and its group on the blog's list by series.
type PostSeries struct {
	Title string
	Group document.Anchor
}

func (PersonRecord) record() {}
func (WorkRecord) record()   {}
func (PostRecord) record()   {}

func personRecord(profile content.Profile) PersonRecord {
	person := profile.Person
	profiles := make([]string, 0, len(person.Links))
	for _, link := range person.Links {
		profiles = append(profiles, link.URL)
	}
	languages := make([]string, 0, len(profile.Languages))
	for _, language := range profile.Languages {
		languages = append(languages, language.Name)
	}
	return PersonRecord{
		Name:           plainText(person.Name),
		AlternateNames: person.AlternateNames,
		Title:          plainText(person.Title),
		Email:          person.Email,
		Employer:       employer(profile.Experience),
		Skills:         skillsInUse(profile.SkillGroups),
		Languages:      languages,
		Profiles:       profiles,
		Page:           document.PageOf(document.PageResume),
	}
}

// employer is where the author works now: the company of the first role that goes on, none where none does.
func employer(roles []content.Role) string {
	for _, role := range roles {
		if role.End.Kind() == content.MonthCurrent {
			return plainText(role.Company)
		}
	}
	return ""
}

// skillsInUse are the names of the skills the author uses, in the profile's order: not those still learning.
func skillsInUse(groups []content.SkillGroup) []string {
	names := []string{}
	for _, group := range groups {
		for _, skill := range group.Skills {
			if !skill.Learning {
				names = append(names, plainText(skill.Name))
			}
		}
	}
	return names
}

func workRecord(project content.Project, author Author) WorkRecord {
	return WorkRecord{
		Name:        plainText(project.Name),
		Description: plainText(project.Description),
		Language:    project.Language,
		Repository:  project.Repo,
		Author:      author,
		Page:        projectPage(project),
	}
}

// authorOf is the author of every work a record names: the profile's person, whose page is the résumé.
func authorOf(person content.Person) Author {
	return Author{Name: plainText(person.Name), Page: document.PageOf(document.PageResume)}
}

// postRecord is the record of a post's current version, by its author, in the blog's language, naming the series its
// page shows.
func (read blog) postRecord(post content.Post, author Author) PostRecord {
	var series *PostSeries
	if read.seriesShownOn(post) != noSeries {
		shown := read.seriesOf[post.Slug]
		series = &PostSeries{Title: shown.Title, Group: read.seriesGroup(shown)}
	}
	return PostRecord{
		Title:       post.Title,
		Description: post.Description,
		Page:        postPage(post),
		Published:   post.Published.At(),
		VersionOut:  versionOutAt(post),
		Author:      author,
		Language:    read.words.Language(),
		Tags:        post.Tags,
		Series:      series,
	}
}

// plainText is a text of the profile with its marks read and gone.
func plainText(marked string) string {
	return document.PlainText(markedSpans(marked))
}
