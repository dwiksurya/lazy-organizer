# lazy-organizer v1.0.0

CLI + Desktop app to organize messy folders. Cross-platform: Linux, macOS, Windows.

## What's New

### CLI (`lazy-organizer`)
- **Smart classifier** — extension + keyword matching (configurable via `categories.yaml`)
- **Interactive mode** — choose category per file before organizing
- **TUI config editor** — edit categories directly in terminal
- **Undo support** — full history log, reverse all moves
- **Cross-platform** — Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64)

### Desktop App (`lazy-organizer-gui`)
- **Native file table** with checkboxes per file
- **Filter bar** — search by name, extension, or category
- **Select All / None / Invert** — bulk selection controls
- **Category picker** — click category cell to change
- **Config editor** — add/edit/delete categories in GUI
- **Organize with confirmation** — only moves checked files

## Downloads

| File | Platform | Size |
|------|----------|------|
| `lazy-organizer-linux-amd64.tar.gz` | Linux x86_64 (CLI) | 1.1MB |
| `lazy-organizer-linux-arm64.tar.gz` | Linux ARM64 (CLI) | 961KB |
| `lazy-organizer-darwin-amd64.tar.gz` | macOS Intel (CLI) | 1.1MB |
| `lazy-organizer-darwin-arm64.tar.gz` | macOS Apple Silicon (CLI) | 985KB |
| `lazy-organizer-windows-amd64.tar.gz` | Windows x86_64 (CLI) | 1.1MB |
| `lazy-organizer-gui-linux-amd64.tar.gz` | Linux x86_64 (GUI) | 11MB |

## Quick Start

### CLI
```bash
# Extract
tar xzf lazy-organizer-linux-amd64.tar.gz
cd pkg-lazy-organizer-linux-amd64

# Generate config
./lazy-organizer -init-config

# Preview
./lazy-organizer -dir ~/Downloads -dry-run

# Organize
./lazy-organizer -dir ~/Downloads

# Undo
./lazy-organizer -dir ~/Downloads -undo
```

### Desktop App (Linux)
```bash
# Requires: libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev
tar xzf lazy-organizer-gui-linux-amd64.tar.gz
cd pkg-lazy-organizer-gui-linux-amd64
./lazy-organizer-gui
```

## Build from Source

```bash
git clone https://github.com/dwiksurya/lazy-organizer.git
cd lazy-organizer

# CLI
go build -o lazy-organizer .

# GUI (requires OpenGL dev libs)
go build -o lazy-organizer-gui ./cmd/gui/

# Tests
go test ./...
```

## Classifier

Files are classified by:
1. **Extension** — `.pdf` → Documents, `.jpg` → Images, etc.
2. **Keyword** — filename substring match. `invoice_march.bin` → Documents (keyword "invoice")

Edit `~/.config/lazy-organizer/categories.yaml` to customize.
