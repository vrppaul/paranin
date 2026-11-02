---
paths:
  - "**/*.go"
  - "go.mod"
---

# Go

How a package is built is said in its package comment. The rules for code in any language are in `.claude/rules/code.md`.

## The layers

- `cmd` → `hosting` → the faces and `brain` → `compose` and `routes` → `content`, `document`, `look`, as ADR-0003 draws them; `inbox` stands on the foundation beside `compose` and `routes`. Each builds only on the layers after it.
- `routes` reads `document` alone; `compose` reaches `routes` only to read a post's links back into targets.
- The parts of one layer stand side by side: one face never builds on another, nor one host on another, nor `content`, `document` and `look` on each other. What two of them share goes down a layer.
- `content`, `document`, `look`, `compose` and `routes` know no face: no web server, no template engine, nothing of Charm.
- Every part but the composition root lives under the module's `internal` directory. depguard holds all of it, each rule saying why (`.golangci.yml`).

## The language

- Write the Go that `go.mod` names: generics, `any`, `min` and `max`, range over a function or an integer where they say it plainer (modernize).
- `gofumpt` formats every file, tests included (`golangci-lint fmt`).
- Names: `err` and `ctx` are the only abbreviations kept (D11), since `error` and `context` are taken; a receiver is a word, `(model Model)`; a test is `test *testing.T`.
- No name hides one of Go's own, `max` or `len` (predeclared).
- An enumeration is a named type with its constants; a `switch` over one names every value (exhaustive).
- A choice between kinds is a closed set of types behind an interface with an unexported method, marked `//sumtype:decl`; a type switch over it names every kind (gochecksumtype).
- A struct of our own is built with every field written out (exhaustruct), so a new field reaches every place that makes one.
- A function literal handed straight to a call, returned or set as a field is one simple statement. One given a name is a helper, however short; so is one of more statements (coderules).
- The body handed to a test's `Run`, `Fuzz` or `synctest.Test` is the test, and an `iter.Seq` a function returns is the value it makes: neither is a helper.
- No `init`, no variable at package level: a fixed table is a function that returns it, a value that never changes is a constant.
- An interface is declared where it is used, as small as its user needs; a constructor returns the concrete type (ireturn).
- Data with a shape is a struct, never a `map[string]any`: JSON, YAML and JSON-LD come in and go out through typed structs.

## Errors

- An error says what failed and with what, wrapped with `%w` as it goes up: `fmt.Errorf("load post %q: %w", slug, err)`.
- An error is compared with `errors.Is` or `errors.As`, never by its text or with `==` on a wrapped one.
- A refusal the outside world sees is an error of one kind, and one table turns a kind into what the caller gets: a status, an exit code.
- Panic only at the start of the program, where it cannot run without what it lacks; never in a request or a session, except to pass a panic already caught on to the recovery that ends the session (forbidigo).

## Time, logging and the terminal

- Code reads time and waits with the `time` package itself; a test that holds a moment or a wait runs inside `synctest.Test`, whose clock moves only when every goroutine waits (D15).
- Nothing prints but the composition root and the scripts: forbidigo refuses `fmt.Print…`, `os.Stdout`, `os.Stderr` and the package-level loggers elsewhere. A part logs through a structured logger it is handed.
- Whatever came from outside — a path, a header, a form, a key typed over SSH — reaches a log or a terminal quoted (`%q`) or stripped of control characters, never as it came.

## Generated code

- Generated code (templ, the design tokens, the client data's types) is never edited by hand: change its source and run its generator. A hook fails when the two disagree.
- A generated file opens with the line `// Code generated … DO NOT EDIT.`, so every linter and `scripts/coderules` pass over it.
