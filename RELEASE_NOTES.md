# lazy-organizer v1.1.0

Windows desktop GUI release.

## What's New

- **Windows desktop app** — first official Windows GUI build (`LazyOrganizer.exe`, x86_64, portable, no install)
- App icon + version metadata embedded (FyneApp.toml)

## Downloads

| File | Platform | Size |
|------|----------|------|
| `LazyOrganizer.exe` | Windows x86_64 (GUI) | 12MB |
| `LazyOrganizer-windows-amd64.tar.gz` | Windows x86_64 (GUI) | ~10MB |

CLI binaries for Linux (amd64/arm64), macOS (amd64/arm64), and Windows (amd64) are in the v1.0.0 release.

## Desktop App Features

- Native file table with checkboxes per file
- Filter bar — search by name, extension, or category
- Select All / None / Invert — bulk selection controls
- Category picker — click category cell to change
- Config editor — add/edit/delete categories in GUI
- Organize with confirmation — only moves checked files

## Quick Start (Windows)

1. Download `LazyOrganizer.exe`
2. Double-click to run
3. Pick a folder → review the table → **Organize**

Config lives at `%APPDATA%\lazy-organizer\categories.yaml`.
