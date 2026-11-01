// Package content holds the facts every face draws: the profile, read from a folder on disk, checked whole and
// held as a snapshot that nothing changes after. It knows no face: how a fact is shown is the faces' own
// (docs/ARCHITECTURE.md, ADR-0003). A load reports every fault it finds at once, each at its field, so whoever
// edits the content fixes it in one pass: first every line the decoder cannot read, and once the file reads,
// every fault of what it says.
package content

import (
	"bytes"
	"errors"
	"io"
	"io/fs"

	"go.yaml.in/yaml/v3"
)

// The file of the folder that holds the profile
const profileFile = "profile.yaml"

// Snapshot is the content as one load read it.
type Snapshot struct {
	Profile Profile
}

// Load reads the content folder and checks it whole: a snapshot, or Faults naming everything wrong in it.
func Load(folder fs.FS) (*Snapshot, error) {
	profile, faults := readProfile(folder)
	if len(faults) == 0 {
		faults = inspectProfile(profile)
	}
	if len(faults) > 0 {
		return nil, faults
	}
	return &Snapshot{Profile: profile}, nil
}

// readProfile decodes the profile strictly: a key the schema lacks is a fault, so a typo is never dropped unsaid.
func readProfile(folder fs.FS) (Profile, Faults) {
	var profile Profile
	data, err := fs.ReadFile(folder, profileFile)
	if err != nil {
		return profile, Faults{unreadFault(err)}
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	err = decoder.Decode(&profile)
	if errors.Is(err, io.EOF) {
		return profile, Faults{{File: profileFile, Field: "", Problem: emptyProblem}}
	}
	if err != nil {
		return profile, decodingFaults(err)
	}
	return profile, nil
}

// decodingFaults are the faults of a profile the decoder could not read: one for each line it named.
func decodingFaults(err error) Faults {
	typeError, isTypeError := errors.AsType[*yaml.TypeError](err)
	if !isTypeError {
		return Faults{fileFault(err)}
	}
	faults := make(Faults, 0, len(typeError.Errors))
	for _, problem := range typeError.Errors {
		faults = append(faults, Fault{File: profileFile, Field: "", Problem: problem})
	}
	return faults
}

func fileFault(err error) Fault {
	return Fault{File: profileFile, Field: "", Problem: err.Error()}
}

// unreadFault is the fault of a file that could not be read. The fault names the file already, so where the error
// names its path as well, only the cause is kept.
func unreadFault(err error) Fault {
	if pathError, isPathError := errors.AsType[*fs.PathError](err); isPathError {
		return fileFault(pathError.Err)
	}
	return fileFault(err)
}
