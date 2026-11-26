package content

import (
	"time"

	"go.yaml.in/yaml/v3"
)

// How the content writes a month, and the end of a role still held
const (
	monthLayout = "2006-01"
	currentWord = "current"
)

// MonthKind is what a month written in the content turned out to be. A snapshot holds the two exported kinds
// alone: a load refuses a month left out and one it could not read, so only content's own checks meet those.
type MonthKind int

const (
	// The zero kind is a month the content leaves out or leaves empty, which yaml never hands over: no check
	// needs its name, as it is neither known nor current
	_ MonthKind = iota
	// monthUnread is text that is none of the kinds below; a check names it
	monthUnread
	// MonthKnown is a month of a year
	MonthKnown
	// MonthCurrent is the end of what still goes on
	MonthCurrent
)

// Month is a month of a year as the content writes it, or the end of what still goes on ("current"). It keeps what
// was written, so a check can name a month it could not read, and a load reports every such month at once.
type Month struct {
	written string
	kind    MonthKind
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
		return Month{written: written, kind: MonthCurrent, start: time.Time{}}
	}
	start, err := time.Parse(monthLayout, written)
	if err != nil {
		return Month{written: written, kind: monthUnread, start: time.Time{}}
	}
	return Month{written: written, kind: MonthKnown, start: start}
}

// Kind is what the month turned out to be.
func (month Month) Kind() MonthKind {
	return month.kind
}

// Written is the month as the content writes it.
func (month Month) Written() string {
	return month.written
}

// Start is the first moment of a known month; for any other kind it is the zero time.
func (month Month) Start() time.Time {
	return month.start
}

// Before tells whether a known month comes before another known one.
func (month Month) Before(other Month) bool {
	return month.start.Before(other.start)
}
