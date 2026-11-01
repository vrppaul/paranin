package text

import "github.com/vrppaul/paranin/internal/document"

// How a month of a year is written
const monthLayout = "01/2006"

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
