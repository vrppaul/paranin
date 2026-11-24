package text

import (
	"time"

	"github.com/vrppaul/paranin/internal/document"
)

// How a month of a year is written, and a day: in digits alone, the day's as ISO 8601 writes it, which needs no
// language's name of a month
const (
	monthLayout = "01/2006"
	dayLayout   = time.DateOnly
)

// versionText writes the version a page's head says is read: the day a version of a post came out, the current one or
// an older one, or the site's word for a draft, which has not; nothing on any other page, which has none.
func versionText(head document.Head) string {
	switch version := head.Version.(type) {
	case document.VersionOut:
		return dayText(version.Day)
	case document.VersionOlder:
		return dayText(version.Day)
	case document.VersionDraft:
		return version.Word
	}
	return ""
}

// newerSpans are what an older version of a post says of the newer one that stands over it, each face writing the link
// in its own format: the site's words round the current version's day, a link to it; none on any other page.
func newerSpans(head document.Head) []document.Span {
	older, isOlder := head.Version.(document.VersionOlder)
	if !isOlder {
		return nil
	}
	return older.NewerSpans(dayText(older.Newer.Day))
}

// dayText writes the day a moment falls on, in the zone it holds; nothing for the zero time, which is none.
func dayText(moment time.Time) string {
	if moment.IsZero() {
		return ""
	}
	return moment.Format(dayLayout)
}

// periodText writes a period as text, its start and its end: each face escapes it as its format needs.
func periodText(period document.Period) string {
	return monthText(period.Start) + rangeDash + monthText(period.End)
}

func monthText(month document.Month) string {
	switch typed := month.(type) {
	case document.KnownMonth:
		return typed.Start.Format(monthLayout)
	case document.GapMonth:
		return typed.Text
	case document.CurrentMonth:
		return typed.Text
	}
	return ""
}
