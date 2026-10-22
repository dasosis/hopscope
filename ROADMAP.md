# Roadmap

Each milestone lists its **deliverables**, a **done when** check that can be
verified objectively, and the **concept** to be able to explain in an interview.

Milestones marked **[write it yourself]** contain the core algorithms and
low-level networking. They are the parts most worth building by hand, with AI
help limited to review and explanation. See [docs/ai-usage.md](docs/ai-usage.md).

Status: ✅ done · 🚧 in progress · ⬜ not started

---

## Phase 0: Foundations ✅

**Deliverables:** Go module, Cobra CLI with `hopscope version`, package layout,
Makefile, golangci-lint, CI (lint, test on Windows/Linux/macOS, cross-compile),
ADRs 0001 to 0003, MIT license.

**Done when:** CI is green on `main`.

**Concept:** Why a CLI command tree is built by constructors instead of package
globals (testability); how `-ldflags -X` injects build metadata; why
`CGO_ENABLED=0` makes cross-compilation trivial.

---

## Phase 1: Probing primitives

### 1.1 Statistics core (`internal/stats`) ⬜ [write it yourself]
- **Deliverables:** `Sample` (seq, send time, RTT, lost) and `Summary`
  (sent, received, loss %, min, avg, max, p50, p95, stddev, jitter). Pick a
  documented percentile method (nearest-rank). Define jitter as the mean absolute
  difference between consecutive RTTs (the instantaneous form behind RFC 3550's
  smoothed estimator).
- **Done when:** table-driven tests cover empty input, all lost, a single sample,
  duplicates and out-of-order replies; coverage of the package is 100%; a
  benchmark exists for 10k samples.
- **Concept:** Why p95 matters more than the average (one 500 ms spike ruins a
  call but barely moves the mean). The different definitions of jitter. Why lost
  probes count toward loss but not toward RTT stats. How unreliable percentiles
  are on small samples.

### 1.2 ICMP echo on Linux and macOS (`internal/probe`) ⬜ [write it yourself]
- **Deliverables:** a `Pinger` built on `golang.org/x/net/icmp`. Try an
  unprivileged datagram socket (`icmp.ListenPacket("udp4", ...)`) first and fall
  back to a raw socket (`"ip4:icmp"`). Encode and decode echo messages, match
  replies by (ID, seq) and drop stray packets. Use per-probe timeouts through
  read deadlines and honor context cancellation. Add `hopscope ping <host> -c N -i interval`.
  IPv4 first, IPv6 after.
- **Done when:** `hopscope ping 1.1.1.1` works as a normal user on Ubuntu, and
  as raw ICMP when the binary has `setcap cap_net_raw+ep`. The output states
  which mode was used. Unit tests cover encoding, decoding and reply matching
  against a fake `net.PacketConn`, with no real network.
- **Concept:** ICMP header layout (type, code, checksum, ID, seq). Raw vs
  datagram ICMP sockets. `net.ipv4.ping_group_range`. Why the kernel overwrites
  the echo ID on datagram sockets. Linux capabilities and file capabilities.
  See [ADR 0003](docs/decisions/0003-icmp-privilege-strategy.md).

### 1.3 ICMP echo on Windows without admin ⬜ [write it yourself]
- **Deliverables:** `probe_windows.go` calling `IcmpSendEcho2` /
  `Icmp6SendEcho2` from `iphlpapi.dll` through `golang.org/x/sys/windows`,
  decoding `ICMP_ECHO_REPLY` into the same `Pinger` result type.
- **Done when:** `hopscope ping` works from a non-elevated PowerShell. The
  reply-struct decoding is unit-tested in the Windows CI job.
- **Concept:** Calling Win32 APIs from Go (lazy DLL loading, struct layout and
  alignment on 32-bit vs 64-bit, `unsafe.Pointer` rules). Why raw sockets need
  admin on Windows. The API's millisecond RTT resolution and how to compensate.

