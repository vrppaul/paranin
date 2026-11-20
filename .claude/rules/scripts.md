---
paths:
  - "scripts/**"
---

# The scripts

The gates of our own, which no linter here can be told to keep, and the generators.

- A script says at its top why it exists and how it is run, what it prints and every code it exits with.
- It is run in the repository's root and does its whole job in one process start.
- `scripts/coderules` and `scripts/commitmessage` are Go programs of this module, and `scripts/history` what they share: every gate reads them as it reads the site, and their tests run their rules on samples.
- A gate taken from claude-review in part says so at its top, and what it left out.
- `scripts/check_instruction_files.py` is claude-review's, taken whole and byte for byte, so it says nothing of this project: it is changed there and copied here, and its tests are there.
- A gate never changes the working tree: what it would change, it changes in a copy.
- A generator writes its output only when asked to, and checks it otherwise: `scripts/tokens` writes `internal/look/palette.go`, `web/src/tokens/tokens.css` and `web/src/tokens/times.ts` with `-write`, and its hook runs it bare.
