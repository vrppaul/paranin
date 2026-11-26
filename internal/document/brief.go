package document

import (
	"iter"
	"slices"
)

// The most facts a brief holds (DESIGN.md, Facts): the first the first section's details list, in the order the
// profile writes them, which is the author's choice
const briefFacts = 3

// Briefed is a document's sections read for the brief a résumé opens with: the brief, the first three of the first
// details the first section holds; that section's title; the sections after it; the ways to reach the author, the mail
// first, which each face places as it chooses; and the roles the sections hold, in order, which CareerOf lays on a
// line. What else the first section holds stands on no tab, the brief standing for it.
// Its lists share the sections' arrays clipped to their length, so a face that appends to one writes no section.
type Briefed struct {
	Brief Facts
	Title []Span
	Rest  []Section
	Ways  []Fact
	Roles []Entry
}

// Address is the first way to reach the author, the mail, as its value; none where the sections give no way.
func (briefed *Briefed) Address() []Span {
	if len(briefed.Ways) == 0 {
		return nil
	}
	return briefed.Ways[0].Value
}

// OtherWays are the ways to reach the author after the first, the mail, which a face that heads the brief with the
// address places apart; none where the sections give no other.
func (briefed *Briefed) OtherWays() []Fact {
	if len(briefed.Ways) == 0 {
		return nil
	}
	return briefed.Ways[1:]
}

// BriefOf is sections read for their brief, the first details the first section holds; nothing where it holds none,
// so the ways to reach the author never stand in for it.
func BriefOf(sections []Section) *Briefed {
	if len(sections) == 0 {
		return nil
	}
	first := sections[0]
	for _, block := range first.Blocks {
		if facts, isFacts := block.(Facts); isFacts && facts.Kind == FactsDetails && len(facts.List) > 0 {
			return &Briefed{
				Brief: Facts{Kind: facts.Kind, List: slices.Clip(facts.List[:min(briefFacts, len(facts.List))])},
				Title: first.Title,
				Rest:  slices.Clip(sections[1:]),
				Ways:  waysOf(sections),
				Roles: rolesOf(sections),
			}
		}
	}
	return nil
}

// SectionsOf parts a document's blocks into those that are not sections and its sections, each in their order.
func SectionsOf(blocks []Block) ([]Block, []Section) {
	others, sections := []Block{}, []Section{}
	for _, block := range blocks {
		if section, isSection := block.(Section); isSection {
			sections = append(sections, section)
		} else {
			others = append(others, block)
		}
	}
	return others, sections
}

// rolesOf is the roles sections hold, their entries, in order.
func rolesOf(sections []Section) []Entry {
	roles := []Entry{}
	for block := range sectionBlocks(sections) {
		if entry, isEntry := block.(Entry); isEntry {
			roles = append(roles, entry)
		}
	}
	return roles
}

// waysOf are the ways to reach the author sections hold: the first facts of the ways' kind, the mail first; none
// where no section holds them.
func waysOf(sections []Section) []Fact {
	for block := range sectionBlocks(sections) {
		if facts, isFacts := block.(Facts); isFacts && facts.Kind == FactsWays && len(facts.List) > 0 {
			return slices.Clip(facts.List)
		}
	}
	return nil
}

// sectionBlocks are the blocks each section holds, in order, not those of the sections nested in them.
func sectionBlocks(sections []Section) iter.Seq[Block] {
	return func(yield func(Block) bool) {
		for _, section := range sections {
			for _, block := range section.Blocks {
				if !yield(block) {
					return
				}
			}
		}
	}
}
