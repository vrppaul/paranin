// Command coderules holds the rules of AGENTS.md and .claude/rules/code.md that no linter here can be told to
// keep, each an analyzer of its own:
//
//   - wholenames: every name the code declares is made of whole words. golangci-lint's varnamelen passes a
//     short name used near where it is declared and never reads a struct's fields.
//   - nonestedhelpers: a helper is a function of its package, not a literal with logic inside another function.
//   - norecordarrays: a record is a struct, not an array of words.
//   - nomarkup and nocolours: markup lives in a template or a component, a colour in the design tokens.
//   - nonumbers: a number with a meaning is a named constant. golangci-lint's mnd passes a number assigned to a
//     variable.
//
// The first two are the Go form of the first two rules of claude-review's scripts/check_code_rules.py;
// its rules on pydantic's models, asyncio's processes, click's commands and the kinds of refusal are
// Python's and are left out. The rest are this project's. Generated files are their generator's and are
// passed over. Test files are read for names, helpers and records, and may hold the markup, the colours and
// the numbers they expect.
//
// Run it in the repository's root, as the pre-commit hook does:
//
//	go run ./scripts/coderules ./...
//
// It prints each finding as file:line:column: message. Built and run itself, it exits 0 when every rule holds,
// 3 when it reported a finding, and 1 when the code could not be loaded or checked; under go run, as the hook
// runs it, any failure exits 1.
package main

import "golang.org/x/tools/go/analysis/multichecker"

func main() {
	multichecker.Main(
		wholeNamesAnalyzer(),
		noNestedHelpersAnalyzer(),
		noRecordArraysAnalyzer(),
		noMarkupAnalyzer(),
		noColoursAnalyzer(),
		noNumbersAnalyzer(),
	)
}
