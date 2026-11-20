// Package history holds the words that tell what code was, which the gates on comments and on commit messages
// refuse: prose says what stands, and git says what changed (.claude/rules/prose.md). Taken from stint's
// scripts/history.py.
package history

import "regexp"

// The phrases that tell what code was, as stint's scripts/history.py lists them, any case
const toldAsHistory = `(?i)\b(no longer(?P<measure> than)?|anymore|formerly|the old way|used to be` +
	`|(?:was|were|has been|have been) renamed|renamed from|was changed to|now uses)\b`

// The group that matches the "than" of "no longer than", a measure that RE2, with no lookahead, matches to pass
const measureGroup = "measure"

// Where a find's whole phrase starts and ends among its submatches' indices, two to a group
const (
	phraseStart     = 0
	phraseEnd       = 1
	indicesPerGroup = 2
)

// Found is a phrase that tells what code was, and the byte after its last in the text read.
type Found struct {
	Phrase string
	End    int
}

// Reader finds the phrases that tell what code was, and passes the measure by its start among a find's indices.
type Reader struct {
	told         *regexp.Regexp
	measureStart int
}

// NewReader is a reader of the phrases, compiled once for every text it reads.
func NewReader() Reader {
	told := regexp.MustCompile(toldAsHistory)
	return Reader{told: told, measureStart: told.SubexpIndex(measureGroup) * indicesPerGroup}
}

// Within are the phrases in a text that tell what code was, in the order they stand.
func (reader Reader) Within(text string) []Found {
	found := []Found{}
	for _, indices := range reader.told.FindAllStringSubmatchIndex(text, -1) {
		if indices[reader.measureStart] >= 0 {
			continue
		}
		found = append(found, Found{Phrase: text[indices[phraseStart]:indices[phraseEnd]], End: indices[phraseEnd]})
	}
	return found
}
