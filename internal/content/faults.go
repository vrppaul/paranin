package content

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Fault is one thing wrong with the content: in a file, at a field of it where the fault has one.
type Fault struct {
	File    string
	Field   string
	Problem string
}

// The quotes strconv writes a character and a string between
const (
	characterQuote = "'"
	stringQuote    = `"`
)

// How many bytes utf8 reads for a byte of no character, which it decodes as the replacement character: the
// replacement character itself takes three
const byteOfNoCharacter = 1

// String writes a fault as a person reads it, on a line of its own: the file, the field, what is wrong.
func (fault Fault) String() string {
	if fault.Field == "" {
		return escaped(fmt.Sprintf("%s: %s", fault.File, fault.Problem))
	}
	return escaped(fmt.Sprintf("%s: %s %s", fault.File, fault.Field, fault.Problem))
}

// escaped writes a fault's line with every character a reader acts on, and every byte of no character, as %q writes
// it, and the rest as it is: a key and the decoder's words come from the file, and reach a terminal.
func escaped(line string) string {
	var written strings.Builder
	for remaining := line; remaining != ""; {
		character, width := utf8.DecodeRuneInString(remaining)
		switch {
		case character == utf8.RuneError && width == byteOfNoCharacter:
			written.WriteString(strings.Trim(strconv.Quote(remaining[:width]), stringQuote))
		case actedOn(character):
			written.WriteString(strings.Trim(strconv.QuoteRune(character), characterQuote))
		default:
			written.WriteRune(character)
		}
		remaining = remaining[width:]
	}
	return written.String()
}

// Faults are every fault a load found. As an error it lists them, one a line.
type Faults []Fault

func (faults Faults) Error() string {
	lines := make([]string, 0, len(faults))
	for _, fault := range faults {
		lines = append(lines, fault.String())
	}
	return strings.Join(lines, "\n")
}
