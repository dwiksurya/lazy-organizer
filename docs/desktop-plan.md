# lazy-organizer Desktop App — Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Desktop GUI app for lazy-organizer using Fyne. Reuses existing core logic (config, scanner, mover). Users can drag-drop folders, preview file organization in a table, change categories per file, and organize with one click.

**Architecture:** Extract reusable core logic into `internal/core/` package. CLI stays at root, GUI becomes `cmd/gui/main.go`. Both share the same core. Fyne v2 for native cross-platform widgets.

**Tech Stack:** Go + `fyne.io/fyne/v2` (GUI toolkit, requires CGO + OpenGL). Existing core logic from `internal/core/`.

**Platform requirements:**
- Linux: `libgl1-mesa-dev xorg-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev`
- macOS: Xcode CLI tools (`xcode-select --install`)
- Windows: MinGW-w64 (for CGO)

**Cross-compile notes:**
- Fyne uses CGO → cannot use `GOOS` cross-compile alone
- Linux: build natively or with Docker
- macOS: build on macOS (or use `fyne-cross` Docker images)
- Windows: build with MinGW or `fyne-cross`
- `fyne-cross` tool handles all platforms from Linux: `go install github.com/fyne-io/fyne-cross@latest`

---

## App Layout

```
┌─────────────────────────────────────────────────────────┐
│  lazy-organizer                              [─][□][×]  │
├─────────────────────────────────────────────────────────┤
│  [📂 Select Folder]        ~/Downloads                  │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │ Name                    │ Ext  │ Category ▼       │  │
│  ├─────────────────────────┼──────┼──────────────────┤  │
│  │ report.pdf              │ .pdf │ [Documents    ▼] │  │
│  │ photo.jpg               │ .jpg │ [Images       ▼] │  │
│  │ invoice_march.bin       │ .bin │ [Documents    ▼] │  │
│  │ screenshot_bug.xyz      │ .xyz │ [Images       ▼] │  │
│  │ my_script.py            │ .py  │ [Code         ▼] │  │
│  │ song.mp3                │ .mp3 │ [Music        ▼] │  │
│  │ ...                     │ ...  │ [...]            │  │
│  └───────────────────────────────────────────────────┘  │
│                                                         │
├─────────────────────────────────────────────────────────┤
│  [▶ Organize (20 files)]  [↩ Undo]  [⚙ Config]        │
│  Status: Ready                                          │
└─────────────────────────────────────────────────────────┘
```

## File Structure (after refactoring)

```
lazy-organizer/
├── internal/
│   └── core/
│       ├── config.go        ← moved from root
│       ├── config_test.go   ← moved from root
│       ├── scanner.go       ← moved from root
│       ├── scanner_test.go  ← moved from root
│       ├── mover.go         ← moved from root
│       └── mover_test.go    ← moved from root
├── cmd/
│   └── gui/
│       └── main.go          ← Fyne desktop app
├── main.go                  ← CLI (updated imports)
├── interactive.go           ← CLI-only (updated imports)
├── tui.go                   ← CLI-only (updated imports)
├── main_test.go             ← CLI integration tests
├── go.mod
├── go.sum
└── README.md
```

---

## Task 1: Install Fyne dependencies

**Objective:** Install system libraries and Go module for Fyne.

**Step 1: Install system deps (Linux)**

```bash
sudo apt-get update && sudo apt-get install -y \
  libgl1-mesa-dev xorg-dev \
  libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev
```

**Step 2: Verify CGO works**

```bash
export DISPLAY=:99
go env CGO_ENABLED
# should print: 1
```

**Step 3: Verify**

```bash
# tiny Fyne test
cd /tmp && mkdir fyne-test && cd fyne-test
go mod init fyne-test
cat > main.go << 'EOF'
package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Hello")
	w.SetContent(widget.NewLabel("Fyne works!"))
	w.ShowAndRun()
}
EOF
go mod tidy
timeout 5 go run main.go 2>&1 || echo "(window opened and closed — OK)"
cd / && rm -rf /tmp/fyne-test
```

**Step 4: Commit**

```bash
# just verification, no files to commit
```

---

## Task 2: Extract core logic to internal/core/

**Objective:** Move `config.go`, `scanner.go`, `mover.go` into `internal/core/` package so both CLI and GUI can import them.

**Files:**
- Move: `config.go` → `internal/core/config.go`
- Move: `scanner.go` → `internal/core/scanner.go`
- Move: `mover.go` → `internal/core/mover.go`
- Move: `main_test.go` → `internal/core/core_test.go`

**Step 1: Create directory**

```bash
mkdir -p internal/core
```

