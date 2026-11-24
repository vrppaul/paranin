# paranin

A résumé served as an SSH TUI, a web page, markdown, a man page for `curl` and a PDF, from one source of facts; a blog drawn the same way.

The decisions, D1 onward, and the order of work are in `docs/FOUNDATION.md`; each architectural decision is an ADR in `docs/ARCHITECTURE.md`; where the two differ the ADR holds. Planned work: `TODO.md`.

## Commands

```bash
go test ./...                                 # the tests
go run ./scripts/coderules ./...              # the code rules no linter here holds
pnpm --dir web install                        # the page's tools, as web/pnpm-lock.yaml pins them
pnpm --dir web build                          # the page's build into build/web; `watch` builds on each change, and the site reloads
pnpm --dir web test                           # the page's tests, in jsdom
go run ./cmd/paranin serve content build/web  # the site: curl localhost:8080, ssh -p 23234 localhost
go generate ./internal/look                   # write the design's tokens, Go and CSS, from design/tokens.yaml
go generate ./internal/face/html              # format the templ templates and write their Go
pre-commit run --all-files                    # every gate a commit runs, over every tracked file
pre-commit run --files $(git ls-files -co --exclude-standard)   # the same over files not yet tracked
```

- The gates are one list, `.pre-commit-config.yaml`, each with its command. A hook that cannot run fails: leave one out only on purpose, by naming it in `SKIP`.
- CI, `.github/workflows/gates.yml`, runs the hooks again on what a commit carries, and adds govulncheck, which needs the network, pinned there.
- What a command makes, a command makes: a Go dependency by `go get` and `go mod tidy`, a tool of the page by `pnpm --dir web add`, never a hand edit.
- A tool the hooks run is pinned in `.pre-commit-config.yaml`; a generator whose output must match its library, as templ's, is a `tool` of `go.mod`.

## How we work

- **The owner's word first.** Nothing is changed, run against the repository or committed before the owner says to go ahead, and his word covers the step he named, not the next.
- **Talk short.** A few lines to the owner; details only when asked.
- **Small chunks.** One concept at a time, read in one sitting, working and checked by itself. Never a run through a plan of several phases.
- **The coordinator**, the session the owner talks to, reads `.claude/coordinator.md` first and alone merges into `main`.
- **Step back first.** Coordinator, author and reviewer each say first what the piece is for, which whole it serves, whether it is worth doing, why each part it lives in exists and why nothing simpler does.
- Each weighs it from more than one side, against how the language's and the libraries' own docs do it.
- A part with no reason that holds is a finding, however old. A why you cannot answer is a question to the owner, never code.
- **Each hands over:** an author the change and ten lines at most: what, why, how checked, how long and its slowest step, where to look; a reviewer a verdict and its findings.
- **Understand before writing.** No code without full understanding of what exists and where the change belongs.
- **Tests hold behaviour.** A behaviour someone can name gets a test first, at the cheapest level: `.claude/rules/tests.md`.
- **Review before presenting.** Read every changed file: the right layer? redundant work? the non-obvious commented? every line needed?
- **Weigh, then decide** a dependency, the architecture, the content's or the document's model: name the ways weighed; ask where none is clearly best. A right way and a wrong one are no choice: take the right one and say so.
- **The owner approves** a rewrite of what stands, and any change to `DESIGN.md`, `docs/FOUNDATION.md` or `docs/ARCHITECTURE.md`.
- Never commit secrets, keys or credentials.
- Planned work is tracked in `TODO.md`, where an item done is removed, not ticked off; what changed, in `CHANGELOG.md`.

## Only better

Every change leaves the repository better than it found it: cleaner, easier to read, more modular, its architecture more fitting. In every part it touches, not on balance. A change that only works is not done.

- **Construction first.** Before code, settle who owns each state and for how long, which layer it lives in, and the contract. Build nothing nobody asked for.
- **Simplest first.** Build the simplest thing that works. A new state, layer, table, gate or package says why nothing simpler does; a reviewer's fix is a proposal. Unsure how? Ask the owner, never build it complicated.
- **Root cause first.** Before a fix, say why the fault can happen and change that construction. A flag, guard or filter hiding a wrong value is a symptom fix: it needs a reason.
- **Show it with facts, not adjectives.** Say what a reader can check: what was written several times and is now written once, which state has one owner now, which rule a tool now holds. Say as plainly what did not get better.
- **The verdict is somebody else's.** A chunk is done when an independent reviewer (`.claude/agents/`) has measured its whole diff against this rule and listed every place that got worse. Its author never calls it better.
- **A regression found becomes a gate** a tool enforces at the commit, not a test; a gate of our own is built on the owner's word. A form a checker refused is never answered with a form it cannot see.
- **What you can name as worse is not committed.** Fix it, or stop and ask.

## Rules by area

Before you change or review code in an area, read its file whole; a brief to another agent names the files.

- `.claude/rules/code.md` — any code: names, values, construction, kinds, where a thing lives.
- `.claude/rules/go.md` — Go: the layers, the language, errors, time, logging, what reaches a terminal, generated code.
- `.claude/rules/tests.md` — any test: when one is written, at which level, how, and which run when.
- `.claude/rules/prose.md` — comments, documents, commit messages, reports.
- `.claude/rules/tui.md` — the TUI: what a key does and does not, its words.
- `.claude/rules/page.md` — the page under `web/`: its build, its stylesheets, what its policy refuses.
- `.claude/rules/design.md` — how a look is designed, for the page and the TUI alike.
- `.claude/rules/scripts.md` — anything under `scripts/`: the gates of our own and the generators.
- `.claude/rules/instructions.md` — the instruction files, this one among them: where a line belongs.
