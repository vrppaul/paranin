# Requirements

Scope of this document: **Phase 1 — the SSH TUI.** Later phases are sketched in
[ROADMAP](ROADMAP.md) and will get their own requirements when started.

Each requirement has an ID and a priority:
- **[MVP]** — in the first working version.
- **[v1]** — required before going public on `paran.in`.
- **[later]** — desirable, not blocking v1.

---

## 1. Purpose & success criteria

`paranin` presents the author's professional profile as an interactive terminal
UI reachable over SSH. Success for Phase 1:

- A stranger can run `ssh paran.in`, with no setup, and explore a polished,
  responsive TUI of the author's background.
- It is memorable enough to be a differentiator for recruiters/engineers.
- It is safe to expose publicly (see §5).

## 2. Personas

- **P1 — Technical evaluator.** Engineer / eng-manager / technical recruiter who
  tries `ssh paran.in` out of curiosity. Wants fast, clear, navigable info.
- **P2 — The author (dogfooding).** Runs locally, edits content, iterates.
- **P3 — The adversary.** Bots and randoms scanning the open port. *Not a user* —
  but the driver of §5. Assume hostile, automated, and high-volume.

## 3. Functional requirements

### 3.1 Access & connection
- **FR-1 [MVP]** Anyone can connect anonymously over SSH; no account, password,
  or pre-shared key required. Any/empty credentials are accepted for the purpose
  of establishing the session.
- **FR-2 [MVP]** The server only ever presents the TUI. Shell, command exec,
  SFTP/SCP, and port/agent/X11 forwarding are **not** offered (see §5).
- **FR-3 [MVP]** A connection without a usable interactive terminal (PTY) is
  rejected gracefully.
- **FR-4 [v1]** Bind host/port/host-key path are configurable via environment.

### 3.2 Navigation & content
- **FR-5 [MVP]** On connect, the user sees a home view with a navigable menu of
  sections: **About, Experience, Projects, Skills, Contact.**
- **FR-6 [MVP]** Keyboard navigation: `↑/↓` and `j/k` to move; `enter` to open a
  section; `esc` to return home; `q` / `ctrl+c` to quit; `?` toggles help.
- **FR-7 [MVP]** Section content is rendered from a single `content` source of
  truth, not hardcoded into views.
- **FR-8 [MVP]** Long content scrolls within the view.
- **FR-9 [v1]** Contact section exposes real, public contact channels (site,
  GitHub, email) and is easy to copy/read.
- **FR-10 [later]** A `cv` / download affordance (ties into Phase 5).
- **FR-11 [later]** A `chat`/`ask` command (ties into Phase 3/4).

### 3.3 Presentation
- **FR-12 [MVP]** The UI is styled (color theme, layout) and uses the alternate
  screen so it restores the user's terminal on exit.
- **FR-13 [MVP]** The UI adapts to the terminal size and to live window resizes.
- **FR-14 [v1]** Graceful degradation on limited terminals (low/no color);
  remains readable without truecolor.
- **FR-15 [v1]** A visible footer/help line shows the key bindings.

### 3.4 Operability
- **FR-16 [v1]** Structured logs for connect/disconnect and errors.
- **FR-17 [v1]** Clean startup and graceful shutdown (drains/【closes sessions).

## 4. Non-functional requirements

### 4.1 Performance
- **NFR-1 [MVP]** Time-to-first-render after connect is effectively instant
  (< ~200 ms on a normal connection); rendering uses the v2 renderer.
- **NFR-2 [v1]** Input-to-redraw latency is imperceptible for navigation.

### 4.2 Scalability & resources
- **NFR-3 [v1]** Handle a realistic burst of concurrent sessions (target: ≥ 100)
  on a small VPS without degradation.
- **NFR-4 [v1]** Bounded memory/goroutines per session; no leaks across the
  lifetime of the process (sessions fully cleaned up on disconnect).

### 4.3 Availability
- **NFR-5 [v1]** Single-host best-effort availability; a crash in one session
  never takes down the server (panic isolation). Process is restarted by
  systemd on failure.

### 4.4 Usability & accessibility
- **NFR-6 [v1]** Works across common SSH clients and terminals (incl. mobile
  clients like Termius), and across window sizes down to ~80×24.
- **NFR-7 [later]** Keyboard-only operation throughout; consider a plain/low-fi
  mode for screen readers and constrained terminals.

### 4.5 Maintainability & learnability
- **NFR-8 [MVP]** Idiomatic, well-organized Go (`gofmt`, `go vet` clean);
  separation between `server`, `tui`, and `content`.
- **NFR-9 [MVP]** Because a goal is *learning Go*, code favors clarity over
  cleverness and is commented where it teaches a Go/TUI concept.
- **NFR-10 [v1]** Unit tests for TUI logic and content; CI runs build/vet/test.

### 4.6 Portability & cost
- **NFR-11 [v1]** Ships as a single static binary; containerizable; runs on a
  small (cheap) Hetzner VPS.
- **NFR-12** No paid third-party dependencies required for Phase 1.

### 4.7 Privacy
- **NFR-13 [v1]** Anonymous by design; collect the minimum. If connection
  metadata (IP, key fingerprint, timestamp) is logged, it is for operational/
  security purposes only, documented, and retention-limited.

## 5. Security requirements & threat model

### 5.1 Trust model
The server is **public, anonymous, and faces wholly untrusted clients** (P3).
Every connection is potentially hostile and automated. The guiding principle is
**"kiosk, not a shell": deny everything except rendering the TUI.**

### 5.2 Assets to protect
- The host machine and its OS (no code/command execution by visitors).
- The server process availability (resist DoS).
- The SSH **host private key** (integrity/confidentiality).
- Author/visitor data minimization (privacy).

### 5.3 Threats → controls

