package content

import (
	"io/fs"
	"maps"
	"path"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
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
// foot line, the way to it beside a post's other posts, the word before how it groups its posts, its grouping by year
// and by series, and its group of the posts in no series; and a post's versions: the title of the window that lists
// them, and the words that mark the current one and the one read; and the language: the label of a post's language,
// the language's own name for itself, which a post's Switch shows, and the locale Open Graph marks a page in; the
// manual's own words, the section of the manual its pages stand in and the names of the sections each opens with and
// of the notes it ends with; and what a translation older than its original says of it.
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
	WordVersions
	WordCurrent
	WordVersionRead
	WordLanguage
	WordLanguageName
	WordLocale
	WordSeriesGrouping
	WordManualSection
	WordManualName
	WordManualDescription
	WordManualNotes
	WordOutdated
)

// Phrase is one of the site's own phrases, words with a place for each value the site fills in: a command's button's
// name and what the toast says once it has copied, each with the command; a code block's button's name, with the
// block's number; and of a series, a part's place in it, with its number and the count of parts, the panes' heading,
// with the series' title, and a part's name, with its number; and of the blog's list, the list's name in a grouping,
// with its title and the grouping's word, and the filter of a list filtered by a tag, with the tag, and that filter as
// the way back to the whole list, with its words; what an older version of a post says of the newer, with that
// version's day; and the title of the blog's feed, with the author's name, and of a series', with the series' title
// and the blog's feed's. A catalogue writes each place as phrasePlace, where its language puts the value, the values
// in the order the site hands them.
type Phrase int

const (
	// The zero phrase names none
	_ Phrase = iota
	PhraseCopyCommand
	PhraseCopiedCommand
	PhraseCopyCode
	PhrasePartOf
	PhraseOtherParts
	PhrasePart
	PhraseGroupedList
	PhraseTagFilter
	PhraseFilterWayOut
	PhraseNewerVersion
	PhraseBlogFeed
	PhraseSeriesFeed
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
		PhrasePart:          {key: "part", values: oneValue},
		PhraseGroupedList:   {key: "grouped_list", values: twoValues},
		PhraseTagFilter:     {key: "tag_filter", values: oneValue},
		PhraseFilterWayOut:  {key: "filter_way_out", values: oneValue},
		PhraseNewerVersion:  {key: "newer_version", values: oneValue},
		PhraseBlogFeed:      {key: "blog_feed", values: oneValue},
		PhraseSeriesFeed:    {key: "series_feed", values: twoValues},
	}
}

// wordKeys are the keys a catalogue writes each word under, each named once. exhaustive holds the literal to
// every Word, so a word added to the constants has its key here too.
func wordKeys() map[Word]string {
	return map[Word]string{
		WordAbout:             "about",
		WordGreeting:          "greeting",
		WordExperience:        "experience",
		WordProjects:          "projects",
		WordSkills:            "skills",
		WordEducation:         "education",
		WordLanguages:         "languages",
		WordContact:           "contact",
		WordMail:              "mail",
		WordSSH:               "ssh",
		WordRepository:        "repository",
		WordInstall:           "install",
		WordPresent:           "present",
		WordSide:              "side",
		WordLearning:          "learning",
		WordAuthor:            "author",
		WordResume:            "resume",
		WordHealth:            "health",
		WordDarkTheme:         "dark_theme",
		WordCopy:              "copy",
		WordCopied:            "copied",
		WordWriteTo:           "write_to",
		WordSSHGives:          "ssh_gives",
		WordReadingGives:      "reading_gives",
		WordBlog:              "blog",
		WordVersion:           "version",
		WordDraft:             "draft",
		WordTags:              "tags",
		WordCopiedCode:        "copied_code",
		WordSeries:            "series",
		WordReading:           "reading",
		WordNext:              "next",
		WordSoon:              "soon",
		WordPrevious:          "previous",
		WordOlder:             "older",
		WordNewer:             "newer",
		WordFirstPart:         "first_part",
		WordLastPart:          "last_part",
		WordFirstPost:         "first_post",
		WordNewestPost:        "newest_post",
		WordOtherPosts:        "other_posts",
		WordDiscuss:           "discuss",
		WordContents:          "contents",
		WordBlogLine:          "blog_line",
		WordListWindow:        "list_window",
		WordAllPosts:          "all_posts",
		WordBy:                "by",
		WordYears:             "years",
		WordOutsideSeries:     "outside_series",
		WordVersions:          "versions",
		WordCurrent:           "current",
		WordVersionRead:       "version_read",
		WordLanguage:          "language",
		WordLanguageName:      "language_name",
		WordLocale:            "locale",
		WordSeriesGrouping:    "series_grouping",
		WordManualSection:     "manual_section",
		WordManualName:        "manual_name",
		WordManualDescription: "manual_description",
		WordManualNotes:       "manual_notes",
		WordOutdated:          "outdated",
	}
}

