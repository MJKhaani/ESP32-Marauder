# Marauder Studio — Project Specification

> Working title: **Marauder Studio** (rename before release if desired).  
> Status: implementation specification; not a claim that the application exists.  
> Updated: 2026-11-14.  
> Purpose: an offline-first, cross-platform desktop control and analysis application for **ESP32 Marauder** devices in an expressly authorized educational wireless lab.

## 1. Mission

Build a visually polished, responsive, reliable **macOS / Windows / Linux** desktop application that automatically detects compatible ESP32 Marauder boards over native USB serial or a USB-to-UART bridge, identifies the **actual hardware and installed firmware**, discovers the functions the connected firmware exposes, and presents those functions through a consistent graphical interface. Users can perform authorized Wi-Fi/BLE surveys, defensive monitoring, packet capture, review of field artifacts, device administration, and safe firmware upgrades without living in a serial terminal.

The product must **not** claim universal support for features absent from a specific board or absent from that firmware's CLI. Provide a visible support explanation and a fallback to on-device operation when a firmware feature lacks host-side control.

### 1.1 Product principles

1. **Go owns the device.** All transport, firmware management, capability discovery, jobs, persistence, protocol parsing, and validation live in Go. The UI never opens a serial port, shells out, or directly controls hardware.
2. **Live capability detection beats static assumptions.** A built-in versioned catalog helps, but observations from the connected device and verified hardware take precedence.
3. **Safety is part of the user journey.** Passive inspection is the default; any potentially disruptive operation is separate, scoped to an authorized lab target, manually confirmed, and stopped on limit/disconnect.
4. **Never freeze.** Serial reads, scans, file processing, chart aggregation, downloads, and flashing run in supervised cancellable background jobs.
5. **Local and private by default.** No account, cloud service, telemetry, or network access is required for normal operation.
6. **Honest feature presentation.** A feature may be supported by firmware yet not controllable by its serial CLI; show this distinction.
7. **Recoverability first.** Firmware changes require explicit identity checks, backup warnings, validated image layouts, and post-flash verification.
8. **Accessible, native-feeling, responsive desktop UI.** Keyboard navigation, clear device state, usable scaling, light/dark themes, and platform HIG conventions.

## 2. Supported platforms and technology decisions

| Layer | Choice | Reason / constraints |
|---|---|---|
| Application / hardware logic | **Go 1.26+**, stable pinned toolchain | Cross-platform concurrency, testability, bounded workers, strong packages. Update minimum only after CI verifies dependencies. |
| Desktop framework | **Wails v3 stable** (pin tested stable release, initially v3.1.x or later verified release) | Go application backend; lightweight platform WebView; tray, menus, events, generated frontend bindings, multiwindow capabilities. |
| UI | **Svelte + TypeScript + Tailwind CSS** plus accessible component primitives; charts as a separately tested library | Highly responsive, custom contemporary design. **This UI is web technology inside Wails**, not pure Go; all application/device logic remains Go. |
| Alternate all-Go UI | **Fyne** only if a future strict Go-only/no web UI requirement supersedes the modern-design choice | Do not maintain both UI implementations. |
| Serial communication | `go.bug.st/serial` plus `go.bug.st/serial/enumerator` | Portable serial I/O and USB VID/PID enumeration. Use OS-aware fallbacks. |
| Platform enumeration | Thin adapters to Windows device notifications, macOS IOKit/serial discovery, Linux udev where practical | OS event-driven hotplug and bounded polling fallback. |
| Firmware flashing | Isolated, version-pinned **Espressif esptool** helper/backend adapter; document its packaging and licenses | Upstream-supported flash/read/verify semantics for multiple ESP32 families. Do not invent low-level flashing protocol. |
| Persistence | Embedded SQLite (chosen driver verified for target platforms), schema migrations | Device profiles, settings, job summaries, metadata, annotations; bulk PCAP remains on disk. |
| Structured logs | Go `log/slog`; JSON export with sanitization | Supportability and auditability without leaking captured identifiers by default. |
| Artifact handling | Streaming parsers for PCAP/PCAPNG (as present), CSV/JSON/GPX/NMEA, bounded memory | Analysis must work with large captures and malformed inputs. |
| Builds | GitHub Actions native OS runners; tagged releases | Platform-compatible compiled artifacts, signatures where configured, checksums, SBOM. |

