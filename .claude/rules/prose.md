---
paths:
  - "**/*"
---

# Prose

Every word that is not code: a comment, a document, a commit message, a report, a note handed over. The owner's own rules on writing, the "Writing" section of his global instructions, hold first.

- What is not obvious gets a comment: one line that says why. What the code says, it does not say again.
- A package's comment says why the package exists, in a few lines; any other, what the thing is for, in one. How it works is read in the code.
- What a change writes is held to these two; a longer comment already there stays until a change rewrites it.
- A comment or a document says what stands, never what was: no "no longer", "used to", "where it was". What changed is said in a commit and in `CHANGELOG.md`.
- No date, probe, review, session or commit hash in a comment.
- No measurement there but the number a constant or a choice rests on, said once where it is used.
- A way not taken is one line, where the choice is made.
- A commit follows Conventional Commits and does the one thing its subject says; the subject says nothing of what it replaces.
- A commit's body is for a why the subject cannot hold, in a few lines: no list of files, no tally of tests or lines.
- A decision a document, a report or a brief records says whether the owner gave it his explicit yes: `(owner,
  2026-11-24)` where he did, `(not approved)` where he did not.
- Every sentence of a commit message or a report is a claim: look before you write it, and read a count off a command's output.
- A sentence a reader loses nothing without is cut.
