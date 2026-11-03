package compose

import "github.com/vrppaul/paranin/internal/content"

// ThemeWords are the words of a reader's choice between the light theme and the dark one, which a face lays beside
// the documents it draws: the label of the choice, then each theme's.
type ThemeWords struct {
	Label string
	Light string
	Dark  string
}

// ThemeWordsOf are the words of the choice of theme, in the site's language.
func ThemeWordsOf(snapshot *content.Snapshot) ThemeWords {
	words := snapshot.Words
	return ThemeWords{
		Label: words.Text(content.WordTheme),
		Light: words.Text(content.WordLight),
		Dark:  words.Text(content.WordDark),
	}
}
