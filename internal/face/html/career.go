package html

import (
	"strconv"

	"github.com/vrppaul/paranin/internal/document"
)

// How the career line stands in its SVG (DESIGN.md, CareerLine), in ems of the caption the stylesheet sets as its font
// size: the baselines of the two rows its names alternate on, so a short role keeps its name; the top of the spans
// and how tall they are; and the line's whole height. Along the line a place is a share of it, from the data
const (
	upperNameBaseline = "1em"
	lowerNameBaseline = "2.5em"
	spanTop           = "3.375em"
	spanHeight        = "0.625em"
	careerLineHeight  = "4em"
)

// How a share of the line is written, a percentage to a tenth; and the whole of a length, as wide as the SVG's column
// a drawing stands in, and as tall as the box of a role's bar in the phone's list, which the stylesheet sizes
const (
	percentSign     = "%"
	wholePercent    = 100
	percentDecimals = 1
	floatBitSize    = 64
	wholeLength     = "100%"
)

// The class of the role still held, which the stylesheet marks out
const currentRoleClass = "career-now"

// careerDrawing is the résumé's career as the first screen draws it: each role, the year the line starts in, and
// what stands at its end, the site's word for the present where a role is still held, else the year it ends in.
type careerDrawing struct {
	roles []careerRole
	start string
	end   string
}

// careerRole is a role as the line and the phone's list draw it: its company and its years; where its span starts
// along the line and how long it is, and how long it is beside the longest role, each a share as a percentage; the
// baseline of its name; and whether it is still held.
type careerRole struct {
	company  string
	years    string
	left     string
	width    string
	share    string
	baseline string
	current  bool
}

// careerDrawingOf is a career as the first screen draws it; no roles where it holds none.
func careerDrawingOf(career document.Career) careerDrawing {
	longest := 1
	for _, span := range career.Spans {
		longest = max(longest, span.Months())
	}
	baselines := nameBaselines()
	roles := make([]careerRole, 0, len(career.Spans))
	for index, span := range career.Spans {
		roles = append(roles, careerRole{
			company:  span.Company,
			years:    yearsOf(career, span),
			left:     percentOf(career.MonthsTo(span.Start), career.Months()),
			width:    percentOf(span.Months(), career.Months()),
			share:    percentOf(span.Months(), longest),
			baseline: baselines[index%len(baselines)],
			current:  span.Current,
		})
	}
	end := career.Present
	if end == "" {
		end = strconv.Itoa(career.Last.Year())
	}
	return careerDrawing{roles: roles, start: strconv.Itoa(career.First.Year()), end: end}
}

// nameBaselines are the rows the roles' names alternate on, in turn.
func nameBaselines() []string {
	return []string{upperNameBaseline, lowerNameBaseline}
}

// yearsOf is the years a role ran, as the phone's list writes them: the year it started, and the year its last month
// falls in, or the site's word for the present where it is still held.
func yearsOf(career document.Career, span document.CareerSpan) string {
	end := career.Present
	if !span.Current {
		end = strconv.Itoa(span.End.AddDate(0, -1, 0).Year())
	}
	return strconv.Itoa(span.Start.Year()) + rangeDash + end
}

// percentOf is a part of a whole as a percentage, as an SVG writes a length.
func percentOf(part, whole int) string {
	return strconv.FormatFloat(float64(part)*wholePercent/float64(whole), 'f', percentDecimals, floatBitSize) + percentSign
}
