# Roadmap

The project is sequenced by **coolness × simplicity**: start with the most
distinctive, lowest-effort win, defer the complex/risky/less-cool work.

## Phases

| Phase | Feature | Cool | Effort | Status |
| --- | --- | --- | --- | --- |
| 1 | **SSH TUI** | ★★★★★ | ★★★ | 🔁 first version built; being rebuilt |
| 2 | Web résumé (normies / vim modes) | ★★★★ | ★★★ | 🔁 first version built; being rebuilt |
| 2b | Blog: articles with widgets, release notes, read in every face | ★★★★ | ★★★ | planned |
| 3 | The brain (LLM chat) | ★★★★ | ★★★ | planned |
| 4 | Chat over SSH | ★★★★★ | ★★ | planned |
| 5 | Dynamic CV generator | ★★★ | ★★★★ | planned |
| 6 | LinkedIn helper | ★★ | ★★★★★ | deferred (ToS/ethics risk) |

Since 2026-11-01 the first version of phases 1 and 2 is being rebuilt on the
construction of [FOUNDATION](FOUNDATION.md); its §8 holds the order of work, the
blog included.

Each phase is independently shippable. Nothing in a later phase blocks an
earlier one. The content, the one source of facts, is shared by all phases.

## Phase 1 milestones (SSH TUI)

- **M1 — Hello TUI (local).** Wish server serves a Bubble Tea program on
  `localhost:23234`; `ssh` in and see a styled screen. ✅ target of the MVP
- **M2 — Navigable content.** Menu + sections (About, Experience, Projects,
  Skills, Contact) driven by the `content` package. ✅ target of the MVP
- **M3 — Polish.** Theme, layout, responsive to window resize, help view.
- **M4 — Hardening.** Idle/absolute timeouts, per-IP connection caps, panic
  recovery, rate limiting, structured logging. (See REQUIREMENTS §5.)
- **M5 — Deploy.** Static binary → Docker → Hetzner VPS; DNS for `paran.in`
  (SSH strategy per ARCHITECTURE), systemd unit, host-key persistence.

The **MVP** delivered first covers M1 + M2 with the easy parts of M3 and the
cheap parts of M4 (PTY required, panic recovery, timeouts).

## Later phases (summary)

- **Phase 2 — Web résumé (normies / vim modes).** One Go server renders the
  profile as HTML with a normies mode and a keyboard-driven vim mode, as
  markdown for agents, as a man page for `curl`, and as `cv.pdf`. Has a
  contact form. Behind Cloudflare. Reuses `content`. Contract:
  [WEB](WEB.md).
- **Phase 3 — The brain.** LLM (Claude API) grounded in `content`, with
  guardrails (on-topic, no invented facts, not a free LLM proxy), rate/cost caps.
- **Phase 4 — Chat over SSH.** Surface the brain inside the SSH TUI.
- **Phase 5 — Dynamic CV.** Recruiter pastes a job description → tailored PDF
  generated from the master profile (never inventing experience).
- **Phase 6 — LinkedIn helper.** Draft-and-approve assistant for recruiter
  messages. Deferred due to LinkedIn ToS, ban risk, and disclosure/ethics.
