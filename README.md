# lazy-organizer

CLI + native Windows desktop app to organize messy folders.

## Downloads

[![GitHub Release](https://img.shields.io/github/v/release/dwiksurya/lazy-organizer?color=blue)](https://github.com/dwiksurya/lazy-organizer/releases/latest)

**Windows desktop app** — native Win32 GUI, no install, no GPU/OpenGL, no extra DLLs.

| File | Platform |
|------|----------|
| [LazyOrganizer.exe](https://github.com/dwiksurya/lazy-organizer/releases/latest/download/LazyOrganizer.exe) | Windows x86_64 |
| [LazyOrganizer-windows-amd64.tar.gz](https://github.com/dwiksurya/lazy-organizer/releases/latest/download/LazyOrganizer-windows-amd64.tar.gz) | Windows x86_64 |

CLI binaries for Linux/macOS/Windows are on the [Releases page](https://github.com/dwiksurya/lazy-organizer/releases).

## Quick Start — Desktop App

**Windows:** download `LazyOrganizer.exe`, double-click. Click **Select Folder...** → pick a folder → review → **Organize**.

## Quick Start — CLI

```bash
# Build
go build -o lazy-organizer .

# Generate config
./lazy-organizer -init-config

# Preview (nothing is moved)
./lazy-organizer -dir ~/Downloads -dry-run

# Organize
./lazy-organizer -dir ~/Downloads

# Interactive mode (choose category per file)
./lazy-organizer -dir ~/Downloads -interactive

# Undo
./lazy-organizer -dir ~/Downloads -undo

# TUI config editor
./lazy-organizer -gui
```

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-dir` | `.` | Target folder |
| `-dry-run` | false | Preview without moving |
| `-interactive` | false | Per-file: choose/change category |
| `-yes` | false | Skip confirmation |
| `-undo` | false | Undo last move |
| `-config` | OS-specific | Custom config path |
| `-init-config` | false | Generate default config and exit |
| `-gui` | false | Open TUI config editor |

## Desktop App Features

- **📂 Select Folder** — native folder picker dialog
- **File Table** — checkbox per file, Name / Ext / Category columns
- **Category change** — double-click a row to pick a category
- **Filter bar** — search by name, extension, or category
- **Select All / None / Invert** — bulk selection
- **▶ Organize** — move checked files with confirmation
- **↩ Undo** — revert last organization
- **⚙ Config Editor** — edit/add/delete categories in the GUI

## Classifier

Files are classified by two methods (first match wins):

1. **Extension** — `.pdf` → Documents, `.jpg` → Images, etc.
2. **Keyword** — filename substring match. `invoice_march.bin` → Documents (keyword "invoice")

## Project Structure

```
lazy-organizer/
├── internal/core/          ← shared logic (config, scanner, mover)
│   ├── config.go           ← YAML config + classifier
│   ├── scanner.go          ← directory scanner
│   ├── mover.go            ← file mover + undo
│   └── core_test.go        ← tests
├── cmd/win/               ← native Windows GUI (lxn/walk, pure Go)
│   └── main.go            ← Win32 GUI: table, filter, config editor
├── cmd/gui/               ← legacy Fyne GUI (needs OpenGL)
│   └── main.go
├── main.go                 ← CLI entry point
├── interactive.go          ← CLI interactive mode
├── tui.go                  ← CLI TUI config editor
└── Makefile                ← build targets
```

## Build from Source

```bash
git clone https://github.com/dwiksurya/lazy-organizer.git
cd lazy-organizer

# CLI (pure Go)
go build -o lazy-organizer .

# Desktop GUI (native Windows — pure Go, no CGO)
GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui" -o LazyOrganizer.exe ./cmd/win/

# Legacy Fyne GUI (needs OpenGL dev libs)
go build -o lazy-organizer-gui ./cmd/gui/

# Tests
go test ./...
```

## Custom Categories

Edit `categories.yaml` (path depends on OS):
- Linux: `~/.config/lazy-organizer/categories.yaml`
- macOS: `~/Library/Application Support/lazy-organizer/categories.yaml`
- Windows: `%APPDATA%\lazy-organizer\categories.yaml`

```yaml
categories:
  Documents:
    extensions: [.pdf, .doc, .docx]
    keywords: [invoice, receipt, report]
  Images:
    extensions: [.jpg, .png, .gif]
    keywords: [screenshot, photo]
  MyCategory:
    extensions: [.custom1, .custom2]
    keywords: [mykeyword]

fallback: Others
```

## License

MIT
