# Security

paranin serves a terminal app over SSH to anyone who connects, and a web page. A fault in either reaches a public
server, so please report it privately.

## Reporting

Open a private report through this repository's **Security → Report a vulnerability** on GitHub. Say what you
did, what happened and what you expected; a command that shows it is the best report.

Please do not open a public issue for a vulnerability, and do not test against `paran.in` beyond what shows the
fault: no load, no scans of the host.

## What is in scope

- The SSH face: anything that gets past the TUI to a shell, a file, another session, the host key, or that
  holds the server's resources past the bounds of one session in `docs/REQUIREMENTS.md` §5.3. A flood of
  connections is known: its caps, SEC-6, come with hosting.
- The web face: the answers by content type, the headers, anything that runs a script the page did not ship.
- The content check: a value that passes `paranin check` and reaches a reader's terminal or browser as more
  than text.
- The gates and the build, where they would let a secret or a key into the repository.

The threat model of the SSH face is `docs/REQUIREMENTS.md` §5.