### 2.1 Supported desktop targets

- macOS: Intel amd64 and Apple Silicon arm64; signed/notarized `.app`/`.dmg` when credentials exist.
- Windows: amd64 and arm64, including WebView2 runtime checks; signed installer when credentials exist.
- Linux: amd64 and arm64; document GTK4/WebKitGTK runtime requirement and supported distributions; offer AppImage or distro-specific packages after native testing.
- Mobile/iOS/Android, remote serial bridges over TCP, and headless remote servers: **not in v1 scope**. Do not represent Wails experimental mobile support as production support.

### 2.2 Dependencies and licensing

- Check upstream licenses before pinning; the official ESP32 Marauder repository carries an **MIT** license at the time of writing. Include its notice when redistributing any upstream files.
- Keep the desktop product's own license explicitly selected by the maintainer rather than presuming a license from this document.
- Do not embed copyrighted upstream artwork/firmware or redistribute esptool/Python runtimes without documented license/notice and supply-chain review.
- Pin Go modules, npm dependencies, flashing helper, and downloaded firmware by version and digest.

## 3. Supported devices and capability model

The upstream supported-hardware list is the **starting catalog**, not evidence that the desktop can execute every menu feature on every device.

### 3.1 Hardware families to recognize

- Official generations: Marauder v4, v6, v6.1/v6.2, v7, v8; Marauder Mini and Mini v3; Marauder Kit.
- Flipper variants: Flipper Zero WiFi Dev Board (ESP32-S2), Flipper Zero Multi Board S3, Marauder Dev Board Pro, BFFB (different revisions have different firmware targets).
- Community/DIY targets where officially listed: ESP32 LDDB/headless, CYD variants, M5StickC / Plus 2, M5 Cardputer, M5 Nano C6, ESP32-C5 DevKit, Dual Mini C5, T-Dongle C5, and other targets added to official releases.
- Chip families as applicable: classic ESP32, S2, S3, C5, C6. **Do not infer Bluetooth, dual-band, GPS, storage, screens, or capture from chip family alone.**

### 3.2 Detection and connection requirements

1. Enumerate native USB CDC/USB-serial adapters and USB-to-UART bridges automatically on launch and on hotplug.
2. Provide names and stable identifiers where exposed: path, VID, PID, USB serial number, manufacturer, product string, device topology, driver, and last seen.
3. Identify common ESP32-related serial bridges as **candidates**, never as guaranteed Marauder devices: CP210x, CH340/CH910x, FTDI, and native USB serial. Other vendors/bridges remain selectable.
4. Use safe, bounded, passive discovery by default. Only open a candidate after user consent to auto-connect or an explicit connect action; avoid disrupting unrelated serial equipment.
5. Default Marauder serial link: **115200 baud**, configurable for advanced users. On connection, use documented `info` and `help` responses to fingerprint the **running firmware**, rather than blind command probing of every port.
6. Account for boot logs, echo, partial UTF-8, CR/LF variants, slow response, disconnects, permission errors, and devices already held by another program.
7. Show reconnect, retry/backoff, and **manual port selection** even when auto-detection fails. Identify ports by stable key rather than path alone; paths may change across reboot/flash.
8. Distinguish **runtime mode** vs **ROM download/bootloader mode**. A blank board can be flashable even if the Marauder CLI is unavailable.
9. Treat a Flipper USB/UART bridge as a special transport that may require explicit setup on the Flipper; document what channel and rate are required.
10. Support multiple simultaneously attached devices with independent session state, locks, job history, and capability lists. Never broadcast a command to all devices by accident.
11. Suspend normal CLI traffic while flashing; exclusive per-device ownership of the port is mandatory.
12. Show USB permission troubleshooting (Windows driver, Linux dialout/udev, macOS device path) without silently requesting elevated privileges.

### 3.3 Capability registry (core requirement)

For each attached device, build a **capability matrix**, with at least:

