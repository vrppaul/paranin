package compose

import (
	"slices"
	"time"

	"github.com/vrppaul/paranin/internal/content"
)

// The place of the next change among the moments a site changes at, past the reader's own, which comes first
const nextChangePlace = 1

// Audience is whom a site is composed for (docs/FOUNDATION.md §5.4, Status): the moment they read it at, before
// which a post dated later is not out, and whether they read the drafts, as the author does under
// `paranin serve --drafts`.
type Audience struct {
	At     time.Time
	Drafts bool
}

// comesLater is whether a post published or unlisted is dated after the audience's moment, which keeps it a draft
// until then. exhaustive holds the switch to every status.
func (audience Audience) comesLater(post content.Post) bool {
	switch post.Status {
	case content.StatusPublished, content.StatusUnlisted:
		return post.Published.At().After(audience.At)
	case content.StatusDraft:
	}
	return false
}

// statusOf is a post's status as the audience meets it: its own, but a draft while it comes later.
func (audience Audience) statusOf(post content.Post) content.Status {
	if audience.comesLater(post) {
		return content.StatusDraft
	}
	return post.Status
}

// serves is whether the site serves a post at its address to the audience: published or unlisted once it is out, and
// a draft to its author alone. exhaustive holds the switch to every status.
func (audience Audience) serves(post content.Post) bool {
	switch audience.statusOf(post) {
	case content.StatusPublished, content.StatusUnlisted:
		return true
	case content.StatusDraft:
		return audience.Drafts
	}
	return false
}

// lists is whether the blog's lists and its series' parts name a post for the audience: one published and out alone.
// exhaustive holds the switch to every status.
func (audience Audience) lists(post content.Post) bool {
	switch audience.statusOf(post) {
	case content.StatusPublished:
		return true
	case content.StatusDraft, content.StatusUnlisted:
	}
	return false
}

// readsVersions is whether the audience reads a post's older versions: of a post published or unlisted and out, whose
// versions a reader read; a draft's none, which no reader read yet, though its revisions stay for when it is out.
// exhaustive holds the switch to every status.
func (audience Audience) readsVersions(post content.Post) bool {
	switch audience.statusOf(post) {
	case content.StatusPublished, content.StatusUnlisted:
		return true
	case content.StatusDraft:
	}
	return false
}

// showsItsSeries is whether a post's own page shows its series, itself the part read, though no list names it: a
// draft's does, so its author knows which part they read; an unlisted post's never, since its series leaves it out for
// good. exhaustive holds the switch to every status.
func (audience Audience) showsItsSeries(post content.Post) bool {
	switch audience.statusOf(post) {
	case content.StatusPublished, content.StatusDraft:
		return true
	case content.StatusUnlisted:
	}
	return false
}

// changes are the moments the site the audience reads changes at, from theirs on: theirs, then each moment a post
// that comes later comes out, in order, each once.
func (audience Audience) changes(posts []content.Post) []time.Time {
	moments := []time.Time{audience.At}
	for _, post := range posts {
		if audience.comesLater(post) {
			moments = append(moments, post.Published.At())
		}
	}
	slices.SortFunc(moments, time.Time.Compare)
	return slices.CompactFunc(moments, time.Time.Equal)
}

// nextChange is when the site the audience reads changes next, the first of its changes after their moment; the zero
// time where none comes later.
func (audience Audience) nextChange(posts []content.Post) time.Time {
	changes := audience.changes(posts)
	if len(changes) <= nextChangePlace {
		return time.Time{}
	}
	return changes[nextChangePlace]
}

// LastChange is when the site the public reads from a moment on last changes, as the last of a snapshot's posts dated
// later comes out; the moment itself where none comes later. No post leaves the site as time passes, so the site the
// public reads then holds every page it reads from the moment on.
func LastChange(snapshot *content.Snapshot, from time.Time) time.Time {
	changes := Audience{At: from, Drafts: false}.changes(allPosts(snapshot))
	return changes[len(changes)-1]
}
