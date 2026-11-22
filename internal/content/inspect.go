package content

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// How a slug is written: small letters and digits, words joined by hyphens
const slugPattern = `^[a-z0-9]+(-[a-z0-9]+)*$`

// How a number alone is written, which a post's or a series' name may not be: blog/12 reads as a post's number
const numberPattern = `^[0-9]+$`

// How a fault says what a slug is, a slug's and a tag's alike
const slugForm = "small letters and digits, joined by hyphens"

// What is said of each kind of fault, as a format for the values it names; a fault of a value written quotes it
const (
	emptyProblem         = "is empty"
	lineProblem          = "holds more than one line"
	addressProblem       = "%q is no absolute web address"
	insecureProblem      = "%q is no https address"
	slashProblem         = "%q ends in a slash, which every address adds itself"
	mailProblem          = "%q is no bare mail address"
	unknownMonth         = "%q is no month, YYYY-MM"
	unknownEnd           = unknownMonth + ", nor " + currentWord
	unknownMonthOrGap    = unknownMonth + ", nor a " + GapOpening + "gap" + GapClosing
	unknownDuration      = "%q is no span, as 1y 7m, nor " + sideWord + ", " + learningWord + " or a " + GapOpening + "gap" + GapClosing
	unflaggedLearning    = "%q says the skill is still being learnt, but learning is not true"
	orderProblem         = "%q comes before the start, %q"
	slugProblem          = "%q is no slug: " + slugForm
	numberNameProblem    = "%q is digits alone, which an address would read as a number"
	sharedSlugProblem    = "%q names another project too"
	sharedGroupProblem   = "%q names another group too"
	stackLanguageProblem = "%q is the project's language already"
	emptyGroupProblem    = "holds no skill"
	noneOfProblem        = "%q is none of %s"
	positiveProblem      = "%d is not above 0"
	rangeProblem         = "%d is not above min, %d"
	unknownWordProblem   = "is no word of the site"
	placeProblem         = "%q writes its place, %s, %d times, where its values ask for %d"
	characterProblem     = "%q holds a control character, a mark of direction or bytes of no character, which every face would pass on"
	sharedNumberProblem  = "%d is another post's number too"
	unknownMoment        = "%q is no date and time with its zone, as 2026-11-10T09:00+01:00"
	unknownDay           = "%q is no date, YYYY-MM-DD"
	updatedProblem       = "%q comes before the day the post was published, %q"
	tagProblem           = "%q is no tag: " + slugForm
	sharedTagProblem     = "%q is a tag of the post already"
	strayPostProblem     = "is no post's folder: the posts' folder holds a folder for each post"
	translationProblem   = "is a translation, and translations come later"
	revisionsProblem     = "holds revisions, which come with their own step"
	widgetsProblem       = "holds widgets, which come with the first island"
	strayPartProblem     = "is no part of a post: a post's folder holds its " + englishPostFile
	frontMatterProblem   = "opens with no front matter: a line " + frontMatterFence + ", the post's fields, and another"
	unclosedProblem      = "opens front matter that no line " + frontMatterFence + " closes"
	rawHTMLProblem       = "%q is raw HTML, which a terminal cannot draw and the page's policy would not run"
	titleLevelProblem    = "is a heading of level 1, the post's title's alone"
	skippedLevelProblem  = "is a heading of level %d after one of %d, which skips a level"
	widgetProblem        = "is a widget, and widgets come with the first island"
	uncitedProblem       = "%q is a footnote no reference cites"
	unansweredProblem    = "%q cites no footnote"
	noteFormProblem      = "%q is no one paragraph, which a note set beside the line it is cited on holds"
	noteCitingProblem    = "cites a footnote from inside a footnote, whose words stand beside the line alone"
	noteHeadingProblem   = "cites a footnote from a heading, which names a part and is no line a source stands beside"
	kindProblem          = "is a node of goldmark's kind %s, which no post holds: no face draws it"
	pictureProblem       = "is a picture, and pictures come with the step that serves them from the post's folder"
	breakProblem         = "is a thematic break, which no face draws: a heading parts a post"
	extraCellsProblem    = "is a table's row of more cells than its header, whose cells past the header's GFM drops"
	lineBreakProblem     = "breaks its line hard, which no face draws: a paragraph's lines join, and a list parts them"
	linkProblem          = "%q is no absolute web address and no mailto: of a bare mail address"
	linkTitleProblem     = "%q is a link's title, which no face shows"
	straySeriesProblem   = "is no series: the series' folder holds a file <name>" + yamlEnding + " for each"
	emptySeriesProblem   = "holds no post"
	unknownPostProblem   = "%q names no post"
	sharedPartProblem    = "%q is a part of the series already"
	otherSeriesProblem   = "%q is a part of another series, %q"
)

// What separates the kinds a fault lists where a value is none of them
const kindSeparator = ", "

// The field of the site's address, which two checks read
const baseAddressField = "site.base_url"

