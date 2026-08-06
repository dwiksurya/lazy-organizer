package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"time"

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
	files       []core.FileInfo       // all files from scan
	filtered    []int                 // indices into files after filter
	selected    map[int]bool          // index → checked
	table       *widget.Table
	statusLbl   *widget.Label
	filterEntry *widget.Entry
	organizeBtn *widget.Button
	undoBtn     *widget.Button
}

func logLine(msg string) {
	logPath := filepath.Join(os.TempDir(), "lazy-organizer.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	fmt.Fprintf(f, "%s %s\n", time.Now().Format(time.RFC3339), msg)
	f.Close()
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			logLine(fmt.Sprintf("PANIC: %v\n%s", r, debug.Stack()))
		}
	}()
	logLine("starting")
	a := app.NewWithID("com.lazy-organizer")
	w := a.NewWindow("lazy-organizer")
	w.Resize(fyne.NewSize(900, 600))

	cfgPath := core.DefaultConfigPath()
	cfg, err := core.LoadConfig(cfgPath)
	if err != nil {
		cfg = core.DefaultConfig()
	}

	myApp := &App{
		window:   w,
		cfg:      cfg,
		cfgPath:  cfgPath,
		selected: make(map[int]bool),
		filtered: []int{},
	}

	// Toolbar
	folderBtn := widget.NewButton("Select Folder", myApp.selectFolder)
	folderBtn.Importance = widget.HighImportance

	myApp.organizeBtn = widget.NewButton("Organize (0)", myApp.organize)
	myApp.organizeBtn.Disable()
	myApp.undoBtn = widget.NewButton("Undo", myApp.undo)
	myApp.undoBtn.Disable()
	configBtn := widget.NewButton("Config", myApp.showConfigEditor)

	toolbar := container.NewHBox(folderBtn, widget.NewSeparator(), myApp.organizeBtn, myApp.undoBtn, widget.NewSeparator(), configBtn)

	// Filter bar
	myApp.filterEntry = widget.NewEntry()
	myApp.filterEntry.SetPlaceHolder("Filter by name, ext, or category...")
	myApp.filterEntry.OnChanged = func(text string) {
		myApp.applyFilter(text)
	}

	selectAllBtn := widget.NewButton("All", func() {
		for _, idx := range myApp.filtered {
			myApp.selected[idx] = true
		}
		myApp.table.Refresh()
		myApp.updateOrganizeCount()
	})
	noneBtn := widget.NewButton("None", func() {
		myApp.selected = make(map[int]bool)
		myApp.table.Refresh()
		myApp.updateOrganizeCount()
	})
	invertBtn := widget.NewButton("Invert", func() {
		for _, idx := range myApp.filtered {
			myApp.selected[idx] = !myApp.selected[idx]
		}
		myApp.table.Refresh()
		myApp.updateOrganizeCount()
	})

	filterBar := container.NewBorder(nil, nil,
		container.NewHBox(selectAllBtn, noneBtn, invertBtn),
		nil,
		myApp.filterEntry,
	)

	// Table
	myApp.table = myApp.makeTable()

	// Status
	myApp.statusLbl = widget.NewLabel("Select a folder to begin")

	// Layout
	content := container.NewBorder(
		container.NewVBox(toolbar, widget.NewSeparator(), filterBar, widget.NewSeparator()),
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
	a.selected = make(map[int]bool)
	for i := range a.files {
		a.selected[i] = true
	}
	a.filterEntry.SetText("")
	a.applyFilter("")
	a.organizeBtn.Enable()
	a.undoBtn.Enable()
	a.updateStatus()
}

func (a *App) applyFilter(text string) {
	text = strings.ToLower(strings.TrimSpace(text))
	a.filtered = []int{}
	for i, f := range a.files {
		if text == "" {
			a.filtered = append(a.filtered, i)
			continue
		}
		if strings.Contains(strings.ToLower(f.Name), text) ||
			strings.Contains(strings.ToLower(f.Ext), text) ||
			strings.Contains(strings.ToLower(f.Category), text) {
			a.filtered = append(a.filtered, i)
		}
	}
	a.table.Refresh()
	a.updateOrganizeCount()
	a.updateStatus()
}

func (a *App) updateOrganizeCount() {
	count := 0
	for _, idx := range a.filtered {
		if a.selected[idx] {
			count++
		}
	}
	a.organizeBtn.SetText(fmt.Sprintf("Organize (%d)", count))
}

