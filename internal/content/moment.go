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

// At is the moment, in the zone it was written in; the zero time where none was written or it did not read.
func (moment Moment) At() time.Time {
	return moment.at
}

// Written is the moment as the front matter writes it.
func (moment Moment) Written() string {
	return moment.written
}

// day is the day the moment falls on in its own zone, as a Day's start holds one: its first moment in UTC.
//
//wrapper:names the moment's own day, which the checks of a post's days and its versions' read
func (moment Moment) day() time.Time {
	return moment.dayAt(moment.at)
}

// dayAt is the day a time falls on in the zone the moment was written in, as a Day's start holds one.
func (moment Moment) dayAt(at time.Time) time.Time {
	year, month, day := at.In(moment.at.Location()).Date()
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

// Written is the day as the content writes it.
func (day Day) Written() string {
	return day.written
}

// Start is the day's first moment in UTC; the zero time where none was written or it did not read.
func (day Day) Start() time.Time {
	return day.start
}
