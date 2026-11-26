// Package content holds the facts every face draws, and the site's own words beside them: read from a folder on
// disk, checked whole and held as a snapshot that nothing changes after. It knows no face: how a fact is shown is
// the faces' own (docs/ARCHITECTURE.md, ADR-0003). A load reports every fault it finds at once, each at its
// field, so whoever edits the content fixes it in one pass: first every line the decoder cannot read, and once
// the files read, every fault of what they say.
package content

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"path"

	"go.yaml.in/yaml/v3"
)

// The files of the folder: the profile, and the site's words in each language, English, the site's own (D12), and
// each a post may be translated into, in the folder of the catalogues, in a file named by the language's tag
// (wordsFile); a file of YAML ends as yamlEnding says, a catalogue and a series alike
const (
	profileFile = "profile.yaml"
	english     = "en"
	wordsFolder = "i18n"
	yamlEnding  = ".yaml"
)

// wordsFile is the file of the site's words in a language, by its tag.
func wordsFile(language string) string {
	return path.Join(wordsFolder, language+yamlEnding)
}

// Snapshot is the content as one load read it.
type Snapshot struct {
	Profile Profile
	// Own is the content in the site's own language, English, which the résumé is written in and every post first
	Own Edition
	// Translations are the content in each other language the site has words for, in their tags' order
	Translations []Edition
	// Warnings are what the content holds wrong that refuses nothing: each translation older than its original
	Warnings Faults
}

// Edition is the content in one language: the site's words in it, the posts written in it, in the order of their
// slugs, which no edit of a post's fields changes, and the series in it, in the order of their names, as their folder
// lists them.
type Edition struct {
	Words  Words
	Posts  []Post
	Series []Series
}

// Editions are the content in every language, English first.
func (snapshot *Snapshot) Editions() []Edition {
	return append([]Edition{snapshot.Own}, snapshot.Translations...)
}

// TranslatedInto are the languages the site is translated into, by their tags, in their order.
func (snapshot *Snapshot) TranslatedInto() []string {
	tags := []string{}
	for _, edition := range snapshot.Translations {
		tags = append(tags, edition.Words.Language())
	}
	return tags
}

// Post is the post of a slug in the language it is translated into, the original where it names none, if any.
func (snapshot *Snapshot) Post(translation, slug string) (Post, bool) {
	for _, edition := range snapshot.Editions() {
		for _, post := range edition.Posts {
			if post.Slug == slug && post.Translation == translation {
				return post, true
			}
		}
	}
	var none Post
	return none, false
}

// Load reads the content folder and checks it whole: a snapshot, or Faults naming everything wrong in it.
func Load(folder fs.FS) (*Snapshot, error) {
	profile, profileDocument, faults := readFile[Profile](folder, profileFile)
	catalogues, unreadWords, misplacedWords := readCatalogues(folder)
	translated := catalogues.translatedLanguages()
	posts, unreadPosts, misplacedPosts := readPosts(folder, translated)
	series, unreadSeries, misplacedSeries := readSeries(folder, translated)
	faults = append(faults, unreadWords...)
	faults = append(faults, unreadPosts...)
	faults = append(faults, unreadSeries...)
	if len(faults) > 0 {
		return nil, faults
	}
	words, wordFaults := inspectCatalogues(catalogues)
	faults = inspectProfile(profile)
	faults = append(faults, readClips(folder, profile.Projects)...)
	faults = append(faults, inspectCharacters(profileFile, profileDocument)...)
	faults = append(faults, wordFaults...)
	faults = append(faults, misplacedWords...)
	faults = append(faults, misplacedPosts...)
	inspected := inspectPosts(folder, profile.Site.BaseURL, posts)
	faults = append(faults, inspected.faults...)
	faults = append(faults, misplacedSeries...)
	faults = append(faults, inspectSeries(series, posts)...)
	if len(faults) > 0 {
		return nil, faults
	}
	editions := []Edition{}
	for _, language := range translated {
		editions = append(editions, Edition{Words: words[language], Posts: inspected.translations[language], Series: series.seriesIn(language)})
	}
	own := Edition{Words: words[english], Posts: inspected.originals, Series: series.seriesIn(english)}
	return &Snapshot{Profile: profile, Own: own, Translations: editions, Warnings: inspected.warnings}, nil
}

// readFile reads a file of the folder and decodes it strictly, as decodeStrictly does.
func readFile[Target any](folder fs.FS, file string) (Target, *yaml.Node, Faults) {
	data, err := fs.ReadFile(folder, file)
	if err != nil {
		var target Target
		return target, nil, Faults{unreadFault(file, err)}
	}
	return decodeStrictly[Target](file, data)
}

// decodeStrictly decodes a file's YAML strictly: a key a struct lacks is a fault, so a typo is never dropped
// unsaid. A map takes any key, for its reader to check. It hands back the file's document too, every value as the
// file writes it, for the checks no type knows of. A node cannot be decoded strictly, so the YAML is decoded twice.
func decodeStrictly[Target any](file string, data []byte) (Target, *yaml.Node, Faults) {
	var target Target
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	err := decoder.Decode(&target)
	if errors.Is(err, io.EOF) {
		return target, nil, Faults{fileFault(file, emptyProblem)}
	}
	if err != nil {
		return target, nil, decodingFaults(file, err)
	}
	var document yaml.Node
	err = yaml.Unmarshal(data, &document)
	if err != nil {
		return target, nil, Faults{fileFault(file, err.Error())}
	}
	return target, &document, nil
}

// decodingFaults are the faults of a file the decoder could not read: one for each line it named.
func decodingFaults(file string, err error) Faults {
	typeError, isTypeError := errors.AsType[*yaml.TypeError](err)
	if !isTypeError {
		return Faults{fileFault(file, err.Error())}
	}
	faults := make(Faults, 0, len(typeError.Errors))
	for _, problem := range typeError.Errors {
		faults = append(faults, fileFault(file, problem))
	}
	return faults
}

// fileFault is a fault of a file, or a folder, as a whole: at no field and no line.
func fileFault(file, problem string) Fault {
	return Fault{File: file, Field: "", Line: 0, Problem: problem}
}

// lineFault is a fault at a line of a file, outside any inspection: where the front matter does not open or close.
func lineFault(file string, line int, problem string) Fault {
	return Fault{File: file, Field: "", Line: line, Problem: problem}
}

// unreadFault is the fault of a file that could not be read, which names the file already.
func unreadFault(file string, err error) Fault {
	return fileFault(file, causeOf(err).Error())
}

// causeOf is why a file could not be read: where the error names the file's path as well, its cause alone, since
// whoever reports it names the file.
func causeOf(err error) error {
	if pathError, isPathError := errors.AsType[*fs.PathError](err); isPathError {
		return pathError.Err
	}
	return err
}

// readFolder lists a folder of the content that may be left out, as the posts' and the series' are: none listed
// where it is.
func readFolder(folder fs.FS, name string) ([]fs.DirEntry, Faults) {
	entries, err := fs.ReadDir(folder, name)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, Faults{unreadFault(name, err)}
	}
	return entries, nil
}
