# lazy-organizer v1.3.0

Native Windows desktop app — no web, no browser.

## What's New

- **GUI rebuilt as native Win32 app** (`cmd/win`, lxn/walk) — real Windows window, native folder picker, native dialogs.
- **Why:** the previous browser-based GUI was flagged by antivirus (localhost server + PowerShell picker + browser auto-open = classic AV heuristic). The native app has no network, no PowerShell, no spawned processes — standard Win32 API only.
- **No OpenGL, no GPU, no extra DLLs** — works on any Windows x86_64 (VM/RDP included).
- Same features: file table with checkboxes, filter bar, All/None/Invert, category change (double-click row), organize with confirmation, undo, config editor.
- 5.7MB binary.

## Downloads

| File | Platform | Size |
|------|----------|------|
| `LazyOrganizer.exe` | Windows x86_64 | 5.7MB |
| `LazyOrganizer-windows-amd64.tar.gz` | Windows x86_64 | 2MB |

## Quick Start (Windows)

1. Download `LazyOrganizer.exe`
2. Double-click → **Select Folder...** → pick a folder → review → **Organize**

Config lives at `%APPDATA%\lazy-organizer\categories.yaml`.

## Note on SmartScreen

If Windows shows "Windows protected your PC": click **More info → Run anyway** (or right-click → Properties → Unblock). That's the "downloaded from internet" mark, not a virus detection. The binary itself is a normal unsigned Win32 app.
