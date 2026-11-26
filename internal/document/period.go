package document

import "time"

// Period is the time an entry spans: from a known month to one, or to what still goes on.
type Period struct {
	Start KnownMonth
	End   Month
}

// GoesOn is whether the period ends in what still goes on: a role still held.
func (period Period) GoesOn() bool {
	_, current := period.End.(CurrentMonth)
	return current
}

// Month is one kind of month an entry starts or ends in.
//
//sumtype:decl
type Month interface {
	month()
}

// KnownMonth is a month of a year, by its first moment.
type KnownMonth struct {
	Start time.Time
}

// CurrentMonth is the end of what still goes on, with the site's word for it.
type CurrentMonth struct {
	Text string
}

func (KnownMonth) month()   {}
func (CurrentMonth) month() {}