**Step 2: Copy and update package name**

For each file, copy to `internal/core/` and change `package main` → `package core`.

Also export any unexported functions/types that the GUI needs:
- `historyPath` → `HistoryPath` (needed if GUI wants to check)
- `scanner` var in interactive.go → stays CLI-only

**Step 3: Update `config.go`**

```go
package core

// ... (same content, just package name change)
```

Key changes:
- `package main` → `package core`
- Everything else stays the same (all types/functions already exported)

**Step 4: Update `scanner.go`**

```go
package core
// same content, package name change only
```

**Step 5: Update `mover.go`**

```go
package core
// same content, package name change only
// Also export: historyPath → HistoryPath
```

**Step 6: Move tests**

```bash
cp main_test.go internal/core/core_test.go
# Update: package main → package core
# Update: imports — no external imports needed
```

**Step 7: Run tests**

```bash
go test ./internal/core/ -v
```

Expected: All 7 tests pass.

**Step 8: Commit**

```bash
git add internal/core/
git commit -m "refactor: extract core logic to internal/core package"
```

---

## Task 3: Update CLI to use internal/core

**Objective:** Update `main.go`, `interactive.go`, `tui.go` to import from `internal/core`.

**Files:**
- Modify: `main.go`
- Modify: `interactive.go`
- Modify: `tui.go`
- Modify: `main_test.go`

**Step 1: Update `main.go`**

```go
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"lazy-organizer/internal/core"
)

func main() {
	// ... same flags ...

	cfg, err := core.LoadConfig(cfgPath)
	if err != nil {
		cfg = core.DefaultConfig()
		// ...
	}

	files, err := core.Scan(absDir, cfg)
	// ...
}
```

Key changes:
- Add `"lazy-organizer/internal/core"` import
- Prefix all core calls: `core.LoadConfig`, `core.DefaultConfig`, `core.Scan`, `core.Move`, `core.UndoAll`, `core.GenerateConfig`, `core.DefaultConfigPath`
- `FileInfo` → `core.FileInfo`
- `Config` → `*core.Config`

**Step 2: Update `interactive.go`**

```go
package main

import (
	"lazy-organizer/internal/core"
)

func InteractiveClassify(files []core.FileInfo, cfg *core.Config) []core.FileInfo {
	// ... same logic ...
}
```

**Step 3: Update `tui.go`**

```go
package main

import (
	"lazy-organizer/internal/core"
)

func RunTUI(cfgPath string) error {
	cfg, err := core.LoadConfig(cfgPath)
	// ...
}
```

Also update all type references: `core.CategoryRule`, `core.Config`, etc.

**Step 4: Update `main_test.go`**

```go
package main

import (
	"lazy-organizer/internal/core"
)

// Tests that were in main_test.go that test core logic → move to core_test.go
// Keep only integration tests here that test the CLI wiring
```

**Step 5: Build and test**

```bash
go build -o lazy-organizer .
go test ./... -v
```

**Step 6: Commit**

```bash
git add -A
git commit -m "refactor: CLI uses internal/core package"
```

---

## Task 4: GUI skeleton — window + folder picker

**Objective:** Create `cmd/gui/main.go` with a Fyne window, folder picker button, and file list.

**Files:**
- Create: `cmd/gui/main.go`

**Step 1: Implement**

```go
package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"lazy-organizer/internal/core"
)

type App struct {
	window    fyne.Window
	cfg       *core.Config
	dir       string
	files     []core.FileInfo
	table     *widget.Table
	statusLbl *widget.Label
	organizeBtn *widget.Button
	undoBtn     *widget.Button
}

func main() {
	a := app.New()
	w := a.NewWindow("lazy-organizer")
	w.Resize(fyne.NewSize(800, 600))

	cfgPath := core.DefaultConfigPath()
	cfg, err := core.LoadConfig(cfgPath)
	if err != nil {
		cfg = core.DefaultConfig()
	}

	app := &App{
		window: w,
		cfg:    cfg,
	}

	// Toolbar
	folderBtn := widget.NewButton("📂 Select Folder", app.selectFolder)
	app.organizeBtn = widget.NewButton("▶ Organize", app.organize)
	app.organizeBtn.Disable()
	app.undoBtn = widget.NewButton("↩ Undo", app.undo)
	app.undoBtn.Disable()

	toolbar := container.NewHBox(folderBtn, app.organizeBtn, app.undoBtn)

	// Table
	app.table = app.makeTable()

	// Status
	app.statusLbl = widget.NewLabel("Select a folder to begin")

	// Layout
	content := container.NewBorder(toolbar, app.statusLbl, nil, nil, app.table)
	w.SetContent(content)
	w.ShowAndRun()
}

func (a *App) selectFolder() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		a.dir = uri.Path()
		a.scanDir()
	}, a.window)
}

func (a *App) scanDir() {
	files, err := core.Scan(a.dir, a.cfg)
	if err != nil {
		a.statusLbl.SetText("Error: " + err.Error())
		return
	}
	a.files = files
	a.table.Refresh()
	a.organizeBtn.Enable()
	a.undoBtn.Enable()
	a.statusLbl.SetText(fmt.Sprintf("%d files found in %s", len(files), a.dir))
}

func (a *App) organize() {
	moved := 0
	for _, f := range a.files {
		if err := core.Move(a.dir, f); err != nil {
			continue
		}
		moved++
	}
	a.statusLbl.SetText(fmt.Sprintf("✓ Moved %d/%d files", moved, len(files)))
	a.scanDir() // refresh
}

func (a *App) undo() {
	if err := core.UndoAll(a.dir); err != nil {
		a.statusLbl.SetText("Undo error: " + err.Error())
		return
	}
	a.statusLbl.SetText("✓ Undo complete")
	a.scanDir()
}

func (a *App) makeTable() *widget.Table {
	// Table with columns: Name, Extension, Category
	// Category column uses a Select widget per row
	// ...
}
```