// Words are the site's words in one language, every one of them there (D12): its tag, and the language as CLDR knows
// it, whose plural rule picks a count's form; its words and phrases, its counts in each form a whole number takes in
// it, and the names of the months as a day writes them.
type Words struct {
	language string
	plurals  language.Tag
	texts    map[Word]string
	phrases  map[Phrase]string
	counts   map[Count]map[plural.Form]string
	months   []string
}

// Text is a word as this language writes it.
func (words Words) Text(word Word) string {
	return words.texts[word]
}

// Phrase is a phrase as this language writes it, each value in its place, in their order; a place written in a value
// stays the value's.
func (words Words) Phrase(phrase Phrase, values ...string) string {
	return filled(words.phrases[phrase], values)
}

// filled is a phrase as a catalogue writes it with each value in its place, in their order.
func filled(written string, values []string) string {
	pieces := strings.Split(written, phrasePlace)
	var builder strings.Builder
	for index, piece := range pieces {
		builder.WriteString(piece)
		if index < len(values) && index < len(pieces)-1 {
			builder.WriteString(values[index])
		}
	}
	return builder.String()
}

// Around are the words a phrase of one value sets before its place and after it, for a value a face writes its own
// way, as a day.
func (words Words) Around(phrase Phrase) (before, after string) {
	before, after, _ = strings.Cut(words.phrases[phrase], phrasePlace)
	return before, after
}

// Language is the tag of the language the words are written in, as BCP 47 writes it: "en".
func (words Words) Language() string {
	return words.language
}

// Translation is the language a page in these words is translated into: their tag, none for English, the site's own.
func (words Words) Translation() string {
	if words.language == english {
		return ""
	}
	return words.language
}

// How a language's tag is written, as ISO 639 writes a language: two or three small letters
const languagePattern = `^[a-z]{2,3}$`

// writtenWords are a catalogue's words as its file writes them, each by its key: a word or a phrase a text, a count a
// text for each plural form, and the months a list.
type writtenWords map[string]yaml.Node

// writtenCatalogue is a catalogue as its file writes it: its language, its words by their keys, and its document.
type writtenCatalogue struct {
	language string
	written  writtenWords
	document *yaml.Node
}

// catalogues are the site's words as their files write them: in English, its own language, and in each other.
type catalogues struct {
	own          writtenCatalogue
	translations []writtenCatalogue
}

// all are the catalogues in every language, English first.
func (written catalogues) all() []writtenCatalogue {
	return append([]writtenCatalogue{written.own}, written.translations...)
}