// inspection gathers the faults of one file, each at its field, so every one is reported at once.
type inspection struct {
	file   string
	faults Faults
}

// add writes a fault from its problem's format and the values it names. It takes the format whole, as Sprintf
// does, so vet refuses a format whose values are missing.
func (inspection *inspection) add(field, format string, values ...any) {
	problem := fmt.Sprintf(format, values...)
	inspection.faults = append(inspection.faults, Fault{File: inspection.file, Field: field, Line: 0, Problem: problem})
}

// addAtLine writes a fault at a line of the file, as add writes one at a field.
func (inspection *inspection) addAtLine(line int, format string, values ...any) {
	problem := fmt.Sprintf(format, values...)
	inspection.faults = append(inspection.faults, Fault{File: inspection.file, Field: "", Line: line, Problem: problem})
}

func (inspection *inspection) require(field, value string) {
	if blank(value) {
		inspection.add(field, emptyProblem)
	}
}

// requireLine holds a value to some text on one line: a title, a line about a post or a series, a site's name.
func (inspection *inspection) requireLine(field, value string) {
	switch {
	case blank(value):
		inspection.add(field, emptyProblem)
	case strings.ContainsRune(value, lineFeed):
		inspection.add(field, lineProblem)
	}
}

// requireName holds the name a post's folder or a series' file gives it, which an address carries, to a slug that is
// not a number alone; the fault stands at the folder or the file it names.
func (inspection *inspection) requireName(name string) {
	switch {
	case !regexp.MustCompile(slugPattern).MatchString(name):
		inspection.add("", slugProblem, name)
	case regexp.MustCompile(numberPattern).MatchString(name):
		inspection.add("", numberNameProblem, name)
	}
}

// blank is whether a value of the content says nothing: it holds no text but spaces.
func blank(value string) bool {
	return strings.TrimSpace(value) == ""
}

func (inspection *inspection) requireWebAddress(field string, address WebAddress) {
	if !address.absolute() {
		inspection.add(field, addressProblem, address.Written())
	}
}

// requireSecureWebAddress holds an address to an absolute one of https.
func (inspection *inspection) requireSecureWebAddress(field string, address WebAddress) {
	switch {
	case !address.absolute():
		inspection.add(field, addressProblem, address.Written())
	case !address.secure:
		inspection.add(field, insecureProblem, address.Written())
	}
}

// inspectProfile checks what the schema cannot say: what must be there, how it is written, and how its parts
// agree with each other.
func inspectProfile(profile Profile) Faults {
	inspection := &inspection{file: profileFile, faults: nil}
	inspectSite(inspection, profile.Site)
	inspectPerson(inspection, profile.Person)
	inspectExperience(inspection, profile.Experience)
	inspectEducation(inspection, profile.Education)
	inspectProjects(inspection, profile.Projects)
	inspectSkills(inspection, profile.SkillGroups)
	inspectHealth(inspection, profile.Health)
	inspectSalary(inspection, profile.Contact.Salary)
	return inspection.faults
}

func inspectSite(inspection *inspection, site Site) {
	inspection.requireWebAddress(baseAddressField, site.BaseURL)
	if strings.HasSuffix(site.BaseURL.Written(), "/") {
		inspection.add(baseAddressField, slashProblem, site.BaseURL.Written())
	}
	inspection.require("site.ssh", site.SSH)
}

func inspectPerson(inspection *inspection, person Person) {
	inspection.require("person.name", person.Name)
	inspection.require("person.title", person.Title)
	if !bareMail(person.Email) {
		inspection.add("person.email", mailProblem, person.Email)
	}
	for index, link := range person.Links {
		inspection.require(fmt.Sprintf("person.links[%d].label", index), link.Label)
		inspection.require(fmt.Sprintf("person.links[%d].text", index), link.Text)
		inspection.requireWebAddress(fmt.Sprintf("person.links[%d].url", index), readWebAddress(link.URL))
	}
}

// inspectExperience holds each role to known months, its end a month or current: a role is dated, never a gap.
func inspectExperience(inspection *inspection, roles []Role) {
	for index, role := range roles {
		inspection.require(fmt.Sprintf("experience[%d].company", index), role.Company)
		inspection.require(fmt.Sprintf("experience[%d].title", index), role.Title)
		start := fmt.Sprintf("experience[%d].start", index)
		if role.Start.Kind() != MonthKnown {
			inspection.add(start, unknownMonth, role.Start.Written())
		}
		end := fmt.Sprintf("experience[%d].end", index)
		switch {
		case role.End.Kind() == MonthCurrent:
		case role.End.Kind() != MonthKnown:
			inspection.add(end, unknownEnd, role.End.Written())
		case role.Start.Kind() == MonthKnown && role.End.Before(role.Start):
			inspection.add(end, orderProblem, role.End.Written(), role.Start.Written())
		}
	}
}

