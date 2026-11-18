package content

import (
	"maps"
	"slices"
)

// Word is one of the site's own words, which the faces show beside the facts: the title of a section of the
// résumé or of the index, the greeting before the author's name, the label of a way to reach the author or of a
// detail of a project, the end of what still goes on, how long a skill was used where it was used on the side or is
// still being learnt, the title of the TUI's health report, or the name of the page's theme button.
type Word int

const (
	// The zero word names none
	_ Word = iota
	WordAbout
	WordGreeting
	WordExperience
	WordProjects
	WordSkills
	WordEducation
	WordLanguages
	WordContact
	WordMail
	WordSSH
	WordRepository
	WordInstall
	WordPresent
	WordSide
	WordLearning
	WordAuthor
	WordResume
	WordHealth
	WordDarkTheme
)

// wordKeys are the keys a catalogue writes each word under, each named once. exhaustive holds the literal to
// every Word, so a word added to the constants has its key here too.
func wordKeys() map[Word]string {
	return map[Word]string{
		WordAbout:      "about",
		WordGreeting:   "greeting",
		WordExperience: "experience",
		WordProjects:   "projects",
		WordSkills:     "skills",
		WordEducation:  "education",
		WordLanguages:  "languages",
		WordContact:    "contact",
		WordMail:       "mail",
		WordSSH:        "ssh",
		WordRepository: "repository",
		WordInstall:    "install",
		WordPresent:    "present",
		WordSide:       "side",
		WordLearning:   "learning",
		WordAuthor:     "author",
		WordResume:     "resume",
		WordHealth:     "health",
		WordDarkTheme:  "dark_theme",
	}
}

// Words are the site's words in one language, every one of them there (D12).
type Words struct {
	language string
	texts    map[Word]string
}

// Text is a word as this language writes it.
func (words Words) Text(word Word) string {
	return words.texts[word]
}

// Language is the tag of the language the words are written in, as BCP 47 writes it: "en".
func (words Words) Language() string {
	return words.language
}

// readWords takes a catalogue in a language as the file writes it: every word under its key and none empty, and
// no key the site has no word for. The faults come in the order of the words, then of the keys.
func readWords(language string, written map[string]string) (Words, Faults) {
	inspection := &inspection{file: wordsFile(language), faults: nil}
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
	return Words{language: language, texts: texts}, inspection.faults
}
