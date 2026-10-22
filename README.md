# hopscope

[![CI](https://github.com/dasosis/hopscope/actions/workflows/ci.yml/badge.svg)](https://github.com/dasosis/hopscope/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Speed tests tell you how fast your internet is. hopscope tells you why it feels slow.**

> **Status: early development.** Only `hopscope version` works today. The
> features below are planned; see [ROADMAP.md](ROADMAP.md) for progress.

## Why

A "200 Mbps" connection can still make video calls stutter and games lag.
Throughput is rarely the real problem. The usual culprits are:

- **Bufferbloat**: latency that jumps from 20 ms to 400 ms the moment someone starts an upload
- **Jitter**: inconsistent delay that breaks real-time audio and video
- **Packet loss**, at a specific hop that you can identify
- **Slow DNS**: every new site waits on a slow resolver before anything loads
- **Weak Wi-Fi**: low signal, a congested 2.4 GHz channel, or the wrong band

hopscope is a local agent that walks the network path hop by hop and layer by
layer:

```
device  ->  Wi-Fi / LAN  ->  router  ->  ISP  ->  DNS  ->  destination
```

It finds the layer at fault and explains it in plain English, with the
evidence behind the verdict and a suggested fix.

```console
$ hopscope diagnose --profile calls      # planned
Verdict: your connection is fine at idle but suffers from bufferbloat.

  Idle latency     18 ms
  Loaded latency  412 ms   (+394 ms while uploading)
  Jitter (loaded)  61 ms   (calls profile limit: 30 ms)

Likely cause: your router's upload buffer fills up when anything uploads.
Suggested fix: enable SQM / Smart Queue Management (fq_codel or CAKE) on your
router and set the upload limit to ~90% of your measured upload speed.
```

## Planned features

| Phase | Feature |
|---|---|
| 1 | ICMP ping (privileged and unprivileged, including Windows without admin), TCP-connect fallback, DNS timing across resolvers, gateway discovery; min/avg/p95, jitter, loss |
| 2 | Its own TTL-based traceroute with MTR-style per-hop stats that does not report false loss from routers that rate-limit ICMP |
| 3 | Diagnosis engine: layered checks, usage profiles (`gaming`, `calls`, `streaming`, `browsing`), custom targets |
| 4 | Wi-Fi signal/band/channel, bufferbloat test, HTTP timing breakdown (connect / TLS / TTFB), MTU black-hole detection |
| 5 | Background service with SQLite history, retention and downsampling, anomaly detection against a time-of-day baseline |
| 6 | Localhost-only, token-protected web dashboard, desktop alerts, one-click "ISP outage report" with per-hop evidence |
| 7 | MCP server exposing diagnostics and history summaries to AI assistants |
| 8 | Release builds, demo, benchmarks |

## Design principles

- **Low-level networking, written here.** Probing and traceroute are built
  directly on `golang.org/x/net/icmp` and the standard library, not on
  high-level ping or traceroute libraries.
- **Least privilege.** Unprivileged ICMP wherever the OS allows it. On Linux,
  at most `CAP_NET_RAW`, never full root. On Windows, no administrator rights.
  See [docs/decisions](docs/decisions/).
- **Private by default.** No telemetry. Results stay on your machine. The
  dashboard binds to localhost only.
- **Evidence, not vibes.** Every verdict shows the measurements that led to it.
- **Cross-platform.** Windows and Linux are first-class; macOS is supported where practical.

## Building

Requires Go 1.26 or newer.

```sh
go build -o bin/hopscope ./cmd/hopscope
./bin/hopscope version
```

With `make` (Linux, macOS, WSL or Git Bash):

```sh
make build   # host binary in bin/
make test    # unit tests
make lint    # golangci-lint
make cross   # windows/linux/darwin binaries in dist/
```

## Project docs

- [ROADMAP.md](ROADMAP.md): milestones and what "done" means for each
- [docs/decisions/](docs/decisions/): architecture decision records
- [docs/ai-usage.md](docs/ai-usage.md): what was built with AI assistance and how it was reviewed

## License

[MIT](LICENSE)
