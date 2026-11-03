// Package comments holds comment lines of every width the rule tells apart: none here is reported but where a
// line says so.
package comments

// This line of prose is exactly as wide as a comment's line may be, one hundred and twenty columns, so it still passes.

// This line holds a dash — three bytes of it, but one letter — and a dot · as well, and is just one hundred and twenty.

// This line of prose runs past the width a comment's line may take, as a line nobody thought to wrap at the edge does, every time. // want `a comment's line is 1[0-9]{2} columns wide`

func wrapped() int {
	// Opened by a tab, which counts as a column, as golangci-lint's lll counts it, this line is a hundred and twenty long.
	return 0
}

func unwrapped() int {
	// Opened by a tab, this comment line still runs past the width that a line of prose may take in the code, by a word or two. // want `a comment's line is 1[0-9]{2} columns wide`
	return 0
}

var wideCode = []string{"a line of code may run past the width", "as a literal with every field written does"} // a note past the code

/* A block comment opening its line is held by its first line, and this one runs past the width a line of prose may take, too. // want `a comment's line is 1[0-9]{2} columns wide`
*/