- `feature_id`, display label, category, risk level, required hardware and firmware range (if known).
- `hardware_supported`, `firmware_reported`, `cli_exposed`, `desktop_implemented`, `currently_available`.
- `state`: **Available**, **Requires hardware**, **Firmware too old/new**, **Firmware-present / on-device only**, **Not exposed through CLI**, **Unsupported**, **Unknown/needs verification**, **Temporarily busy**.
- `evidence`: which source proves each status (`info`, `help`, device profile, release metadata, manual user confirmation, or observed success).
- Diagnostic reason and actionable next step.

**Discovery precedence:** validated runtime evidence > matched exact hardware manifest > upstream release manifest > version-specific compatibility fixtures > generic catalog. Contradictions are flagged, never silently resolved in favor of an optimistic claim. Unknown is distinct from unsupported.

The app must refresh capabilities after a firmware upgrade, reconnect, device swap, or known storage/peripheral change. Preserve an exportable snapshot for bug reports. Device capability badges are informative; do not imply all on-device UI functions have corresponding serial verbs.

## 4. Feature set

### 4.1 Device dashboard and diagnostics — P0

- Connection state, board/revision, chip family, firmware version/build, CLI API/known protocol profile, USB port, SD/GPS/BLE support, flash capacity when reliably known, signal/health status where exposed.
- Firmware release comparison, hardware match confidence, update eligibility, USB/serial connection diagnostics, permission hints.
- Commands visible in live `help`, unavailable functions, parsed settings, raw read-only serial console with sanitized export.
- Device self-test: read-only checks for CLI responsiveness, SD availability, GPS status, BLE readiness and storage list **only if the firmware supports these checks**.
- Quick actions: connect/disconnect, refresh status, start passive survey, stop active job, open capability matrix.
- Per-device profile and history (nicknames, last used, evidence provenance, last successful firmware verification).

### 4.2 Wi-Fi AP and station survey — P0

- AP and station survey initiated via the command(s) supported by the **actual** firmware. Modern firmware (v1.20 documented CLI) uses `scanall`; older `scanap` / `scansta` are **not** universally valid.
- Live AP grid: SSID, BSSID, channel, band, RSSI, security information when observable, first/last seen, observed frame counts, optional vendor/OUI attribution from an offline dataset.
- Station grid and observed AP association where available; unknown association remains unknown.
- Filters by SSID/channel/signal/security/band and **authorized target allowlist**; charts update without manipulating radio state.
- Saved snapshots, diff across sessions, duplicate merging, channel occupancy estimates from samples (label as estimates rather than spectrum analyzer readings).
- Export filtered reports to CSV/JSON and sanitized HTML/PDF only if a renderer is implemented and tested.
- Local session markers and annotations; no external enrichment API by default.

### 4.3 Channel monitoring and defensive monitoring — P0

- Live channel histogram and timeseries: AP counts, signal distributions, and observed event rates; show sample interval and missing data.
- Channel selection, hopping where exposed, band support by hardware, visibility of the radio state.
- Beacon/probe analysis and deauthentication **detection/alerts** using firmware-supported sniffers.
- Detection panels for suspicious management frames and anomalies; avoid claiming definitive attack attribution from incomplete samples.
- Signal-strength tracking / Fox Hunt / MAC Track if compiled and controllable. Use non-person-identifying language in results.
- Optional alert rules for lab-configured observation thresholds. **Alerts never autonomously initiate RF transmissions.**

### 4.4 Packet sniffing, capture, and offline analysis — P0

- Start/stop supported passive sniffing modes; expose only available filters and capture destinations.
- Capture progress: elapsed time, packet/frame counts if provided, storage remaining if reported, warnings when SD is required.
- Preserve original captures and metadata. Pull PCAPs from device **only where the transport/firmware provides a real file-transfer path**; otherwise guide the user to SD-card import. Never assert that serial CLI supports file download without proof.
- Import PCAP/PCAPNG locally; display frame type/subtype, time, channel, signal if recorded, addresses, and safe protocol summary. Unavailable capture metadata is labeled accordingly.
- Offline management-frame timelines, beacon counts, deauth-event counts, channel heatmaps, and permitted lab session comparisons.
- Support firmware-specific PMKID/EAPOL/SAE **observation/capture presence** with clear privacy warnings. No password cracking, credential recovery, or automatic export to attack tools.
- Streaming file reads, indices or paging for large files, size limits, cancellation, and malformed-file handling; memory usage remains bounded.
- Optional launch/open-in-Wireshark action for saved captures, enabled only when user explicitly chooses it.

