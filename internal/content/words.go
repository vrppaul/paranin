package content

import (
	"maps"
	"slices"
)

// Word is one of the site's own words, which the faces show beside the facts: the title of a section of the
// résumé, or the label of a way to reach the author or of a detail of a project.
type Word int

const (
	// The zero word names none
	_ Word = iota
	WordFacts
	WordExperience
	WordProjects
	WordSkills
	WordEducation
	WordLanguages
	WordMail
	WordSSH
	WordRepository
	WordInstall
)

// wordKeys are the keys a catalogue writes each word under, each named once. exhaustive holds the literal to
// every Word, so a word added to the constants has its key here too.
func wordKeys() map[Word]string {
	return map[Word]string{
		WordFacts:      "facts",
		WordExperience: "experience",
		WordProjects:   "projects",
		WordSkills:     "skills",
		WordEducation:  "education",
		WordLanguages:  "languages",
		WordMail:       "mail",
		WordSSH:        "ssh",
		WordRepository: "repository",
		WordInstall:    "install",
	}
}

// Words are the site's words in one language, every one of them there (D12).
type Words struct {
	texts map[Word]string
}

// Text is a word as this language writes it.
func (words Words) Text(word Word) string {
	return words.texts[word]
}

// readWords takes a catalogue as the file writes it: every word under its key and none empty, and no key the
// site has no word for. The faults come in the order of the words, then of the keys.
func readWords(written map[string]string) (Words, Faults) {
	inspection := &inspection{file: englishWordsFile, faults: nil}
	keys := wordKeys()
	texts := make(map[Word]string, len(keys))
	for _, word := range slices.Sorted(maps.Keys(keys)) {
		texts[word] = written[keys[word]]
		inspection.require(keys[word], texts[word])
	}
	known := slices.Collect(maps.Values(keys))
	for _, key := range slices.Sorted(maps.Keys(written)) {
		if !slices.Contains(known, key) {
			inspection.add(key, unknownWordProblem)
		}
	}
	return Words{texts: texts}, inspection.faults
}
