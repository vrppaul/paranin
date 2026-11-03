package compose

import (
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// Extras are the pages hidden behind commands, as the first version's reports were: the health report and the list
// of skills. routes gives them no address, so a host hands them only to a face that opens them by a command
// (ADR-0003).
func Extras(snapshot *content.Snapshot) []document.Document {
	return []document.Document{Health(snapshot), Skills(snapshot)}
}

// Skills is the list of every skill, after lazy.nvim's list of plugins, under the site's word for skills: those the
// author uses first, then those they are still learning, as lazy.nvim lists the plugins loaded before the rest,
// each in the profile's order.
func Skills(snapshot *content.Snapshot) document.Document {
	return document.Document{
		Head: document.Head{
			Page:        document.Page{Kind: document.PageSkills, Slug: ""},
			Language:    snapshot.Words.Language(),
			Title:       textSpans(snapshot.Words.Text(content.WordSkills)),
			Subtitle:    nil,
			Place:       "",
			Lead:        nil,
			Ways:        nil,
			Description: nil,
		},
		Blocks: []document.Block{document.Skills{List: skillList(snapshot.Profile.SkillGroups)}},
	}
}

// skillList is every skill of the groups, those the author uses before those they are still learning.
func skillList(groups []content.SkillGroup) []document.Skill {
	used, learning := []document.Skill{}, []document.Skill{}
	for _, group := range groups {
		for _, skill := range group.Skills {
			if skill.Learning {
				learning = append(learning, listedSkill(skill))
			} else {
				used = append(used, listedSkill(skill))
			}
		}
	}
	return append(used, learning...)
}

// Health is the health report, after Neovim's :checkhealth, under the site's word for it: each group of checks the
// profile lists, in its order.
func Health(snapshot *content.Snapshot) document.Document {
	return document.Document{
		Head: document.Head{
			Page:        document.Page{Kind: document.PageHealth, Slug: ""},
			Language:    snapshot.Words.Language(),
			Title:       textSpans(snapshot.Words.Text(content.WordHealth)),
			Subtitle:    nil,
			Place:       "",
			Lead:        nil,
			Ways:        nil,
			Description: nil,
		},
		Blocks: checkGroups(snapshot.Profile.Health),
	}
}

func checkGroups(sections []content.HealthSection) []document.Block {
	levels := checkLevels()
	groups := make([]document.Block, 0, len(sections))
	for _, section := range sections {
		checks := make([]document.Check, 0, len(section.Checks))
		for _, check := range section.Checks {
			checks = append(checks, documentCheck(check, levels))
		}
		groups = append(groups, document.Checks{Title: textSpans(section.Name), List: checks})
	}
	return groups
}

// documentCheck is a check of the profile as the document holds it, at the document's level for its own.
func documentCheck(check content.HealthCheck, levels map[content.Level]document.CheckLevel) document.Check {
	return document.Check{Level: levels[check.Level], Spans: markedSpans(check.Text), Advice: markedSpans(check.Advice)}
}

// checkLevels are the levels of the document a check of the profile reads at. exhaustive holds the literal to
// every level, so a level the profile gains reaches the document too.
func checkLevels() map[content.Level]document.CheckLevel {
	return map[content.Level]document.CheckLevel{
		content.LevelOK:      document.CheckOK,
		content.LevelWarning: document.CheckWarning,
		content.LevelError:   document.CheckError,
	}
}