### 4.5 Bluetooth/BLE observation — P1

- Discover BLE device observations and advertisement metadata when the connected hardware+firmware support Bluetooth.
- BLE device timeline, RSSI charts, service/manufacturer identifiers if actually provided, duplicate/rotating identifier caveats.
- Firmware-specific defensive scanners and named detection functions (e.g., supported trackers/advertisement categories) presented with evidence and privacy notes.
- BLE capture/import where genuinely exposed; make no promise of full protocol-level BLE packet capture on targets that only expose advertisements.
- **Advertisement spam, impersonation, spoofing, and unsolicited notification operations** belong in the high-risk capability inventory and are **disabled by default**; do not silently implement or trigger them.

### 4.6 GPS, Recon, and field artifacts — P1

- GPS fix, quality, UTC time, location accuracy/status, NMEA diagnostics if exposed.
- GPS route import, GPX visualization and timestamps; geofencing and permitted collection boundaries, if supported by firmware.
- Recon/wardrive file import, offline review, deduplication, filtering, session comparison and export.
- If map tiles require Internet, show an explicit request, use opt-in cached/offline tiles when licensed, and prevent inadvertent upload of sensitive coordinates.
- Remote ID passive scan/telemetry viewer on supported hardware; avoid assumptions about identity or owner attribution.
- SPIFFS / SD browsing and backup/restore **only when safe, supported, and verifiable**. Warn about secrets in files and backups.

### 4.7 Authorized active lab tests — P1; isolated, opt-in

A separate **Authorized Lab** workspace covers controlled, potentially disruptive RF test capabilities that the upstream firmware may expose (including deauthentication testing). It is not part of the ordinary passive survey flow.

- Feature availability is still derived from real hardware/firmware commands; the product does not invent an unsupported attack operation.
- Require explicit acknowledgement of ownership/authorization, named lab environment, **user-selected allowlisted BSSID / target identifiers**, scope, and a short bounded session duration. Store a local audit record.
- Add a conspicuous **Stop All** action, per-session timeout, one active RF operation per radio, and no unattended scripts, bulk target selection, automatic neighbor selection, or repeating campaigns.
- Never automatically switch from scan/detection to injection. Enforce scope in the Go service layer, not only the UI.
- Reset/stop on disconnect, cancellation, scope change, timeout, or application exit when communication is still available. UI must report when a disconnected device could not be confirmed stopped.
- Show local lab guidance emphasizing isolation and minimizing unintended RF effects; physical radio signals can extend beyond an intended room.
- **No credential-harvesting portal templates, mass-target workflows, evasion, persistence, or amplification modes in v1.** Document those firmware features as unimplemented/disabled where relevant instead of falsely claiming full remote operation.
- Keep active-test command implementations in a separately reviewed package with explicit authorization-policy and device-specific integration tests. Do **not** treat a legal disclaimer as permission to ignore safety restrictions.

### 4.8 Firmware manager and flashing — P0

