package content

import (
	"regexp"
	"strconv"

	"go.yaml.in/yaml/v3"
)

// How the content writes how long a skill was used: a span, as "1y 7m", of years, of months below a year, or of both,
// years first and one space between, each count above nought and written with no sign and no nought before it; or a
// word for a skill used for no span of its own, on the side or still being learnt. The pattern's groups are the
// years, the months after them, and the months of a span of months alone
const (
	spanPattern      = `^(?:(?P<years>[1-9][0-9]?)y(?: (?P<months>[1-9]|1[01])m)?|(?P<monthsAlone>[1-9]|1[01])m)$`
	yearsGroup       = "years"
	monthsGroup      = "months"
	monthsAloneGroup = "monthsAlone"
	monthsInAYear    = 12
	sideWord         = "side"
	learningWord     = "learning"
)

// DurationKind is what a skill's duration written in the content turned out to be. A snapshot holds the exported
// kinds alone: a load refuses a duration left out and one it could not read, so only content's own check meets the
// unread kind.
type DurationKind int

const (
	// durationUnread is a duration left out, which yaml never hands over, or text that is none of the kinds below; a
	// check names it
	durationUnread DurationKind = iota
	// DurationMeasured is a span measured in years and months
	DurationMeasured
	// DurationSide is a skill used on the side, for no span of its own
	DurationSide
	// DurationLearning is a skill still being learnt
	DurationLearning
	// DurationGap is a duration the author has still to fill
	DurationGap
)

// Duration is how long a skill was used, as the content writes it: a span, on the side, still being learnt, or a gap
// still to fill. It keeps what was written, which a check names and the faces show of a span and a gap, and a span's
// months, which a face may draw.
type Duration struct {
	written string
	kind    DurationKind
	months  int
}

// UnmarshalYAML reads a duration, and never fails on its text: what is no duration is a fault for the check to name.
func (duration *Duration) UnmarshalYAML(node *yaml.Node) error {
	*duration = readDuration(node.Value)
	return nil
}

func readDuration(written string) Duration {
	switch {
	case written == sideWord:
		return Duration{written: written, kind: DurationSide, months: 0}
	case written == learningWord:
		return Duration{written: written, kind: DurationLearning, months: 0}
	case isGap(written):
		return Duration{written: written, kind: DurationGap, months: 0}
	}
	months, isSpan := spanMonths(written)
	if !isSpan {
		return Duration{written: written, kind: durationUnread, months: 0}
	}
	return Duration{written: written, kind: DurationMeasured, months: months}
}

// spanMonths are the months a span comes to, and whether the text is a span, written as spanPattern says.
func spanMonths(written string) (int, bool) {
	pattern := regexp.MustCompile(spanPattern)
	match := pattern.FindStringSubmatch(written)
	if match == nil {
		return 0, false
	}
	years := spanCount(match[pattern.SubexpIndex(yearsGroup)])
	months := spanCount(match[pattern.SubexpIndex(monthsGroup)]) + spanCount(match[pattern.SubexpIndex(monthsAloneGroup)])
	return years*monthsInAYear + months, true
}

// spanCount is the count a group of a span's match writes, none where the span leaves the group out.
func spanCount(matched string) int {
	if matched == "" {
		return 0
	}
	// the pattern lets only one or two digits into a group, which always read as a count
	count, _ := strconv.Atoi(matched)
	return count
}

// Kind is what the duration turned out to be.
func (duration Duration) Kind() DurationKind {
	return duration.kind
}

// Written is the duration as the content writes it.
func (duration Duration) Written() string {
	return duration.written
}

// Months are the months a span comes to; none for any other kind.
func (duration Duration) Months() int {
	return duration.months
}
