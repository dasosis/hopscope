# CLAUDE.md

Guidance for AI coding assistants working on hopscope. Read this before making
changes.

## Project context

hopscope is a local network diagnostics agent written in Go: it explains *why*
a connection feels slow (bufferbloat, jitter, loss, slow DNS, weak Wi-Fi), layer
by layer. It is a **portfolio project**. The author must be able to understand
and defend every line in a technical interview. Optimize for clarity and
correctness over cleverness.

- Module: `github.com/dasosis/hopscope`, Go ≥ 1.26 (see `go.mod`).
- Targets: Windows and Linux are first-class; macOS is best effort.
- Plan and progress: `ROADMAP.md`. Design decisions: `docs/decisions/`.

## How to work with the author

- **Explain non-obvious decisions as you go**: why this approach, what the
  alternatives were, what the trade-off is. Assume the author will be asked
  "why did you do it this way?" in an interview.
- **Respect "write it yourself" milestones** (marked in `ROADMAP.md`). Do not
  implement them unless explicitly asked. Act as a reviewer and tutor instead:
  explain concepts, point out bugs, suggest tests, and show small examples
  rather than full implementations.
- Record significant decisions as a new ADR in `docs/decisions/` (copy
  `0000-template.md`). Significant means a new dependency, a privilege or
  security trade-off, a storage format, or a public CLI/JSON contract.
- **Append an entry to `docs/ai-usage.md` at the end of every session**: what
  was delegated, what the author should review, and what the author wrote.
  Never claim the author reviewed something; leave that for them to fill in.
- Keep changes scoped to the current milestone. Do not scaffold future phases
  early.

## Hard rules

1. **No high-level networking libraries.** ICMP, traceroute, DNS-query and
   gateway code is written in this repo on top of the standard library,
   `golang.org/x/net` (`icmp`, `ipv4`, `ipv6`, `dns/dnsmessage`, `route`) and
   `golang.org/x/sys`. Do not add ping, traceroute, DNS-client or netlink
   wrapper packages (e.g. `go-ping`, `pro-bing`, `miekg/dns`). The low-level
   networking is the point of the project.
2. **Least privilege.** Never require root or Administrator when an
   unprivileged path exists. On Linux the ceiling is `CAP_NET_RAW`. Never
   self-escalate (no `sudo` calls, no UAC prompts). Always report which probe
   mode was used. See ADR 0003.
3. **No telemetry.** Nothing leaves the machine except probes to the targets
   the user chose or the documented defaults. The dashboard binds to loopback
   only.
4. **No cgo.** Builds use `CGO_ENABLED=0` so every platform cross-compiles from
   one machine. That rules out dependencies that require cgo (see ADR 0002).
5. **New dependencies need justification** in the PR/commit message, plus an
   ADR if they are significant. Prefer the standard library.

## Go style

- `gofmt`/`goimports` formatting, enforced by golangci-lint (`.golangci.yml`).
  Imports are grouped as stdlib, third-party, then `github.com/dasosis/hopscope/...`.
- Follow [Effective Go](https://go.dev/doc/effective_go) and the
  [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments).
- All code lives under `internal/`; `cmd/hopscope` only wires the CLI together.
  CLI code (`internal/cli`) parses flags and renders output. It does not
  contain measurement or diagnosis logic.
- **No mutable package-level state.** Build things with constructors and pass
  dependencies explicitly. Link-time variables in `internal/version` are the
  one exception.
- Small interfaces, defined by the consumer (e.g. `diagnose` defines the
  `Pinger` it needs; `probe` provides concrete types).
- Any function that does I/O takes `ctx context.Context` as its first
  parameter and honors cancellation. Every network operation has a timeout.
- Durations are `time.Duration`, not ints of milliseconds. Use the monotonic
  clock (`time.Since`) for RTTs, never wall-clock subtraction.
- Platform-specific code uses filename suffixes (`_linux.go`, `_windows.go`,
  `_darwin.go`, `_unix.go`) plus a fallback that returns
  `ErrUnsupportedPlatform`, so every target always compiles. CI cross-compiles
  all of them.
- Comments explain *why*, not *what*. Exported identifiers have doc comments.
  Networking code cites the RFC section it implements.
- Logging uses `log/slog` to stderr. Results go to stdout, so
  `hopscope ... --json | jq` works.

## Error handling

- Wrap errors with context: `fmt.Errorf("resolve %s: %w", host, err)`.
  Messages are lowercase and have no trailing punctuation.
- Inspect errors with `errors.Is` / `errors.As`, never string matching
  (enforced by `errorlint`).
- Expected conditions are sentinel errors or typed errors that callers can
  handle, e.g. `ErrPermission` so the CLI can suggest
  `setcap cap_net_raw+ep`.
- A lost probe is **data, not an error**: record it as a lost sample. Return
  an error only when measurement itself is impossible.
- No `panic` in library code. `main` is the only place that calls `os.Exit`.
- Errors shown to users say what to do next (e.g. "run as a user in
  net.ipv4.ping_group_range, or grant CAP_NET_RAW").

## Testing

- Every package with logic has tests. Stats, parsers and diagnosis rules are
  pure functions with **table-driven tests**.
- **Unit tests never touch the real network.** Inject fakes (`net.PacketConn`,
  clocks, probers) behind small interfaces. Packet parsers are tested against
  captured bytes in `testdata/`.
- Integration tests use the `//go:build integration` tag and run in Docker,
  with faults injected by `tc netem`. They assert statistical tolerances, not
  exact values.
- Rendered output (CLI tables, reports) uses golden files in `testdata/` with
  an `-update` flag.
- Tests must pass with `-race` and must not leak goroutines.
- Run `go test ./...` and `golangci-lint run` before declaring work done. If
  either fails, say so; never hide failures.

## Commands

```sh
go build ./... && go test ./...      # works everywhere, including Windows without make
make test | make lint | make cross   # Linux, macOS, WSL, Git Bash
```

Local linting without installing golangci-lint:

```sh
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run
```

## Commits

Conventional, imperative subject lines (`probe: add unprivileged ICMP echo on Linux`),
prefixed with the package or area. Explain the *why* in the body.
