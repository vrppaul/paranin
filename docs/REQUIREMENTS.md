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

*Holds* says whether the code meets a requirement and where to read how. A
bound's value and its reason are its constant's, named there.

| ID | Threat (STRIDE) | Requirement | Holds | Priority |
| --- | --- | --- | --- | --- |
| **SEC-1** | *Elevation* — escape TUI to a shell / run commands | Serve the TUI alone: no shell, exec, SFTP or subsystem; a command a client sends is never run | Yes: `internal/hosting/ssh` | [MVP] |
| **SEC-2** | *Elevation* — tunnel via the SSH connection | No TCP/IP forwarding, local or remote, no X11, no agent forwarding | Yes: the server opens `session` channels alone and none to the client, and refuses a request for remote forwarding. The session refuses `x11-req`; `charm.land/ssh` accepts the agent's request, which opens nothing | [v1] |
| **SEC-3** | *DoS* — automated probes without a terminal | Require an active PTY; reject otherwise | Yes: `activeterm` (`internal/hosting/ssh/ssh.go`) | [MVP] |
| **SEC-4** | *DoS* — crash via malformed input/edge case | A panic ends its own session, never the server | Yes: `recovering` and `watchedModel` (`internal/hosting/ssh/ssh.go`); ARCHITECTURE, Phase 1 | [MVP] |
| **SEC-5** | *DoS* — idle/slowloris connections hold resources | Close a connection that starts no session in time, its handshake counted, and a session idle too long or open too long | Yes: `openingTimeout`, `sessionIdleTimeout`, `sessionMaxTimeout` (`cmd/paranin/main.go`); ARCHITECTURE, Phase 1 | [MVP] |
| **SEC-6** | *DoS* — connection floods | Per-IP concurrent-connection cap + global max sessions; connection rate limiting | No: at hosting (`TODO.md`). SEC-7 bounds what one connection costs, not how many connect | [v1] |
| **SEC-7** | *DoS* — resource exhaustion (memory/goroutines) | Bound what each session takes: its terminal's size, one session a connection, its input's pace, the line typed, its environment; leave no goroutine once it closes | Yes: `widestTerminal`, `tallestTerminal`, `inputRate`, `inputBurst`, `colourVariables`, `longestColourValue` (`internal/hosting/ssh/channels.go`), `typedLimit` (`internal/face/tui/keys.go`), the goroutines in `internal/hosting/ssh/ssh_test.go`; ARCHITECTURE, Phase 1 | [v1] |
| **SEC-8** | *Tampering/EoP* — host compromise blast radius | Run as **non-root**, least privilege; dedicated user; container/systemd sandboxing (`NoNewPrivileges`, read-only FS, etc.); host firewall | No: at deploy (`TODO.md`, hosting) | [v1] |
| **SEC-9** | *Info disclosure* — leak host FS / env / secrets via UI | UI renders only `content`; never reflects server FS/env; no secrets in process env beyond what's needed | Yes: the server reads only its addresses and the host key's path from its environment (ARCHITECTURE, Configuration) | [v1] |
| **SEC-10** | *Spoofing* — host-key churn erodes client trust (TOFU) | Persist a stable ed25519 host key on a fixed path; back it up; never commit it | Yes but the backup, which is deploy's (`TODO.md`): `PARANIN_SSH_HOST_KEY` (ARCHITECTURE, Configuration) | [v1] |
| **SEC-11** | *DoS at network edge* — volumetric attacks on SSH port | Hetzner Cloud Firewall; `fail2ban`/throttling; note Cloudflare proxy does **not** cover plain SSH (see ARCHITECTURE) | No: at deploy (`TODO.md`, hosting) | [v1] |
| **SEC-12** | *Tampering* — supply-chain / vulnerable deps | Pin versions; run `govulncheck` and dependency updates in CI | Yes but the updates, which come with the repository's remote (`TODO.md`): `go.mod`, `web/pnpm-lock.yaml`, `.github/workflows/gates.yml`, `.pre-commit-config.yaml` | [v1] |
| **SEC-13** | *Repudiation* — no audit trail for abuse | Structured logs of connects/disconnects/errors with timestamps (privacy-limited per NFR-13) | In part: each session as it opens and closes, and its refused requests (`internal/hosting/ssh`); the server's other refusals not yet (`TODO.md`) | [v1] |
| **SEC-14** | *Tampering* — secrets leak via repo | No secrets in git; `.gitignore` covers keys/.env; config via env | Yes: `.gitignore`, and the `detect-private-key` hook (`.pre-commit-config.yaml`) | [MVP] |
| **SEC-15** | *Tampering* — a client's text forges lines of the log | A client's text reaches the log quoted and cut short | Yes: `longestLoggedText` (`internal/hosting/ssh/ssh.go`), and slog's text handler (`cmd/paranin/main.go`) | [v1] |
| **SEC-16** | *Tampering* — a value of the content acts on a reader's terminal or browser | No value of the content holds a character a terminal or a browser acts on rather than shows | Yes: every load of the content, `paranin check` among them (`internal/content/characters.go`) | [v1] |

