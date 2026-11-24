package content

import (
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// The folder of the series, each a file its name names, and what joins its name to the language a translation of it is
// written in, reading.ru.yaml, as a post's text in a language is named
const (
	seriesFolder        = "series"
	seriesLanguageJoint = "."
)

// Series is a run of posts read in order (D25), named by its file, in one language: its title, its line and its parts
// to come as its file in that language writes them, or its original's where it has none there.
type Series struct {
	Name  string `yaml:"-"`
	Title string `yaml:"title"`
	// Line is a line about the series
	Line string `yaml:"line"`
	// Posts are its posts' slugs in their order, written once, so a part put in the middle renumbers nothing
	Posts []string `yaml:"posts"`
	// ToCome are the titles of the parts announced and not yet out, a drafted part's among them until it is out
	ToCome []string `yaml:"to_come"`
}

// seriesText is a series' own words in another language, as its translation's file writes them: its title, its line
// and its parts to come there; its posts are its original's.
type seriesText struct {
	Title  string   `yaml:"title"`
	Line   string   `yaml:"line"`
	ToCome []string `yaml:"to_come"`
}

// writtenSeries is a series as its file writes it: the file, the series, and the file's document, which the check of
// characters walks.
type writtenSeries struct {
	file     string
	series   Series
	document *yaml.Node
}

// writtenSeriesText is a series' translation as its file writes it: the file, the series' name and the language it
// is in, its words, and the file's document.
type writtenSeriesText struct {
	file     string
	name     string
	language string
	text     seriesText
	document *yaml.Node
}

// seriesFiles are the series as their files write them, in their names' order: the originals, and their translations.
type seriesFiles struct {
	originals    []writtenSeries
	translations []writtenSeriesText
}

// readSeries reads every series' file, and each translation of one into a language the site is translated into, in
// their names' order, with the faults that kept a series from being read.
//
// The faults of the files that are no series come apart, as readPosts hands back those of its folders.
func readSeries(folder fs.FS, languages []string) (seriesFiles, Faults, Faults) {
	entries, unreadable := readFolder(folder, seriesFolder)
	written := seriesFiles{originals: nil, translations: nil}
	misplaced := Faults{}
	for _, entry := range entries {
		file := path.Join(seriesFolder, entry.Name())
		stem, isYAML := strings.CutSuffix(entry.Name(), yamlEnding)
		name, language, isTranslation := strings.Cut(stem, seriesLanguageJoint)
		switch {
		case !isYAML:
			misplaced = append(misplaced, fileFault(file, straySeriesProblem))
		case isTranslation && language == english:
			misplaced = append(misplaced, fileFault(file, englishSeriesProblem))
		case isTranslation && !slices.Contains(languages, language):
			misplaced = append(misplaced, fileFault(file, noWordsProblem))
		case isTranslation:
			text, document, textFaults := readFile[seriesText](folder, file)
			written.translations = append(written.translations, writtenSeriesText{file: file, name: name, language: language, text: text, document: document})
			unreadable = append(unreadable, textFaults...)
		default:
			series, document, seriesFaults := readFile[Series](folder, file)
			series.Name = name
			written.originals = append(written.originals, writtenSeries{file: file, series: series, document: document})
			unreadable = append(unreadable, seriesFaults...)
		}
	}
	return written, unreadable, misplaced
}

// inspectSeries checks every series by itself, and that its parts are posts, each in one series alone; then each
// translation, which translates a series of the content.
func inspectSeries(series seriesFiles, posts []writtenFolder) Faults {
	slugs := map[string]bool{}
	for _, written := range posts {
		slugs[written.slug] = true
	}
	seriesOfPost := map[string]string{}
	faults := Faults{}
	names := map[string]bool{}
	for _, written := range series.originals {
		inspection := &inspection{file: written.file, faults: nil}
		inspection.requireName(written.series.Name)
		inspectSeriesText(inspection, seriesText{Title: written.series.Title, Line: written.series.Line, ToCome: written.series.ToCome})
		inspectParts(inspection, written.series, slugs, seriesOfPost)
		faults = slices.Concat(faults, inspection.faults, inspectCharacters(written.file, written.document))
		names[written.series.Name] = true
	}
	for _, written := range series.translations {
		inspection := &inspection{file: written.file, faults: nil}
		inspection.requireName(written.name)
		if !names[written.name] {
			inspection.add("", untranslatedSeriesProblem, path.Join(seriesFolder, written.name+yamlEnding))
		}
		inspectSeriesText(inspection, written.text)
		faults = slices.Concat(faults, inspection.faults, inspectCharacters(written.file, written.document))
	}
	return faults
}

// inspectSeriesText holds a series' words in a language to a title and a line, and each part to come to a title.
func inspectSeriesText(inspection *inspection, text seriesText) {
	inspection.requireLine(titleField, text.Title)
	inspection.requireLine("line", text.Line)
	for index, title := range text.ToCome {
		inspection.requireLine(fmt.Sprintf("to_come[%d]", index), title)
	}
}

// inspectParts holds a series to a post at least, and each of its parts to a post of the content, in this series
// once and in no other. The series each post is a part of gathers in seriesOfPost as the series are inspected.
func inspectParts(inspection *inspection, series Series, slugs map[string]bool, seriesOfPost map[string]string) {
	if len(series.Posts) == 0 {
		inspection.add("posts", emptySeriesProblem)
	}
	for index, slug := range series.Posts {
		field := fmt.Sprintf("posts[%d]", index)
		other, isPart := seriesOfPost[slug]
		switch {
		case !slugs[slug]:
			inspection.add(field, unknownPostProblem, slug)
		case isPart && other == series.Name:
			inspection.add(field, sharedPartProblem, slug)
		case isPart:
			inspection.add(field, otherSeriesProblem, slug, other)
		default:
			seriesOfPost[slug] = series.Name
		}
	}
}

// seriesIn are the series a load read as a language reads them, in their names' order: each with its own words in the
// language where its translation writes them, and its original's where none does; English's own are the originals.
func (written seriesFiles) seriesIn(language string) []Series {
	series := make([]Series, 0, len(written.originals))
	for _, original := range written.originals {
		read := original.series
		for _, translation := range written.translations {
			if translation.name == read.Name && translation.language == language {
				read.Title, read.Line, read.ToCome = translation.text.Title, translation.text.Line, translation.text.ToCome
			}
		}
		series = append(series, read)
	}
	return series
}
