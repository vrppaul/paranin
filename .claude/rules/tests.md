---
paths:
  - "**/*_test.go"
  - "**/testdata/**"
  - "web/tests/**"
---

# Tests

## When a test is written

- A test holds a behaviour someone can name: it does one thing — a call, a request, a key, a command — and checks what came of it.
- Its name states that behaviour: `TestAPostComesOutAtItsMoment`, never `TestStore2`.
- First find the test that already holds the behaviour and extend it. A new test only where none does.
- It is written before its code and seen failing. A refactor gets no new test: the old ones pass unchanged.
- A test rewritten is seen failing too: with the line of code it holds broken, then restored.
- No test written to a change: that a line was edited, or that what the code did before it no longer does. A test holds what the code does.
- Never how it looks: no colour, size or position read back. The text a text face writes is what a terminal reader gets, so a golden file of it is behaviour.
- The drawings the HTML face writes in place of words, the block letters' rectangles and each role's share of the career line, are behaviour too: its goldens hold them.
- The one position a test reads back is the terminal's cursor: it stands where the next letter typed goes. Its shape and colour are looks.
- No test of how an error or a log line is worded: a test reads which fault came, not its words.
- No test for a declaration, a library's own behaviour, internals or a test helper.
- A rule about how code is written is a gate in pre-commit, never a test.
- A test deleted leaves no debt: nothing is written in its place, neither a test nor a gate.

## At which level

- The cheapest level that sees the behaviour: a package's own test, with no network.
- A behaviour that hangs on time — a date passed, a timeout, a debounce — runs inside `synctest.Test`, never against the real clock and never with a sleep.
- A text face is checked against a golden file in its package's testdata directory, rewritten only by `go test -update` and read in the diff.
- The TUI is tested through its model: keys in, the screen's text out, its colours stripped.
- A server is tested through its handler with `httptest`; a real socket only where the behaviour is the socket's.

## How a test is written

- Tests run on fixture content in a testdata directory, never on the real profile or posts: editing the résumé breaks no test. One test checks the real content loads.
- Cases of one rule are one table test, a struct per case with named fields, each run by the table's `Run`.
- A test holds no branch but its assertion: no `if` that chooses what to check, no filter, no switch. A helper that depends on a kind is a lookup table.
- Of `.claude/rules/code.md` a test keeps the names, the records and no helper inside a function.
- A failure says what was asked, what came and what was wanted, with `%q` for text.

## Runs

- The tests of the packages a change touches, never the whole suite: no commit and no gate runs all of them.
- The cards' tests draw with Typst, which `go run ./scripts/typst` puts in the folder it is given, ~/.local/bin.
- At a chunk's end the author runs the tests of every package it touched once, and the reviewer reads that run and
  repeats only what it doubts.
- CI runs every gate again on what the commit carries.
- A review's findings about tests are two: a test to delete (a look, a twin, internals, one that cannot fail) and a behaviour no test holds.
