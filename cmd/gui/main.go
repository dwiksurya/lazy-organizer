package main

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"lazy-organizer/internal/core"
)

type App struct {
	window      fyne.Window
	cfg         *core.Config
	cfgPath     string
	dir         string
	files       []core.FileInfo
	table       *widget.Table
	statusLbl   *widget.Label
	organizeBtn *widget.Button
	undoBtn     *widget.Button
}

func main() {
	a := app.NewWithID("com.lazy-organizer")
	w := a.NewWindow("lazy-organizer")
	w.Resize(fyne.NewSize(900, 600))

	cfgPath := core.DefaultConfigPath()
	cfg, err := core.LoadConfig(cfgPath)
	if err != nil {
		cfg = core.DefaultConfig()
	}

	myApp := &App{
		window:  w,
		cfg:     cfg,
		cfgPath: cfgPath,
	}

	// Toolbar
	folderBtn := widget.NewButton("📂 Select Folder", myApp.selectFolder)
	folderBtn.Importance = widget.HighImportance

	myApp.organizeBtn = widget.NewButton("▶ Organize", myApp.organize)
	myApp.organizeBtn.Disable()
	myApp.undoBtn = widget.NewButton("↩ Undo", myApp.undo)
	myApp.undoBtn.Disable()
	configBtn := widget.NewButton("⚙ Config", myApp.showConfigEditor)

	toolbar := container.NewHBox(folderBtn, widget.NewSeparator(), myApp.organizeBtn, myApp.undoBtn, widget.NewSeparator(), configBtn)

	// Table
	myApp.table = myApp.makeTable()

	// Status
	myApp.statusLbl = widget.NewLabel("Select a folder to begin")

	// Layout
	content := container.NewBorder(
		container.NewVBox(toolbar, widget.NewSeparator()),
		container.NewHBox(myApp.statusLbl),
		nil, nil,
		myApp.table,
	)
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

func (a *App) undo() {
	if err := core.UndoAll(a.dir); err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.statusLbl.SetText("✓ Undo complete")
	a.scanDir()
}

func (a *App) makeTable() *widget.Table {
	table := widget.NewTable(
		func() (int, int) {
			return len(a.files) + 1, 3 // +1 for header
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("template")
			label.Wrapping = fyne.TextTruncate
			return label
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row == 0 {
				// Header row
				switch id.Col {
				case 0:
					label.SetText("Name")
					label.TextStyle = fyne.TextStyle{Bold: true}
				case 1:
					label.SetText("Ext")
					label.TextStyle = fyne.TextStyle{Bold: true}
				case 2:
					label.SetText("Category")
					label.TextStyle = fyne.TextStyle{Bold: true}
				}
				return
			}
			// Data rows
			label.TextStyle = fyne.TextStyle{}
			if id.Row-1 >= len(a.files) {
				label.SetText("")
				return
			}
			f := a.files[id.Row-1]
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

	table.SetColumnWidth(0, 350)
	table.SetColumnWidth(1, 80)
	table.SetColumnWidth(2, 200)

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			table.UnselectAll()
			return
		}
		if id.Col == 2 && id.Row-1 < len(a.files) {
			a.showCategoryPicker(id.Row - 1)
		}
		table.UnselectAll()
	}

	return table
}

func (a *App) showCategoryPicker(fileIdx int) {
	categories := a.cfg.ListCategories()
	sort.Strings(categories)

	f := a.files[fileIdx]
	selected := f.Category
	sel := widget.NewSelect(categories, func(val string) {
		selected = val
	})
	sel.SetSelected(f.Category)

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Category for: %s", f.Name)),
		sel,
	)

	dialog.ShowCustomConfirm("Change Category", "OK", "Cancel", content, func(ok bool) {
		if ok && selected != "" {
			a.files[fileIdx].Category = selected
			a.table.Refresh()
		}
	}, a.window)
}

func (a *App) showConfigEditor() {
	categories := a.cfg.Categories
	orderedCats := a.cfg.ListCategories()

	list := widget.NewList(
		func() int { return len(orderedCats) },
		func() fyne.CanvasObject {
			return widget.NewLabel("category name")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(orderedCats[id])
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		name := orderedCats[id]
		if name == a.cfg.Fallback {
			dialog.ShowInformation("Fallback", fmt.Sprintf("Fallback category: %s\n\nEdit in the 'Edit Fallback' option.", a.cfg.Fallback), a.window)
			list.UnselectAll()
			return
		}
		rule := categories[name]
		a.showCategoryEditor(name, rule)
		list.UnselectAll()
	}

	// Add category button
	addBtn := widget.NewButton("+ Add Category", func() {
		a.showAddCategoryDialog()
	})

	// Edit fallback button
	fallbackBtn := widget.NewButton(fmt.Sprintf("Fallback: %s", a.cfg.Fallback), func() {
		a.showEditFallbackDialog()
	})

	// Reset button
	resetBtn := widget.NewButton("Reset to Defaults", func() {
		dialog.ShowConfirm("Reset", "Reset all categories to defaults?", func(ok bool) {
			if ok {
				a.cfg = core.DefaultConfig()
				a.saveConfig()
				dialog.ShowInformation("Done", "Config reset to defaults", a.window)
			}
		}, a.window)
	})

	content := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Click a category to edit:"),
			container.NewHBox(addBtn, fallbackBtn, resetBtn),
		),
		nil, nil, nil,
		list,
	)

	dialog.ShowCustom("Config Editor", "Close", content, a.window)
}

