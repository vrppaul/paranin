package compose

import (
	"slices"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// postVersion is a version of a post as compose reads it (docs/FOUNDATION.md §5.4, Revisions): the post; the version's
// own text, its current one or a revision; and the page it is read at, the post's own for its current version.
type postVersion struct {
	post content.Post
	text content.Revision
	page document.Page
}

// currentOf is a post's current version, at the post's own page.
func currentOf(post content.Post) postVersion {
	return postVersion{post: post, text: post.Current(), page: postPage(post)}
}

// olderVersionsOf are a post's older versions its audience reads, newest first, each at its revision's page: none of a
// post not out, whose versions no reader read yet.
func (read blog) olderVersionsOf(post content.Post) []postVersion {
	if !read.audience.readsVersions(post) {
		return nil
	}
	versions := make([]postVersion, 0, len(post.Revisions))
	for _, revision := range slices.Backward(post.Revisions) {
		versions = append(versions, postVersion{post: post, text: revision, page: postPage(post).OlderVersion(revision.Day.Written())})
	}
	return versions
}

// versionsOf are the versions of a post out its reader may go to, read at a page of it, newest first: its current one,
// then each older one its audience reads.
func (read blog) versionsOf(post content.Post, shown document.Page) []document.PostVersion {
	listed := []document.PostVersion{listedVersion(currentOf(post), shown)}
	for _, version := range read.olderVersionsOf(post) {
		listed = append(listed, listedVersion(version, shown))
	}
	return listed
}

// listedVersion is a version as a post's versions list it, read at a page: its page, the day it came out and its name,
// that day as its revision writes it, and whether it is the current version and the version read.
func listedVersion(version postVersion, shown document.Page) document.PostVersion {
	day := version.text.Day
	return document.PostVersion{Page: version.page, Day: day.Start(), Name: day.Written(), Current: version.page.Revision == "", Read: version.page == shown}
}

// versionOf is the version of a post its reader reads, among the languages it is read in: a draft, by the site's word
// for one; or a version out, by the day it came out, with every version its reader may go to, the current one of a
// translation older than its original saying so, an older one with the current one it says stands over it, in the
// site's words round that version's day. exhaustive holds the switch to every status.
func (read blog) versionOf(version postVersion, words content.Words, languages []document.Language) document.Version {
	switch version.post.Status {
	case content.StatusDraft:
		return document.VersionDraft{Word: words.Text(content.WordDraft)}
	case content.StatusPreview, content.StatusPublished, content.StatusUnlisted:
	}
	day, versions := version.text.Day.Start(), read.versionsOf(version.post, version.page)
	if version.page.Revision == "" {
		return document.VersionOut{Day: day, Versions: versions, Outdated: outdatedOf(version.post, words, languages)}
	}
	newer := listedVersion(currentOf(version.post), version.page)
	before, after := words.Around(content.PhraseNewerVersion)
	return document.VersionOlder{Day: day, Newer: newer, NewerWords: document.Around{Before: before, After: after}, Versions: versions}
}

// outdatedOf is what a translation older than its original says, leading to the original where its Switch offers the
// original, among the languages it is read in; none where it is no stale translation, or its Switch leads nowhere.
func outdatedOf(post content.Post, words content.Words, languages []document.Language) *document.Outdated {
	if !post.Stale {
		return nil
	}
	for _, offered := range languages {
		if offered.Page.Translation == document.Untranslated {
			return &document.Outdated{Original: offered.Page, Language: offered.Tag, Words: words.Text(content.WordOutdated)}
		}
	}
	return nil
}

// olderDocuments are the older versions of the posts the blog serves its audience, each a page of its own, the posts in
// the snapshot's order and each one's versions newest first; each composed as postDocument composes a post, its links
// read among the pages served and the feeds.
func olderDocuments(prepared *Prepared, read blog, paths map[string]document.Target) []document.Document {
	documents := []document.Document{}
	for _, post := range read.served {
		for _, version := range read.olderVersionsOf(post) {
			documents = append(documents, postDocument(version, read, prepared, paths))
		}
	}
	return documents
}
