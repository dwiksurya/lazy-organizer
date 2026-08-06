# lazy-organizer v1.1.1

Windows startup diagnostic fix.

## What's New

- **Startup logging** — app now writes `lazy-organizer.log` (and panic stack traces) to `%TEMP%` on Windows. If the window doesn't open, the log shows why.

## Downloads

| File | Platform | Description |
|------|----------|-------------|
| `LazyOrganizer.exe` | Windows x86_64 | Desktop GUI (portable) |
| `LazyOrganizer-windows-amd64.tar.gz` | Windows x86_64 | Desktop GUI (tar.gz) |
| `LazyOrganizer-debug.exe` | Windows x86_64 | Console build — run in cmd to see errors |

## Quick Start (Windows)

1. Download `LazyOrganizer.exe`
2. Double-click to run
3. Pick a folder → review the table → **Organize**

Config lives at `%APPDATA%\lazy-organizer\categories.yaml`.
Log lives at `%TEMP%\lazy-organizer.log`.
