package names

import (
	"errors"
	text "strings" // a name given to an import is a name too

	s "strconv"   // want `"s" is not a whole word \(in "s"\)`
	str "unicode" // want `"str" is not a whole word \(in "str"\)`
)

type place struct {
	row      int
	at       int // a short word that still says what it means
	ID       int // an initialism in capitals
	Id       int // want `"Id" spells an initialism in mixed case: write "ID"`
	userID   int // a short word inside a longer name
	isOK     bool
	x        int // want `"x" is not a whole word \(in "x"\)`
	xy       int // want `"xy" is not a whole word \(in "xy"\)`
	segWidth int // want `"seg" is not a whole word \(in "segWidth"\)`
	kOK      int // want `"k" is not a whole word \(in "kOK"\)`
}

// Go writes an initialism in one case, its plural with a small s, and a few names of things as their owners
// spell them: each is one word.
type initialisms struct {
	userIDs     []int
	URLs        []string
	APIs        []string
	ETag        string
	staticETags []string
	isIPv6      bool
	OAuthToken  string
	clientIP    string
	uiState     int
	osName      string
}

// A word cut short is refused in the plural too, and the forms Go calls wrong are refused.
type plurals struct {
	segs    []int    // want `"segs" is not a whole word \(in "segs"\)`
	srvName string   // want `"srv" is not a whole word \(in "srvName"\)`
	userIds []int    // want `"userIds" spells an initialism in mixed case: write "IDs"`
	Urls    []string // want `"Urls" spells an initialism in mixed case: write "URLs"`
	errs    []error  // want `"errs" is not a whole word \(in "errs"\)`
}

type HTTPServer struct{} // a run of capitals is one word

// A digit stays with the word before it, and is no letter of it.
func digits() int {
	utf8Text := 1
	i0 := 2  // want `"i0" is not a whole word \(in "i0"\)`
	p99 := 3 // want `"p99" is not a whole word \(in "p99"\)`
	return utf8Text + i0 + p99
}

// An initialism of two letters stands inside a name, and alone only as "id", as D11 and D14 have it.
func alone() string {
	id := "one"
	ip := "two"   // want `"ip" is not a whole word \(in "ip"\)`
	os := "three" // want `"os" is not a whole word \(in "os"\)`
	return id + ip + os
}

// Words of English that look like an abbreviation's plural are words.
func optics(lens []int, curs []string) int { return len(lens) + len(curs) }

// The standard library's names of cryptography are kept; the plural of a kept initialism stands alone; an
// initialism ending in s is no plural.
func ciphers(ed25519Key, hostKeyEd25519, md5Sum, x509Pool, css string, ids []int) int {
	ips := ids // want `"ips" is not a whole word \(in "ips"\)`
	return len(ed25519Key+hostKeyEd25519+md5Sum+x509Pool+css) + len(ips)
}

// Each part of a name written with underscores starts as a name starts.
func Example_url() {}

func Test_id() {}

func (p place) column() int { // want `"p" is not a whole word \(in "p"\)`
	return p.row + p.at + p.ID + p.userID + p.x + p.xy + p.segWidth + p.kOK
}

func measure(input string) (count int) {
	for i := range len(input) { // want `"i" is not a whole word \(in "i"\)`
		count += i
	}
	ok := text.Contains(input, s.Itoa(count)) && str.IsUpper('A')
	_ = ok
	return count
}

func identity[T any](value T) T { // want `"T" is not a whole word \(in "T"\)`
	return value
}

func failing(ctx any) error {
	err := errors.New("failed")
	_ = ctx
	return err
}

func kind(value any) string {
	switch v := value.(type) { // want `"v" is not a whole word \(in "v"\)`
	case int:
		return "a number"
	default:
		_ = v
		return "something else"
	}
}

// A small word of English may stand inside a longer name, never alone.
func oneOfAKind() int {
	a := 1 // want `"a" is not a whole word \(in "a"\)`
	return a
}
