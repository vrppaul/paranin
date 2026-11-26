package document

import (
	"iter"
	"slices"
)

// Briefed is a document's sections read for the brief a résumé opens with: the brief, the first facts the first
// section holds; that section's title, and the blocks it holds before the brief and after it, but the ways to reach
// the author, which each face places as it chooses; the sections after it; those ways, the mail first; and the roles
// the sections hold, in order, which CareerOf lays on a line.
// Its lists are their own or share the sections' arrays clipped to their length, so a face that appends to one writes
// no section.
type Briefed struct {
	Brief  Facts
	Title  []Span
	Before []Block
	After  []Block
	Rest   []Section
	Ways   []Fact
	Roles  []Entry
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

// BriefOf is sections read for their brief; nothing where the first holds no facts.
func BriefOf(sections []Section) *Briefed {
	if len(sections) == 0 {
		return nil
	}
	first := sections[0]
	for index, block := range first.Blocks {
		if facts, isFacts := block.(Facts); isFacts && len(facts.List) > 0 {
			return &Briefed{
				Brief:  facts,
				Title:  first.Title,
				Before: slices.DeleteFunc(slices.Clone(first.Blocks[:index]), isWays),
				After:  slices.DeleteFunc(slices.Clone(first.Blocks[index+1:]), isWays),
				Rest:   slices.Clip(sections[1:]),
				Ways:   waysOf(sections),
				Roles:  rolesOf(sections),
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

// isWays is whether a block is the ways to reach the author.
func isWays(block Block) bool {
	facts, isFacts := block.(Facts)
	return isFacts && facts.Kind == FactsWays
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