// readCatalogues reads the site's words in English, then in each other language, with their faults and the strays'.
func readCatalogues(folder fs.FS) (catalogues, Faults, Faults) {
	written, document, unreadable := readFile[writtenWords](folder, wordsFile(english))
	read := catalogues{own: writtenCatalogue{language: english, written: written, document: document}, translations: nil}
	entries, unreadFolder := readFolder(folder, wordsFolder)
	unreadable = append(unreadable, unreadFolder...)
	languageForm := regexp.MustCompile(languagePattern)
	misplaced := Faults{}
	for _, entry := range entries {
		language, isYAML := strings.CutSuffix(entry.Name(), yamlEnding)
		switch {
		case entry.IsDir() || !isYAML || !languageForm.MatchString(language):
			misplaced = append(misplaced, fileFault(path.Join(wordsFolder, entry.Name()), strayWordsProblem))
		case language != english:
			written, document, faults := readFile[writtenWords](folder, wordsFile(language))
			read.translations = append(read.translations, writtenCatalogue{language: language, written: written, document: document})
			unreadable = append(unreadable, faults...)
		}
	}
	return read, unreadable, misplaced
}

// translatedLanguages are the languages of the catalogues but English: those a post may be translated into.
func (written catalogues) translatedLanguages() []string {
	languages := []string{}
	for _, catalogue := range written.translations {
		languages = append(languages, catalogue.language)
	}
	return languages
}

// inspectCatalogues takes each catalogue as readWords does: the words by their language, and the faults found.
func inspectCatalogues(written catalogues) (map[string]Words, Faults) {
	words := map[string]Words{}
	faults := Faults{}
	for _, catalogue := range written.all() {
		read, wordFaults := readWords(catalogue.language, catalogue.written)
		words[catalogue.language] = read
		faults = slices.Concat(faults, wordFaults, inspectCharacters(wordsFile(catalogue.language), catalogue.document))
	}
	return words, faults
}

// readWords takes a catalogue in a language as the file writes it: every word and phrase under its key, a text and none
// empty, each phrase with a place for each of its values; every count, in each form the language gives a whole number
// alone; the months; and no key the site has no word for. The faults come in the order of the words, then of the
// phrases, the counts, the months, then of the keys.
func readWords(language string, written writtenWords) (Words, Faults) {
	inspection := &inspection{file: wordsFile(language), faults: nil}
	plurals, isKnown := pluralRuleOf(language)
	if !isKnown {
		inspection.add("", unknownLanguage, language)
	}
	keys := wordKeys()
	texts := make(map[Word]string, len(keys))
	for _, word := range slices.Sorted(maps.Keys(keys)) {
		texts[word] = inspection.text(written, keys[word])
	}
	forms := phraseForms()
	phrases := make(map[Phrase]string, len(forms))
	known := append(slices.Collect(maps.Values(keys)), monthsKey)
	for _, phrase := range slices.Sorted(maps.Keys(forms)) {
		form := forms[phrase]
		phrases[phrase] = inspection.text(written, form.key)
		inspection.requirePlaces(form.key, form.values, phrases[phrase])
		known = append(known, form.key)
	}
	whole := wholeForms(plurals)
	counted := countForms()
	counts := make(map[Count]map[plural.Form]string, len(counted))
	for _, count := range slices.Sorted(maps.Keys(counted)) {
		counts[count] = inspection.count(written, counted[count], whole)
		known = append(known, counted[count].key)
	}
	months := inspection.months(written)
	for _, key := range slices.Sorted(maps.Keys(written)) {
		if !slices.Contains(known, key) {
			inspection.add(key, unknownWordProblem)
		}
	}
	return Words{language: language, plurals: plurals, texts: texts, phrases: phrases, counts: counts, months: months}, inspection.faults
}

// text is the text a catalogue writes under a key, a fault where it writes none, or something else there.
func (inspection *inspection) text(written writtenWords, key string) string {
	node, isWritten := written[key]
	var text string
	switch {
	case !isWritten:
		inspection.add(key, emptyProblem)
	case node.Decode(&text) != nil:
		inspection.add(key, textProblem)
	case blank(text):
		inspection.add(key, emptyProblem)
	}
	return text
}

// requirePlaces finds a phrase at a field that holds a count of places other than the count of values it takes; an
// empty one is a fault already.
func (inspection *inspection) requirePlaces(field string, values int, phrase string) {
	if places := strings.Count(phrase, phrasePlace); !blank(phrase) && places != values {
		inspection.add(field, placeProblem, phrase, phrasePlace, places, values)
	}
}
