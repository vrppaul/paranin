package compose

import "github.com/vrppaul/paranin/internal/content"

// FrameWords are the site's words a face lays round the documents it draws, on every page alike: the name of the
// theme button, which is pressed while the page is dark.
type FrameWords struct {
	DarkTheme string
}

// FrameWordsOf are the words of a page's frame, in the site's language.
func FrameWordsOf(snapshot *content.Snapshot) FrameWords {
	return FrameWords{DarkTheme: snapshot.Words.Text(content.WordDarkTheme)}
}
