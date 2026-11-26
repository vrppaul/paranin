package document

import "strings"

// Project is what a project's page says of it beside its name and lines (DESIGN.md, ProjectHead): what gets the
// project, where its code lives, how it is installed and when another way is taken instead; and what tells about it,
// what it is made of, where its releases are read and its licence, each as far as the profile gives it. Then where its
// reader goes next (DESIGN.md, ProjectEnd): the author's other projects, the way to them all, and the author's mail.
type Project struct {
	// Repository is where the project's code lives; none where the profile names no place
	Repository *Repository
	// Install is the command that installs the project; none where the profile gives none
	Install string
	// Alternatives are the other ways to install it, each with a note on when to take it instead
	Alternatives []Alternative
	// Stack is what the project is made of, each name whole: the language it is written in, then what it is built
	// on; none where the profile names nothing
	Stack []string
	// Release is where the project's releases are read; none where the profile gives none
	Release *Release
	// Licence is the project's licence, a word; none where the profile names none
	Licence string
	// Others are the author's other projects with a page of their own, in the profile's order
	Others []OtherProject
	// AllProjects is the way to every project, the résumé's tab of them, by the site's words
	AllProjects Link
	// Mail is where the reader writes to the author
	Mail Mail
	// Labels are the site's words over each part, for a face that labels them
	Labels ProjectLabels
}

// OtherProject is another project of the author's as a project's page leads to it: its name, its tagline, what it is
// made of, each name whole, and its page.
type OtherProject struct {
	Name    []Span
	Tagline []Span
	Stack   []string
	Page    Page
}

// Alternative is a way to install a project instead of its install: a note on when to take it, and its command.
type Alternative struct {
	Note    []Span
	Command string
}

// Release is where a project's releases are read: a link to the newest, and one to their history, each by the site's
// word.
type Release struct {
	Latest    Link
	Changelog Link
}

// ProjectLabels are the site's words over the parts of a project's page that a face labels: its head's, and the
// heading over its other projects.
type ProjectLabels struct {
	Install string
	Stack   string
	Release string
	Licence string
	Others  string
}

// EndSection is where a project's reader goes next as a face that writes it as one more of the page's sections: under
// the heading over its other projects, the way to them all, each other project, where there is one, a link by its name
// with its tagline, then the author's mail.
func (project Project) EndSection() Section {
	blocks := []Block{Paragraph{Spans: []Span{project.AllProjects}}}
	if len(project.Others) > 0 {
		blocks = append(blocks, Contents{Listings: project.OtherListings()})
	}
	mail := Link{Target: project.Mail, Spans: []Span{Text{Text: project.Mail.Address}}}
	return Section{Title: []Span{Text{Text: project.Labels.Others}}, Name: "", Blocks: append(blocks, Paragraph{Spans: []Span{mail}})}
}

// OtherListings are the other projects as a face that lists pages by their titles and lines lists them: each by its
// name, a link to its page, and its tagline.
func (project Project) OtherListings() []Listing {
	listings := make([]Listing, 0, len(project.Others))
	for _, other := range project.Others {
		listings = append(listings, Listing{Page: other.Page, Title: other.Name, Note: other.Tagline})
	}
	return listings
}

// InstallFact is the install as a face that lists commands among facts writes it, under the site's word; none where
// the project gives no install.
func (project Project) InstallFact() (Fact, bool) {
	if project.Install == "" {
		return Fact{Label: "", Value: nil}, false
	}
	return CommandFact(project.Labels.Install, project.Install), true
}

// AboutFacts are what tells about the project as a face that writes it as text lists it among facts, each under the
// site's word, as far as the project gives it: what it is made of, its names joined by the face's own joint; its
// release's two links, the same joint between them; its licence.
func (project Project) AboutFacts(joint string) []Fact {
	facts := []Fact{}
	if len(project.Stack) > 0 {
		facts = append(facts, Fact{Label: project.Labels.Stack, Value: []Span{Text{Text: strings.Join(project.Stack, joint)}}})
	}
	if release := project.Release; release != nil {
		facts = append(facts, Fact{Label: project.Labels.Release, Value: []Span{release.Latest, Text{Text: joint}, release.Changelog}})
	}
	if project.Licence != "" {
		facts = append(facts, Fact{Label: project.Labels.Licence, Value: []Span{Text{Text: project.Licence}}})
	}
	return facts
}
