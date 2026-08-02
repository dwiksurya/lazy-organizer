# file-organizer

CLI tool to organize messy folders. Categories customizable via config file. Cross-platform: macOS, Linux, Windows.

## Install

```bash
go build -o file-organizer .

# cross-compile:
GOOS=linux   GOARCH=amd64 go build -o file-organizer-linux-amd64 .
GOOS=darwin  GOARCH=arm64 go build -o file-organizer-darwin-arm64 .
GOOS=windows GOARCH=amd64 go build -o file-organizer-windows-amd64.exe .
```

## Quick Start

```bash
# 1. Generate config (edit as needed)
./file-organizer -init-config
# config path per OS:
#   Linux:   ~/.config/file-organizer/categories.yaml
#   macOS:   ~/Library/Application Support/file-organizer/categories.yaml
#   Windows: %APPDATA%\file-organizer\categories.yaml

# 2. Preview (nothing is moved)
./file-organizer -dir ~/Downloads -dry-run

# 3. Organize
./file-organizer -dir ~/Downloads

# 4. Interactive mode (choose category per file)
./file-organizer -dir ~/Downloads -interactive

# 5. Undo
./file-organizer -dir ~/Downloads -undo

# 6. TUI config editor
./file-organizer -gui
```

## Flags

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

## Classifier

Files are classified by two methods (first match wins):

1. **Extension** — `.pdf` → Documents, `.jpg` → Images, etc.
2. **Keyword** — filename substring match. `invoice_march.bin` → Documents (keyword "invoice"), `screenshot_bug.xyz` → Images (keyword "screenshot")

## Custom Categories

Edit `categories.yaml` (path depends on OS):

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
