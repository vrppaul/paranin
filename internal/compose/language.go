package compose

import (
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// The fewest languages a post's Switch offers: the one read, and another to go to
const switchedLanguages = 2

// language is the blog in one language as a post's Switch reads it: the language's words, and the posts its lists name.
type language struct {
	words  content.Words
	listed map[string]bool
}

// blogsOf are a snapshot's blog in the site's own language and in each it is translated into, as an audience reads
// them, each knowing every language the blog is read in.
func blogsOf(snapshot *content.Snapshot, audience Audience) (blog, []blog) {
	own := blogOf(snapshot, snapshot.Own, audience)
	translations := make([]blog, 0, len(snapshot.Translations))
	languages := []language{{words: own.words, listed: own.listedSlugs()}}
	for _, edition := range snapshot.Translations {
		read := blogOf(snapshot, edition, audience)
		translations = append(translations, read)
		languages = append(languages, language{words: edition.Words, listed: read.listedSlugs()})
	}
	own.languages = languages
	for index := range translations {
		translations[index].languages = languages
	}
	return own, translations
}

// listedSlugs are the slugs of the posts the blog's lists name.
func (read blog) listedSlugs() map[string]bool {
	slugs := make(map[string]bool, len(read.listed))
	for _, post := range read.listed {
		slugs[post.Slug] = true
	}
	return slugs
}

// languagesOf are the languages a version of a post is read in, as its Switch offers them: the one read, at the
// version's page, and each other the lists name the post in; none where it is read in one alone.
func (read blog) languagesOf(version postVersion) []document.Language {
	offered := []document.Language{}
	for _, other := range read.languages {
		words := other.words
		isRead := words.Language() == read.words.Language()
		page := version.page
		if !isRead {
			page = postPage(version.post).InLanguage(words.Translation())
		}
		if isRead || other.listed[version.post.Slug] {
			offered = append(offered, document.Language{
				Tag: words.Language(), Name: words.Text(content.WordLanguageName), Locale: words.Text(content.WordLocale), Page: page,
				Read: isRead,
			})
		}
	}
	if len(offered) < switchedLanguages {
		return nil
	}
	return offered
}