**Step 2: Build**

```bash
mkdir -p cmd/gui
go build -o lazy-organizer-gui ./cmd/gui/
```

**Step 3: Test**

```bash
DISPLAY=:99 timeout 5 ./lazy-organizer-gui 2>&1 || echo "(window opened — OK)"
```

**Step 4: Commit**

```bash
git add cmd/gui/
git commit -m "feat: add Fyne desktop GUI skeleton"
```

---

## Task 5: File table with category dropdown

**Objective:** Implement the file table with per-row category selector.

**Files:**
- Modify: `cmd/gui/main.go`

**Step 1: Implement table**

```go
func (a *App) makeTable() *widget.Table {
	categories := a.cfg.ListCategories()

	table := widget.NewTable(
		// data func: return cell value
		func() (int, int) {
			return len(a.files), 3 // rows, cols
		},
		func() fyne.CanvasObject {
			// template for each cell type
			return widget.NewLabel("template")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			if id.Row >= len(a.files) {
				return
			}
			f := a.files[id.Row]
			label := obj.(*widget.Label)
			switch id.Col {
			case 0:
				label.SetText(f.Name)
			case 1:
				label.SetText(f.Ext)
			case 2:
				label.SetText(f.Category)
			}
		},
	)

	table.SetColumnWidth(0, 300) // Name
	table.SetColumnWidth(1, 80)  // Extension
	table.SetColumnWidth(2, 200) // Category

	return table
}
```

**Step 2: Add category change via dialog**

When user clicks a cell in the Category column, show a dialog to select a new category:

```go
table.OnSelected = func(id widget.TableCellID) {
	if id.Col == 2 && id.Row < len(a.files) {
		categories := a.cfg.ListCategories()
		dialog.ShowSelect("Category", categories, func(selected string) {
			if selected != "" {
				a.files[id.Row].Category = selected
				table.Refresh()
			}
		}, a.window)
	}
	table.UnselectAll()
}
```

**Step 3: Build and test**

```bash
go build -o lazy-organizer-gui ./cmd/gui/
DISPLAY=:99 timeout 5 ./lazy-organizer-gui
```

**Step 4: Commit**

```bash
git add cmd/gui/
git commit -m "feat: file table with per-row category selector"
```

---

## Task 6: Status bar + organize/undo flow

**Objective:** Wire up organize and undo buttons with proper status feedback.

**Files:**
- Modify: `cmd/gui/main.go`

**Step 1: Implement organize with confirmation**

```go
func (a *App) organize() {
	if len(a.files) == 0 {
		a.statusLbl.SetText("No files to organize")
		return
	}

	dialog.ShowConfirm("Organize",
		fmt.Sprintf("Move %d files into category folders?", len(a.files)),
		func(ok bool) {
			if !ok {
				return
			}
			moved := 0
			for _, f := range a.files {
				if err := core.Move(a.dir, f); err != nil {
					continue
				}
				moved++
			}
			a.statusLbl.SetText(fmt.Sprintf("✓ Moved %d/%d files (use Undo to revert)", moved, len(a.files)))
			a.scanDir()
		},
		a.window,
	)
}
```

**Step 2: Implement undo**

```go
func (a *App) undo() {
	if err := core.UndoAll(a.dir); err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.statusLbl.SetText("✓ Undo complete")
	a.scanDir()
}
```

