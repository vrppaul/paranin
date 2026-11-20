package compose

import (
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// Record is what compose makes of a page for whoever reads it as data rather than as text, a search engine
// reading JSON-LD among them (ADR-0003): the facts a vocabulary of data names, each as plain text. A face writes
// the vocabulary; compose picks the facts, as it does for a document. Every page a person reads has one.
//
//sumtype:decl
type Record interface {
	record()
}

// PersonRecord is the résumé's record: who it is about, the names they also go by, their title and mail, where
// they work now, none where no role goes on, the schools they went to, the skills they use, not those still
// learning, the languages they speak, the addresses of their public profiles, and the page that is theirs.
type PersonRecord struct {
	Name           string
	AlternateNames []string
	Title          string
	Email          string
	Employer       string
	Schools        []string
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
	Author      string
	AuthorPage  document.Page
	Page        document.Page
}

func (PersonRecord) record() {}
func (WorkRecord) record()   {}

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
		Schools:        schools(profile.Education),
		Skills:         skillsInUse(profile.SkillGroups),
		Languages:      languages,
		Profiles:       profiles,
		Page:           resumePage(),
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

func schools(education []content.School) []string {
	names := make([]string, 0, len(education))
	for _, school := range education {
		names = append(names, plainText(school.Name))
	}
	return names
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

func workRecord(project content.Project, author content.Person) WorkRecord {
	return WorkRecord{
		Name:        plainText(project.Name),
		Description: plainText(project.Description),
		Language:    project.Language,
		Repository:  project.Repo,
		Author:      plainText(author.Name),
		AuthorPage:  resumePage(),
		Page:        projectPage(project),
	}
}

// plainText is a text of the profile with its marks read and gone.
func plainText(marked string) string {
	return document.PlainText(markedSpans(marked))
}
