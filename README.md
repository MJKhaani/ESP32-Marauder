<div align="center">

# ⚡ MARAUDER STUDIO

### Next-Generation Desktop Control & Defensive Analysis Suite for ESP32 Marauder

[![GitHub Release](https://img.shields.io/github/v/release/MJKhaani/ESP32-Marauder?style=for-the-badge&logo=github&color=0ea5e9&labelColor=0f172a)](https://github.com/MJKhaani/ESP32-Marauder/releases)
[![CI Status](https://img.shields.io/github/actions/workflow/status/MJKhaani/ESP32-Marauder/ci.yml?branch=main&style=for-the-badge&logo=githubactions&logoColor=white&labelColor=0f172a)](https://github.com/MJKhaani/ESP32-Marauder/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=for-the-badge&logo=go&logoColor=white&labelColor=0f172a)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platforms-macOS%20%7C%20Windows%20%7C%20Linux-818cf8?style=for-the-badge&labelColor=0f172a)](#platform-downloads)
[![License: MIT](https://img.shields.io/badge/License-MIT-38bdf8?style=for-the-badge&labelColor=0f172a)](LICENSE)

<br/>

**Marauder Studio** is an offline-first, cross-platform technical instrument that interfaces seamlessly with **ESP32 Marauder** hardware over native USB serial and UART bridges. It replaces raw terminal guesswork with automatic hardware fingerprinting, live evidence-backed capability matrices, real-time RF observation charts, streaming PCAP analysis, and guarded hardware-matched firmware flashing.

[Downloads](#-platform-downloads) • [Features](#-core-capabilities) • [Hardware Matrix](#-supported-hardware-catalog) • [Authorized Lab](#-authorized-lab--safety-policy) • [Architecture](#-architecture)

</div>

---

## 🖥️ User Interface Overview

```text
┌────────────────────────────────────────────────────────────────────────────────────────┐
│ [M] Marauder Studio v1.0.0      Target: Marauder v6 [Connected]     [Quick Survey] [STOP ALL] │
├──────────────┬─────────────────────────────────────────────────────────────────────────┤
│  Overview    │  DISCOVERED ACCESS POINTS (38 Active)                   Filter: [ Lab_5G  ]     │
│  Devices     ├─────────────────────────────────────────────────────────────────────────┤
│ >Wi-Fi Survey│  BSSID              SSID              CH  BAND     RSSI     SECURITY VENDOR │
│  Channels    │  00:11:22:33:44:55  Lab_Secure_5G     36  5 GHz    -45 dBm  WPA2     Cisco  │
│  Captures    │  aa:bb:cc:dd:ee:ff  Guest_Open         6  2.4 GHz  -72 dBm  OPEN     Intel  │
│  Bluetooth   │  00:11:22:aa:bb:cc  Lab_IoT_Sensors    1  2.4 GHz  -60 dBm  WPA3     Cisco  │
│  GPS & Recon ├─────────────────────────────────────────────────────────────────────────┤
│  Auth Lab ⚠️ │  DEFENSIVE ANOMALY MONITOR                                              │
│  Firmware    │  [CRIT] 13:42:01 — Deauth Storm: >15 frames/sec from aa:bb:cc:xx:xx:xx  │
│  Reports     │  [WARN] 13:41:12 — Evil Twin Alert: Duplicate SSID 'Guest_Open' on CH 1 │
│  Console     ├─────────────────────────────────────────────────────────────────────────┤
│  Settings    │  PACKET STREAM: [||||||||||||||||||||||||||||||||] 142 pkts/s (Bounded) │
└──────────────┴─────────────────────────────────────────────────────────────────────────┘
```

---

## 📦 Platform Downloads

Automated GitHub Actions builds and packages native desktop releases for **every tag**:

| Platform | Architecture | Package Format | Download Link |
|:---|:---|:---|:---|
| **macOS** | Universal (Apple Silicon M1/M2/M3/M4 & Intel x86_64) | `.zip` (`.app` bundle) / `.dmg` | [Download macOS](https://github.com/MJKhaani/ESP32-Marauder/releases/latest) |
| **Windows** | Windows 10 / 11 (64-bit x86_64) | `.zip` (`.exe` binary) | [Download Windows](https://github.com/MJKhaani/ESP32-Marauder/releases/latest) |
| **Linux** | Ubuntu / Debian / Fedora / Arch (64-bit x86_64) | `.tar.gz` (ELF binary) | [Download Linux](https://github.com/MJKhaani/ESP32-Marauder/releases/latest) |

Every release includes cryptographic **`SHA256SUMS.txt`** checksums generated during build.

---

## ✨ Core Capabilities

### ⚡ 1. Zero-Config Device Discovery & Fingerprinting
- **Universal Port Enumeration**: Detects native ESP32-S2/S3/C3/C5/C6 CDC ports, CP210x, CH340/CH910x, and FTDI USB-to-UART bridges.
- **Evidence-Backed Matrix**: Interrogates connected boards (`info` and `help`) on connect to build a real-time capability table (`Available`, `RequiresHardware`, `FirmwareTooOld`, `NotExposedCLI`, `Unsupported`).
- **No GUI Freezes**: Dedicated non-blocking serial reader goroutine with bounded scanner buffers and thread-safe dispatch queues.

### 📡 2. Passive Wi-Fi Survey & Station Mapping
- **Adaptive Version Compatibility**: Automatically detects and leverages modern v1.20+ `scanall` or falls back to legacy `scanap`/`scansta`.
- **Live Signal Visualizer**: Sortable tables tracking SSID, BSSID, Channel, Band (2.4 GHz vs 5 GHz), RSSI dBm signal level, Encryption (WPA2/WPA3/OPEN), and OUI vendor attribution.
- **Client Association**: Maps discovered stations to associated AP BSSIDs in real time.
- **Export Formats**: Instant sanitized CSV and JSON report exports.

### 🛡️ 3. Defensive Monitoring & Anomaly Detection
- **Channel Occupancy**: Live distribution bar charts across channels 1–14 and 5 GHz UNII bands.
- **Rogue AP / Evil Twin Detector**: Flags conflicting BSSIDs advertising identical SSIDs with mismatched security or channels.
- **Deauth Burst Detection**: Surfaces management frame storms (>10 deauths/5s) in real time.
- **Zero RF Transmissions**: Passive observation mode is guaranteed never to emit 802.11 frames.

### 📦 4. Streaming PCAP / PCAPNG Analysis
- **Bounded Memory Ingestion**: Streams multi-gigabyte packet captures without loading entire files into RAM.
- **Protocol Decoding**: Fast 802.11 management/control/data frame classifier (Beacons, Probe Requests, Deauthentications, PMKID handshakes).
- **MicroSD Card Import**: Direct scanning and ingestion of capture files from mounted MicroSD volumes.

### 📍 5. Bluetooth Trackers & Field Recon
- **BLE Inspector**: Real-time advertisement decoding with RSSI proximity tracking.
- **Tracker Detection**: Automated identification of Apple AirTag / FindMy, Tile, and Samsung SmartTag payloads.
- **GPS & Wardriving**: Decodes live NMEA ($GPRMC, $GPGGA) satellite telemetry, visualizes tracks via GPX export, and imports WiGLE wardriving CSV archives.

### 💾 6. Hardware-Matched Firmware Flashing
- **Layout Safety Preflights**: Validates chip family, flash capacity, and partition table offsets before writing.
- **ESP32-C5 Bricking Prevention**: Strict enforcement preventing 8MB images (Marauder v8 / Mini v3) from flashing onto 16MB hardware (LilyGO T-Dongle C5) and vice versa.
- **Dry-Run Simulation**: Simulates flash verification steps prior to physical execution.

---

## 🔒 Authorized Lab & Safety Policy

> ⚠️ **ETHICAL USE MANDATE**: Marauder Studio is designed strictly for authorized penetration testing, security validation, educational research, and owner-authorized laboratory testing.

```text
               User Action (Authorized Lab Workspace)
                                │
                                ▼
        ┌────────────────────────────────────────────────┐
        │            Go Core PolicyGate                  │
        │  [1] Signed Operator & Lab Facility Consent?   │── No ──► REJECT (ErrConsentRequired)
        │  [2] Target BSSID in Explicit Allowlist?       │── No ──► REJECT (ErrAllowlistEmpty)
        │  [3] Non-Wildcard Target (No Broadcast / '*')? │── No ──► REJECT (ErrWildcardProhibited)
        │  [4] Duration <= 120s Bounded Maximum?         │── No ──► REJECT (ErrDurationExceeded)
        │  [5] Log Immutable Local Audit Record          │
        └───────────────────────┬────────────────────────┘
                                │ Pass
                                ▼
                 Dispatch Bounded Radio Job
                    [STOP ALL Abort Active]
```

- **Isolated Workspace**: Potentially disruptive test capabilities are quarantined in the **Authorized Lab** view.
- **Immutable Audit Trail**: Every active test records operator credentials, facility name, targeted BSSIDs, duration, and timestamp.
- **Global Emergency Stop**: A prominent crimson `STOP ALL` button immediately halts all serial commands and active radio routines.

---

## 📋 Supported Hardware Catalog

| Hardware Target | MCU Family | Flash Size | Partition Layout | MicroSD | GPS Fix | BLE |
|:---|:---|:---:|:---:|:---:|:---:|:---:|
| **Marauder v4** | ESP32-WROOM-32 | 4 MB | `min_spiffs` | ✅ | Optional | ✅ |
| **Marauder v6 / v6.1** | ESP32-WROVER-E | 4 MB / 16 MB | `huge_app` | ✅ | ✅ | ✅ |
| **Marauder v7** | ESP32-S3 | 16 MB | `default_16MB` | ✅ | ✅ | ✅ |
| **Marauder v8 (C5)** | ESP32-C5 | 8 MB | `default_8MB` | ✅ | ✅ | ✅ |
| **Marauder Mini v3** | ESP32-C5 | 8 MB | `default_8MB` | ✅ | Optional | ✅ |
| **LilyGO T-Dongle C5** | ESP32-C5 | 16 MB | `default_16MB` | ✅ | ❌ | ✅ |
| **Flipper WiFi Dev Board** | ESP32-S2 | 4 MB | `min_spiffs` | Via Flipper | ❌ | ❌ |
| **Flipper Multi Board S3**| ESP32-S3 | 16 MB | `default_16MB` | ✅ | ✅ | ✅ |

---

## 🏛️ Architecture

```text
┌────────────────────────────────────────────────────────────────────────┐
│ Desktop Application Shell (Wails v2)                                   │
│ Svelte + TypeScript + Tailwind CSS Frontend                            │
│ Virtualized Tables • Live Sparklines • Event Stream Bridges            │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ Typed RPC & Event Bus
┌───────────────────────────────────▼────────────────────────────────────┐
│ Go Application Core (internal/app)                                     │
│ • DeviceManager    • JobManager      • SurveyStore     • MonitorEngine │
│ • PolicyGate       • FirmwareAdapter • StreamingPCAP   • Diagnostics   │
├────────────────────────────────────────────────────────────────────────┤
│ Protocol & Transport Layer                                             │
│ • SerialPort Interface  • PortEnumerator   • WireParser (v1.20+/legacy)│
├───────────────────────────────────┬────────────────────────────────────┤
│ Pure-Go Embedded SQLite Database  │ esptool Process Runner             │
└─────────────────┬─────────────────┴──────────────────┬─────────────────┘
                  │                                    │
             USB Serial                         Bootloader ROM
                  │                                    │
           ESP32 Marauder Hardware (v4 / v6 / v7 / v8 / C5 / DevBoard)
```

---

## 🛠️ Building from Source

### Prerequisites
- **Go**: 1.22 or higher
- **Node.js**: 18+ and `npm`
- **Wails CLI**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Local Development
```bash
# Clone the repository
git clone https://github.com/MJKhaani/ESP32-Marauder.git
cd ESP32-Marauder

# Run full Go test suite
go test -v ./...

# Start Wails live development with hot reload
wails dev
```

### Production Desktop Build
```bash
# Build desktop binary for host OS
wails build -clean

# Output is located in build/bin/
```

---

## 🚀 Automated Release on Tag Workflow

Whenever a Git tag (e.g. `v1.0.0` or `1.0.1`) is pushed to the repository:

```bash
git tag v1.0.1
git push origin v1.0.1
```

GitHub Actions automatically executes `.github/workflows/release.yml`:
1. Compiles native desktop binaries on **macOS**, **Windows**, and **Linux** matrix runners.
2. Compiles standalone CLI helper binaries (`marauder-cli`).
3. Bundles `.zip` and `.tar.gz` packages.
4. Generates a cryptographic `SHA256SUMS.txt` checksum manifest.
5. Publishes a new GitHub Release with all platform assets attached.

---

## 📄 License

Marauder Studio is distributed under the **[MIT License](LICENSE)**.  
Upstream ESP32 Marauder firmware by [justcallmekoko](https://github.com/justcallmekoko/ESP32Marauder) is licensed under the MIT License.
