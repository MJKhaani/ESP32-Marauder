# Security, Threat Model, and Safety Architecture

## Threat Model

### Untrusted Inputs
1. **Serial Port Wire Data**: Input arriving from connected microcontrollers may contain malformed byte streams, ANSI escape injections, buffer overflow attempts, or forged packet counts.
   - *Mitigation*: The serial parser uses bounded byte scanners (`bufio.Scanner` with strict buffer limits), truncates unbounded lines, and ignores unrecognized control sequences.
2. **Imported PCAP / NMEA Files**: Arbitrary files imported by the user may be maliciously crafted to trigger out-of-memory or path traversal vulnerabilities.
   - *Mitigation*: Streaming parsers with fixed-size packet windows, memory-bounded batch allocations, and strict filepath validation (`filepath.Clean` and rejection of `..`).
3. **Firmware Binary Assets**: Downloaded or imported firmware images could be corrupted or mismatched.
   - *Mitigation*: SHA-256 cryptographic digest verification against pinned maintainer manifests; dry-run preflight checks matching chip model and partition geometry.

## Safety Controls & Ethical Policy Gate

```text
User Request (Frontend)
         │
         ▼
┌─────────────────────────────────┐
│       PolicyGate (Go Core)      │
│ 1. Is feature passive?          │─── Yes ───► Dispatch to Device
│ 2. If active / high-risk:       │
│    • User consent active?       │─── No ────► Return PolicyError
│    • Target in allowlist?       │─── No ────► Return PolicyError
│    • Duration <= max limit?     │─── No ────► Return PolicyError
│    • Append to Audit Trail      │─── OK ────► Dispatch Bounded Job
└─────────────────────────────────┘
```

### Safety Rules
1. **Default Passive Mode**: Application launches in passive survey/monitor mode.
2. **Authorized Lab Isolation**: The Authorized Lab workspace requires:
   - Explicit confirmation that the operator owns or has written permission to evaluate the target RF environment.
   - Allowlisted BSSIDs / MAC addresses (wildcards and broadcast targets are strictly rejected).
   - Time-bounded sessions with automatic expiration and radio stop command issuance.
   - Global `Emergency Stop` button active across all views.
3. **Privacy Sanitization**: Logs and exported findings mask MAC addresses and BSSIDs (e.g. `aa:bb:cc:xx:xx:xx`) and truncate GPS coordinates by default unless the user explicitly exports raw artifacts.
