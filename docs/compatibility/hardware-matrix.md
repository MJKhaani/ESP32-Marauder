# Hardware Compatibility Matrix

## Supported Hardware Families

| Board Identifier | Chip Family | Flash Size | Partition Scheme | GPS Support | SD Card | Display |
|---|---|---|---|---|---|---|
| **Marauder v4** | ESP32-WROOM-32 | 4MB | `min_spiffs` | Optional | MicroSD | 2.8" ILI9341 |
| **Marauder v6 / v6.1 / v6.2** | ESP32-WROVER-E | 4MB / 16MB | `huge_app` / `default_16MB` | Optional | MicroSD | 2.8" ILI9341 |
| **Marauder v7** | ESP32-S3 | 16MB | `default_16MB` | Optional | MicroSD | 2.8" ST7789 |
| **Marauder v8 / Mini v3** | ESP32-C5 | 8MB | `default_8MB` | Yes | MicroSD | Color LCD |
| **Marauder Mini** | ESP32-WROOM-32 | 4MB | `min_spiffs` | No | Optional | 1.8" ST7735 |
| **Flipper WiFi Dev Board** | ESP32-S2 | 4MB | `min_spiffs` | No | Via Flipper | None |
| **Flipper Multi Board S3** | ESP32-S3 | 16MB | `default_16MB` | Optional | MicroSD | None |
| **M5StickC / Plus 2** | ESP32-PICO / ESP32-V3 | 4MB / 8MB | `min_spiffs` | No | None | Built-in TFT |
| **M5 Cardputer** | ESP32-S3 | 8MB | `default_8MB` | No | MicroSD | Built-in TFT + KB |
| **T-Dongle C5** | ESP32-C5 | 16MB | `default_16MB` | No | MicroSD | 1.14" ST7789 |

## ESP32-C5 Memory Layout Distinction
> **CRITICAL WARNING**:
> - Marauder v8 and Marauder Mini v3 using ESP32-C5 employ an **8 MB** partition scheme (`default_8MB`).
> - LilyGO T-Dongle C5 uses a **16 MB** partition scheme (`default_16MB`).
> Flashing an 8MB layout onto a 16MB board or vice-versa will render the device unbootable. Marauder Studio's firmware manager checks and enforces the exact partition table and flash size before writing.
