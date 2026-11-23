package main

import (
	"fmt"
	"go/format"
	"strings"
)

// How a block of the page's script's numbers closes: one object of them, read only
const scriptClosing = "} as const;\n"

// namedNumber is a number of the tokens a face's code reads as a constant: its token's name, what it is for, and the
// number, in the unit its file says.
type namedNumber struct {
	name   string
	about  string
	number uint
}

// goConstants is the Go of numbers after the opening of their file, which opens their block of constants: a constant
// of each, under what it is for, the block closed, formatted as gofmt would.
func goConstants(opening string, numbers []namedNumber) ([]byte, error) {
	var builder strings.Builder
	builder.WriteString(opening)
	for _, named := range numbers {
		name := constantName(named.name)
		fmt.Fprintf(&builder, "\t// %s is %s\n\t%s = %d\n", name, named.about, name, named.number)
	}
	builder.WriteString(")\n")
	formatted, err := format.Source([]byte(builder.String()))
	if err != nil {
		return nil, fmt.Errorf("format the generated constants: %w", err)
	}
	return formatted, nil
}

// scriptNumbers is the page's module of numbers after its opening, which opens their object: each a property, the
// number under what it is for, the object closed.
func scriptNumbers(opening string, numbers []namedNumber) []byte {
	var builder strings.Builder
	builder.WriteString(opening)
	for _, named := range numbers {
		fmt.Fprintf(&builder, "  /** %s */\n  %s: %d,\n", named.about, scriptName(named.name), named.number)
	}
	builder.WriteString(scriptClosing)
	return []byte(builder.String())
}
