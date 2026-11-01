package content

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
)

// How a slug is written: small letters and digits, words joined by hyphens
const slugPattern = `^[a-z0-9]+(-[a-z0-9]+)*$`

// The schemes an address of the profile may have
const (
	secureScheme = "https"
	plainScheme  = "http"
)

// What is said of each kind of fault
const (
	emptyProblem      = "is empty"
	addressProblem    = "is no absolute web address"
	unknownMonth      = "%q is no month, YYYY-MM"
	unknownEnd        = "%q is no month, YYYY-MM, nor current"
	unknownMonthOrGap = "%q is no month, YYYY-MM, nor a [gap]"
	orderProblem      = "comes before the start"
)

// The field of the site's address, which two checks read
const baseAddressField = "site.base_url"

// inspection gathers the faults of one profile, each at its field, so every one is reported at once.
type inspection struct {
	faults Faults
}

func (inspection *inspection) add(field, problem string) {
	inspection.faults = append(inspection.faults, Fault{File: profileFile, Field: field, Problem: problem})
}

func (inspection *inspection) require(field, value string) {
	if strings.TrimSpace(value) == "" {
		inspection.add(field, emptyProblem)
	}
}

func (inspection *inspection) requireAddress(field, address string) {
	parsed, err := url.Parse(address)
	if err != nil || parsed.Host == "" || (parsed.Scheme != secureScheme && parsed.Scheme != plainScheme) {
		inspection.add(field, addressProblem)
	}
}

// inspectProfile checks what the schema cannot say: what must be there, how it is written, and how its parts
// agree with each other.
func inspectProfile(profile Profile) Faults {
	inspection := &inspection{faults: nil}
	inspectSite(inspection, profile.Site)
	inspectPerson(inspection, profile.Person)
	inspectExperience(inspection, profile.Experience)
	inspectEducation(inspection, profile.Education)
	inspectProjects(inspection, profile.Projects)
	inspectSkills(inspection, profile.SkillGroups, profile.Skills)
	inspectHealth(inspection, profile.Health)
	inspectSalary(inspection, profile.Contact.Salary)
	return inspection.faults
}

func inspectSite(inspection *inspection, site Site) {
	inspection.requireAddress(baseAddressField, site.BaseURL)
	if strings.HasSuffix(site.BaseURL, "/") {
		inspection.add(baseAddressField, "ends in a slash, which every address adds itself")
	}
	inspection.require("site.ssh", site.SSH)
}

func inspectPerson(inspection *inspection, person Person) {
	inspection.require("person.name", person.Name)
	inspection.require("person.given_name", person.GivenName)
	inspection.require("person.title", person.Title)
	if parsed, err := mail.ParseAddress(person.Email); err != nil || parsed.Address != person.Email {
		inspection.add("person.email", fmt.Sprintf("%q is no bare mail address", person.Email))
	}
	for index, link := range person.Links {
		inspection.require(fmt.Sprintf("person.links[%d].key", index), link.Key)
		inspection.require(fmt.Sprintf("person.links[%d].label", index), link.Label)
		inspection.require(fmt.Sprintf("person.links[%d].text", index), link.Text)
		inspection.requireAddress(fmt.Sprintf("person.links[%d].url", index), link.URL)
	}
}

// inspectExperience holds each role to known months, its end a month or current: a role is dated, never a gap.
func inspectExperience(inspection *inspection, roles []Role) {
	for index, role := range roles {
		inspection.require(fmt.Sprintf("experience[%d].company", index), role.Company)
		inspection.require(fmt.Sprintf("experience[%d].title", index), role.Title)
		start := fmt.Sprintf("experience[%d].start", index)
		if !role.Start.IsKnown() {
			inspection.add(start, fmt.Sprintf(unknownMonth, role.Start.Written()))
		}
		end := fmt.Sprintf("experience[%d].end", index)
		switch {
		case role.Current():
		case !role.End.IsKnown():
			inspection.add(end, fmt.Sprintf(unknownEnd, role.End.Written()))
		case role.Start.IsKnown() && role.End.Before(role.Start):
			inspection.add(end, orderProblem)
		}
	}
}

// inspectEducation lets a school's months be gaps, and holds known ones to their order.
func inspectEducation(inspection *inspection, schools []School) {
	for index, school := range schools {
		inspection.require(fmt.Sprintf("education[%d].name", index), school.Name)
		start := fmt.Sprintf("education[%d].start", index)
		end := fmt.Sprintf("education[%d].end", index)
		if !school.Start.IsKnown() && !school.Start.IsGap() {
			inspection.add(start, fmt.Sprintf(unknownMonthOrGap, school.Start.Written()))
		}
		if !school.End.IsKnown() && !school.End.IsGap() {
			inspection.add(end, fmt.Sprintf(unknownMonthOrGap, school.End.Written()))
		}
		if school.Start.IsKnown() && school.End.IsKnown() && school.End.Before(school.Start) {
			inspection.add(end, orderProblem)
		}
	}
}

func inspectProjects(inspection *inspection, projects []Project) {
	slug := regexp.MustCompile(slugPattern)
	seen := map[string]bool{}
	for index, project := range projects {
		field := fmt.Sprintf("projects[%d].slug", index)
		switch {
		case !slug.MatchString(project.Slug):
			inspection.add(field, "is no slug: small letters and digits, joined by hyphens")
		case seen[project.Slug]:
			inspection.add(field, "names another project too")
		}
		seen[project.Slug] = true
		inspection.require(fmt.Sprintf("projects[%d].name", index), project.Name)
		if project.Repo != "" {
			inspection.requireAddress(fmt.Sprintf("projects[%d].repo", index), project.Repo)
		}
	}
}

// inspectSkills holds every skill to a group the profile lists, and every group to a key of its own.
func inspectSkills(inspection *inspection, groups []SkillGroup, skills []Skill) {
	keys := map[string]bool{}
	for index, group := range groups {
		field := fmt.Sprintf("skill_groups[%d].key", index)
		inspection.require(field, group.Key)
		if keys[group.Key] {
			inspection.add(field, "names another group too")
		}
		keys[group.Key] = true
		inspection.require(fmt.Sprintf("skill_groups[%d].title", index), group.Title)
	}
	for index, skill := range skills {
		inspection.require(fmt.Sprintf("skills[%d].name", index), skill.Name)
		if !keys[skill.Group] {
			inspection.add(fmt.Sprintf("skills[%d].group", index), "is none of skill_groups")
		}
	}
}

func inspectHealth(inspection *inspection, sections []HealthSection) {
	for sectionIndex, section := range sections {
		for checkIndex, check := range section.Checks {
			if !check.Level.IsKnown() {
				inspection.add(fmt.Sprintf("health[%d].checks[%d].level", sectionIndex, checkIndex), "is none of ok, warn, error")
			}
			inspection.require(fmt.Sprintf("health[%d].checks[%d].text", sectionIndex, checkIndex), check.Text)
		}
	}
}

func inspectSalary(inspection *inspection, salary Salary) {
	if salary.Min <= 0 || salary.Step <= 0 || salary.SoftMax <= salary.Min || salary.BigStep < salary.Step ||
		len(salary.Currencies) == 0 {
		inspection.add("contact.salary", "needs min above 0, soft_max above min, step above 0, big_step of a step or more, and a currency")
	}
}