### 5.4 Explicitly out of scope (Phase 1)
- User authentication / authorization (the service is intentionally anonymous).
- Encryption of stored data (no user data is stored).
- The LLM-specific threats (prompt injection, cost-abuse) — addressed in Phase 3.

### 5.5 What the parts still to come must meet
Each part meets these when it is built ([FOUNDATION](FOUNDATION.md) §8).

**Posts and their parser.**
- No raw HTML in a post: its markup reaches a face only through the
  document model.
- A link's scheme is one of a list: `https`, `mailto` and the site's own
  paths, checked once, as the content loads.
- A post holds no character a terminal acts on (SEC-16).
- A post's size and depth are bounded; its front matter is a typed struct,
  read with `KnownFields`.

**Statuses and preview links.**
- The repository is public: a preview hides a post's address, not its text.
  An embargo needs a store outside the repository.
- A preview's token is HMAC-SHA256 over the post's address, its language's
  folder in it, and the moment the link lapses, which the token carries,
  compared with `hmac.Equal`; a link lapsed answers as no page. Not over
  the status: the server checks each request against the post's status as
  it stands, and a link signed over
  `preview` could not send its reader on to the post once it is out. Its
  secret is at least 32 bytes, read from a file alone, since a value in
  the environment is read by every child process and through `/proc`, and
  never logged; a new secret ends every link.
- Every answer in the preview's folder, a refusal or a redirect `net/http`
  writes among them, carries `X-Robots-Tag: noindex`,
  `Cache-Control: private, no-store` and `Referrer-Policy: no-referrer`,
  so its token reaches neither Cloudflare nor a site it links to.
- A post before its date answers 404, not 403, and a status worked out for
  one request never reaches a cached feed.

**Widgets' props.**
- Props are checked against their JSON Schema on every load of a snapshot,
  not in CI alone; a fault refuses the whole snapshot.
- Props reach the page as data, in an `application/json` block read with
  `JSON.parse`; never through `{@html}`.
- An island is a file the page's policy allows as it stands: no
  `unsafe-inline`, no `eval`; the build's size is bounded.

**`inbox` and the form.**
- POST alone, its `Origin` checked; the body read through
  `http.MaxBytesReader`, each field bounded, parsed into a typed struct.
- A rate for each client and one for all. Behind Cloudflare the client is
  `CF-Connecting-IP`, trusted only from Cloudflare's addresses, and the
  origin's HTTP open to them alone, since anyone can write
  `X-Forwarded-For`.
- The Telegram token comes from the environment or a file and never
  reaches a log or an error; a message goes to Telegram with no
  `parse_mode`.
- The messages' file is 0600, outside the content folder and bounded in
  size; the log says a message came, never what it says.

**Hosting**, beside SEC-8, SEC-10, SEC-11 and SEC-6, which holds before
port 22 is opened.
- One process holds both faces, so what stops one stops both: systemd's
  `MemoryMax`, `TasksMax` and `LimitNOFILE` bound it, and it restarts on
  failure.
- `ProtectSystem=strict`; `CAP_NET_BIND_SERVICE` for port 22, the host's own
  `sshd` on another port behind the firewall.
- The host key on a lasting volume, never in the image; the
  `.dockerignore` leaves out `.paranin/`, `.ssh/`, `_old/` and `build/`.
- Cloudflare ignores `Vary: Accept`: no "Cache Everything" on an address
  whose format the `Accept` header chooses, or a browser gets the manual
  cached for `curl`. HSTS at the edge.
- The log is rotated.

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