| ID | Threat (STRIDE) | Control / requirement | Priority |
| --- | --- | --- | --- |
| **SEC-1** | *Elevation* — escape TUI to a shell / run commands | Only `bubbletea` middleware; **no** shell/exec/SFTP/subsystem handlers; no PTY-shell fallthrough | [MVP] |
| **SEC-2** | *Elevation* — tunnel via the SSH connection | Disable TCP/IP (local & remote) forwarding, X11, agent forwarding | [v1] |
| **SEC-3** | *DoS* — automated probes without a terminal | Require an active PTY (`activeterm`); reject otherwise | [MVP] |
| **SEC-4** | *DoS* — crash via malformed input/edge case | Per-session `recover` middleware; one session cannot crash the server | [MVP] |
| **SEC-5** | *DoS* — idle/slowloris connections hold resources | Idle timeout + absolute max session duration | [MVP] |
| **SEC-6** | *DoS* — connection floods | Per-IP concurrent-connection cap + global max sessions; connection rate limiting | [v1] |
| **SEC-7** | *DoS* — resource exhaustion (memory/goroutines) | Bounded per-session resources; verified no goroutine leak on disconnect | [v1] |
| **SEC-8** | *Tampering/EoP* — host compromise blast radius | Run as **non-root**, least privilege; dedicated user; container/systemd sandboxing (`NoNewPrivileges`, read-only FS, etc.); host firewall | [v1] |
| **SEC-9** | *Info disclosure* — leak host FS / env / secrets via UI | UI renders only `content`; never reflects server FS/env; no secrets in process env beyond what's needed | [v1] |
| **SEC-10** | *Spoofing* — host-key churn erodes client trust (TOFU) | Persist a stable ed25519 host key on a fixed path; back it up; never commit it | [v1] |
| **SEC-11** | *DoS at network edge* — volumetric attacks on SSH port | Hetzner Cloud Firewall; `fail2ban`/throttling; note Cloudflare proxy does **not** cover plain SSH (see ARCHITECTURE) | [v1] |
| **SEC-12** | *Tampering* — supply-chain / vulnerable deps | Pin versions (`go.sum`); run `govulncheck` and dependency updates in CI | [v1] |
| **SEC-13** | *Repudiation* — no audit trail for abuse | Structured logs of connects/disconnects/errors with timestamps (privacy-limited per NFR-13) | [v1] |
| **SEC-14** | *Tampering* — secrets leak via repo | No secrets in git; `.gitignore` covers keys/.env; config via env | [MVP] |

### 5.4 Explicitly out of scope (Phase 1)
- User authentication / authorization (the service is intentionally anonymous).
- Encryption of stored data (no user data is stored).
- The LLM-specific threats (prompt injection, cost-abuse) — addressed in Phase 3.

### 5.5 Security review outcomes (2026-06-02)
A self-review (independent source audit of the Wish / charmbracelet-ssh modules
plus empirical probes against a running server) confirmed the core posture and
surfaced a few tracked items:

- **Confirmed — "kiosk, not a shell" holds (SEC-1/SEC-2).** With only the
  bubbletea/activeterm/recover/logging middleware, the stack denies *by default*:
  interactive shell, command exec, SFTP/subsystems, local/remote/dynamic port
  forwarding, direct-tcpip/ProxyJump, X11 and agent forwarding. Client-supplied
  commands are logged but never executed (no `os/exec` in the tree). Verified in
  source and empirically (`ssh host whoami` → "Requires an active PTY";
  `ssh -t host id` → TUI, `id` not run; `sftp` → "subsystem request failed").
- **SEC-6 [v1] — gate for public exposure.** No per-IP / global connection caps
  or rate limiting yet; a connection flood can exhaust the VPS. Add before
  opening `paran.in` (e.g. `charm.land/wish/v2/ratelimiter` + a session cap).
  MVP is local, so this does not block it.
- **SEC-15 [v1] — log injection (low).** The logging middleware records the
  client-supplied command/term strings unsanitized; a hostile client can embed
  newlines/control bytes to forge log lines. Sanitize before public launch.
- **SEC-12 — supply-chain scan run & fixed.** `govulncheck` initially flagged 7
  vulnerabilities reachable via the SSH accept path (all in `golang.org/x/crypto`
  ≤ v0.50.0, incl. a client-triggerable transport-layer panic that sits below
  the recover middleware); upgraded to v0.52.0 → scan clean. Wire `govulncheck`
  into CI for [v1].
- **Deploy note (SEC-8).** `docker-compose.yml` builds the `dev` stage, which
  runs as root with the source bind-mounted — dev only. Production must deploy
  the distroless, non-root `runtime` image target.

## 6. Constraints & assumptions
- **C-1** Implementation language is Go (also a learning goal).
- **C-2** Stack is Charm v2 (`charm.land/*`); see ADR-0001.
- **C-3** Single developer; favor simplicity and incremental delivery.
- **C-4** Hosting: Hetzner VPS; Cloudflare in front of the web face only.
- **A-1** Visitor volume is modest (personal site), but bursts/abuse are possible.

## 7. Open questions / decisions log
- **OQ-1** SSH exposure strategy at deploy (DNS-only vs Spectrum vs Tunnel).
  *Leaning: DNS-only + host hardening.* (ARCHITECTURE §Deployment)
- **OQ-2** Public TUI port: standard `22` (max discoverability) vs a custom port
  (less scan noise) with `22` reserved for admin. *Leaning: TUI on 22, admin SSH
  moved + locked down.*
- **OQ-3** Identify repeat visitors by key fingerprint (lightweight analytics)
  vs zero correlation? *Leaning: anonymous, minimal logging for now.*
- **OQ-4** Real profile content — to be supplied by the author (MVP ships with
  clearly-marked placeholder content).
