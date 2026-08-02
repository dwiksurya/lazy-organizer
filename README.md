# lazy-organizer

CLI + Desktop app to organize messy folders. Cross-platform: macOS, Linux, Windows.

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

## Quick Start — Desktop App

```bash
# Install system deps (Linux only)
sudo apt-get install -y libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev

# Build
go build -o lazy-organizer-gui ./cmd/gui/

# Run
./lazy-organizer-gui
```

**macOS:** No extra deps (Xcode CLI tools sufficient)
**Windows:** Requires MinGW-w64 for CGO

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

- **📂 Select Folder** — native folder picker
- **File Table** — shows all files with Name, Extension, Category
- **Category Dropdown** — click category cell to change
- **▶ Organize** — move files with confirmation dialog
- **↩ Undo** — revert last organization
- **⚙ Config Editor** — edit/add/delete categories in GUI

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
│   └── core_test.go        ← 7 tests
├── cmd/gui/                ← desktop app (Fyne)
│   └── main.go             ← GUI with file table + config editor
├── main.go                 ← CLI entry point
├── interactive.go          ← CLI interactive mode
├── tui.go                  ← CLI TUI config editor
├── Makefile                ← build targets
└── README.md
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

## Cross-Compile

```bash
# CLI (pure Go — no CGO)
make cross-cli

# GUI (requires CGO — use fyne-cross)
go install github.com/fyne-io/fyne-cross@latest
fyne-cross linux -arch amd64,arm64
fyne-cross darwin -arch amd64,arm64
fyne-cross windows -arch amd64
```

## License

MIT
