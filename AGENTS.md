# paranin

A résumé served as an SSH TUI, a web page, markdown, a man page for `curl` and a PDF, from one source of facts; a blog drawn the same way.

The review the rebuild began from, the decisions, D1 onward, and the order of work are in `docs/FOUNDATION.md`; each architectural decision is an ADR in `docs/ARCHITECTURE.md`; where the two differ the ADR holds. Planned work: `TODO.md`.

## Commands

```bash
go test ./...                                           # the tests
go run ./scripts/coderules ./...                        # the code rules no linter here holds
pnpm --dir web install                                  # the page's tools, as web/pnpm-lock.yaml pins them
pnpm --dir web build                                    # the page's build into build/web; `watch` builds again on each change
go run ./cmd/paranin serve content build/web            # the site: curl localhost:8080, ssh -p 23234 localhost
go generate ./internal/look                             # write the design's colours, Go and CSS, from design/tokens.yaml
go generate ./internal/face/html                        # format the templ templates and write their Go
pre-commit run --all-files                              # every gate a commit runs, over every tracked file
pre-commit run --files $(git ls-files -co --exclude-standard)   # the same over files not yet tracked
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 fmt   # write the formatting the gate asks for
```

- The gates are one list, `.pre-commit-config.yaml`, each with its command. A hook that cannot run fails: leave one out only on purpose, by naming it in `SKIP`.
- CI, `.github/workflows/gates.yml`, runs the hooks again on what a commit carries, and adds govulncheck, which needs the network, pinned there.
- Add a Go dependency with `go get`, and let `go mod tidy` write `go.mod`: never a hand edit. A tool the hooks run is pinned in `.pre-commit-config.yaml`; a generator whose output must match its library, as templ's, is a `tool` of `go.mod`.
- Add a tool of the page with `pnpm --dir web add`, which writes `web/package.json` and its lockfile: never a hand edit.

## How we work

- **The owner's word first.** Nothing is changed, run against the repository or committed before the owner says to go ahead, and his word covers the step he named, not the next.
- **Small chunks.** One logical piece at a time, presented for review. Never a run through a plan of several phases.
- **Understand before writing.** No code without full understanding of what exists and where the change belongs.
- **Tests hold behaviour.** A behaviour someone can name gets a test first, at the cheapest level; a look, a refactor or a declaration gets none.
- **Review before presenting.** Read every changed file: the right layer? redundant work? the non-obvious commented? every line needed?
- **Test runs follow the change.** Which tests run when: `.claude/rules/tests.md`.
- **Speed is measured.** Say with each piece how long it took and what its slowest step was.
- **Ask first** before adding a dependency, changing the architecture, or changing what the content or the document model is made of, where two ways are equally good or the choice is a real trade-off.
- **A right way and a wrong one are no choice.** Take the right one and say what changed; never bring the owner a question you can answer.
- Never commit secrets, keys or credentials.
- A commit follows Conventional Commits, `type(scope): description`.
- Every sentence of a commit message is a claim: look before you write it, and read a count off a command's output.
- Planned work is tracked in `TODO.md`, where an item done is removed, not ticked off; what changed, in `CHANGELOG.md`.

## Only better

Every change leaves the repository better than it found it: cleaner, easier to read, more modular, its architecture more fitting. In every part it touches, not on balance. A change that only works is not done.

- **Construction first.** Before code, settle who owns each state and for how long, which layer it lives in, and the contract. Build nothing nobody asked for.
- **Simplest first.** Build the simplest thing that works. A new state, layer, table, gate or package says why nothing simpler does; a reviewer's fix is a proposal. Unsure how? Ask the owner, never build it complicated.
- **Root cause first.** Before a fix, say why the fault can happen and change that construction. A flag, guard or filter hiding a wrong value is a symptom fix: it needs a reason.
- **Show it with facts, not adjectives.** Say what a reader can check: what was written several times and is now written once, which state has one owner now, which rule a tool now holds. Say as plainly what did not get better.
- **A review steps back.** Ask why each part the change lives in exists at all, from more than one side, against how the language's and the libraries' own docs do it. One with no reason that holds is a finding, however old.
- **The verdict is somebody else's.** A chunk is done when an independent reviewer, `.claude/agents/reviewer.md`, has measured its whole diff against this rule and listed every place that got worse. Its author never calls it better.
- **A regression found becomes a gate.** What a review catches turns into a rule a tool enforces at the commit, not a test. A form a checker refused is never answered with a form it cannot see.
- **What you can name as worse is not committed.** Fix it, or stop and ask.

## Rules by area

Before you change or review code in an area, read its file whole; a brief to another agent names the files.

- `.claude/rules/code.md` — any code: names, literals, helpers, records, markup, state, dependencies, facts.
- `.claude/rules/go.md` — Go: the layers, the language, errors, time, logging, what reaches a terminal, generated code.
- `.claude/rules/tests.md` — any test: when one is written, at which level, how, and which run when.
- `.claude/rules/scripts.md` — anything under `scripts/`: the gates of our own.
- `.claude/rules/instructions.md` — this file and the rules of each area.
- `.claude/rules/tui.md` — the TUI: what is laid out and painted when, where its colours come from, its keys.
- `.claude/rules/page.md` — the page under `web/`: its build, its stylesheets, what its policy refuses.
- The rules of the design arrive with its code.
