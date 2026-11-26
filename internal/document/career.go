package document

import (
	"slices"
	"time"
)

// The months of a year, which a line of months counts by
const monthsInYear = 12

// Career is roles on one line of months, as the TUI and the page both draw it under the brief: the spans of the roles,
// in the order of their starts; where the line begins, the first role's start, or the day where every role starts
// after it; where it ends, the day, or the last month a role reaches past it; and the site's word for the present,
// where a role is still held.
type Career struct {
	Spans   []CareerSpan
	First   time.Time
	Last    time.Time
	Present string
}

// CareerSpan is a role on the line: its company, from its first month to the end of its last, or to the day where it
// is still held, and whether it is.
type CareerSpan struct {
	Company string
	Start   time.Time
	End     time.Time
	Current bool
}

// CareerOf is roles on a line of months on a day: a role that ended runs to the end of its last month, and one still
// held to the day.
func CareerOf(roles []Entry, today time.Time) Career {
	career := Career{Spans: []CareerSpan{}, First: today, Last: today, Present: ""}
	for _, role := range roles {
		start, known := role.Period.Start.(KnownMonth)
		if !known {
			continue
		}
		span := CareerSpan{Company: PlainText(role.Title), Start: start.Start, End: today, Current: false}
		switch end := role.Period.End.(type) {
		case KnownMonth:
			span.End = end.Start.AddDate(0, 1, 0)
		case CurrentMonth:
			span.Current, career.Present = true, end.Text
		}
		career.Spans = append(career.Spans, span)
		career.First = minTime(career.First, span.Start)
		career.Last = maxTime(career.Last, maxTime(span.Start, span.End))
	}
	slices.SortFunc(career.Spans, func(first, second CareerSpan) int { return first.Start.Compare(second.Start) })
	return career
}

// Months is how many months the line runs, one at the least, so a place along it is a share of them.
func (career Career) Months() int {
	return max(career.MonthsTo(career.Last), 1)
}

// MonthsTo is how many months lie from the line's first month to another.
//
//wrapper:where a month stands on the line, which each face reads to place a role
func (career Career) MonthsTo(month time.Time) int {
	return monthsBetween(career.First, month)
}

// Months is how many months a role's span runs.
//
//wrapper:a role's length, which the page reads to size its share of the line
func (span CareerSpan) Months() int {
	return monthsBetween(span.Start, span.End)
}

// monthsBetween is how many months lie from one month to another.
func monthsBetween(from, to time.Time) int {
	return (to.Year()-from.Year())*monthsInYear + int(to.Month()-from.Month())
}

// minTime is the earlier of two times.
func minTime(first, second time.Time) time.Time {
	if second.Before(first) {
		return second
	}
	return first
}

// maxTime is the later of two times.
func maxTime(first, second time.Time) time.Time {
	if second.After(first) {
		return second
	}
	return first
}
