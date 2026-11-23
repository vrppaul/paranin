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

// comesLater is whether a post published or unlisted is dated after the audience's moment, which keeps it a preview
// until then. exhaustive holds the switch to every status.
func (audience Audience) comesLater(post content.Post) bool {
	switch post.Status {
	case content.StatusPublished, content.StatusUnlisted:
		return post.Published.At().After(audience.At)
	case content.StatusDraft, content.StatusPreview:
	}
	return false
}

// statusOf is a post's status as the audience meets it: its own, but a preview while it comes later.
func (audience Audience) statusOf(post content.Post) content.Status {
	if audience.comesLater(post) {
		return content.StatusPreview
	}
	return post.Status
}

// serves is whether the site serves a post at its address to the audience: published or unlisted once it is out, and
// a draft to its author alone. A preview has no address, only its link (previews). exhaustive holds the switch to
// every status.
func (audience Audience) serves(post content.Post) bool {
	switch audience.statusOf(post) {
	case content.StatusPublished, content.StatusUnlisted:
		return true
	case content.StatusDraft:
		return audience.Drafts
	case content.StatusPreview:
	}
	return false
}

// previews is whether the site serves a post to the audience at its preview link alone: a preview, or a post published
// or unlisted that comes later, until it is out; never a draft, which its author alone reads.
func (audience Audience) previews(post content.Post) bool {
	return audience.statusOf(post) == content.StatusPreview
}

// lists is whether the blog's lists and its series' parts name a post for the audience: one published and out alone.
// exhaustive holds the switch to every status.
func (audience Audience) lists(post content.Post) bool {
	switch audience.statusOf(post) {
	case content.StatusPublished:
		return true
	case content.StatusDraft, content.StatusPreview, content.StatusUnlisted:
	}
	return false
}

// showsItsSeries is whether a post's own page shows its series, itself the part read, though no list names it: a
// draft's and a preview's do, so their reader knows which part they read; an unlisted post's never, since its series
// leaves it out for good. exhaustive holds the switch to every status.
func (audience Audience) showsItsSeries(post content.Post) bool {
	switch audience.statusOf(post) {
	case content.StatusPublished, content.StatusDraft, content.StatusPreview:
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
