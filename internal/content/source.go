package content

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
)

// How a translation names the text of its original it was made from (docs/FOUNDATION.md §5.4, Translations),
// en@3f9a1c0b2d4e: its field; what parts the original's language from the hash; and the hash, SHA-256's first twelve
// hexadecimal digits, as git abbreviates a commit, enough that no two versions of one post share one
const (
	sourceField      = "source"
	sourceJoint      = "@"
	sourceHashLength = 12
)

// sourceOf is how a translation names a version of its original: the original's language, then the hash of the
// version's text.
func sourceOf(version Revision) string {
	digest := sha256.Sum256(version.text())
	return english + sourceJoint + hex.EncodeToString(digest[:])[:sourceHashLength]
}

// sourceHashForm is how a source's hash is written: hexadecimal digits in small letters, as many as sourceOf writes.
func sourceHashForm() *regexp.Regexp {
	return regexp.MustCompile(`^[0-9a-f]{` + strconv.Itoa(sourceHashLength) + `}$`)
}

// inspectSource holds a translation's source to the form a source takes, a fault where not, and finds it stale where
// it names a text of its original other than the one that stands, a warning, which refuses nothing: a version a
// revision keeps, or one edited since and kept nowhere. It says whether the translation is stale.
func inspectSource(inspection, warnings *inspection, original Post, source string) bool {
	current := sourceOf(original.Current())
	language, hash, isParted := strings.Cut(source, sourceJoint)
	switch {
	case blank(source):
		inspection.add(sourceField, noSourceProblem, current)
	case !isParted || language != english || !sourceHashForm().MatchString(hash):
		inspection.add(sourceField, sourceProblem, source, current)
	case source != current:
		warnings.add(sourceField, staleProblem, source, madeFrom(original, source), current)
		return true
	}
	return false
}

// madeFrom says which version of its original a stale translation was made from: the day of the one a revision keeps,
// or that none keeps it.
func madeFrom(original Post, source string) string {
	for _, revision := range original.Revisions {
		if sourceOf(revision) == source {
			return olderVersionSource + revision.Day.written
		}
	}
	return editedSource
}
