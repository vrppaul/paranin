package content

import (
	"maps"
	"slices"
	"strings"
)

// Word is one of the site's own words, which the faces show beside the facts: the title of a section of the
// résumé or of the index, the greeting before the author's name, the label of a way to reach the author or of a
// detail of a project, the end of what still goes on, how long a skill was used where it was used on the side or is
// still being learnt, the title of the TUI's health report, or a word of the page's frame: the name of its theme
// button, a command's button that copies it: the word on it and the word it turns to once it has, the first
// screen's button that writes to the author, what each way a terminal reaches the site gives, which its hint says,
// the blog's session, the labels of a post's facts, what a draft's version says it is, what the toast says once a
// code block is copied, and the words of a post's series and the panes at its foot: where a part stands for the
// reader, each side's name, what a side with none says, the panes' heading outside a series, the word before where
// the post is discussed, and the name of its outline; and the blog's list: the line it opens with, its window in the
// foot line, the way to it beside a post's other posts, the word before how it groups its posts, its grouping by year,
// and its group of the posts in no series.
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
	WordCopy
	WordCopied
	WordWriteTo
	WordSSHGives
	WordReadingGives
	WordBlog
	WordVersion
	WordDraft
	WordTags
	WordCopiedCode
	WordSeries
	WordReading
	WordNext
	WordSoon
	WordPrevious
	WordOlder
	WordNewer
	WordFirstPart
	WordLastPart
	WordFirstPost
	WordNewestPost
	WordOtherPosts
	WordDiscuss
	WordContents
	WordBlogLine
	WordListWindow
	WordAllPosts
	WordBy
	WordYears
	WordOutsideSeries
)

// Phrase is one of the site's own phrases, words with a place for each value the site fills in: a command's button's
// name and what the toast says once it has copied, each with the command; a code block's button's name, with the
// block's number; and of a series, a part's place in it, with its number and the count of parts, the panes' heading,
// with the series' title, the pick that lists every part, with their count, and a part's name, with its number; and
// of the blog's list, how many posts a group holds, one or more, how many of a series' parts are out, with that
// count and the count of parts, and the list's name in a grouping, with its title and the grouping's word. A
// catalogue writes each place as phrasePlace, where its language puts the value, the values in the order the site
// hands them.
type Phrase int

const (
	// The zero phrase names none
	_ Phrase = iota
	PhraseCopyCommand
	PhraseCopiedCommand
	PhraseCopyCode
	PhrasePartOf
	PhraseOtherParts
	PhraseAllParts
	PhrasePart
	PhraseOnePost
	PhrasePosts
	PhrasePartsOut
	PhraseGroupedList
)

// How a phrase writes the place of a value
const phrasePlace = "{}"

// The count of values a phrase takes, most take one
const (
	oneValue  = 1
	twoValues = 2
)

// phraseForm is how a catalogue writes a phrase: under its key, with a place for each of its values.
type phraseForm struct {
	key    string
	values int
}

// phraseForms are the forms a catalogue writes each phrase in, each named once, as wordKeys names the words.
func phraseForms() map[Phrase]phraseForm {
	return map[Phrase]phraseForm{
		PhraseCopyCommand:   {key: "copy_command", values: oneValue},
		PhraseCopyCode:      {key: "copy_code", values: oneValue},
		PhraseCopiedCommand: {key: "copied_command", values: oneValue},
		PhrasePartOf:        {key: "part_of", values: twoValues},
		PhraseOtherParts:    {key: "other_parts", values: oneValue},
		PhraseAllParts:      {key: "all_parts", values: oneValue},
		PhrasePart:          {key: "part", values: oneValue},
		PhraseOnePost:       {key: "one_post", values: oneValue},
		PhrasePosts:         {key: "posts_count", values: oneValue},
		PhrasePartsOut:      {key: "parts_out", values: twoValues},
		PhraseGroupedList:   {key: "grouped_list", values: twoValues},
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
		WordDarkTheme:     "dark_theme",
		WordCopy:          "copy",
		WordCopied:        "copied",
		WordWriteTo:       "write_to",
		WordSSHGives:      "ssh_gives",
		WordReadingGives:  "reading_gives",
		WordBlog:          "blog",
		WordVersion:       "version",
		WordDraft:         "draft",
		WordTags:          "tags",
		WordCopiedCode:    "copied_code",
		WordSeries:        "series",
		WordReading:       "reading",
		WordNext:          "next",
		WordSoon:          "soon",
		WordPrevious:      "previous",
		WordOlder:         "older",
		WordNewer:         "newer",
		WordFirstPart:     "first_part",
		WordLastPart:      "last_part",
		WordFirstPost:     "first_post",
		WordNewestPost:    "newest_post",
		WordOtherPosts:    "other_posts",
		WordDiscuss:       "discuss",
		WordContents:      "contents",
		WordBlogLine:      "blog_line",
		WordListWindow:    "list_window",
		WordAllPosts:      "all_posts",
		WordBy:            "by",
		WordYears:         "years",
		WordOutsideSeries: "outside_series",
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

// Phrase is a phrase as this language writes it, each value in its place, in their order; a place written in a value
// stays the value's.
func (words Words) Phrase(phrase Phrase, values ...string) string {
	pieces := strings.Split(words.phrases[phrase], phrasePlace)
	var builder strings.Builder
	for index, piece := range pieces {
		builder.WriteString(piece)
		if index < len(values) && index < len(pieces)-1 {
			builder.WriteString(values[index])
		}
	}
	return builder.String()
}

// Language is the tag of the language the words are written in, as BCP 47 writes it: "en".
func (words Words) Language() string {
	return words.language
}

// readWords takes a catalogue in a language as the file writes it: every word and phrase under its key and none
// empty, each phrase with a place for each of its values, and no key the site has no word or phrase for. The faults
// come in the order of the words, then of the phrases, then of the keys.
func readWords(language string, written map[string]string) (Words, Faults) {
	inspection := &inspection{file: wordsFile(language), faults: nil}
	keys := wordKeys()
	texts := make(map[Word]string, len(keys))
	for _, word := range slices.Sorted(maps.Keys(keys)) {
		texts[word] = written[keys[word]]
		inspection.require(keys[word], texts[word])
	}
	forms := phraseForms()
	phrases := make(map[Phrase]string, len(forms))
	known := slices.Collect(maps.Values(keys))
	for _, phrase := range slices.Sorted(maps.Keys(forms)) {
		form := forms[phrase]
		phrases[phrase] = written[form.key]
		inspection.requirePhrase(form, phrases[phrase])
		known = append(known, form.key)
	}
	for _, key := range slices.Sorted(maps.Keys(written)) {
		if !slices.Contains(known, key) {
			inspection.add(key, unknownWordProblem)
		}
	}
	return Words{language: language, texts: texts, phrases: phrases}, inspection.faults
}

// requirePhrase finds a phrase empty, or holding a count of places other than the count of values its form takes.
func (inspection *inspection) requirePhrase(form phraseForm, phrase string) {
	if blank(phrase) {
		inspection.add(form.key, emptyProblem)
		return
	}
	if places := strings.Count(phrase, phrasePlace); places != form.values {
		inspection.add(form.key, placeProblem, phrase, phrasePlace, places, form.values)
	}
}
