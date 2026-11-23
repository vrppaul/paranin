package document

import "time"

// Version is which version of a post its reader reads: one out on a day, or a draft, which is not out.
//
//sumtype:decl
type Version interface {
	version()
}

// VersionOut is a version out: when it came out, in the zone it was written in.
type VersionOut struct {
	Published time.Time
}

// VersionDraft is a draft, which only its author reads: the site's word for one, which a face writes where a version
// out gives its day.
type VersionDraft struct {
	Word string
}

func (VersionOut) version()   {}
func (VersionDraft) version() {}
