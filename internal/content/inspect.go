package content

import (
	"fmt"
	"net/mail"
	"regexp"
	"slices"
	"strings"
)

// How a slug is written: small letters and digits, words joined by hyphens
const slugPattern = `^[a-z0-9]+(-[a-z0-9]+)*$`

// What is said of each kind of fault, as a format for the values it names; a fault of a value written quotes it
const (
	emptyProblem       = "is empty"
	addressProblem     = "%q is no absolute web address"
	slashProblem       = "%q ends in a slash, which every address adds itself"
	mailProblem        = "%q is no bare mail address"
	unknownMonth       = "%q is no month, YYYY-MM"
	unknownEnd         = unknownMonth + ", nor " + currentWord
	unknownMonthOrGap  = unknownMonth + ", nor a " + GapOpening + "gap" + GapClosing
	orderProblem       = "%q comes before the start, %q"
	slugProblem        = "%q is no slug: small letters and digits, joined by hyphens"
	sharedSlugProblem  = "%q names another project too"
	sharedGroupProblem = "%q names another group too"
	emptyGroupProblem  = "holds no skill"
	levelProblem       = "%q is none of %s"
	positiveProblem    = "%d is not above 0"
	rangeProblem       = "%d is not above min, %d"
	unknownWordProblem = "is no word of the site"
)

// What separates the levels a level fault lists
const levelSeparator = ", "

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
	inspection.faults = append(inspection.faults, Fault{File: inspection.file, Field: field, Problem: problem})
}

func (inspection *inspection) require(field, value string) {
	if strings.TrimSpace(value) == "" {
		inspection.add(field, emptyProblem)
	}
}

func (inspection *inspection) requireWebAddress(field string, address WebAddress) {
	if !address.absolute() {
		inspection.add(field, addressProblem, address.Written())
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
	inspection.require("person.given_name", person.GivenName)
	inspection.require("person.title", person.Title)
	if parsed, err := mail.ParseAddress(person.Email); err != nil || parsed.Address != person.Email {
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
	}
}

// inspectSkills holds every group to a title of its own and to a skill at least, and every skill to its name.
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
			inspection.require(fmt.Sprintf("skill_groups[%d].skills[%d].name", groupIndex, skillIndex), skill.Name)
		}
	}
}

func inspectHealth(inspection *inspection, sections []HealthSection) {
	for sectionIndex, section := range sections {
		for checkIndex, check := range section.Checks {
			if !check.Level.IsKnown() {
				field := fmt.Sprintf("health[%d].checks[%d].level", sectionIndex, checkIndex)
				inspection.add(field, levelProblem, check.Level, levelList())
			}
			inspection.require(fmt.Sprintf("health[%d].checks[%d].text", sectionIndex, checkIndex), check.Text)
		}
	}
}

// levelList writes out every level a check may read as, in the order of the alphabet, the same on every run.
func levelList() string {
	words := []string{}
	for level := range levels() {
		words = append(words, string(level))
	}
	slices.Sort(words)
	return strings.Join(words, levelSeparator)
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