### 1.4 TCP-connect ping ⬜
- **Deliverables:** time a TCP handshake to `host:port` (default 443) with a
  `net.Dialer` and context timeout. A `connection refused` (RST) is a *valid* RTT
  sample; a timeout counts as loss. Use it automatically when ICMP is unavailable
  or filtered.
- **Done when:** works with zero privileges on every OS; unit tests use a local
  listener and a closed port.
- **Concept:** The three-way handshake. Why an RST still proves reachability.
  Why the timeout must stay below the kernel's SYN retransmit (~1 s on Linux),
  or a retransmitted SYN skews the RTT. Why ICMP and TCP RTTs can differ.

### 1.5 DNS timing ⬜
- **Deliverables:** build queries with `golang.org/x/net/dns/dnsmessage` and send
  them over UDP to each resolver: the system-configured ones, 1.1.1.1, 8.8.8.8
  and 9.9.9.9. Discover the system resolvers (Linux: `/etc/resolv.conf`, noting
  the systemd-resolved `127.0.0.53` stub; Windows: `GetAdaptersAddresses`).
  Separate cached ("warm") from uncached ("cold") lookups. Handle the TC bit by
  retrying over TCP. Add `hopscope dns`.
- **Done when:** a comparison table with min/avg/p95 per resolver names the
  fastest. Message encoding and decoding are tested against canned packets in
  `testdata/`.
- **Concept:** Recursive vs authoritative resolution, caching and TTLs, why
  your "system DNS" is usually your router, UDP vs TCP for DNS, and why query
  IDs and source ports are randomized (Kaminsky attack).

### 1.6 Default gateway discovery ⬜
- **Deliverables:** Linux parses `/proc/net/route` (hex, little-endian).
  Windows calls `GetBestRoute2` or `GetIpForwardTable2`. macOS uses
  `golang.org/x/net/route`. Add `hopscope gateway`.
- **Done when:** prints gateway IP and interface on Windows and Linux (and macOS);
  the Linux parser is tested on captured files.
- **Concept:** Routing tables, the default route `0.0.0.0/0`, longest-prefix
  match, byte order.

### 1.7 Fault-injection test harness ⬜
- **Deliverables:** a Docker Compose topology of `client ↔ router ↔ target` on
  two bridge networks, with the router forwarding. Go integration tests
  (`//go:build integration`) apply `tc qdisc ... netem delay 100ms 10ms loss 5%`
  and assert on hopscope's results. Add `make integration` and a CI job on
  Linux.
- **Done when:** the CI job is green, measured average RTT is within ±15% of the
  injected delay, and measured loss falls inside a binomial confidence interval.
- **Concept:** Queueing disciplines and netem. Why the tests assert *statistical
  tolerances*: with 5% loss and 200 probes, the observed loss is roughly 2% to 8%.
  `NET_ADMIN` vs `NET_RAW` container capabilities.

---

## Phase 2: Traceroute (`internal/trace`)

### 2.1 TTL probing and ICMP error parsing ⬜ [write it yourself]
- **Deliverables:** send probes with increasing TTL (`ipv4.PacketConn.SetTTL`
  from `golang.org/x/net/ipv4`). Parse ICMP Time Exceeded (type 11) and
  Destination Unreachable (type 3) replies, including the original IP header
  plus 8 bytes that the router quotes back, and use that quote to match the reply
  to its probe. Stop at the destination or max hops. Windows uses `IcmpSendEcho2`
  with `IP_OPTION_INFORMATION.Ttl` and treats `IP_TTL_EXPIRED_TRANSIT` as a hop
  (this is how `tracert` works). How Linux runs unprivileged (raw socket vs
  `IP_RECVERR`) is decided in a new ADR.
- **Done when:** `hopscope trace 1.1.1.1` matches the hops from
  `traceroute`/`tracert` on a real network. Decoding is unit-tested against
  captured packets in `testdata/`. The integration test sees the router container
  as hop 1.
