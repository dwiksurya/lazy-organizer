package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lxn/walk"
	"gopkg.in/yaml.v3"

	"lazy-organizer/internal/core"
)

type File struct {
	Name     string
	Ext      string
	Category string
	Checked  bool
}

type FileModel struct {
	walk.TableModelBase
	files []*File
}

func (m *FileModel) RowCount() int { return len(m.files) }

func (m *FileModel) Value(row, col int) interface{} {
	f := m.files[row]
	switch col {
	case 1:
		return f.Name
	case 2:
		return f.Ext
	case 3:
		return f.Category
	}
	return ""
}

func (m *FileModel) Checked(row int) bool { return m.files[row].Checked }

func (m *FileModel) SetChecked(row int, checked bool) error {
	m.files[row].Checked = checked
	return nil
}

type App struct {
	*walk.MainWindow
	cfg     *core.Config
	cfgPath string
	dir     string
	all     []*File
	model   *FileModel
	table   *walk.TableView
	filter  *walk.LineEdit
	status  *walk.Label
	orgBtn  *walk.PushButton
}

func main() {
	mw, err := walk.NewMainWindow()
	if err != nil {
		panic(err)
	}
	cfgPath := core.DefaultConfigPath()
	cfg, err := core.LoadConfig(cfgPath)
	if err != nil {
		cfg = core.DefaultConfig()
		core.GenerateConfig(cfgPath)
	}
	a := &App{MainWindow: mw, cfg: cfg, cfgPath: cfgPath}
	if err := a.buildUI(); err != nil {
		panic(err)
	}
	a.Run()
}

func (a *App) buildUI() error {
	a.SetTitle("Lazy Organizer")
	a.SetClientSize(walk.Size{Width: 900, Height: 600})

	root, err := walk.NewComposite(a.MainWindow)
	if err != nil {
		return err
	}
	if err := root.SetLayout(walk.NewVBoxLayout()); err != nil {
		return err
	}

	// Toolbar
	toolbar, err := walk.NewComposite(root)
	if err != nil {
		return err
	}
	if err := toolbar.SetLayout(walk.NewHBoxLayout()); err != nil {
		return err
	}
	folderBtn, err := walk.NewPushButton(toolbar)
	if err != nil {
		return err
	}
	folderBtn.SetText("Select Folder...")
	folderBtn.Clicked().Attach(a.selectFolder)

	a.orgBtn, err = walk.NewPushButton(toolbar)
	if err != nil {
		return err
	}
	a.orgBtn.SetText("Organize (0)")
	a.orgBtn.Clicked().Attach(a.organize)

	undoBtn, err := walk.NewPushButton(toolbar)
	if err != nil {
		return err
	}
	undoBtn.SetText("Undo")
	undoBtn.Clicked().Attach(a.undo)

	cfgBtn, err := walk.NewPushButton(toolbar)
	if err != nil {
		return err
	}
	cfgBtn.SetText("Config...")
	cfgBtn.Clicked().Attach(a.showConfig)

	// Filter bar
	filterBar, err := walk.NewComposite(root)
	if err != nil {
		return err
	}
	if err := filterBar.SetLayout(walk.NewHBoxLayout()); err != nil {
		return err
	}
	a.filter, err = walk.NewLineEdit(filterBar)
	if err != nil {
		return err
	}
	a.filter.SetCueBanner("Filter by name, ext, or category...")
	a.filter.TextChanged().Attach(a.applyFilter)

	allBtn, err := walk.NewPushButton(filterBar)
	if err != nil {
		return err
	}
	allBtn.SetText("All")
	allBtn.Clicked().Attach(func() { a.setAll(true) })

	noneBtn, err := walk.NewPushButton(filterBar)
	if err != nil {
		return err
	}
	noneBtn.SetText("None")
	noneBtn.Clicked().Attach(func() { a.setAll(false) })

	invBtn, err := walk.NewPushButton(filterBar)
	if err != nil {
		return err
	}
	invBtn.SetText("Invert")
	invBtn.Clicked().Attach(a.invert)

	// Table
	a.table, err = walk.NewTableView(root)
	if err != nil {
		return err
	}
	a.table.SetCheckBoxes(true)
	a.table.SetLastColumnStretched(true)
	for _, c := range []struct {
		title string
		width int
	}{{"Name", 320}, {"Ext", 90}, {"Category", 200}} {
		col := walk.NewTableViewColumn()
		col.SetTitle(c.title)
		col.SetWidth(c.width)
		if err := a.table.Columns().Add(col); err != nil {
			return err
		}
	}
	a.table.ItemActivated().Attach(a.changeCategory)

	a.model = &FileModel{}
	if err := a.table.SetModel(a.model); err != nil {
		return err
	}

	// Status
	a.status, err = walk.NewLabel(root)
	if err != nil {
		return err
	}
	a.status.SetText("Select a folder to begin.")

	return nil
}

