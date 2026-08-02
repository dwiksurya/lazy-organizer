package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "categories.yaml")
	os.WriteFile(cfgPath, []byte(`
categories:
  Documents:
    extensions: [.pdf, .txt]
    keywords: [invoice, report]
  Images:
    extensions: [.jpg, .png]
    keywords: [screenshot, photo]
  Game:
    extensions: [.rom, .iso]
    keywords: []
fallback: Others
`), 0644)

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	if got := cfg.Classify(".pdf", "report.pdf"); got != "Documents" {
		t.Errorf("Classify .pdf = %q, want Documents", got)
	}
	if got := cfg.Classify(".jpg", "pic.jpg"); got != "Images" {
		t.Errorf("Classify .jpg = %q, want Images", got)
	}
	if got := cfg.Classify(".rom", "game.rom"); got != "Game" {
		t.Errorf("Classify .rom = %q, want Game", got)
	}
	if got := cfg.Classify(".xyz", "invoice_2024.xyz"); got != "Documents" {
		t.Errorf("Classify 'invoice_2024.xyz' = %q, want Documents", got)
	}
	if got := cfg.Classify(".abc", "screenshot_now.abc"); got != "Images" {
		t.Errorf("Classify 'screenshot_now.abc' = %q, want Images", got)
	}
	if got := cfg.Classify("", "my_invoice_final"); got != "Documents" {
		t.Errorf("Classify 'my_invoice_final' = %q, want Documents", got)
	}
	if got := cfg.Classify(".xyz", "random_file.xyz"); got != "Others" {
		t.Errorf("Classify 'random_file.xyz' = %q, want Others", got)
	}
	if got := cfg.Classify("", "nothing_here"); got != "Others" {
		t.Errorf("Classify empty = %q, want Others", got)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if got := cfg.Classify(".pdf", "doc.pdf"); got != "Documents" {
		t.Errorf(".pdf = %q, want Documents", got)
	}
	if got := cfg.Classify(".bin", "invoice_march.bin"); got != "Documents" {
		t.Errorf("invoice_march.bin = %q, want Documents", got)
	}
	if got := cfg.Classify(".bin", "screenshot_2024.bin"); got != "Images" {
		t.Errorf("screenshot_2024.bin = %q, want Images", got)
	}
	if got := cfg.Classify(".xyz", "random.xyz"); got != "Others" {
		t.Errorf("random.xyz = %q, want Others", got)
	}
}

func TestGenerateConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "categories.yaml")
	if err := GenerateConfig(cfgPath); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("generated config is empty")
	}
}

func TestScan(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "report.pdf"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, "photo.jpg"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, "song.mp3"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(dir, "invoice_unknown.xyz"), []byte("x"), 0644)
	os.Mkdir(filepath.Join(dir, "subdir"), 0755)

	cfg := DefaultConfig()
	files, err := Scan(dir, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 4 {
		t.Fatalf("got %d files, want 4", len(files))
	}
	for _, f := range files {
		switch f.Name {
		case "report.pdf":
			if f.Category != "Documents" {
				t.Errorf("pdf classified as %q", f.Category)
			}
		case "invoice_unknown.xyz":
			if f.Category != "Documents" {
				t.Errorf("invoice_unknown.xyz classified as %q, want Documents", f.Category)
			}
		}
	}
}

func TestMove(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "report.pdf")
	os.WriteFile(src, []byte("hello"), 0644)

	fi := FileInfo{Name: "report.pdf", Path: src, Ext: ".pdf", Category: "Documents"}
	err := Move(dir, fi)
	if err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(dir, "Documents", "report.pdf")
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("file not moved: %v", err)
	}

	h, err := LoadHistory(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(h.Moves) != 1 || h.Moves[0].From != src {
		t.Fatalf("history wrong: %+v", h.Moves)
	}
}

func TestUndo(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "report.pdf")
	os.WriteFile(src, []byte("hello"), 0644)

	fi := FileInfo{Name: "report.pdf", Path: src, Ext: ".pdf", Category: "Documents"}
	Move(dir, fi)

	if err := UndoAll(dir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(src); err != nil {
		t.Fatal("file not restored")
	}
	if _, err := os.Stat(filepath.Join(dir, "Documents")); !os.IsNotExist(err) {
		t.Fatal("empty category dir not cleaned")
	}
}

func TestDuplicateMove(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("first"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("second"), 0644)

	fi1 := FileInfo{Name: "a.txt", Path: filepath.Join(dir, "a.txt"), Ext: ".txt", Category: "Documents"}
	fi2 := FileInfo{Name: "b.txt", Path: filepath.Join(dir, "b.txt"), Ext: ".txt", Category: "Documents"}

	// Move both files with same target name into same category
	Move(dir, fi1)
	// Rename b.txt to a.txt before second move to simulate collision
	os.Rename(filepath.Join(dir, "b.txt"), filepath.Join(dir, "a_copy.txt"))
	fi2.Name = "a_copy.txt"
	fi2.Path = filepath.Join(dir, "a_copy.txt")
	Move(dir, fi2)

	h, _ := LoadHistory(dir)
	if len(h.Moves) != 2 {
		t.Fatalf("expected 2 moves, got %d", len(h.Moves))
	}
}