**Step 3: Build and test**

**Step 4: Commit**

---

## Task 7: Config editor dialog

**Objective:** Add a Config button that opens a dialog to edit categories.

**Files:**
- Modify: `cmd/gui/main.go`

**Step 1: Implement config editor**

```go
func (a *App) showConfigEditor() {
	// Build form with current categories
	// Each category: name entry, extensions entry, keywords entry
	// Buttons: Save, Cancel, Reset to Defaults

	categories := a.cfg.Categories
	orderedCats := a.cfg.orderedCats

	// Use a List widget with edit buttons
	// Or a Form with entries per category

	// Simplest: show category list, click to edit
	list := widget.NewList(
		func() int { return len(orderedCats) },
		func() fyne.CanvasObject { return widget.NewLabel("category") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(orderedCats[id])
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		name := orderedCats[id]
		rule := categories[name]
		a.showCategoryEditor(name, rule)
		list.UnselectAll()
	}

	dialog.ShowCustom("Config Editor", "Close", list, a.window)
}
```

**Step 2: Implement category editor sub-dialog**

```go
func (a *App) showCategoryEditor(name string, rule core.CategoryRule) {
	extEntry := widget.NewEntry()
	extEntry.SetText(strings.Join(rule.Extensions, ", "))
	kwEntry := widget.NewEntry()
	kwEntry.SetText(strings.Join(rule.Keywords, ", "))

	form := widget.NewForm(
		widget.NewFormItem("Extensions", extEntry),
		widget.NewFormItem("Keywords", kwEntry),
	)

	dialog.ShowForm("Edit: "+name, "Save", "Cancel", form.Items, func(ok bool) {
		if !ok {
			return
		}
		rule.Extensions = parseCSV(extEntry.Text)
		rule.Keywords = parseCSV(kwEntry.Text)
		a.cfg.Categories[name] = rule
		a.cfg.buildOrder() // need to export this
		a.saveConfig()
	}, a.window)
}
```

**Step 3: Export `buildOrder` in core**

```go
// In internal/core/config.go
func (c *Config) BuildOrder() {
	c.buildOrder()
}
```

Or make `buildOrder` public → `BuildOrder`.

**Step 4: Build and test**

**Step 5: Commit**

---

## Task 8: Polish + build

**Objective:** Final touches, cross-compile, update README.

**Files:**
- Modify: `README.md`
- Modify: `go.mod` (if needed)

**Step 1: Add Makefile**

```makefile
.PHONY: cli gui test clean

cli:
	go build -o lazy-organizer .

gui:
	go build -o lazy-organizer-gui ./cmd/gui/

test:
	go test ./... -v

clean:
	rm -f lazy-organizer lazy-organizer-gui

# Cross-compile CLI (pure Go, no CGO)
cross-cli:
	GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o dist/lazy-organizer-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags="-s -w" -o dist/lazy-organizer-linux-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o dist/lazy-organizer-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o dist/lazy-organizer-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/lazy-organizer-windows-amd64.exe .
```

**Step 2: Update README**

```markdown
# lazy-organizer

CLI + Desktop app to organize messy folders.

## CLI

```bash
go build -o lazy-organizer .
./lazy-organizer -dir ~/Downloads -dry-run
```

## Desktop App

```bash
go build -o lazy-organizer-gui ./cmd/gui/
./lazy-organizer-gui
```

Requires OpenGL libraries (Linux: `libgl1-mesa-dev xorg-dev`).
```

**Step 3: Build both**

```bash
make cli gui
```

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: desktop app + Makefile + updated README"
```

---

## Summary

| Task | Description | Est. |
|------|-------------|------|
| 1 | Install Fyne dependencies | 3 min |
| 2 | Extract core to internal/core/ | 5 min |
| 3 | Update CLI imports | 5 min |
| 4 | GUI skeleton — window + folder picker | 5 min |
| 5 | File table with category dropdown | 8 min |
| 6 | Status bar + organize/undo flow | 5 min |
| 7 | Config editor dialog | 8 min |
| 8 | Polish + Makefile + README | 3 min |

**Total: ~42 min, 1 new dep (fyne.io/fyne/v2), reuses all existing core logic.**

## Risks

- **CGO cross-compile**: Fyne requires CGO. Cross-compiling for macOS/Windows from Linux needs `fyne-cross` Docker or native builds. CLI stays pure Go (no CGO).
- **OpenGL on server**: Fyne needs a display. Xvfb (:99) works for testing but the GUI binary is meant for desktop use.
- **Fyne version**: Use `fyne.io/fyne/v2` (stable). Check latest version at https://github.com/fyne-io/fyne/releases.