func (a *App) selectFolder() {
	dlg := new(walk.FileDialog)
	dlg.Title = "Select Folder"
	dlg.FilePath = a.dir
	if ok, err := dlg.ShowBrowseFolder(a.MainWindow); err != nil {
		return
	} else if !ok || dlg.FilePath == "" {
		return
	}
	a.dir = dlg.FilePath
	a.scanDir()
}

func (a *App) scanDir() {
	files, err := core.Scan(a.dir, a.cfg)
	if err != nil {
		a.status.SetText("Error: " + err.Error())
		return
	}
	a.all = make([]*File, 0, len(files))
	for _, f := range files {
		a.all = append(a.all, &File{Name: f.Name, Ext: f.Ext, Category: f.Category, Checked: true})
	}
	a.filter.SetText("")
	a.applyFilter()
}

func (a *App) visible() []*File {
	return a.model.files
}

func (a *App) applyFilter() {
	q := strings.ToLower(strings.TrimSpace(a.filter.Text()))
	var out []*File
	for _, f := range a.all {
		if q == "" ||
			strings.Contains(strings.ToLower(f.Name), q) ||
			strings.Contains(strings.ToLower(f.Ext), q) ||
			strings.Contains(strings.ToLower(f.Category), q) {
			out = append(out, f)
		}
	}
	a.model.files = out
	a.model.PublishRowsReset()
	a.updateCounts()
}

func (a *App) setAll(v bool) {
	for _, f := range a.visible() {
		f.Checked = v
	}
	a.model.PublishRowsReset()
	a.updateCounts()
}

func (a *App) invert() {
	for _, f := range a.visible() {
		f.Checked = !f.Checked
	}
	a.model.PublishRowsReset()
	a.updateCounts()
}

func (a *App) updateCounts() {
	total := len(a.all)
	shown := len(a.visible())
	checked := 0
	for _, f := range a.visible() {
		if f.Checked {
			checked++
		}
	}
	a.orgBtn.SetText(fmt.Sprintf("Organize (%d)", checked))
	if total == shown {
		a.status.SetText(fmt.Sprintf("%d files, %d selected", total, checked))
	} else {
		a.status.SetText(fmt.Sprintf("%d files shown (%d total), %d selected", shown, total, checked))
	}
}

func (a *App) organize() {
	var picks []*File
	for _, f := range a.visible() {
		if f.Checked {
			picks = append(picks, f)
		}
	}
	if len(picks) == 0 {
		return
	}
	res := walk.MsgBox(a.MainWindow, "Organize",
		fmt.Sprintf("Move %d selected files into category folders?", len(picks)),
		walk.MsgBoxOKCancel|walk.MsgBoxIconQuestion)
	if res != walk.DlgCmdOK {
		return
	}
	moved := 0
	for _, f := range picks {
		fi := core.FileInfo{Name: f.Name, Path: filepath.Join(a.dir, f.Name), Ext: f.Ext, Category: f.Category}
		if err := core.Move(a.dir, fi); err == nil {
			moved++
		}
	}
	a.status.SetText(fmt.Sprintf("Moved %d/%d files (use Undo to revert)", moved, len(picks)))
	a.scanDir()
}

