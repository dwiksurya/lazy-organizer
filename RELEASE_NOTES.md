# lazy-organizer v1.2.0

New browser-based desktop GUI — works on ANY Windows machine (no GPU/OpenGL required).

## What's New

- **GUI rebuilt as local web app** (`cmd/web`) — the exe starts a local server and opens your browser. Pure Go, no CGO, no OpenGL, no extra DLLs. Fixes the blank-window issue on VMs / RDP / machines without OpenGL 2.0.
- **Smaller binary** — 6MB (was 12MB)
- **Windows, Linux, macOS** — same GUI everywhere
- Same features: file table with checkboxes, filter bar, category picker, config editor, organize with confirmation, undo

## Downloads

| File | Platform | Size |
|------|----------|------|
| `LazyOrganizer.exe` | Windows x86_64 | 6MB |
| `LazyOrganizer-windows-amd64.tar.gz` | Windows x86_64 | 2.5MB |
| `LazyOrganizer-linux-amd64.tar.gz` | Linux x86_64 | 2.4MB |
| `LazyOrganizer-darwin-arm64.tar.gz` | macOS Apple Silicon | 2.3MB |
| `LazyOrganizer-darwin-amd64.tar.gz` | macOS Intel | 2.4MB |

## Quick Start (Windows)

1. Download `LazyOrganizer.exe`
2. Double-click — your browser opens with the organizer
3. Type a folder path (e.g. `C:\Users\you\Downloads`) → Scan → review → **Organize**

Config lives at `%APPDATA%\lazy-organizer\categories.yaml` (Linux: `~/.config/lazy-organizer/`, macOS: `~/Library/Application Support/lazy-organizer/`).
