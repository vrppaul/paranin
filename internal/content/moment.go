package content

import (
	"time"

	"go.yaml.in/yaml/v3"
)

// How a post writes when it was published, to the minute and with its zone, and the day it was last updated
const (
	momentLayout = "2006-01-02T15:04Z07:00"
	dayLayout    = time.DateOnly
)

// Moment is a date and time with its zone, as a post writes when it was published. It keeps what was written, so a
// check can tell one left out from one it could not read, and name the second.
type Moment struct {
	written string
	at      time.Time
	known   bool
}

// UnmarshalYAML reads a moment, and never fails on its text: what is no moment is a fault for the check to name.
func (moment *Moment) UnmarshalYAML(node *yaml.Node) error {
	*moment = readMoment(node.Value)
	return nil
}

func readMoment(written string) Moment {
	at, err := time.Parse(momentLayout, written)
	return Moment{written: written, at: at, known: err == nil}
}

// day is the day the moment falls on in its own zone, as a Day's start holds one: its first moment in UTC.
func (moment Moment) day() time.Time {
	year, month, day := moment.at.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// Day is a date, as a post writes the day it was last updated. It keeps what was written, as a Moment does.
type Day struct {
	written string
	start   time.Time
	known   bool
}

// UnmarshalYAML reads a day, and never fails on its text: what is no day is a fault for the check to name.
func (day *Day) UnmarshalYAML(node *yaml.Node) error {
	*day = readDay(node.Value)
	return nil
}

func readDay(written string) Day {
	start, err := time.Parse(dayLayout, written)
	return Day{written: written, start: start, known: err == nil}
}
