package document

import (
	"iter"
	"slices"
)

// Briefed is a document's sections read for the brief a résumé opens with: the brief, the first facts the first
// section holds; that section's title, and the blocks it holds before the brief and after it; the sections after it;
// the address, the first way to reach the author they give, the mail; and the roles they hold, in order, which
// CareerOf lays on a line.
// Its lists share the sections' arrays, each clipped to its length, so a face that appends to one writes no section.
type Briefed struct {
	Brief   Facts
	Title   []Span
	Before  []Block
	After   []Block
	Rest    []Section
	Address []Span
	Roles   []Entry
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
				Brief:   facts,
				Title:   first.Title,
				Before:  slices.Clip(first.Blocks[:index]),
				After:   slices.Clip(first.Blocks[index+1:]),
				Rest:    slices.Clip(sections[1:]),
				Address: addressOf(sections),
				Roles:   rolesOf(sections),
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

// rolesOf is the roles sections hold, the entries whose kind is a role, in order.
func rolesOf(sections []Section) []Entry {
	roles := []Entry{}
	for block := range sectionBlocks(sections) {
		if entry, isEntry := block.(Entry); isEntry && entry.Kind == EntryRole {
			roles = append(roles, entry)
		}
	}
	return roles
}

// addressOf is the first way to reach the author sections hold, the mail, as its value: the first fact of the first
// facts of the ways' kind; none where no section holds one.
func addressOf(sections []Section) []Span {
	for block := range sectionBlocks(sections) {
		if facts, isFacts := block.(Facts); isFacts && facts.Kind == FactsWays && len(facts.List) > 0 {
			return facts.List[0].Value
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
