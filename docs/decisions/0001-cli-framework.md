# 0001. Use Cobra for the command-line interface

- **Status:** Accepted
- **Date:** 2026-10-22

## Context

hopscope is a multi-command tool: `version`, `ping`, `dns`, `gateway`,
`trace`, `diagnose`, `wifi`, `agent`, `report`, `mcp`, and more to come. These
commands share some flags (`--json`, `--timeout`, `--profile`) and have their
own flags too. Users on Windows mostly use PowerShell; Linux users use bash or
zsh. The CLI is the product's main interface, so help text and shell completion
matter.

## Options considered

1. **Standard library `flag`.** No dependencies. But there are no subcommands:
   you dispatch on `os.Args[1]` and hand-build one `FlagSet` per command. Flags
   are Go-style (`-json`) rather than GNU-style (`--json`, `-c 5`). There is no
   generated help tree and no shell completion. That is fine for a
   three-command tool and tedious for twelve.
2. **`urfave/cli` (v3).** Mature, with subcommands and completion. A smaller
   community in infrastructure tooling, and its API changed substantially
   between major versions.
3. **`alecthomas/kong`.** Declares the CLI as structs with tags. Very concise,
   but the struct-tag "magic" is harder to explain at a glance and less widely
   known.
4. **`spf13/cobra`.** The de facto standard for Go infrastructure CLIs
   (`kubectl`, `docker`, `gh`, `helm`, `hugo`, `etcdctl`). Nested
   subcommands, GNU-style flags via `pflag`, persistent (inherited) flags,
   generated help, and shell completion for bash, zsh, fish **and
   PowerShell**.

## Decision

Use **Cobra** without Viper. Configuration loading will be decided separately
if it is ever needed.

Commands are built by constructor functions (`NewRootCmd()`,
`newVersionCmd()`) instead of the package-level variables plus `init()` that
Cobra's generator and many tutorials use. Each call returns an independent
command tree, so tests can run commands in isolation with `cmd.SetArgs` and
`cmd.SetOut`, and no flag values leak between tests.

## Consequences

- Two small transitive dependencies: `spf13/pflag`, and
  `inconshreveable/mousetrap` (Windows only: it detects when the binary was
  double-clicked from Explorer and explains that it is a command-line tool).
- PowerShell completion comes for free, which matters for a Windows-first tool.
- Reviewers and interviewers who know `kubectl` or `gh` will recognize the
  structure.
- Cobra adds some boilerplate per command. That is acceptable at this size.
- Revisit if the binary size or dependency count becomes a concern. The
  `cobra.Command` usage is confined to `internal/cli`, so swapping frameworks
  would not touch measurement or diagnosis code.