func (a *App) updateStatus() {
	total := len(a.files)
	shown := len(a.filtered)
	checked := 0
	for _, idx := range a.filtered {
		if a.selected[idx] {
			checked++
		}
	}
	if total == shown {
		a.statusLbl.SetText(fmt.Sprintf("%d files, %d selected", total, checked))
	} else {
		a.statusLbl.SetText(fmt.Sprintf("%d files shown (%d total), %d selected", shown, total, checked))
	}
}

func (a *App) organize() {
	var toMove []core.FileInfo
	for _, idx := range a.filtered {
		if a.selected[idx] {
			toMove = append(toMove, a.files[idx])
		}
	}
	if len(toMove) == 0 {
		a.statusLbl.SetText("No files selected")
		return
	}
	dialog.ShowConfirm("Organize",
		fmt.Sprintf("Move %d selected files into category folders?", len(toMove)),
		func(ok bool) {
			if !ok {
				return
			}
			moved := 0
			for _, f := range toMove {
				if err := core.Move(a.dir, f); err != nil {
					continue
				}
				moved++
			}
			a.statusLbl.SetText(fmt.Sprintf("Moved %d/%d files (use Undo to revert)", moved, len(toMove)))
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
	a.statusLbl.SetText("Undo complete")
	a.scanDir()
}

func (a *App) makeTable() *widget.Table {
	table := widget.NewTable(
		func() (int, int) {
			return len(a.filtered) + 1, 4
		},
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			label := widget.NewLabel("template")
			label.Wrapping = fyne.TextTruncate
			return container.NewStack(check, label)
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			cell := obj.(*fyne.Container)
			check := cell.Objects[0].(*widget.Check)
			label := cell.Objects[1].(*widget.Label)

			if id.Row == 0 {
				check.Hide()
				label.Show()
				label.TextStyle = fyne.TextStyle{Bold: true}
				switch id.Col {
				case 0:
					label.SetText("")
				case 1:
					label.SetText("Name")
				case 2:
					label.SetText("Ext")
				case 3:
					label.SetText("Category")
				}
				return
			}

			if id.Row-1 >= len(a.filtered) {
				check.Hide()
				label.Show()
				label.SetText("")
				return
			}
			fileIdx := a.filtered[id.Row-1]
			f := a.files[fileIdx]

			switch id.Col {
			case 0:
				label.Hide()
				check.Show()
				check.OnChanged = func(val bool) {
					a.selected[fileIdx] = val
					a.updateOrganizeCount()
					a.updateStatus()
				}
				check.SetChecked(a.selected[fileIdx])
			case 1:
				check.Hide()
				label.Show()
				label.TextStyle = fyne.TextStyle{}
				label.SetText(f.Name)
			case 2:
				check.Hide()
				label.Show()
				label.TextStyle = fyne.TextStyle{}
				label.SetText(f.Ext)
			case 3:
				check.Hide()
				label.Show()
				label.TextStyle = fyne.TextStyle{}
				label.SetText(f.Category)
			}
		},
	)

	table.SetColumnWidth(0, 40)
	table.SetColumnWidth(1, 320)
	table.SetColumnWidth(2, 80)
	table.SetColumnWidth(3, 200)

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			table.UnselectAll()
			return
		}
		if id.Col == 3 && id.Row-1 < len(a.filtered) {
			fileIdx := a.filtered[id.Row-1]
			a.showCategoryPicker(fileIdx)
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
			label := widget.NewLabel("category name")
			label.Wrapping = fyne.TextTruncate
			return label
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			name := orderedCats[id]
			rule := categories[name]
			detail := fmt.Sprintf("%s  (%d ext, %d kw)", name, len(rule.Extensions), len(rule.Keywords))
			if name == a.cfg.Fallback {
				detail = fmt.Sprintf("%s  [fallback]", name)
			}
			obj.(*widget.Label).SetText(detail)
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		name := orderedCats[id]
		if name == a.cfg.Fallback {
			a.showEditFallbackDialog()
			list.UnselectAll()
			return
		}
		rule := categories[name]
		a.showCategoryEditor(name, rule, func() {
			orderedCats = a.cfg.ListCategories()
			categories = a.cfg.Categories
			list.Refresh()
		})
		list.UnselectAll()
	}

	addBtn := widget.NewButton("Add Category", func() {
		a.showAddCategoryDialog(func() {
			orderedCats = a.cfg.ListCategories()
			categories = a.cfg.Categories
			list.Refresh()
		})
	})

	fallbackBtn := widget.NewButton(fmt.Sprintf("Fallback: %s", a.cfg.Fallback), func() {
		a.showEditFallbackDialog()
	})

	resetBtn := widget.NewButton("Reset to Defaults", func() {
		dialog.ShowConfirm("Reset", "Reset all categories to defaults?", func(ok bool) {
			if ok {
				a.cfg = core.DefaultConfig()
				a.saveConfig()
				orderedCats = a.cfg.ListCategories()
				categories = a.cfg.Categories
				list.Refresh()
				fallbackBtn.SetText(fmt.Sprintf("Fallback: %s", a.cfg.Fallback))
			}
		}, a.window)
	})

	header := container.NewVBox(
		widget.NewLabel("Click a category to edit it"),
		container.NewHBox(addBtn, fallbackBtn, resetBtn),
		widget.NewSeparator(),
	)

	content := container.NewBorder(header, nil, nil, nil, list)
	content.Resize(fyne.NewSize(600, 500))

	d := dialog.NewCustom("Config Editor", "Close", content, a.window)
	d.Resize(fyne.NewSize(600, 500))
	d.Show()
}

