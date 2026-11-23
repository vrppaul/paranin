package html

import (
	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/routes"
)

// How a page's record is written as linked data: the media type of its script, the vocabulary it is in, and the
// vocabulary's kinds of thing it names
const (
	linkedDataType   = "application/ld+json"
	vocabulary       = "https://schema.org"
	profilePageKind  = "ProfilePage"
	personKind       = "Person"
	organizationKind = "Organization"
	schoolKind       = "EducationalOrganization"
	sourceCodeKind   = "SoftwareSourceCode"
)

// profilePage is the résumé as schema.org has a page about one person.
type profilePage struct {
	Context    string `json:"@context"`
	Kind       string `json:"@type"`
	URL        string `json:"url"`
	MainEntity person `json:"mainEntity"`
}

// person is who a page is about, as schema.org names them; a fact the record lacks is left out.
type person struct {
	Kind          string         `json:"@type"`
	Name          string         `json:"name"`
	AlternateName []string       `json:"alternateName,omitempty"`
	JobTitle      string         `json:"jobTitle,omitempty"`
	Email         string         `json:"email,omitempty"`
	URL           string         `json:"url"`
	WorksFor      organization   `json:"worksFor,omitzero"`
	AlumniOf      []organization `json:"alumniOf,omitempty"`
	KnowsAbout    []string       `json:"knowsAbout,omitempty"`
	KnowsLanguage []string       `json:"knowsLanguage,omitempty"`
	SameAs        []string       `json:"sameAs,omitempty"`
}

// organization is a company or a school, by its kind and name.
type organization struct {
	Kind string `json:"@type"`
	Name string `json:"name"`
}

// sourceCode is a project's page as schema.org has a piece of code, with its author.
type sourceCode struct {
	Context             string `json:"@context"`
	Kind                string `json:"@type"`
	Name                string `json:"name"`
	Description         string `json:"description,omitempty"`
	ProgrammingLanguage string `json:"programmingLanguage,omitempty"`
	URL                 string `json:"url"`
	CodeRepository      string `json:"codeRepository,omitempty"`
	Author              author `json:"author"`
}

// author is who wrote a piece of code, by their name and their page.
type author struct {
	Kind string `json:"@type"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// linkedData is the script of a page's record, in schema.org's words, for the page at an address.
func (writer PageWriter) linkedData(record compose.Record, address string) templ.Component {
	switch typed := record.(type) {
	case compose.PersonRecord:
		return linkedScript(writer.profilePageOf(typed, address))
	case compose.WorkRecord:
		return linkedScript(writer.sourceCodeOf(typed))
	}
	// compose makes every page of the résumé and the projects with its record; a post's page is drawn with none until
	// its BlogPosting comes with the site's SEO (TODO.md)
	return templ.NopComponent
}

// linkedScript is the script that holds an entity of schema.org as linked data.
func linkedScript(entity any) templ.Component {
	return templ.JSONScript("", entity).WithType(linkedDataType)
}

// profilePageOf is a page at an address about the person a record names, who is at the page that is theirs.
func (writer PageWriter) profilePageOf(record compose.PersonRecord, address string) profilePage {
	return profilePage{
		Context: vocabulary,
		Kind:    profilePageKind,
		URL:     address,
		MainEntity: person{
			Kind:          personKind,
			Name:          record.Name,
			AlternateName: record.AlternateNames,
			JobTitle:      record.Title,
			Email:         record.Email,
			URL:           writer.frame.site.Address(record.Page, routes.FormatChosen),
			WorksFor:      employerOf(record.Employer),
			AlumniOf:      schoolsOf(record.Schools),
			KnowsAbout:    record.Skills,
			KnowsLanguage: record.Languages,
			SameAs:        record.Profiles,
		},
	}
}

// employerOf is where a person works now, as an organization, or none where they work nowhere now.
func employerOf(name string) organization {
	if name == "" {
		return organization{Kind: "", Name: ""}
	}
	return organization{Kind: organizationKind, Name: name}
}

func schoolsOf(names []string) []organization {
	schools := make([]organization, 0, len(names))
	for _, name := range names {
		schools = append(schools, organization{Kind: schoolKind, Name: name})
	}
	return schools
}

func (writer PageWriter) sourceCodeOf(record compose.WorkRecord) sourceCode {
	return sourceCode{
		Context:             vocabulary,
		Kind:                sourceCodeKind,
		Name:                record.Name,
		Description:         record.Description,
		ProgrammingLanguage: record.Language,
		URL:                 writer.frame.site.Address(record.Page, routes.FormatChosen),
		CodeRepository:      record.Repository,
		Author: author{
			Kind: personKind,
			Name: record.Author,
			URL:  writer.frame.site.Address(record.AuthorPage, routes.FormatChosen),
		},
	}
}