- Recognize devices in firmware **runtime** and **download/bootloader** modes, including devices without a working Marauder installation.
- Show upstream stable releases and exact board target binaries; work offline with imported artifacts after one-time catalog acquisition.
- Strict target catalog keyed by **exact board model, PCB revision, chip family, flash size, partition-layout ID, bootloader, partitions, OTA metadata, application binary, offsets, and required migrations**.
- Verify the GitHub release source and available SHA-256 digests; if upstream does not publish trustworthy signed checksums, transparently mark artifact provenance and validate a manifest pinned by this application's maintainers. Never describe HTTPS alone as a cryptographic signature.
- Offer **safe normal update** only if compatible layout is known; **full install/recovery** is a separate explicit workflow.
- Do not mistake an application-only `.bin` for a complete bare-device installation. Distinguish full install from app-only upgrade.
- **ESP32-C5 warning:** v8/Mini v3 and several C5 variants use an 8 MB `default_8MB` layout, while the T-Dongle C5 has a different 16 MB layout. Never swap layouts or offsets based only on C5 family.
- Preflight: exclusive port, stable connection, model confidence, image provenance and checksum, partition plan, necessary prerequisites, battery/power warning, backup reminder, disk space, and confirmation of destructive erase behavior.
- Streaming step view: prepare > identify bootloader > check chip > obtain images > verify manifest > flash > read-back/verify if supported > reboot > reconnect > run `info` > compare expected firmware+hardware > refresh capabilities.
- Flash jobs must support progress, cancellation **where safe**, and explicit indications that interrupting an in-progress write can leave a device unbootable. Never claim rollback unless the device actually supports it.
- Guardrails: refuse a contradictory chip/model/flash-size combination. Unknown board revisions require manual expert confirmation; no unsafe 'guess nearest firmware'.
- Advanced: import custom images with a full user-provided manifest and a prominent 'unsupported / risk of bricking' warning; keep this outside one-click flows.
- Record firmware job logs and helpful recovery instructions; never include secrets in public bug reports.
- Prefer upstream official installer paths or a pinned, tested esptool adapter; do not promise browser-only methods for unsupported boards.

### 4.9 Console, logs, reports, and settings — P1

- Read-only console by default with timestamped, color-coded parsed messages and optional raw view.
- Advanced free-form terminal may be enabled in an expert mode, with warnings and clear visibility that direct commands bypass normal feature UX; do not use it for unattended automation.
- Per-session event log, no raw identifiers in generic diagnostics, opt-in sanitized bug-report bundle.
- Local report builder with scope, hardware/firmware, session window, annotated findings, charts, sanitized identifiers, capture references and hashes.
- Preferences: theme, UI density, language-ready i18n, serial defaults, reconnect policy, retention days, storage directory, file size limits, diagnostics level, offline mode, external map opt-in, and permission troubleshooting.
- Accessible hotkeys (connect, stop, filter, export), screen-reader labels, reduced-motion setting and keyboard-only workflows.

## 5. Information architecture and design system

### 5.1 Primary navigation

1. **Overview** — device cards, session status, quick passive actions, warnings.
2. **Devices** — discovered USB serial devices, board details, compatibility/capability explorer.
3. **Wi-Fi Survey** — AP/station tables, network detail drawer, saved snapshots.
4. **Channels & Alerts** — live graphs, passive event monitor, detection rules.
5. **Captures** — capture controller, files, PCAP analysis.
6. **Bluetooth** — available BLE features, observations and charts.
7. **GPS & Recon** — route/wardrive/Remote ID artifacts where supported.
8. **Authorized Lab** — separately gated manual active-test workspace.
9. **Firmware** — releases, flash wizard, recovery.
10. **Sessions & Reports** — annotations and sanitized exports.
11. **Console & Diagnostics** — serial transcript, USB troubleshooting, audit.
12. **Settings** — global preferences, privacy, retention, about.

Hide irrelevant modules intelligently but keep an inspectable **Unsupported Features** view so users can answer "why can't I use this?" without guessing.

### 5.2 Visual language

- Modern technical-instrument style: restrained neutral palette, one primary accent, consistent spacing/type scale, high contrast, clean cards, refined data tables, sparklines and confidence states.
- Dark and light themes; respect OS preference, independent override, and reduced motion.
- Three-pane desktop mode on large windows (navigation / data / detail), two-pane compact mode, stacked navigation on narrow windows; never require a fixed 1440px viewport.
- Tables must virtualize thousands of rows and support sorting, filtering, pinned identifiers, copy, export, persisted column layouts.
- UI always shows device selection, selected device connection state, current job, and the global Stop All control.
- Use non-color indicators alongside status colors (icon/label/description).
- Draw only data actually observed. Display age of data and lost-sample indicators.
- Prefer native OS menus, file dialogs, notifications, tray behavior, and shortcuts through Wails services.

## 6. Suggested architecture

