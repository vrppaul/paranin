package document

import "time"

// Version is which version of a post its reader reads: the current one, out on a day, an older one, or a draft, which
// is not out.
//
//sumtype:decl
type Version interface {
	version()
}

// VersionOut is a post's current version, out: the day it came out, the day its text last changed, in the zone it was
// written in; every version of the post, newest first, itself first, alone where it has no older one; and, on a
// translation older than its original, what it says of that, none on any other.
type VersionOut struct {
	Day time.Time
	// Published is when the post itself came out, its first version, as the blog's list dates it, in its own zone
	Published time.Time
	Versions  []PostVersion
	Outdated  *Outdated
}

// Outdated is what a translation older than its original says under its day (docs/FOUNDATION.md §5.4, Translations):
// the original's page and the tag of its language, and the site's words that say the translation is older.
type Outdated struct {
	Original Page
	Language string
	Words    string
}

// Spans are what a translation older than its original says, each face writing the link in its own format: the site's
// words, a link to the original.
func (outdated Outdated) Spans() []Span {
	return []Span{Link{Target: outdated.Original, Spans: []Span{Text{Text: outdated.Words}}}}
}

// VersionOlder is an older version of a post, over which a newer stands (docs/FOUNDATION.md §5.4, Revisions): the day
// it came out; the current version, and the site's words that say it stands, round its day; and every version of the
// post, newest first, itself among them.
type VersionOlder struct {
	Day        time.Time
	Newer      PostVersion
	NewerWords Around
	Versions   []PostVersion
}

// Around are the words a phrase of the site sets round a value each face writes its own way, as a day: those before
// it and those after it.
type Around struct {
	Before string
	After  string
}

// NewerSpans are what an older version says of the newer one that stands over it, its day as a face writes one: the
// site's words round the day, a link to that version.
func (older VersionOlder) NewerSpans(day string) []Span {
	link := Link{Target: older.Newer.Page, Spans: []Span{Text{Text: day}}}
	return []Span{Text{Text: older.NewerWords.Before}, link, Text{Text: older.NewerWords.After}}
}

// VersionDraft is a draft, which only its author reads: the site's word for one, which a face writes where a version
// out gives its day.
type VersionDraft struct {
	Word string
}

// PostVersion is a version of a post as its versions list it: its page; the day it came out, and its name, that day as
// its address writes it, YYYY-MM-DD; and whether it is the post's current version, and the version read.
type PostVersion struct {
	Page    Page
	Day     time.Time
	Name    string
	Current bool
	Read    bool
}

func (VersionOut) version()   {}
func (VersionOlder) version() {}
func (VersionDraft) version() {}

// VersionsOf are the versions of a post a head lists, newest first, the version read among them; none where its
// reader reads a draft. gochecksumtype holds the switch to every kind.
func VersionsOf(version Version) []PostVersion {
	switch typed := version.(type) {
	case VersionOut:
		return typed.Versions
	case VersionOlder:
		return typed.Versions
	case VersionDraft:
	}
	return nil
}