func (a *App) undo() {
	if a.dir == "" {
		return
	}
	if err := core.UndoAll(a.dir); err != nil {
		walk.MsgBox(a.MainWindow, "Undo", err.Error(), walk.MsgBoxIconError)
		return
	}
	a.status.SetText("Undo complete")
	a.scanDir()
}

func (a *App) changeCategory() {
	idx := a.table.CurrentIndex()
	if idx < 0 || idx >= len(a.visible()) {
		return
	}
	f := a.visible()[idx]

	dlg, err := walk.NewDialog(a.MainWindow)
	if err != nil {
		return
	}
	dlg.SetTitle("Change Category")
	dlg.SetClientSize(walk.Size{Width: 320, Height: 140})
	if err := dlg.SetLayout(walk.NewVBoxLayout()); err != nil {
		return
	}
	if _, err := walk.NewLabel(dlg); err != nil {
		return
	}
	combo, err := walk.NewComboBox(dlg)
	if err != nil {
		return
	}
	names := a.cfg.ListCategories()
	sort.Strings(names)
	if err := combo.SetModel(names); err != nil {
		return
	}
	combo.SetCurrentIndex(indexOf(names, f.Category))

	btnRow, _ := walk.NewComposite(dlg)
	btnRow.SetLayout(walk.NewHBoxLayout())
	okBtn, _ := walk.NewPushButton(btnRow)
	okBtn.SetText("OK")
	okBtn.Clicked().Attach(func() {
		if idx := combo.CurrentIndex(); idx >= 0 {
			f.Category = names[idx]
		}
		dlg.Accept()
	})
	cancelBtn, _ := walk.NewPushButton(btnRow)
	cancelBtn.SetText("Cancel")
	cancelBtn.Clicked().Attach(func() { dlg.Cancel() })

	dlg.Run()
	a.model.PublishRowsReset()
	a.updateCounts()
}

func indexOf(list []string, v string) int {
	for i, s := range list {
		if s == v {
			return i
		}
	}
	return 0
}

