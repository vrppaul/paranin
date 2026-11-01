package main

import (
	"regexp"

	"golang.org/x/tools/go/analysis"
)

// What only markup writes: a tag with an attribute, a closing tag, or a tag that closes itself. A bare <word>
// is left alone, since usage text writes one as a word to fill in.
const markupTagPattern = `<[A-Za-z][\w-]*\s+[\w:-]+\s*=|</[A-Za-z][\w-]*\s*>|<[A-Za-z][\w-]*(?:\s[^<>]*)?/>`

const markupMessage = "markup lives in a template or a component, not in a string"

const noMarkupDescription = `refuse markup written in a string

A string that holds a tag with an attribute, a closing tag or a tag that closes itself is HTML kept in code,
where no template checks it and no stylesheet sees it. HTML lives in a templ component. A bare word in angle
brackets ("<slug>") is a word of usage text, whatever the word, and markdown's autolink is a link. A string
put together from pieces, or by Sprintf, is past what the rule sees. A test may hold the markup it expects a
page to carry.`

func noMarkupAnalyzer() *analysis.Analyzer {
	tag := regexp.MustCompile(markupTagPattern)
	return &analysis.Analyzer{
		Name: "nomarkup",
		Doc:  noMarkupDescription,
		Run:  func(pass *analysis.Pass) (any, error) { return runStringRule(pass, tag, markupMessage) },
	}
}