- **Concept:** TTL semantics. What an ICMP error quotes (RFC 792, RFC 4884
  extensions). Why some hops show `*`. Hidden MPLS hops.

### 2.2 Flow-stable probes (Paris traceroute) ⬜
- **Deliverables:** keep the fields that load balancers hash on constant across
  probes. For ICMP echo that means the ID and the checksum: vary the sequence
  number and compensate in the payload so the checksum does not change.
- **Done when:** a unit test proves the checksum is constant across all probes.
- **Concept:** ECMP and per-flow hashing, and why classic traceroute can draw
  links that do not exist.

### 2.3 MTR-style rounds and per-hop stats ⬜
- **Deliverables:** repeated rounds with per-hop sent/received/loss/min/avg/p95/jitter
  (reusing `internal/stats`), tracking every responder seen at each TTL.
  Probes are paced so they do not trip rate limits. Add `hopscope trace --rounds N`.
- **Done when:** a live-updating table runs for 50 rounds with no goroutine leaks
  (checked in tests).
- **Concept:** The concurrency design. One receiver goroutine demultiplexes
  replies to in-flight probes keyed by (TTL, seq), with timeouts and cancellation.

### 2.4 Honest loss attribution ⬜ [write it yourself]
- **Deliverables:** loss at hop *N* counts only if it persists through every
  later hop to the destination. Loss seen only at an intermediate hop is labeled
  "this router rate-limits ICMP replies; it does not affect your traffic". The
  same rule applies to latency spikes that do not carry forward.
- **Done when:** unit tests on synthetic hop tables cover a rate-limited middle
  hop (no false alarm), real loss starting at hop 3 (blamed on the link into
  hop 3), destination-only loss, and `* * *` hops mid-path.
- **Concept:** **Data plane vs control plane.** Routers forward transit packets
  in hardware but generate ICMP replies on the CPU, which Control Plane Policing
  (CoPP) rate-limits. So "50% loss at hop 5, 0% at hop 6" cannot be real loss.

---

## Phase 3: Diagnosis engine (`internal/diagnose`)

### 3.1 Evidence model and check interface ⬜
- **Deliverables:** `Layer` (local link, gateway, ISP, DNS, destination),
  `Finding` (layer, severity, verdict, evidence, fix), a small `Check` interface,
  and probers injected through interfaces.
- **Done when:** every check runs in unit tests on fixed inputs with no network.
- **Concept:** Keeping measurement separate from interpretation, so the
  interpretation is pure functions over data and easy to test. Dependency
  injection with small Go interfaces.

### 3.2 Layered root-cause logic ⬜ [write it yourself]
- **Deliverables:** evaluate from the nearest layer outward. A fault at an inner
  layer explains symptoms further out instead of producing duplicate verdicts.
  The ISP layer is the first hop beyond the gateway. Detect double NAT and
  CGNAT (`100.64.0.0/10`).
- **Done when:** table-driven tests show each injected fault produces exactly
  one primary verdict, at the right layer.
- **Concept:** Fault localization by dividing the path. Why the innermost
  failing layer is the root cause.

### 3.3 Usage profiles ⬜
- **Deliverables:** `gaming`, `calls`, `streaming`, `browsing` with thresholds
  for p95 RTT, jitter, loss, DNS time and TTFB, kept in one table with a source
  cited for each value (e.g. ITU-T G.114 for voice delay). Add `--profile`.
- **Done when:** `hopscope diagnose --profile calls` runs end to end, and
  `docs/profiles.md` explains every threshold.
- **Concept:** Why applications tolerate impairments differently: real-time UDP
  with jitter buffers vs TCP retransmission vs buffered streaming.

### 3.4 Output, user targets, end-to-end tests ⬜
- **Deliverables:** human-readable output plus `--json` with a stable schema.
  User-defined targets via `--target`. Golden-file tests for rendered output.
- **Done when:** netem on the router produces an "ISP / beyond your router"
  verdict, and netem on the client link produces a "local network" verdict, in CI.
