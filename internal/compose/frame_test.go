package compose

import (
	"testing"

	"github.com/vrppaul/paranin/internal/fixtures"
)

// The site is named after its person, as the profile writes the name, its marks read and gone, in every language's
// frame alike.
func TestTheSiteIsNamedAfterItsPerson(test *testing.T) {
	marked := *fixtures.LoadedContent(test)
	marked.Profile.Person.Name = "==Mira== Example"
	for _, composed := range SiteOf(Prepare(&marked), publicOnTheDay()).Blogs() {
		if got, want := composed.Frame.SiteName, "Mira Example"; got != want {
			test.Errorf("the site's name in the frame of %q = %q, want %q", composed.Frame.Language, got, want)
		}
	}
}
