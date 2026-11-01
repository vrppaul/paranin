package compose

import (
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
)

func periodOf(start, end content.Month) document.Period {
	return document.Period{Start: monthOf(start), End: monthOf(end)}
}

// monthOf carries a month across, by every kind a snapshot holds.
func monthOf(month content.Month) document.Month {
	switch month.Kind() {
	case content.MonthKnown:
		return document.KnownMonth{Start: month.Start()}
	case content.MonthGap:
		return document.GapMonth{Text: month.Written()}
	case content.MonthCurrent:
		return document.CurrentMonth{}
	}
	// A snapshot holds no other kind: a load refuses a month it cannot read and a month left out alike. Shown as
	// written, such a month would say what the content says
	return document.GapMonth{Text: month.Written()}
}