```text
┌───────────────────────────────────────────────────────────────────┐
│ Wails v3 Desktop Shell                                            │
│ Svelte/TypeScript UI: routed views, stores, virtual tables, charts│
└──────────────────────────────┬────────────────────────────────────┘
                               │ Typed generated bindings / events
┌──────────────────────────────▼────────────────────────────────────┐
│ Go Application Services                                           │
│ DeviceService • CapabilityService • JobService • CaptureService   │
│ SurveyService • FirmwareService • Reports • Privacy/Policy        │
├───────────────────────────────────────────────────────────────────┤
│ Protocol and Device Layer                                         │
│ Discovery → Port Lock → Device Session → Command Queue → Parser   │
│ Protocol Profiles / Evidence Cache / Hotplug & Reconnect          │
├───────────────────────────┬───────────────────────────────────────┤
│ File/SQLite Persistence    │ Firmware Adapter (esptool process)    │
└──────────────┬────────────┴──────────────┬────────────────────────┘
               │                           │
        USB CDC / UART                Flash/download mode
               │                           │
          ESP32 Marauder device(s), all peripherals on-device
```

### 6.1 Proposed repository layout

```text
/
  AGENT.md
  PROJECT.md
  README.md
  LICENSE                      # selected explicitly by maintainer
  CHANGELOG.md
  go.mod go.sum
  main.go
  Taskfile.yml
  /cmd/marauder-studio/
  /internal/app/
  /internal/device/             # discovery, state machine, session manager
  /internal/serialtransport/    # OS adapters, reader, writer, port ownership
  /internal/marauder/           # CLI profiles, wire parsers, command registry
  /internal/capabilities/       # evidence and compatibility resolution
  /internal/jobs/               # cancellation, progress, limits, lifecycle
  /internal/survey/
  /internal/monitor/
  /internal/capture/
  /internal/ble/
  /internal/gps/
  /internal/firmware/           # catalog, validation, flash adapter
  /internal/policy/             # authorization, active-lab constraints
  /internal/store/              # migrations, metadata, settings
  /internal/reports/
  /internal/diagnostics/
  /frontend/src/
  /frontend/tests/
  /assets/
  /firmware/catalog/           # hardware and layout manifests, not opaque blobs
  /testdata/                   # sanitized serial transcripts/PCAP fixtures
  /tests/integration/
  /tests/hardware/             # opt-in, local lab only
  /docs/architecture/
  /docs/protocol/
  /docs/compatibility/
  /docs/security/
  /docs/progress/
  /.github/workflows/
```

### 6.2 Transport and event contract

- Exactly **one serial read loop** and one ordered **command dispatcher** per device. State-changing commands have acknowledgements/timeouts; when the firmware lacks a reliable ACK, use a conservative state machine and surface uncertainty.
- Split raw read bytes, framed/machine responses where available, and human text with a version-aware decoder; never rely on one regex for all firmware generations.
- Use typed Go events for `DeviceAdded`, `Connected`, `Disconnected`, `CapabilityChanged`, `ScanObservation`, `CaptureProgress`, `JobStatus`, `Warning`, `FirmwareProgress`.
- Every long-lived job has `context.Context`, a deadline if applicable, cancellation, concurrency limits, and a unique ID. Resources close deterministically.
- Channel buffers are **bounded**; aggregate/debounce high-rate observations for the UI, while optional raw capture is streamed to disk. Show overflow counters instead of silently treating discarded UI events as complete data.
- Never block a serial reader on slow rendering, SQLite writes, file parsing, or network downloads.
- Serialize mutually exclusive actions: survey/capture/channel operations and firmware flash on the same device; support a compatibility-aware operation matrix.
- All frontend calls return promptly with a job ID; progress and results arrive through typed events or paged reads.
- Reconnect must re-run `info`/`help`; never assume the device behind a reused port is the previous one.

### 6.3 Domain types (illustrative, not wire spec)

```go
type FeatureStatus string // available, unsupported, on_device_only, unknown, busy, etc.

type FeatureCapability struct {
    ID                string
    Status            FeatureStatus
    FirmwareReported  bool
    CLIExposed        bool
    DesktopImplemented bool
    Reason            string
    Evidence          []string
}

type DeviceIdentity struct {
    StablePortID string
    Path         string
    ChipFamily   string
    BoardModel   string
    PCBRevision  string
    Firmware     string
    Confidence   string
}

type JobState string // queued, running, stopping, succeeded, failed, interrupted
```

