# 0003. Prefer unprivileged ICMP; never require root or Administrator

- **Status:** Accepted for ping (Phase 1). Traceroute specifics will be
  refined in a Phase 2 ADR.
- **Date:** 2026-10-22

## Context

Ping and traceroute send ICMP packets that the application builds itself,
which TCP and UDP sockets cannot do. Each OS controls this differently.

**Linux**

- *Raw sockets* (`SOCK_RAW, IPPROTO_ICMP`; `icmp.ListenPacket("ip4:icmp", ...)`)
  require `CAP_NET_RAW`. The socket receives a copy of **every** ICMP packet
  arriving at the host, so the program must filter by echo ID itself. In
  return it gets full control: it can choose the ID, and it receives Time
  Exceeded and Destination Unreachable messages, which traceroute needs.
- *ICMP datagram ("ping") sockets* (`SOCK_DGRAM, IPPROTO_ICMP`;
  `icmp.ListenPacket("udp4", ...)`), available since Linux 3.0. Any process
  whose group ID falls inside the sysctl `net.ipv4.ping_group_range` may open
  one. The kernel default `1 0` is an empty range, but systemd (v243+) sets it
  to `0 2147483647`, so most current distributions allow every user. Recent
  Docker versions do the same inside containers. The kernel **overwrites the
  echo ID** with the socket's local "port" and delivers only matching replies,
  so filtering is done for us, but the ID cannot be chosen. ICMP errors (Time
  Exceeded) are not returned by normal reads; they arrive on the socket error
  queue via `IP_RECVERR`.

**macOS:** datagram ICMP sockets are available to unprivileged users, as on
the BSDs. Raw sockets require root. The details differ from Linux, for example
whether reads include the IP header, so tests must cover both.

**Windows**

- *Raw sockets* require Administrator. Client editions of Windows also
  restrict raw socket behavior, and Windows Firewall may filter inbound ICMP.
  Datagram ICMP sockets do not exist, so `x/net/icmp`'s `"udp4"` mode is
  Linux/macOS only.
- *The ICMP helper API*: `IcmpSendEcho2` and `Icmp6SendEcho2` in
  `iphlpapi.dll`. Standard users can call them, and they are what `ping.exe` and
  `tracert.exe` use. The caller passes `IP_OPTION_INFORMATION` (TTL, TOS, the
  Don't Fragment flag) and gets back `ICMP_ECHO_REPLY` structures with a status
  (`IP_SUCCESS`, `IP_TTL_EXPIRED_TRANSIT`, `IP_REQ_TIMED_OUT`, ...), the
  responding address and the RTT. So traceroute and MTU probing are possible
  without admin.

**Everywhere:** a *TCP connect* to an open or closed port needs no privileges
and measures handshake RTT. It measures something different from ICMP: it
passes through firewalls and load balancers, and it is not subject to ICMP
deprioritization.

## Options considered

1. **Require root/Administrator and use raw sockets everywhere.** One code path
   and full control. But it violates least privilege: a diagnostic tool running
   as root parses packets from the network. It is hostile to users and unusable
   in many corporate environments.
2. **TCP-connect only.** No privileges at all, but no traceroute and no
   ICMP-based MTU testing, which defeats the purpose of the project.
3. **Layered strategy: the least-privileged mechanism that works on each OS,
   with explicit fallbacks.**

## Decision

Option 3. hopscope picks a probe mode at runtime in this order and
**always reports which mode it used** (in human output and in `--json`):

| OS | 1st choice | 2nd choice | Fallback |
|---|---|---|---|
| Linux | ICMP datagram socket (unprivileged) | Raw ICMP socket, if `CAP_NET_RAW` is held | TCP connect |
| macOS | ICMP datagram socket (unprivileged) | n/a (raw needs root; not supported) | TCP connect |
| Windows | `IcmpSendEcho2` / `Icmp6SendEcho2` | n/a | TCP connect |

Rules:

- hopscope **never self-escalates**: no `sudo` invocations and no UAC prompts.
  If ICMP is unavailable, it says why and how to fix it (join a group in
  `ping_group_range`, or `setcap cap_net_raw+ep`) and uses TCP connect,
  labeled as such.
- On Linux the privilege ceiling is `CAP_NET_RAW`, granted either as a file
  capability (`setcap cap_net_raw+ep /path/to/hopscope`) or, for the agent, by
  systemd (`AmbientCapabilities=CAP_NET_RAW`, `CapabilityBoundingSet=CAP_NET_RAW`,
  `NoNewPrivileges=yes`). Full root is never needed.
- On Windows, hopscope does **not** use raw sockets even when elevated. The
  helper API covers ping, TTL-limited probes and the DF bit with one code path
  that behaves the same for every user.

### Windows trade-offs to be aware of

- **Resolution.** `ICMP_ECHO_REPLY.RoundTripTime` is in whole milliseconds,
  which is why `ping.exe` prints `<1ms` on a LAN. hopscope also times each call
  with Go's monotonic clock and prefers that value. The measurement then
  includes some API-call overhead, which will be measured and documented in
  Phase 1.
- **Less control.** Windows chooses the echo ID and sequence, and we never see
  the raw ICMP error packet. Phase 2's Paris-traceroute checksum technique is
  therefore Linux/macOS only, and Windows traceroute gets the hop address but
  not the quoted original packet.
- **Blocking calls.** Each synchronous call blocks an OS thread. Go handles this
  (blocked syscalls get their own thread), but concurrency must be bounded,
  for example with a semaphore, so an MTR-style run does not create hundreds
  of threads. The asynchronous event-based mode of `IcmpSendEcho2` is an
  option if this becomes a problem.
- **FFI risk.** Struct layout must match the C definitions exactly. We build
  only 64-bit Windows targets (amd64, arm64), which avoids the
  `ICMP_ECHO_REPLY32` variant. Decoding is unit-tested and runs in the Windows
  CI job.

## Consequences

- Probing goes through a small interface with one implementation per mode.
  Results carry a `Mode` field, and statistics are only compared within the
  same mode.
- More code paths to test: Linux datagram and raw (both in Docker
  integration tests, with and without `CAP_NET_RAW`), Windows helper API
  (Windows CI), and TCP everywhere.
- Linux packaging (Phase 8) must re-apply the file capability after every
  upgrade, because replacing the binary drops it.
- Open for Phase 2: unprivileged traceroute on Linux needs either `IP_RECVERR`
  plus `MSG_ERRQUEUE` (as `tracepath` does) or `CAP_NET_RAW`. This will be
  decided in its own ADR.