- **Concept:** Golden testing. Machine-readable output as an API contract.

---

## Phase 4: Local layer and deeper tests

### 4.1 Wi-Fi link quality (`internal/wifi`) ⬜
- **Deliverables:** report signal (dBm), band, channel, width and PHY rate.
  Windows: compare the native WLAN API (`wlanapi.dll`) with parsing
  `netsh wlan show interfaces`, which is **localized** and so breaks on
  non-English Windows; decide in an ADR. Linux: `iw dev <if> link`, `nmcli`.
  macOS: best effort.
- **Done when:** `hopscope wifi` works on Windows and Linux; parsers are tested
  on captured outputs, including a non-English locale.
- **Concept:** The dBm log scale (−50 excellent, −70 poor). Trade-offs between
  2.4, 5 and 6 GHz. Overlapping channels.

### 4.2 Bufferbloat test (`internal/bufferbloat`) ⬜ [write it yourself]
- **Deliverables:** measure idle latency, then saturate download and upload
  separately against a documented, user-overridable endpoint while still
  probing latency. Report the increase and a grade. Duration and data volume
  are capped, and the test needs an explicit flag because it uses bandwidth.
- **Done when:** in the integration harness, `netem rate 10mbit` with a deep
  queue is diagnosed as bufferbloat, and the same link with `fq_codel` is not.
- **Concept:** Queueing delay = queued bytes ÷ link rate. Why oversized buffers
  exist. Active queue management (CoDel, fq_codel, CAKE) and SQM.

### 4.3 HTTP timing (`internal/httptiming`) ⬜
- **Deliverables:** use `net/http/httptrace` to split a request into DNS,
  connect, TLS and time to first byte, with keep-alive off for cold
  measurements. Add `hopscope http <url>`.
- **Done when:** unit tests against `httptest.NewTLSServer` with an injected
  delay verify each phase.
- **Concept:** The TLS 1.3 handshake (1-RTT), that TTFB includes server think
  time, and connection reuse.

### 4.4 MTU black-hole check (`internal/pmtu`) ⬜
- **Deliverables:** send echo probes with the Don't Fragment bit set (Linux
  `IP_MTU_DISCOVER`; Windows `IP_FLAG_DF`) and binary-search the largest size
  that gets through. Flag a black hole when large packets vanish without an ICMP
  "fragmentation needed" (type 3, code 4).
- **Done when:** the harness router with MTU 1400 that drops ICMP type 3/4 is
  detected as a black hole.
- **Concept:** Path MTU discovery (RFC 1191) and why blanket ICMP filtering
  breaks it. PPPoE's 1492-byte MTU, MSS clamping, PLPMTUD (RFC 8899).

---

## Phase 5: Background service and history

### 5.1 SQLite store (`internal/store`) ⬜
- **Deliverables:** `modernc.org/sqlite` ([ADR 0002](docs/decisions/0002-sqlite-driver.md)),
  schema migrations embedded with `embed`, WAL mode, `busy_timeout`, and a
  single writer.
- **Done when:** store tests run against a temporary database; migrations are
  idempotent.
- **Concept:** WAL vs rollback journal, SQLite's single-writer model, and
  indexing for time-range queries.

### 5.2 Agent and service integration (`internal/monitor`) ⬜
- **Deliverables:** `hopscope agent` runs a scheduler with jittered intervals
  and graceful shutdown. Ship a systemd unit (`AmbientCapabilities=CAP_NET_RAW`,
  `NoNewPrivileges=yes`, unprivileged user) and a Windows service
  (`golang.org/x/sys/windows/svc`).
- **Done when:** a 24-hour soak run shows flat memory and goroutine counts
  (pprof); install docs exist for both OSes.
- **Concept:** Service managers, capabilities under systemd, goroutine
  lifecycles and context trees, and why probe schedules are jittered.

