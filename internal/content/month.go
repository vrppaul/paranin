package content

import (
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

// How the content writes a month, and the end of a role still held
const (
	monthLayout = "2006-01"
	currentWord = "current"
)

// What a gap is written between: what the author has still to fill, "[20XX]"
const (
	gapOpening = "["
	gapClosing = "]"
)

// monthKind is what a month written in the content turned out to be.
type monthKind int

const (
	// The zero kind is a month the content leaves out or leaves empty, which yaml never hands over: no check
	// needs its name, as it is neither known, a gap nor current
	_ monthKind = iota
	// unreadMonth is text that is none of the kinds below; a check names it
	unreadMonth
	knownMonth
	gapMonth
	// currentMonth is the end of what still goes on
	currentMonth
)

// Month is a month of a year as the content writes it, a gap still to fill, or the end of what still goes on
// ("current"). It keeps what was written, so a check can name a month it could not read, and a load reports
// every such month at once.
type Month struct {
	written string
	kind    monthKind
	// start is the first moment of a known month
	start time.Time
}

// UnmarshalYAML reads a month, and never fails on its text: what is no month is a fault for the check to name.
func (month *Month) UnmarshalYAML(node *yaml.Node) error {
	*month = readMonth(node.Value)
	return nil
}

func readMonth(written string) Month {
	if written == currentWord {
		return Month{written: written, kind: currentMonth, start: time.Time{}}
	}
	if isGap(written) {
		return Month{written: written, kind: gapMonth, start: time.Time{}}
	}
	start, err := time.Parse(monthLayout, written)
	if err != nil {
		return Month{written: written, kind: unreadMonth, start: time.Time{}}
	}
	return Month{written: written, kind: knownMonth, start: start}
}

func isGap(written string) bool {
	inside, opened := strings.CutPrefix(written, gapOpening)
	inside, closed := strings.CutSuffix(inside, gapClosing)
	return opened && closed && inside != "" && !strings.ContainsAny(inside, gapOpening+gapClosing)
}

// IsKnown tells a month read as a month of a year.
func (month Month) IsKnown() bool {
	return month.kind == knownMonth
}

// IsGap tells a month the author has still to fill.
func (month Month) IsGap() bool {
	return month.kind == gapMonth
}

// IsCurrent tells the end of what still goes on.
func (month Month) IsCurrent() bool {
	return month.kind == currentMonth
}

// Written is the month as the content writes it.
func (month Month) Written() string {
	return month.written
}

// Before tells whether a known month comes before another known one.
func (month Month) Before(other Month) bool {
	return month.start.Before(other.start)
}