// inspectEducation lets a school's months be gaps, and holds known ones to their order.
func inspectEducation(inspection *inspection, schools []School) {
	for index, school := range schools {
		inspection.require(fmt.Sprintf("education[%d].name", index), school.Name)
		start := fmt.Sprintf("education[%d].start", index)
		end := fmt.Sprintf("education[%d].end", index)
		if !knownOrGap(school.Start) {
			inspection.add(start, unknownMonthOrGap, school.Start.Written())
		}
		if !knownOrGap(school.End) {
			inspection.add(end, unknownMonthOrGap, school.End.Written())
		}
		if school.Start.Kind() == MonthKnown && school.End.Kind() == MonthKnown && school.End.Before(school.Start) {
			inspection.add(end, orderProblem, school.End.Written(), school.Start.Written())
		}
	}
}

// knownOrGap tells a month a school may start or end in.
func knownOrGap(month Month) bool {
	return month.Kind() == MonthKnown || month.Kind() == MonthGap
}

func inspectProjects(inspection *inspection, projects []Project) {
	slug := regexp.MustCompile(slugPattern)
	seen := map[string]bool{}
	for index, project := range projects {
		field := fmt.Sprintf("projects[%d].slug", index)
		switch {
		case !slug.MatchString(project.Slug):
			inspection.add(field, slugProblem, project.Slug)
		case seen[project.Slug]:
			inspection.add(field, sharedSlugProblem, project.Slug)
		}
		seen[project.Slug] = true
		inspection.require(fmt.Sprintf("projects[%d].name", index), project.Name)
		if project.Repo != "" {
			inspection.requireWebAddress(fmt.Sprintf("projects[%d].repo", index), readWebAddress(project.Repo))
		}
		inspectStack(inspection, index, project)
	}
}

// inspectStack holds each name of a project's stack to some text, and to another than the project's language.
func inspectStack(inspection *inspection, projectIndex int, project Project) {
	for stackIndex, name := range project.Stack {
		field := fmt.Sprintf("projects[%d].stack[%d]", projectIndex, stackIndex)
		switch {
		case blank(name):
			inspection.add(field, emptyProblem)
		case name == project.Language:
			inspection.add(field, stackLanguageProblem, name)
		}
	}
}

// inspectSkills holds every group to a title of its own and to a skill at least, and every skill as inspectSkill
// does.
func inspectSkills(inspection *inspection, groups []SkillGroup) {
	seen := map[string]bool{}
	for groupIndex, group := range groups {
		title := fmt.Sprintf("skill_groups[%d].title", groupIndex)
		inspection.require(title, group.Title)
		if seen[group.Title] {
			inspection.add(title, sharedGroupProblem, group.Title)
		}
		seen[group.Title] = true
		if len(group.Skills) == 0 {
			inspection.add(fmt.Sprintf("skill_groups[%d].skills", groupIndex), emptyGroupProblem)
		}
		for skillIndex, skill := range group.Skills {
			inspectSkill(inspection, fmt.Sprintf("skill_groups[%d].skills[%d]", groupIndex, skillIndex), skill)
		}
	}
}

// inspectSkill holds a skill, at its place in the profile, to its name and to a duration it reads, which says a
// skill is still being learnt only of one whose learning says so too.
func inspectSkill(inspection *inspection, place string, skill Skill) {
	inspection.require(place+".name", skill.Name)
	switch {
	case skill.Duration.Kind() == durationUnread:
		inspection.add(place+".duration", unknownDuration, skill.Duration.Written())
	case skill.Duration.Kind() == DurationLearning && !skill.Learning:
		inspection.add(place+".duration", unflaggedLearning, skill.Duration.Written())
	}
}

func inspectHealth(inspection *inspection, sections []HealthSection) {
	for sectionIndex, section := range sections {
		for checkIndex, check := range section.Checks {
			if !check.Level.IsKnown() {
				field := fmt.Sprintf("health[%d].checks[%d].level", sectionIndex, checkIndex)
				inspection.add(field, noneOfProblem, check.Level, kindList(levels()))
			}
			inspection.require(fmt.Sprintf("health[%d].checks[%d].text", sectionIndex, checkIndex), check.Text)
		}
	}
}

// kindList writes out every kind of an enumeration, a check's levels or a post's statuses, in the order of the
// alphabet, the same on every run.
func kindList[Kind ~string](kinds map[Kind]bool) string {
	words := []string{}
	for kind := range kinds {
		words = append(words, string(kind))
	}
	slices.Sort(words)
	return strings.Join(words, kindSeparator)
}

func inspectSalary(inspection *inspection, salary Salary) {
	if salary.Min <= 0 {
		inspection.add("contact.salary.min", positiveProblem, salary.Min)
	}
	if salary.SoftMax <= salary.Min {
		inspection.add("contact.salary.soft_max", rangeProblem, salary.SoftMax, salary.Min)
	}
	if len(salary.Currencies) == 0 {
		inspection.add("contact.salary.currencies", emptyProblem)
	}
}
