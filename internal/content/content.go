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

	"go.yaml.in/yaml/v3"
)

// The files of the folder: the profile, and the site's words in English, its one language so far (D12), in the
// folder of the catalogues, in a file named by the language's tag (wordsFile)
const (
	profileFile = "profile.yaml"
	english     = "en"
	wordsFolder = "i18n/"
	wordsEnding = ".yaml"
)

// wordsFile is the file of the site's words in a language, by its tag.
func wordsFile(language string) string {
	return wordsFolder + language + wordsEnding
}

// Snapshot is the content as one load read it.
type Snapshot struct {
	Profile Profile
	Words   Words
}

// Load reads the content folder and checks it whole: a snapshot, or Faults naming everything wrong in it.
func Load(folder fs.FS) (*Snapshot, error) {
	wordsPath := wordsFile(english)
	profile, profileDocument, faults := readFile[Profile](folder, profileFile)
	written, wordsDocument, wordFaults := readFile[map[string]string](folder, wordsPath)
	faults = append(faults, wordFaults...)
	if len(faults) > 0 {
		return nil, faults
	}
	words, wordFaults := readWords(english, written)
	faults = inspectProfile(profile)
	faults = append(faults, inspectCharacters(profileFile, profileDocument)...)
	faults = append(faults, wordFaults...)
	faults = append(faults, inspectCharacters(wordsPath, wordsDocument)...)
	if len(faults) > 0 {
		return nil, faults
	}
	return &Snapshot{Profile: profile, Words: words}, nil
}

// readFile decodes a file of the folder strictly: a key a struct lacks is a fault, so a typo is never dropped
// unsaid. A map takes any key, for its reader to check. It hands back the file's document too, every value as the
// file writes it, for the checks no type knows of. A node cannot be decoded strictly, so the file is decoded twice.
func readFile[Target any](folder fs.FS, file string) (Target, *yaml.Node, Faults) {
	var target Target
	data, err := fs.ReadFile(folder, file)
	if err != nil {
		return target, nil, Faults{unreadFault(file, err)}
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	err = decoder.Decode(&target)
	if errors.Is(err, io.EOF) {
		return target, nil, Faults{{File: file, Field: "", Problem: emptyProblem}}
	}
	if err != nil {
		return target, nil, decodingFaults(file, err)
	}
	var document yaml.Node
	err = yaml.Unmarshal(data, &document)
	if err != nil {
		return target, nil, Faults{fileFault(file, err)}
	}
	return target, &document, nil
}

// decodingFaults are the faults of a file the decoder could not read: one for each line it named.
func decodingFaults(file string, err error) Faults {
	typeError, isTypeError := errors.AsType[*yaml.TypeError](err)
	if !isTypeError {
		return Faults{fileFault(file, err)}
	}
	faults := make(Faults, 0, len(typeError.Errors))
	for _, problem := range typeError.Errors {
		faults = append(faults, Fault{File: file, Field: "", Problem: problem})
	}
	return faults
}

func fileFault(file string, err error) Fault {
	return Fault{File: file, Field: "", Problem: err.Error()}
}

// unreadFault is the fault of a file that could not be read. The fault names the file already, so where the error
// names its path as well, only the cause is kept.
func unreadFault(file string, err error) Fault {
	if pathError, isPathError := errors.AsType[*fs.PathError](err); isPathError {
		return fileFault(file, pathError.Err)
	}
	return fileFault(file, err)
}
