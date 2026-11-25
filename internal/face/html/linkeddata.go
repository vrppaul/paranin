package html

import (
	"time"

	"github.com/a-h/templ"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// How a page's record is written as linked data: the media type of its script, the vocabulary it is in, and the
// vocabulary's kinds of thing it names
const (
	linkedDataType   = "application/ld+json"
	vocabulary       = "https://schema.org"
	profilePageKind  = "ProfilePage"
	webSiteKind      = "WebSite"
	webPageKind      = "WebPage"
	personKind       = "Person"
	organizationKind = "Organization"
	schoolKind       = "EducationalOrganization"
	sourceCodeKind   = "SoftwareSourceCode"
	blogPostingKind  = "BlogPosting"
	seriesKind       = "CreativeWorkSeries"
)

// The name the author goes by in the résumé's linked data, after its address, so every record names the one person
const personNode = "person"

// profilePage is the résumé as schema.org has a page about one person.
type profilePage struct {
	Context    string `json:"@context"`
	Kind       string `json:"@type"`
	URL        string `json:"url"`
	MainEntity person `json:"mainEntity"`
}

// webSite is the site as schema.org has one: the name a search engine shows for it, its host as the name it also goes
// by, and its root.
type webSite struct {
	Context       string `json:"@context"`
	Kind          string `json:"@type"`
	Name          string `json:"name"`
	AlternateName string `json:"alternateName"`
	URL           string `json:"url"`
}

// webPage is a page about the one person that is not their profile, as a tab of the résumé is.
type webPage struct {
	Context string          `json:"@context"`
	Kind    string          `json:"@type"`
	URL     string          `json:"url"`
	About   personReference `json:"about"`
}

// person is who a page is about, as schema.org names them; a fact the record lacks is left out.
type person struct {
	Kind          string         `json:"@type"`
	ID            string         `json:"@id"`
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
	Context             string          `json:"@context"`
	Kind                string          `json:"@type"`
	Name                string          `json:"name"`
	Description         string          `json:"description,omitempty"`
	ProgrammingLanguage string          `json:"programmingLanguage,omitempty"`
	URL                 string          `json:"url"`
	CodeRepository      string          `json:"codeRepository,omitempty"`
	Author              personReference `json:"author"`
}

// personReference names the one person, by their @id, their name and their page: the author of a work, or whom a page
// is about.
type personReference struct {
	Kind string `json:"@type"`
	ID   string `json:"@id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// blogPosting is a post as schema.org has a post of a blog, with its author and the series it is a part of; a fact the
// record lacks is left out.
type blogPosting struct {
	Context       string             `json:"@context"`
	Kind          string             `json:"@type"`
	Headline      string             `json:"headline"`
	Description   string             `json:"description,omitempty"`
	URL           string             `json:"url"`
	DatePublished string             `json:"datePublished,omitempty"`
	DateModified  string             `json:"dateModified,omitempty"`
	Author        personReference    `json:"author"`
	InLanguage    string             `json:"inLanguage"`
	Keywords      []string           `json:"keywords,omitempty"`
	IsPartOf      creativeWorkSeries `json:"isPartOf,omitzero"`
}

// creativeWorkSeries is a series of posts, by its name and the page that lists its parts.
type creativeWorkSeries struct {
	Kind string `json:"@type"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// writtenPost is a post's record as its Open Graph and its linked data both write it, worked out once: when it came
// out and when its current version did, as a machine reads each, empty where it is not known, its author and its tags.
type writtenPost struct {
	published string
	modified  string
	author    personReference
	tags      []string
}

// recordOf is a page's record as the page writes it, for the page shown at an address: its script of linked data, in
// schema.org's words, and a post's as its Open Graph reads it too, none where the record is no post's. The person's
// own page is their profile, with the site beside it, each a script of its own; a tab of it is a page about them.
func (writer PageWriter) recordOf(record compose.Record, shown document.Page, address string) (templ.Component, *writtenPost) {
	switch typed := record.(type) {
	case compose.PersonRecord:
		if shown != typed.Page {
			return linkedScript(writer.webPageOf(typed, address)), nil
		}
		return templ.Join(linkedScript(writer.profilePageOf(typed, address)), linkedScript(writer.webSiteOf())), nil
	case compose.WorkRecord:
		return linkedScript(writer.sourceCodeOf(typed)), nil
	case compose.PostRecord:
		post := writer.writtenPostOf(typed)
		return linkedScript(writer.blogPostingOf(typed, post)), &post
	}
	// a list, filtered or not, and an older version of a post have no record and write none
	return templ.NopComponent, nil
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
			ID:            writer.frame.site.Identifier(record.Page, personNode),
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

// webSiteOf is the site at its root, named by its frame's name and by its host.
func (writer PageWriter) webSiteOf() webSite {
	site := writer.frame.site
	return webSite{
		Context:       vocabulary,
		Kind:          webSiteKind,
		Name:          writer.frame.siteName,
		AlternateName: site.Host(),
		URL:           site.Address(document.PageOf(document.PageResume), routes.FormatChosen),
	}
}

// webPageOf is a page at an address about the person a record names.
func (writer PageWriter) webPageOf(record compose.PersonRecord, address string) webPage {
	return webPage{
		Context: vocabulary,
		Kind:    webPageKind,
		URL:     address,
		About:   writer.personReferenceOf(compose.Author{Name: record.Name, Page: record.Page}),
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
		Author:              writer.personReferenceOf(record.Author),
	}
}

// writtenPostOf is a post's record as its page writes it.
func (writer PageWriter) writtenPostOf(record compose.PostRecord) writtenPost {
	return writtenPost{
		published: machineTime(record.Published, time.RFC3339),
		modified:  machineTime(record.VersionOut, time.RFC3339),
		author:    writer.personReferenceOf(record.Author),
		tags:      record.Tags,
	}
}

// blogPostingOf is a post at its page, as its page writes it, in the series its record names, where it names one.
func (writer PageWriter) blogPostingOf(record compose.PostRecord, post writtenPost) blogPosting {
	site := writer.frame.site
	var series creativeWorkSeries
	if record.Series != nil {
		series = creativeWorkSeries{Kind: seriesKind, Name: record.Series.Title, URL: site.Address(record.Series.Group, routes.FormatChosen)}
	}
	return blogPosting{
		Context:       vocabulary,
		Kind:          blogPostingKind,
		Headline:      record.Title,
		Description:   record.Description,
		URL:           site.Address(record.Page, routes.FormatChosen),
		DatePublished: post.published,
		DateModified:  post.modified,
		Author:        post.author,
		InLanguage:    record.Language,
		Keywords:      post.tags,
		IsPartOf:      series,
	}
}

// personReferenceOf is the one person as a record names them: by the IRI every record names them by, their name and
// their page.
func (writer PageWriter) personReferenceOf(named compose.Author) personReference {
	site := writer.frame.site
	return personReference{
		Kind: personKind, ID: site.Identifier(named.Page, personNode), Name: named.Name, URL: site.Address(named.Page, routes.FormatChosen),
	}
}

// machineTime is a time as a machine reads it, in a layout; empty where the time is not known.
func machineTime(moment time.Time, layout string) string {
	if moment.IsZero() {
		return ""
	}
	return moment.Format(layout)
}
