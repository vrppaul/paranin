package compose

import (
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

// periodOf carries a role's months across: it starts in a known month, as the content check holds, and ends in one or
// in what still goes on.
func periodOf(start, end content.Month, words content.Words) document.Period {
	return document.Period{Start: document.KnownMonth{Start: start.Start()}, End: monthOf(end, words)}
}

// monthOf carries a month across, by every kind a snapshot holds; what still goes on ends in the site's word.
func monthOf(month content.Month, words content.Words) document.Month {
	switch month.Kind() {
	case content.MonthKnown:
		return document.KnownMonth{Start: month.Start()}
	case content.MonthCurrent:
		return document.CurrentMonth{Text: words.Text(content.WordPresent)}
	}
	// A snapshot holds no other kind, as a load refuses a month it cannot read and a month left out alike: such a month
	// would be the zero time
	return document.KnownMonth{Start: time.Time{}}
}
