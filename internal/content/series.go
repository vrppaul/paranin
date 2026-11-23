package content

import (
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// The folder of the series, each a file its name names
const seriesFolder = "series"

// Series is a run of posts read in order (D25), named by its file.
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

// writtenSeries is a series as its file writes it: the file, the series, and the file's document, which the check of
// characters walks.
type writtenSeries struct {
	file     string
	series   Series
	document *yaml.Node
}

// readSeries reads every series' file, in their names' order, with the faults that kept a series from being read.
//
// The faults of the files that are no series come apart, as readPosts hands back those of its folders.
func readSeries(folder fs.FS) ([]writtenSeries, Faults, Faults) {
	entries, unreadable := readFolder(folder, seriesFolder)
	written := make([]writtenSeries, 0, len(entries))
	misplaced := Faults{}
	for _, entry := range entries {
		file := path.Join(seriesFolder, entry.Name())
		name, isYAML := strings.CutSuffix(entry.Name(), yamlEnding)
		if !isYAML {
			misplaced = append(misplaced, fileFault(file, straySeriesProblem))
			continue
		}
		series, document, seriesFaults := readFile[Series](folder, file)
		series.Name = name
		written = append(written, writtenSeries{file: file, series: series, document: document})
		unreadable = append(unreadable, seriesFaults...)
	}
	return written, unreadable, misplaced
}

// inspectSeries checks every series by itself, and that its parts are posts, each in one series alone.
func inspectSeries(series []writtenSeries, posts []writtenPost) Faults {
	slugs := map[string]bool{}
	for _, written := range posts {
		slugs[written.post.Slug] = true
	}
	seriesOfPost := map[string]string{}
	faults := Faults{}
	for _, written := range series {
		inspection := &inspection{file: written.file, faults: nil}
		inspection.requireName(written.series.Name)
		inspection.requireLine(titleField, written.series.Title)
		inspection.requireLine("line", written.series.Line)
		inspectParts(inspection, written.series, slugs, seriesOfPost)
		for index, title := range written.series.ToCome {
			inspection.requireLine(fmt.Sprintf("to_come[%d]", index), title)
		}
		faults = slices.Concat(faults, inspection.faults, inspectCharacters(written.file, written.document))
	}
	return faults
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

// seriesOf are the series a load read, as the snapshot keeps them.
func seriesOf(written []writtenSeries) []Series {
	series := make([]Series, 0, len(written))
	for _, entry := range written {
		series = append(series, entry.series)
	}
	return series
}
