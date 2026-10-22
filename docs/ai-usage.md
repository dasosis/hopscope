# AI usage log

A running record of how AI coding assistants were used on hopscope: what was
delegated, what I reviewed or corrected, and what I wrote myself. The aim is to
be transparent about authorship and to make sure every part of the project is
something I understand and can defend.

**Rules I follow**

- Milestones marked **[write it yourself]** in [ROADMAP.md](../ROADMAP.md)
  are written by me. AI is used only to explain concepts and review my code.
- Nothing AI-generated is merged until I have read it and can explain it.
- Each entry says honestly what I did and did not review.

---

## 2026-10-22: Project scaffold (Phase 0)

**Assistant:** Claude Code (Claude Opus 5.5)

**What I wrote myself**

- The product idea, scope and the eight-phase plan.
- The engineering requirements given to the assistant: cross-platform
  targets, least privilege, no telemetry, the "no high-level networking
  libraries" rule, the testing strategy (unit tests plus `tc netem` fault
  injection), and pure-Go SQLite.

**What I delegated**

- Checked the development environment and searched GitHub and pkg.go.dev for
  name clashes with "hopscope".
- Repository scaffold: `go.mod`, Cobra CLI with `hopscope version` and its test,
  `internal/` package layout, Makefile, golangci-lint config, GitHub Actions CI,
  Dependabot config, `.gitignore`, `.gitattributes`, LICENSE.
- First drafts of README.md, ROADMAP.md, CLAUDE.md, ADRs 0001 to 0003 and this log.
- `git init`, the initial commit, creating the GitHub repository and checking CI.

**Notable choices the assistant made (to understand and verify)**

- `go.mod` declares Go 1.26 (the previous release) and CI also tests on
  `oldstable`, so both supported Go releases are covered.
- `.gitattributes` forces LF line endings because Git for Windows defaults to
  `core.autocrlf=true`, which would break gofmt checks.
- Release metadata is injected via `-ldflags -X`, with a fallback to
  `runtime/debug.ReadBuildInfo` for `go install` builds. The build date is the
  commit time, so builds are reproducible.
- CI runs with `permissions: contents: read` (least privilege) and uses the race
  detector everywhere except Windows.

**What I reviewed / corrected**

- _To fill in after reviewing: which files I read line by line, what I
  changed, and anything I disagreed with._
