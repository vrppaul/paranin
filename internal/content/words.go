package content

import (
	"maps"
	"slices"
	"strings"
)

// Word is one of the site's own words, which the faces show beside the facts: the title of a section of the
// résumé or of the index, the greeting before the author's name, the label of a way to reach the author or of a
// detail of a project, the end of what still goes on, how long a skill was used where it was used on the side or is
// still being learnt, the title of the TUI's health report, or a word of the page's frame: its link past the top bar
// to the main part, the name of its theme button, and a command's button that copies it: the word on it and the word
// it turns to once it has.
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
	WordSkipToContent
	WordDarkTheme
	WordCopy
	WordCopied
)

// Phrase is one of the site's own phrases, words with a place for one value the site fills in: a command's button's
// name and what the toast says once it has copied, each with the command; and the first screen's button that writes
// to the author, with their given name. A catalogue writes the place as phrasePlace, once, where its language puts
// the value.
type Phrase int

const (
	// The zero phrase names none
	_ Phrase = iota
	PhraseCopyCommand
	PhraseCopiedCommand
	PhraseWriteTo
)

// How a phrase writes the place of its value
const phrasePlace = "{}"

// phraseKeys are the keys a catalogue writes each phrase under, each named once, as wordKeys names the words.
func phraseKeys() map[Phrase]string {
	return map[Phrase]string{
		PhraseCopyCommand:   "copy_command",
		PhraseCopiedCommand: "copied_command",
		PhraseWriteTo:       "write_to",
	}
}

// wordKeys are the keys a catalogue writes each word under, each named once. exhaustive holds the literal to
// every Word, so a word added to the constants has its key here too.
func wordKeys() map[Word]string {
	return map[Word]string{
		WordAbout:         "about",
		WordGreeting:      "greeting",
		WordExperience:    "experience",
		WordProjects:      "projects",
		WordSkills:        "skills",
		WordEducation:     "education",
		WordLanguages:     "languages",
		WordContact:       "contact",
		WordMail:          "mail",
		WordSSH:           "ssh",
		WordRepository:    "repository",
		WordInstall:       "install",
		WordPresent:       "present",
		WordSide:          "side",
		WordLearning:      "learning",
		WordAuthor:        "author",
		WordResume:        "resume",
		WordHealth:        "health",
		WordSkipToContent: "skip_to_content",
		WordDarkTheme:     "dark_theme",
		WordCopy:          "copy",
		WordCopied:        "copied",
	}
}

// Words are the site's words in one language, every one of them there (D12).
type Words struct {
	language string
	texts    map[Word]string
	phrases  map[Phrase]string
}

// Text is a word as this language writes it.
func (words Words) Text(word Word) string {
	return words.texts[word]
}

// Phrase is a phrase as this language writes it, a value in its place.
func (words Words) Phrase(phrase Phrase, value string) string {
	return strings.Replace(words.phrases[phrase], phrasePlace, value, 1)
}

// Language is the tag of the language the words are written in, as BCP 47 writes it: "en".
func (words Words) Language() string {
	return words.language
}

// readWords takes a catalogue in a language as the file writes it: every word and phrase under its key and none
// empty, each phrase with one place for its value, and no key the site has no word or phrase for. The faults come in
// the order of the words, then of the phrases, then of the keys.
func readWords(language string, written map[string]string) (Words, Faults) {
	inspection := &inspection{file: wordsFile(language), faults: nil}
	keys := wordKeys()
	texts := make(map[Word]string, len(keys))
	for _, word := range slices.Sorted(maps.Keys(keys)) {
		texts[word] = written[keys[word]]
		inspection.require(keys[word], texts[word])
	}
	phraseKeyOf := phraseKeys()
	phrases := make(map[Phrase]string, len(phraseKeyOf))
	for _, phrase := range slices.Sorted(maps.Keys(phraseKeyOf)) {
		phrases[phrase] = written[phraseKeyOf[phrase]]
		inspection.requirePhrase(phraseKeyOf[phrase], phrases[phrase])
	}
	known := slices.Concat(slices.Collect(maps.Values(keys)), slices.Collect(maps.Values(phraseKeyOf)))
	for _, key := range slices.Sorted(maps.Keys(written)) {
		if !slices.Contains(known, key) {
			inspection.add(key, unknownWordProblem)
		}
	}
	return Words{language: language, texts: texts, phrases: phrases}, inspection.faults
}

// requirePhrase finds a phrase empty, or holding no place for its value or more than one.
func (inspection *inspection) requirePhrase(field, phrase string) {
	if blank(phrase) {
		inspection.add(field, emptyProblem)
		return
	}
	if strings.Count(phrase, phrasePlace) != 1 {
		inspection.add(field, placeProblem, phrase, phrasePlace)
	}
}
