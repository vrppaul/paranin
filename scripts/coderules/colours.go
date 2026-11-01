package main

import (
	"regexp"

	"golang.org/x/tools/go/analysis"
)

// A colour written out: a string that is a hex colour whole, or that calls one of CSS's colour functions
const colourPattern = `^#(?:[0-9A-Fa-f]{3,4}|[0-9A-Fa-f]{6}|[0-9A-Fa-f]{8})$|\b(?:rgba?|hsla?|hwb|lab|lch|oklab|oklch|color-mix)\(`

const colourMessage = "a colour is a design token: write it once, with the tokens, and name it here"

const noColoursDescription = `refuse a colour written in a string

A colour, a hex value or a call of a CSS colour function, is written once, with the design tokens, for every
face that draws it; the code names the token. The tokens' own generated code is passed over, as all
generated code is. A hash among words ("fixed in #123", "#paranin") is no colour; a string that is a hex
value whole is taken for one, a word made of hex letters too ("#cafe"). A test may hold the colour it
expects a face to draw.`

func noColoursAnalyzer() *analysis.Analyzer {
	colour := regexp.MustCompile(colourPattern)
	return &analysis.Analyzer{
		Name: "nocolours",
		Doc:  noColoursDescription,
		Run:  func(pass *analysis.Pass) (any, error) { return runStringRule(pass, colour, colourMessage) },
	}
}
