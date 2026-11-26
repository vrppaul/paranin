package document

import "strings"

// Project is what a project's page says of it beside its name and lines (DESIGN.md, ProjectHead): what gets the
// project, where its code lives, how it is installed and when another way is taken instead; and what tells about it,
// what it is made of, where its releases are read and its licence, each as far as the profile gives it.
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
	// Labels are the site's words over each part, for a face that labels them
	Labels ProjectLabels
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

// ProjectLabels are the site's words over the parts of a project's head that a face labels.
type ProjectLabels struct {
	Install string
	Stack   string
	Release string
	Licence string
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
