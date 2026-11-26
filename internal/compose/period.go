package compose

import (
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

func periodOf(start, end content.Month, words content.Words) document.Period {
	return document.Period{Start: monthOf(start, words), End: monthOf(end, words)}
}

// monthOf carries a month across, by every kind a snapshot holds; what still goes on ends in the site's word.
func monthOf(month content.Month, words content.Words) document.Month {
	switch month.Kind() {
	case content.MonthCurrent:
		return document.CurrentMonth{Text: words.Text(content.WordPresent)}
	case content.MonthKnown:
	}
	// A snapshot holds no other kind: a load refuses a month it cannot read and a month left out alike
	return document.KnownMonth{Start: month.Start()}
}