func (a *App) showConfig() {
	dlg, err := walk.NewDialog(a.MainWindow)
	if err != nil {
		return
	}
	dlg.SetTitle("Config Editor")
	dlg.SetClientSize(walk.Size{Width: 520, Height: 480})
	if err := dlg.SetLayout(walk.NewVBoxLayout()); err != nil {
		return
	}

	lb, err := walk.NewListBox(dlg)
	if err != nil {
		return
	}
	cats := a.cfg.ListCategories()
	sort.Strings(cats)
	lb.SetModel(cats)

	extEdit, _ := walk.NewTextEdit(dlg)
	extEdit.SetMinMaxSize(walk.Size{Width: 0, Height: 60}, walk.Size{Width: 0, Height: 60})
	kwEdit, _ := walk.NewTextEdit(dlg)
	kwEdit.SetMinMaxSize(walk.Size{Width: 0, Height: 60}, walk.Size{Width: 0, Height: 60})

	loadCat := func() {
		i := lb.CurrentIndex()
		if i < 0 || i >= len(cats) {
			extEdit.SetText("")
			kwEdit.SetText("")
			return
		}
		name := cats[i]
		rule, ok := a.cfg.Categories[name]
		if !ok { // fallback
			extEdit.SetText("")
			kwEdit.SetText("")
			return
		}
		extEdit.SetText(strings.Join(rule.Extensions, "\n"))
		kwEdit.SetText(strings.Join(rule.Keywords, "\n"))
	}
	lb.CurrentIndexChanged().Attach(loadCat)
	if len(cats) > 0 {
		lb.SetCurrentIndex(0)
	}

	fallbackRow, _ := walk.NewComposite(dlg)
	fallbackRow.SetLayout(walk.NewHBoxLayout())
	fallbackLbl, err := walk.NewLabel(fallbackRow)
	if err != nil {
		return
	}
	fallbackLbl.SetText("Fallback:")
	fallbackCombo, _ := walk.NewComboBox(fallbackRow)
	fallbackCombo.SetModel(cats)
	fallbackCombo.SetCurrentIndex(indexOf(cats, a.cfg.Fallback))

	newRow, _ := walk.NewComposite(dlg)
	newRow.SetLayout(walk.NewHBoxLayout())
	nameEdit, _ := walk.NewLineEdit(newRow)
	nameEdit.SetCueBanner("New category name")
	addBtn, _ := walk.NewPushButton(newRow)
	addBtn.SetText("Add")
	addBtn.Clicked().Attach(func() {
		name := strings.TrimSpace(nameEdit.Text())
		if name == "" {
			return
		}
		if _, ok := a.cfg.Categories[name]; ok {
			return
		}
		a.cfg.Categories[name] = core.CategoryRule{}
		a.cfg.BuildOrder()
		cats = a.cfg.ListCategories()
		sort.Strings(cats)
		lb.SetModel(cats)
		fallbackCombo.SetModel(cats)
		lb.SetCurrentIndex(indexOf(cats, name))
		nameEdit.SetText("")
	})

	btnRow, _ := walk.NewComposite(dlg)
	btnRow.SetLayout(walk.NewHBoxLayout())
	saveBtn, _ := walk.NewPushButton(btnRow)
	saveBtn.SetText("Save")
	saveBtn.Clicked().Attach(func() {
		i := lb.CurrentIndex()
		if i >= 0 && i < len(cats) {
			name := cats[i]
			if rule, ok := a.cfg.Categories[name]; ok {
				rule.Extensions = parseLines(extEdit.Text())
				rule.Keywords = parseLines(kwEdit.Text())
				a.cfg.Categories[name] = rule
			}
		}
		if fc := fallbackCombo.CurrentIndex(); fc >= 0 {
			a.cfg.Fallback = cats[fc]
		}
		a.cfg.BuildOrder()
		a.saveConfig()
		if len(a.all) > 0 {
			a.applyFilter()
		}
		dlg.Accept()
	})
	delBtn, _ := walk.NewPushButton(btnRow)
	delBtn.SetText("Delete")
	delBtn.Clicked().Attach(func() {
		i := lb.CurrentIndex()
		if i < 0 || i >= len(cats) {
			return
		}
		name := cats[i]
		if name == a.cfg.Fallback {
			return
		}
		delete(a.cfg.Categories, name)
		a.cfg.BuildOrder()
		cats = a.cfg.ListCategories()
		sort.Strings(cats)
		lb.SetModel(cats)
		fallbackCombo.SetModel(cats)
		loadCat()
	})
	resetBtn, _ := walk.NewPushButton(btnRow)
	resetBtn.SetText("Reset")
	resetBtn.Clicked().Attach(func() {
		a.cfg = core.DefaultConfig()
		cats = a.cfg.ListCategories()
		sort.Strings(cats)
		lb.SetModel(cats)
		fallbackCombo.SetModel(cats)
		loadCat()
	})
	closeBtn, _ := walk.NewPushButton(btnRow)
	closeBtn.SetText("Close")
	closeBtn.Clicked().Attach(func() { dlg.Cancel() })

	dlg.Run()
}

func parseLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func (a *App) saveConfig() {
	data, err := yaml.Marshal(a.cfg)
	if err != nil {
		walk.MsgBox(a.MainWindow, "Config", "Save failed: "+err.Error(), walk.MsgBoxIconError)
		return
	}
	if err := os.MkdirAll(filepath.Dir(a.cfgPath), 0755); err != nil {
		walk.MsgBox(a.MainWindow, "Config", "Save failed: "+err.Error(), walk.MsgBoxIconError)
		return
	}
	if err := os.WriteFile(a.cfgPath, data, 0644); err != nil {
		walk.MsgBox(a.MainWindow, "Config", "Save failed: "+err.Error(), walk.MsgBoxIconError)
	}
}
