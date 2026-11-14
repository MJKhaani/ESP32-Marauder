# Marauder Studio

An offline-first, cross-platform desktop control and analysis application for **ESP32 Marauder** devices in expressly authorized educational wireless research and security validation environments.

Built with **Go**, **Wails**, **Svelte**, **TypeScript**, and **Tailwind CSS**.

---

## Authorization, Acceptable Use, and Safety Notice

> **IMPORTANT**: Marauder Studio is supplied for educational, security validation, and authorized research use only. Users must obtain all permissions required by applicable law and must limit their activities to systems, networks, and radio environments they own or are entitled to test. The developers and contributors do not authorize misuse or knowingly endorse unlawful, disruptive, or privacy-invasive conduct. Operators are exclusively responsible for complying with all applicable laws and third-party rights. Passive inspection is the default; any active testing is isolated, strictly scoped to manually confirmed lab targets, and bounded by safety controls.

---

## Core Principles

1. **Go Owns the Device**: All serial transport, protocol parsing, capability discovery, state machines, background jobs, persistence, and safety gates live in Go. The web UI never opens a serial port, shells out, or talks to hardware directly.
2. **Live Capability Detection**: Rather than blindly assuming features exist, Marauder Studio interrogates the connected device (`info`, `help`, hardware characteristics) to build an evidence-backed capability matrix.
3. **Safety by Design**: Passive discovery and inspection are the default. Potentially disruptive operations are segregated into an **Authorized Lab** workspace with target allowlisting, explicit user consent, strict timeouts, and a global Emergency Stop.
4. **Non-Blocking Architecture**: Serial reads, scans, packet analysis, and flashing run in cancellable background jobs over bounded Go channels. The UI never freezes.
5. **Privacy First**: Local and private by default. No cloud accounts, external telemetry, or unsolicited data uploads. Logs and exported reports sanitize MAC addresses, SSIDs, and GPS coordinates by default.
6. **Hardware-Matched Flashing**: Strict hardware catalog matching chip family, PCB revision, flash size, and partition layout before writing firmware images. Dry-run validation prevents accidental bricks.

---

## Architecture Overview

```text
┌───────────────────────────────────────────────────────────────────┐
│ Wails Desktop Shell (macOS / Windows / Linux)                     │
│ Svelte + TypeScript + Tailwind CSS UI                             │
│ Routed views, stores, virtualized tables, live sparklines/charts  │
└──────────────────────────────┬────────────────────────────────────┘
                               │ Typed Wails Bindings & Event Bus
┌──────────────────────────────▼────────────────────────────────────┐
│ Go Application Layer (internal/app)                               │
│ DeviceService • CapabilityService • JobService • SurveyService    │
│ MonitorService • CaptureService • BLEService • GPSService         │
│ FirmwareService • PolicyGate • Store • Diagnostics                │
├───────────────────────────────────────────────────────────────────┤
│ Protocol & Hardware Layer                                         │
│ DeviceManager • DeviceSession • PortEnumerator • SerialTransport  │
│ Wire Decoder • CLI Profiles (v1.20+ / legacy) • Command Queue     │
├───────────────────────────┬───────────────────────────────────────┤
│ Embedded SQLite Database  │ Flasher Adapter (esptool backend)     │
└──────────────┬────────────┴──────────────┬────────────────────────┘
               │                           │
          USB Serial                  Bootloader / ROM
               │                           │
        ESP32 Marauder Device (v4, v6, Mini, DevBoard, C5, S3, etc.)
```

---

## Repository Structure

```text
├── .github/workflows/       # GitHub Actions CI and multi-platform release on tag
├── cmd/marauder-studio/     # Entry point binary
├── docs/
│   ├── architecture/        # Architecture Decision Records (ADRs)
│   ├── compatibility/       # Hardware and firmware compatibility matrices
│   ├── progress/            # Chronological engineering progress logs
│   ├── protocol/            # CLI wire protocol documentation
│   └── security/            # Threat model, safety controls, and privacy
├── firmware/catalog/        # Validated hardware manifests & partition layouts
├── frontend/                # Svelte + TypeScript + Tailwind frontend
├── internal/
│   ├── app/                 # Application lifecycle, Wails bridges, handlers
│   ├── ble/                 # BLE observation, advertisement analysis
│   ├── capabilities/        # Evidence-backed capability engine
│   ├── capture/             # Streaming PCAP/PCAPNG parser & SD card import
│   ├── device/              # DeviceManager, DeviceSession, state machine
│   ├── diagnostics/         # Structured logging, sanitized bug report export
│   ├── firmware/            # Target catalog, partition validator, flasher adapter
│   ├── gps/                 # NMEA parser, GPX track visualizer, wardrive import
│   ├── jobs/                # Cancellable background JobManager, bounded streams
│   ├── marauder/            # CLI profiles, wire parsers, command registry
│   ├── monitor/             # Channel monitor, beacon/probe/deauth passive alerts
│   ├── policy/              # Lab authorization gates, target allowlisting
│   ├── reports/             # Sanitized HTML/JSON/CSV report generator
│   ├── serialtransport/     # Portable serial ports, USB enumerator, mock port
│   ├── store/               # Embedded SQLite migrations, session persistence
│   └── survey/              # Wi-Fi AP & station survey manager
├── testdata/                # Sanitized serial logs & PCAP test fixtures
├── tests/integration/       # Full pipeline integration tests with mock serial
└── wails.json               # Wails v2 project configuration
```

---

## Getting Started

### Prerequisites

- **Go**: 1.22+ installed
- **Node.js**: 18+ and `npm`
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Operating-system specific webview requirements:
  - macOS: Xcode command-line tools
  - Linux: `libgtk-3-dev`, `libwebkit2gtk-4.0-dev` (or 4.1)
  - Windows: WebView2 runtime (pre-installed on Windows 10/11)

### Development

```bash
# Clone the repository
git clone https://github.com/MJKhaani/ESP32-Marauder.git
cd ESP32-Marauder

# Run test suite
go test -v ./...

# Start Wails live development mode
wails dev
```

### Production Build

```bash
# Build desktop binary for current host platform
wails build -clean
```

---

## Release on Tag

Automated GitHub Actions workflows build native binaries and installers for:
- **macOS**: Universal binary (arm64 + amd64) in `.dmg` and `.zip`
- **Windows**: `x86_64` `.exe` installer and `.zip`
- **Linux**: `x86_64` binary and `.tar.gz`

Whenever a version tag `v*` is pushed:
```bash
git tag v1.0.0
git push origin v1.0.0
```
GitHub Actions compiles all platform artifacts, computes SHA-256 checksums, and publishes a tagged GitHub Release.

---

## License

Licensed under the [MIT License](LICENSE).
Upstream ESP32 Marauder firmware by [justcallmekoko](https://github.com/justcallmekoko/ESP32Marauder) is licensed under the MIT License.