Do not confuse the illustrative type model with an upstream guaranteed machine-readable schema.

## 7. Data, privacy, and security

- Keep all observation data on the user's machine unless a **separately explicit** export/upload occurs. No automatic public map/wardriving upload.
- Mark captures, BSSIDs/MAC addresses, SSIDs, BLE identifiers, GPS/Remote ID telemetry, saved Wi-Fi profiles, and portal submissions as potentially sensitive. Firmware's on-screen **Censor** setting does **not** guarantee redaction of serial output or saved files.
- Sanitized report/export mode defaults to masking addresses, credentials, and precise coordinates, with reversible/irreversible options only if explained and tested.
- Separate internal raw session captures from support logs; do not include raw PCAP or GPS traces in bug reports automatically.
- Protect local databases and files with OS permissions; use OS credential storage for secrets if a future feature requires credentials. Never print tokens/keys into logs.
- Validate untrusted serial bytes, PCAP headers, NMEA/CSV/JSON input, firmware metadata, archive entries, file names, and downloaded assets. Defend against path traversal, excessive sizes, malformed binary records, and UI injection.
- Limit external requests to documented firmware-release fetching and opt-in services; allow strict offline mode.
- Route high-risk active operations through a Go policy gate and append a local audit record (not a remote surveillance system).
- Define retention controls, deletion of session data, secure-enough export warnings, and a README privacy notice.

## 8. Authorization, acceptable use, and liability notice

**Scope:** This application is intended exclusively for educational research, authorized penetration testing, security validation, and investigation of networks/devices the operator owns or has explicit permission to assess. The initial test environment is the author's **own local laboratory AP**.

The application must visibly tell operators that certain wireless transmissions can disrupt connectivity, interfere with unrelated nearby devices, breach legal restrictions, or expose private traffic. An in-app acknowledgment is a friction/notice mechanism, **not proof of authorization**.

### Required proposed notice (review with local counsel before publication)

> Marauder Studio is supplied for educational and authorized security-testing use only. Users must obtain all permissions required by applicable law and must limit their activities to systems and radio environments they are entitled to test. The developers and contributors do not authorize misuse or knowingly endorse unlawful, disruptive, or privacy-invasive conduct. Operators are responsible for their decisions and for complying with all applicable laws and third-party rights. The software is provided “AS IS,” without warranties, to the maximum extent permitted by law; limitations of liability apply only where enforceable.

**Important:** Merely stating "all misuse is the user's responsibility" does not automatically eliminate the author's/developer's legal liability. Do not make an absolute guarantee of immunity. Obtain jurisdiction-specific legal advice for a final release notice, policy, and software license.

## 9. Compatibility and upstream protocol governance

- Official firmware repository: <https://github.com/justcallmekoko/ESP32Marauder>
- Upstream CLI: <https://github.com/justcallmekoko/ESP32Marauder/wiki/CLI>
- Upstream command index: <https://github.com/justcallmekoko/ESP32Marauder/wiki/CLI-Usage>
- Firmware features: <https://github.com/justcallmekoko/ESP32Marauder/wiki/Applications>
- Hardware table: <https://github.com/justcallmekoko/ESP32Marauder/wiki/Supported-Hardware>
- Updates/flashing: <https://github.com/justcallmekoko/ESP32Marauder/wiki/Update-Firmware>
- Releases: <https://github.com/justcallmekoko/ESP32Marauder/releases>
- Privacy: <https://github.com/justcallmekoko/ESP32Marauder/wiki/Privacy-and-Data-Handling>
- Wails: <https://wails.io/>
- Go serial enumeration: <https://pkg.go.dev/go.bug.st/serial/enumerator>
- Espressif esptool: <https://docs.espressif.com/projects/esptool/en/latest/>

For each upstream release, compare command output, hardware targets, release asset suffixes, layouts, firmware feature changes, and privacy notes. Store an auditable compatibility fixture and link to the source commit/tag. Never hard-code that all earlier command names exist in the current firmware.

## 10. Phased roadmap, exit criteria, and priorities

### Phase 0 — Research and scaffolding

- Verify current release/docs and license; build versioned hardware+firmware reference table, acceptance tests, architecture decision records, Wails shell, Go service layer, formatting/lint/test CI.
- **Exit:** app launches on three OSes; mock device visible; source and test provenance documented.

