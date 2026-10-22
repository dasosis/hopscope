# 0002. Use modernc.org/sqlite (pure Go) for local history

- **Status:** Accepted (the dependency is added in Phase 5)
- **Date:** 2026-10-22

## Context

The background agent (Phase 5) stores probe results locally with retention
and downsampling. Requirements:

- An embedded, single-file database with no server to install. SQL makes time-range
  queries and aggregation straightforward.
- Builds stay `CGO_ENABLED=0` so one machine can cross-compile static binaries
  for Windows, Linux and macOS (see the `cross` target in the Makefile and CI).
- Low write volume. For example, 5 targets probed every 10 s is about 0.5 rows/s, and
  24 h of raw data is about 43k rows. Raw database speed is not the bottleneck.

## Options considered

1. **`mattn/go-sqlite3`.** The most widely used driver; it wraps the real
   SQLite C code through **cgo**. It is the fastest option, but it requires
   a C toolchain for every target. Cross-compiling to Windows or macOS from
   Linux needs mingw, osxcross or `zig cc`, which complicates CI and releases.
   It also links dynamically against the C library by default.
2. **`modernc.org/sqlite`.** SQLite's C source mechanically translated to Go.
   It needs no cgo and implements `database/sql`. It is mature and widely used.
   The binary is several MB larger and it is slower than the cgo driver,
   especially on write-heavy workloads.
3. **`ncruces/go-sqlite3`.** SQLite compiled to WebAssembly and run in-process
   by the pure-Go `wazero` runtime. It needs no cgo and is actively developed.
   It is younger, and it adds a Wasm runtime plus a compile step at startup.
4. **Not SQL (`bbolt`, flat files).** Simpler dependencies, but we would
   hand-write the time-range scans and aggregations that SQL provides.

## Decision

Use **`modernc.org/sqlite`** through `database/sql`, with WAL journal mode and
a single writer goroutine.

## Consequences

- `CGO_ENABLED=0` and the one-command cross-compile keep working. CI does not
  need a C toolchain.
- The binary grows by several MB, which is acceptable for a desktop agent.
- Our write rate is several orders of magnitude below where the speed gap
  matters. A Phase 5 benchmark (insert plus downsampling job) will confirm this.
- Code depends only on `database/sql` (plus SQL that SQLite accepts), so
  switching to `ncruces/go-sqlite3` or the cgo driver is a one-line import
  change if needed.
- This decision would be wrong if profiling showed the database dominating
  CPU or latency in the agent. That is the trigger to revisit.