func (a *App) showCategoryEditor(name string, rule core.CategoryRule) {
	extEntry := widget.NewMultiLineEntry()
	extEntry.SetText(strings.Join(rule.Extensions, "\n"))
	extEntry.SetMinRowsVisible(3)

	kwEntry := widget.NewMultiLineEntry()
	kwEntry.SetText(strings.Join(rule.Keywords, "\n"))
	kwEntry.SetMinRowsVisible(3)

	// Delete button
	deleteBtn := widget.NewButton("Delete Category", func() {
		dialog.ShowConfirm("Delete", fmt.Sprintf("Delete category [%s]?", name), func(ok bool) {
			if ok {
				delete(a.cfg.Categories, name)
				a.cfg.BuildOrder()
				a.saveConfig()
			}
		}, a.window)
	})
	deleteBtn.Importance = widget.DangerImportance

	form := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Edit: %s", name)),
		widget.NewSeparator(),
		widget.NewLabel("Extensions (one per line):"),
		extEntry,
		widget.NewLabel("Keywords (one per line):"),
		kwEntry,
		widget.NewSeparator(),
		container.NewHBox(layout.NewSpacer(), deleteBtn),
	)

	dialog.ShowCustomConfirm(fmt.Sprintf("Edit: %s", name), "Save", "Cancel", form, func(ok bool) {
		if !ok {
			return
		}
		rule.Extensions = parseLines(extEntry.Text)
		rule.Keywords = parseLines(kwEntry.Text)
		a.cfg.Categories[name] = rule
		a.cfg.BuildOrder()
		a.saveConfig()
	}, a.window)
}

func (a *App) showAddCategoryDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Category name")

	extEntry := widget.NewMultiLineEntry()
	extEntry.SetPlaceHolder(".ext1\n.ext2")
	extEntry.SetMinRowsVisible(3)

	kwEntry := widget.NewMultiLineEntry()
	kwEntry.SetPlaceHolder("keyword1\nkeyword2")
	kwEntry.SetMinRowsVisible(3)

	form := container.NewVBox(
		widget.NewLabel("Add New Category"),
		widget.NewSeparator(),
		widget.NewLabel("Name:"),
		nameEntry,
		widget.NewLabel("Extensions (one per line):"),
		extEntry,
		widget.NewLabel("Keywords (one per line):"),
		kwEntry,
	)

	dialog.ShowCustomConfirm("Add Category", "Add", "Cancel", form, func(ok bool) {
		if !ok || nameEntry.Text == "" {
			return
		}
		if _, exists := a.cfg.Categories[nameEntry.Text]; exists {
			dialog.ShowError(fmt.Errorf("category '%s' already exists", nameEntry.Text), a.window)
			return
		}
		a.cfg.Categories[nameEntry.Text] = core.CategoryRule{
			Extensions: parseLines(extEntry.Text),
			Keywords:   parseLines(kwEntry.Text),
		}
		a.cfg.BuildOrder()
		a.saveConfig()
	}, a.window)
}

func (a *App) showEditFallbackDialog() {
	entry := widget.NewEntry()
	entry.SetText(a.cfg.Fallback)

	form := container.NewVBox(
		widget.NewLabel("Edit Fallback Category"),
		widget.NewSeparator(),
		widget.NewLabel("Name (for unmatched files):"),
		entry,
	)

	dialog.ShowCustomConfirm("Edit Fallback", "Save", "Cancel", form, func(ok bool) {
		if !ok || entry.Text == "" {
			return
		}
		a.cfg.Fallback = entry.Text
		a.saveConfig()
	}, a.window)
}

func (a *App) saveConfig() {
	if err := core.GenerateConfig(a.cfgPath); err != nil {
		// If generate fails, try manual save
		a.statusLbl.SetText("Warning: could not save config: " + err.Error())
		return
	}
	a.statusLbl.SetText("✓ Config saved")
}

func parseLines(s string) []string {
	parts := strings.Split(s, "\n")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	sort.Strings(result)
	return result
}
