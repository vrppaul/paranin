---
paths:
  - "**/*.go"
  - "scripts/**"
  - "web/src/**"
---

# Code

The rules for code in any language here; Go's own are in `.claude/rules/go.md`, the page's in `.claude/rules/page.md`.

## Names

- A name says what it holds, in whole words: `writer`, `request`, `rowIndex`, never `w`, `r`, `i`, `seg`, `cfg`. Coordinates are no exception: `nodeX`, `offsetY`, never `x`, `y`.
- A few short words are kept, listed once, in `scripts/coderules/names.go`.
- A name the gate lets pass still says what it holds: `data`, `info`, `handle`, `process` and `value2` say nothing.

## Values

- No literal with a meaning in an expression: a number, a colour, a time, a size, an opacity or a key is a named constant, or data, with a comment saying what it is.
- A record is a struct with named fields, never an array or a slice standing in for one (`[2]string`); a place is the thing it names, never its index in a list.
- Data with a shape is typed wherever it goes: in, out, and in tests.

## Construction

- A function does one thing; no function does everything.
- No function inside a function: a helper is a function of its package, with a name and parameters of its own. A short lambda handed to a call is fine; a helper with logic inside another function is not.
- No function only passes its call on, fixed arguments or none, unless it names a notion its call lacks or hides real work: shorter to write is no reason.
- Such a function's reason stands on the line above it as a Go directive, which no doc prints: `//wrapper:` and the reason, no space after the colon, in small letters.
- A method an interface asks for needs no reason.
- No state at package level: a fixed table is a function that returns it, and what changes is held by whoever is handed it.
- Dependencies are explicit, a constructor's or a function's parameters, so each part is tested by itself.
- Something derived is worked out once, where its inputs change, never again in a hot path: a key press, a request, a frame.
- No alias or shim kept for an old name: whoever uses it is changed.
- Shared code comes with its second user, not before; once there are two, it is written once.
- Code nearby that breaks these is no model: write it down in `TODO.md` and tell the owner, never copy it.

## Kinds and faces

- A new face or a new kind of post is added code: a table or a switch that names every kind (exhaustive, gochecksumtype) takes it in, never a default or an `else` that swallows it.

## Where a thing lives

- HTML lives in a templ component, a style in a stylesheet, a colour in the design tokens: never in a string in code. A text face (markdown, the man page, Typst) writes its own format, from the document model.
- Facts live in the content alone: no name, number or word of the résumé in code, a template or a script.
