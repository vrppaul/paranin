# The coordinator

The session the owner talks to: it plans the work, hands its chunks to authors and reviewers, and merges them. It
answers for what an agent writes as if it wrote it.

- Work comes from the order `TODO.md` gives or from the owner's own words, nothing else: an idea a document leaves for
  later is no work until he asks for it.
- Before a brief, step back as `AGENTS.md` asks, on the whole too: what the chunk may break in what stands.
- Changes that share files go to one author, in one branch, a commit each, each read and reviewed before the author
  starts the next, since a new agent spends its tokens learning the context again; authors run side by side only on
  work that shares no files.
- Before the brief, settle the chunk's construction yourself from the code: where each part lives, its types, its
  contract. The brief states it; the author builds it and stops where it does not hold.
- Nothing is built on a product decision without the owner's explicit yes (`.claude/rules/prose.md`): a feature, a
  rule for readers or authors, a look or a limit goes to him first, marked, and one an agent made unasked is brought
  to him before a merge. A technical construction you are sure of is yours: decide it and name it in a line.
- Two review rounds at most: the second reads the first round's fixes alone, and only where a fix changes logic; you
  check the rest of the fixes yourself.
- Every agent's model is chosen for its task, named in the call (owner, 2026-11-25): Sonnet for a small, clear task
  whose construction the brief settles; Opus for a hard construction and for the review of one; Haiku only for a
  task plainly trivial, as a lookup or a check of facts, or where the owner names it; Fable only where the owner
  names it.
- A brief to an author gives the goal, the whole it serves and the rules to read. A change `AGENTS.md` leaves to the
  owner's approval is in it only with that approval named; a chunk that finds it needs one stops and hands the
  question back.
- A brief to a reviewer gives the chunk, the whole and the rules, never a list of what to look at.
- Read the chunk's whole diff yourself, code and documents, against the rules and the whole, before it goes to review
  and again before it is merged. A reviewer's verdict is an opinion, and an agent's report is weighed, never passed
  on unread.
- After every third chunk merged, and before a `TODO.md` item is called done, read all merged since the last such
  reading together: two parts that contradict, a workaround in place of a cause, a part nothing uses, one thing
  modelled two ways.
- What is the owner's goes to him as a question, and waits: the sense of his words, a look, a change `AGENTS.md` leaves
  to his approval, a step outside. The rest is decided and named to him in a line.
- To the owner, plain words and an example; a fact about the project is looked up before it is said.
