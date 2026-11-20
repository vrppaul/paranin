---
paths:
  - "**/*.go"
---

# Go

The rules for code in any language are in `.claude/rules/code.md`.

## The layers

- The parts build on each other in one direction, as ADR-0003 of `docs/ARCHITECTURE.md` draws them; depguard holds each border, saying why (`.golangci.yml`).
- What two parts of one layer share goes down a layer: one face never builds on another, nor one host on another.
- `content`, `document`, `look`, `assets`, `compose`, `routes` and `release` know no face: no web server, no template engine, nothing of Charm. A face never reads `release`: a host hands it what it draws.
- Every part but the composition root lives under the module's `internal` directory.

## The language

- Names: `err` and `ctx` are the only abbreviations kept (D11), since `error` and `context` are taken; a receiver is a word, `(model Model)`; a test is `test *testing.T`.
- A choice between kinds is a closed set of types behind an interface with an unexported method, marked `//sumtype:decl`; a type switch over it names every kind.
- A function literal handed straight to a call, returned or set as a field is one simple statement. One given a name is a helper, however short; so is one of more statements (coderules).
- The body handed to a test's `Run`, `Fuzz` or `synctest.Test` is the test, and an `iter.Seq` a function returns is the value it makes: neither is a helper.
- No variable at package level: a value that never changes is a constant.
- An enumeration is a named type with its constants; a `switch` over one names every value (exhaustive).
- An interface is declared where it is used, as small as its user needs; a constructor returns the concrete type.
- Data with a shape is a struct, never a `map[string]any`: JSON, YAML and JSON-LD come in and go out through typed structs.

## Errors

- An error says what failed and with what, wrapped with `%w` as it goes up: `fmt.Errorf("load post %q: %w", slug, err)`.
- A refusal the outside world sees is an error of one kind, and one table turns a kind into what the caller gets: a status, an exit code.
- Panic only at the start of the program, where it cannot run without what it lacks; never in a request or a session, except to pass a panic already caught on to the recovery that ends the session.

## Time, logging and the terminal

- Code reads time and waits with the `time` package itself; a test that holds a moment or a wait runs inside `synctest.Test`, whose clock moves only when every goroutine waits (D15).
- A part that draws the day, and never waits on it, takes it as a value from the host that reads the clock, as `tui.New` takes the day a session opens; its tests hand it a fixed day. No `Clock` is handed (D15).
- Nothing prints but the composition root and the scripts: a part logs through a structured logger it is handed.
- Whatever came from outside — a path, a header, a form, a key typed over SSH — reaches a log or a terminal quoted (`%q`) or stripped of control characters, never as it came.

## Generated code

- Generated code (templ, the design tokens, the client data's types) is never edited by hand: change its source and run its generator. A hook fails when the two disagree.
