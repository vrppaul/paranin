package text

import (
	"testing"

	"github.com/vrppaul/paranin/internal/routes"
)

// A page the site does not have quotes the path asked for, so no control character in it reaches the reader's
// terminal, and says where the site begins.
func TestAMissingPageQuotesItsPathAndSaysWhereTheSiteBegins(test *testing.T) {
	path := "/nothing\x1b[2J"
	got := NewNotFound(routes.New(fixtureBase), "404 Not Found").For(path)
	want := "404 Not Found: \"/nothing\\x1b[2J\"\n\ncurl https://example.test/\n"
	if got != want {
		test.Errorf("NotFound for %q =\n%q\nwant\n%q", path, got, want)
	}
}