### 5.3 Retention and downsampling ⬜ [write it yourself]
- **Deliverables:** keep raw samples for 24 h and 1-minute aggregates for 30 days.
  Each aggregate stores count, lost, min, max, sum, sum of squares and a
  fixed-bucket latency histogram. The job is idempotent and runs in a transaction.
- **Done when:** tests with a fake clock show raw rows expire, aggregates keep
  count and loss exactly, and p95 stays within one histogram bucket.
- **Concept:** Mergeable aggregates, and why **you cannot average
  percentiles**, which is the reason a histogram is stored.

### 5.4 Anomaly detection ⬜ [write it yourself]
- **Deliverables:** a per-hour-of-week baseline (168 buckets) built from the
  median and MAD of recent weeks. An anomaly needs a robust z-score above the
  threshold for *k* consecutive minutes (hysteresis). Defined cold-start
  behavior.
- **Done when:** synthetic-series tests confirm a normal evening peak is not
  flagged, an outage is, and a single spike is not.
- **Concept:** Seasonality. Robust statistics (median/MAD vs mean/stddev).
  Alert fatigue and flapping.

---

## Phase 6: Dashboard, alerts, reports

### 6.1 Local dashboard (`internal/dashboard`) ⬜
- **Deliverables:** `net/http` bound to `127.0.0.1`/`::1` only. A 256-bit token
  from `crypto/rand` is stored in the user config dir with mode 0600 and
  exchanged for an `HttpOnly; SameSite=Strict` cookie. Requests with an
  unexpected `Host` header are rejected (DNS-rebinding defense). Cross-origin
  POSTs are rejected. The UI uses `html/template` with embedded assets.
- **Done when:** tests prove a request without the token gets 401, a wrong Host
  gets 403, and the listener never binds a non-loopback address.
- **Concept:** Why "localhost only" is not enough by itself: any web page can
  send requests to localhost. DNS rebinding. Constant-time token comparison.

### 6.2 Desktop alerts (`internal/alert`) ⬜
- **Deliverables:** native notifications on Windows and Linux (mechanism chosen
  in an ADR) with deduplication and recovery notices.
- **Done when:** a simulated outage gives exactly one alert and one recovery
  notice.
- **Concept:** Alert state machines and deduplication.

### 6.3 ISP outage report (`internal/report`) ⬜
- **Deliverables:** a self-contained HTML report with a timeline, the failing
  layer, per-hop evidence, and timestamps in both local time and UTC. PDF comes
  from a print stylesheet rather than a heavy PDF library (decided in an ADR).
  Local IPs and SSIDs are redacted.
- **Done when:** `hopscope report --from ... --to ...` produces a report that is
  covered by a golden test.
- **Concept:** What evidence convinces an ISP: loss that starts at their first
  hop and persists to the destination, with timestamps.

---

## Phase 7: MCP server (`internal/mcpserver`) ⬜

- **Deliverables:** `hopscope mcp` serves MCP over stdio, using the official Go
  SDK (this is not measurement code, so a library is acceptable). Read-only
  tools: `run_diagnosis`, `health_summary`, `list_incidents`. They return
  summaries only, never raw samples, LAN addresses or SSIDs.
- **Done when:** it works from an MCP client, and tests assert that tool output
  contains no private addresses or SSIDs.
- **Concept:** MCP tools over JSON-RPC, data minimization, and **prompt
  injection through tool output**: a reverse-DNS hostname is an
  attacker-controlled string.

---

## Phase 8: Release polish ⬜

- **Deliverables:** goreleaser (archives, checksums, SBOM, and optionally
  signing; deb/rpm packages that set `cap_net_raw` on install). A demo GIF
  scripted with `vhs`. Benchmarks with `b.ReportAllocs()` and `benchstat` for
  stats, packet encoding and decoding, and reply matching, written up in
  `docs/benchmarks.md`.
- **Done when:** pushing tag `v1.0.0` publishes a GitHub release, and the README
  shows the GIF.
- **Concept:** Software supply chain (checksums, SBOMs, signing, reproducible
  builds) and benchmarking methodology.