### Phase 1 — Device connection and feature detection (P0)

- USB enumeration/hotplug, manual connect, `info`/`help`, transport state machine, capability matrix, read-only console, error/reconnect UI.
- **Exit:** known and unknown port cases, fake serial streams, unplug/replug, multiport selection pass; no GUI hangs.

### Phase 2 — Passive survey and monitoring (P0)

- Supported AP/station scans, channel charts, passive alerts, sessions and CSV/JSON export.
- **Exit:** compatibility fixture tests for old/current command differences, replay, large-observation load, cancellation.

### Phase 3 — Capture, analysis, and field data (P0/P1)

- Captures, SD-import workflow, PCAP reader, BLE observation, GPS/Recon/wardrive import (as supported), annotations.
- **Exit:** malformed input handling, bounded memory, known-good PCAP comparison, private-data sanitization tests.

### Phase 4 — Firmware manager (P0)

- Exact board catalog, trusted downloads, compatibility and partition validation, pinned flasher adapter, recovery wizard, post-flash discovery.
- **Exit:** dry-run tests for all published target manifests; supervised physical flashing on representative hardware; invalid layout safely rejected.

### Phase 5 — Authorized Lab & advanced diagnostics (P1)

- Separate reviewed active-test UI with manual consent, explicit lab target scope, rate/time bounds, audit, failsafe, tested stop semantics; advanced console and defensive detections.
- **Exit:** prohibited/unscoped operations consistently denied; failures surface their stop-uncertainty; user can halt any supported job quickly.

### Phase 6 — Polish and release

- Platform-native packaging, performance tuning, accessibility, localization readiness, docs, checksums/SBOM, opt-in release artifact audit.
- **Exit:** reproducible versioned release, clean-install smoke tests, documented limitations, no critical test failures.

## 11. Definition of done and acceptance targets

- All implemented features have tests (unit/integration and hardware tests where needed), user-facing documentation, and an updated capability manifest/compatibility table.
- CI: `go test ./...`, `go test -race ./...` where supported, formatting, static analysis, frontend typecheck/tests, build all supported packages, dependency audit, and fixture replay.
- USB unplug or slow firmware response cannot freeze the UI; job cancellation results in bounded, visible state transition; port ownership is deterministic.
- 10,000+ synthetic survey rows can be navigated without bulk DOM rendering; sustained streamed observations remain memory bounded and chart update rate is throttled.
- Imported large captures can be paged/cancelled without loading the whole file into memory. Exact performance targets must be measured and recorded on reference machines.
- No phantom capabilities: every feature claims an evidence source, and unknown hardware reports Unknown.
- Flashing refuses mismatched targets/layouts; post-flash identity is checked; any failure includes a recovery path.
- Authorized Lab operations cannot start without explicit user-scoped authorization state, manual target allowlist, and enforced limits; no automatic campaigns.
- Default diagnostics redact personal/sensitive fields; raw file exports require deliberate action.
- Every completed work item is reflected in `/docs/progress/`, roadmap, changelog (when applicable), tests and examples.

## 12. Explicit non-goals for v1

- Universal remote execution of functions that exist **only on a board's display menu**.
- Controlling arbitrary ESP32 firmware that is not verified as compatible Marauder.
- Remote fleet management, cloud sync, data sharing, unattended lab attack orchestration.
- Password cracking, credential harvesting, evasion, mass disruption, or public-network attacks.
- Guaranteed physical containment of RF transmissions, security attribution from mere observations, or automatic detection of ownership/authorization.
- Guaranteed recovery of a device with broken bootloader/hardware or an unsupported flash layout.

---

## 13. Final product outcome

Users plug in an ESP32 Marauder, see its serial port automatically detected, connect, inspect the board/firmware and precise **Supported / Not supported / On-device only / Unknown** feature map, run permitted passive analyses with live charts and responsive UI, review captures, and flash the exact correct official firmware through a safe guided workflow. A clearly segregated Authorized Lab interface provides narrowly scoped manual test controls subject to device capability and safety review. Every design claim is backed by tests and observed behavior, not assumed command availability.
