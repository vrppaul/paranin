---
paths:
  - "**/*_test.go"
  - "**/testdata/**"
---

# Tests

## When a test is written

- A test holds a behaviour someone can name: it does one thing — a call, a request, a key, a command — and checks what came of it.
- Its name states that behaviour: `TestPreviewLinkDiesWithItsSecret`, never `TestStore2`.
- First find the test that already holds the behaviour and extend it. A new test only where none does.
- It is written before its code and seen failing. A refactor gets no new test: the old ones pass unchanged.
- A test rewritten is seen failing too: with the line of code it holds broken, then restored.
- Never how it looks: no colour, size or position read back. The text a text face writes is what a terminal reader gets, so a golden file of it is behaviour.
- No test for a declaration, a library's own behaviour, internals or a test helper.
- A rule about how code is written is a gate in pre-commit, never a test.
- "No test holds X" is no finding and no TODO item unless X is a behaviour someone can name.
- A test deleted leaves no debt: nothing is written in its place, neither a test nor a gate.

## At which level

- The cheapest level that sees the behaviour: a package's own test, with no network.
- A behaviour that hangs on time — a date passed, a timeout, a debounce — runs inside `synctest.Test`, never against the real clock and never with a sleep.
- A text face is checked against a golden file in its package's testdata directory, rewritten only by `go test -update` and read in the diff.
- The TUI is tested through its model: keys in, the screen's text out.
- A server is tested through its handler with `httptest`; a real socket only where the behaviour is the socket's.

## How a test is written

- Tests run on fixture content in a testdata directory, never on the real profile or posts: editing the résumé breaks no test. One test checks the real content loads.
- Cases of one rule are one table test, a struct per case with named fields, each run by the table's `Run`.
- A test holds no branch but its assertion: no `if` that chooses what to check, no filter, no switch. A helper that depends on a kind is a lookup table.
- Names are whole words here too: `test *testing.T`, `testCase`, `want`, `got`. `scripts/coderules` reads tests for names, nested helpers and records; the body handed to `Run` is the test, not a helper.
- golangci-lint's linters do not read tests (`run.tests` in `.golangci.yml`); its formatter and `scripts/coderules` do.
- A failure says what was asked, what came and what was wanted, with `%q` for text.

## Runs

- Iterating: the tests of the package in hand. A commit runs the whole suite with the gates: it takes seconds.
- At a chunk's end the author runs everything once, and the reviewer reads that run and repeats only what it doubts.
- CI runs every gate again on what the commit carries.
