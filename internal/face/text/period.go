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

// versionText writes the version a page's head says is read: the day a post came out, or the site's word for a draft,
// which has not; nothing on any other page, which has none.
func versionText(head document.Head) string {
	switch version := head.Version.(type) {
	case document.VersionOut:
		return dayText(version.Published)
	case document.VersionDraft:
		return version.Word
	}
	return ""
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