func (a *App) showCategoryEditor(name string, rule core.CategoryRule, onSaved func()) {
	extEntry := widget.NewMultiLineEntry()
	extEntry.SetText(strings.Join(rule.Extensions, "\n"))
	extEntry.SetMinRowsVisible(6)

	kwEntry := widget.NewMultiLineEntry()
	kwEntry.SetText(strings.Join(rule.Keywords, "\n"))
	kwEntry.SetMinRowsVisible(6)

	deleteBtn := widget.NewButton("Delete Category", func() {
		dialog.ShowConfirm("Delete", fmt.Sprintf("Delete category [%s]?", name), func(ok bool) {
			if ok {
				delete(a.cfg.Categories, name)
				a.cfg.BuildOrder()
				a.saveConfig()
				if onSaved != nil {
					onSaved()
				}
			}
		}, a.window)
	})
	deleteBtn.Importance = widget.DangerImportance

	form := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Edit: %s", name)),
		widget.NewSeparator(),
		widget.NewLabel("Extensions (one per line):"),
		extEntry,
		widget.NewSeparator(),
		widget.NewLabel("Keywords (one per line):"),
		kwEntry,
		widget.NewSeparator(),
		container.NewHBox(layout.NewSpacer(), deleteBtn),
	)

	scroll := container.NewVScroll(form)
	scroll.SetMinSize(fyne.NewSize(580, 450))

	d := dialog.NewCustomConfirm(fmt.Sprintf("Edit: %s", name), "Save", "Cancel", scroll, func(ok bool) {
		if !ok {
			return
		}
		rule.Extensions = parseLines(extEntry.Text)
		rule.Keywords = parseLines(kwEntry.Text)
		a.cfg.Categories[name] = rule
		a.cfg.BuildOrder()
		a.saveConfig()
		if onSaved != nil {
			onSaved()
		}
	}, a.window)
	d.Resize(fyne.NewSize(600, 500))
	d.Show()
}

func (a *App) showAddCategoryDialog(onSaved func()) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Category name")

	extEntry := widget.NewMultiLineEntry()
	extEntry.SetPlaceHolder(".ext1\n.ext2")
	extEntry.SetMinRowsVisible(5)

	kwEntry := widget.NewMultiLineEntry()
	kwEntry.SetPlaceHolder("keyword1\nkeyword2")
	kwEntry.SetMinRowsVisible(5)

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

	scroll := container.NewVScroll(form)
	scroll.SetMinSize(fyne.NewSize(580, 400))

	d := dialog.NewCustomConfirm("Add Category", "Add", "Cancel", scroll, func(ok bool) {
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
		if onSaved != nil {
			onSaved()
		}
	}, a.window)
	d.Resize(fyne.NewSize(600, 460))
	d.Show()
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

	d := dialog.NewCustomConfirm("Edit Fallback", "Save", "Cancel", form, func(ok bool) {
		if !ok || entry.Text == "" {
			return
		}
		a.cfg.Fallback = entry.Text
		a.saveConfig()
	}, a.window)
	d.Resize(fyne.NewSize(420, 220))
	d.Show()
}

func (a *App) saveConfig() {
	if err := core.GenerateConfig(a.cfgPath); err != nil {
		a.statusLbl.SetText("Warning: could not save config: " + err.Error())
		return
	}
	a.statusLbl.SetText("Config saved")
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
