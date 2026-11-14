# Changelog

All notable changes to **Marauder Studio** will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-11-14

### Added
- Complete Go application architecture (`internal/device`, `internal/serialtransport`, `internal/marauder`, `internal/capabilities`, `internal/jobs`, `internal/survey`, `internal/monitor`, `internal/capture`, `internal/ble`, `internal/gps`, `internal/firmware`, `internal/policy`, `internal/store`, `internal/reports`, `internal/diagnostics`).
- Serial transport abstraction with automatic USB VID/PID discovery (`go.bug.st/serial/enumerator`), connection state machine, bounded reader goroutine, and high-fidelity mock serial port for testing.
- Marauder wire parser supporting both modern v1.20+ (`scanall`, `sniffpmkid`, `sniffraw`) and legacy firmware syntax (`scanap`, `scansta`).
- Evidence-backed capability matrix engine interrogating runtime `info` and `help` responses.
- Wi-Fi AP and station survey engine with deduplication, RSSI tracking, BSSID vendor lookup, and allowlist filtering.
- Defensive channel monitor with beacon, probe, and deauthentication frame anomaly detectors.
- Streaming PCAP/PCAPNG parser with memory bounds and SD-card capture import.
- BLE advertisement inspector and tracker detection.
- GPS/NMEA stream parser, GPX track visualizer, and wardriving session import.
- Authorized Lab workspace gated by strict policy controls: explicit user consent, target allowlisting, session timeouts, local audit logging, and global Emergency Stop.
- Firmware manager with validated board catalog, partition layout validator, and dry-run flash safety preflights.
- Embedded SQLite persistence engine (`modernc.org/sqlite`) for device profiles, survey snapshots, and audit records.
- Responsive Svelte + TypeScript + Tailwind CSS desktop interface with dark/light themes, keyboard navigation, and virtualized data views.
- GitHub Actions CI matrix and multi-platform Release-on-Tag workflow (macOS, Windows, Linux).
- Progress log `docs/progress/2026-11-14.md` and Architecture Decision Record (ADR-001).
