# ADR 001: Core Architecture and Platform Decisions

## Status
Accepted

## Context
Marauder Studio requires a desktop platform that interacts with hardware serial ports (USB-to-UART bridges and native USB CDC) connected to ESP32 Marauder devices. The application needs to support macOS, Windows, and Linux, with high UI responsiveness, streaming data visualization, bounded resource usage, and safety controls.

## Decisions

### 1. Application Core in Go
- **Decision**: All device control, serial communication, wire protocol parsing, capability discovery, state management, background jobs, persistence, and safety policy enforcement are implemented in Go.
- **Rationale**: Go provides excellent concurrency primitives (`context`, goroutines, channels), cross-platform portability, low memory footprint, and strong typing. The UI never opens serial ports or shells out to system binaries directly.

### 2. Desktop Shell: Wails with Web Frontend
- **Decision**: Wails v2 is chosen as the desktop runtime hosting a Svelte + TypeScript + Tailwind CSS user interface.
- **Rationale**: Wails bundles the Go backend with the native OS webview (WebKit on macOS, WebView2 on Windows, WebKitGTK on Linux), resulting in tiny binaries (<25MB) without embedding Chromium. Svelte gives high rendering performance and fine-grained reactivity required for high-rate Wi-Fi survey and channel monitoring updates.

### 3. Transport Abstraction & Mocking
- **Decision**: Define a clean `SerialPort` interface and `PortEnumerator` interface implemented by `go.bug.st/serial` and `go.bug.st/serial/enumerator`, accompanied by an in-memory `MockPort` for deterministic integration testing.
- **Rationale**: Tests must execute reliably in CI without physical hardware, simulating boot logs, partial reads, timeouts, disconnects, and packet noise.

### 4. Storage: Pure Go Embedded SQLite
- **Decision**: Use `modernc.org/sqlite` instead of CGO-based `mattn/go-sqlite3`.
- **Rationale**: Pure Go SQLite enables smooth cross-compilation across macOS (Darwin amd64/arm64), Windows, and Linux in GitHub Actions runners without requiring complex cross-toolchain C compilers.

### 5. Isolated Authorized Lab Safety Gate
- **Decision**: Any potentially disruptive RF command must pass through an isolated `PolicyGate` in Go that requires:
  1. Signed user authorization acknowledgement
  2. Non-empty allowlist of verified lab BSSIDs / targets
  3. Bounded session duration (max 120s default)
  4. Immediate emergency abort (`Stop All`)
  5. Local audit trail logging
- **Rationale**: Prevents accidental disruption and enforces ethical and legal boundaries programmatically in the Go layer, regardless of frontend actions.
