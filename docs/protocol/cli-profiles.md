# Marauder CLI Wire Protocol & Profiles

## Overview
ESP32 Marauder runs an interactive command-line interface over USB serial (115200 baud, 8N1). Different firmware versions expose varying command sets, output formatting, and capabilities.

## Protocol Profiles

### Profile 1: Modern Firmware (v1.20+)
- **Discovery**: `info` outputs firmware version `v1.20.*` or higher.
- **Survey**:
  - `scanall`: Scans all 2.4 GHz channels for both Access Points and associated Stations simultaneously.
  - `stopscan`: Halts active scanning or sniffing.
- **Sniffing**:
  - `sniffpmkid`: Passively monitors for PMKID key exchange frames.
  - `sniffraw`: Captures raw 802.11 frames to attached SD card or SPIFFS.
  - `sniffbeacon`: Passively inspects beacon frames.
  - `sniffdeauth`: Passive deauthentication frame monitor.
  - `sniffesp`: ESP-NOW packet detection.
- **Bluetooth**:
  - `sniffbt`: BLE advertisement sniffing.
- **Diagnostics**:
  - `channel`: Sets or queries radio channel.
  - `settings`: Displays or toggles configuration flags (e.g. Censor, Channel Hopping).

### Profile 2: Legacy Firmware (<v1.20)
- **Discovery**: Older firmware versions report version strings like `v0.13.*`.
- **Survey**:
  - `scanap`: Scans exclusively for Access Points.
  - `scansta`: Scans exclusively for client stations.
  - `stopscan`: Halts active scan.
- **Sniffing**:
  - `sniffpmkid`: PMKID frame sniffing.
  - `sniffbeacon`: Beacon frame sniffing.

### Wire Framing & Chunking
Marauder CLI responses are streamed as ASCII text terminated by `\r\n` or `\n`. Output may be fragmented across serial buffer reads.
The parser maintains an internal byte buffer, splitting on line terminators and extracting structured data with regular expressions and tokenizers.
Bootloader output, ANSI escapes, and interactive prompts (e.g. `>` or `esp32>`) are stripped or categorized as console diagnostics.
