package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type MoveRecord struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type History struct {
	Moves []MoveRecord `json:"moves"`
}

const historyFile = ".lazy-organizer-history.json"

func historyPath(dir string) string {
	return filepath.Join(dir, historyFile)
}

func LoadHistory(dir string) (*History, error) {
	h := &History{}
	data, err := os.ReadFile(historyPath(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return h, nil
		}
		return nil, err
	}
	err = json.Unmarshal(data, h)
	return h, err
}

func SaveHistory(dir string, h *History) error {
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(historyPath(dir), data, 0644)
}

func Move(dir string, f FileInfo) error {
	catDir := filepath.Join(dir, f.Category)
	if err := os.MkdirAll(catDir, 0755); err != nil {
		return err
	}
	dst := filepath.Join(catDir, f.Name)

	if _, err := os.Stat(dst); err == nil {
		dst = uniquePath(dst)
	}

	if err := os.Rename(f.Path, dst); err != nil {
		if err := copyFile(f.Path, dst); err != nil {
			return fmt.Errorf("move %s: %w", f.Name, err)
		}
		os.Remove(f.Path)
	}

	h, _ := LoadHistory(dir)
	h.Moves = append(h.Moves, MoveRecord{From: f.Path, To: dst})
	return SaveHistory(dir, h)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err == nil {
		os.Chmod(dst, info.Mode())
	}
	return out.Close()
}

func uniquePath(p string) string {
	dir := filepath.Dir(p)
	ext := filepath.Ext(p)
	base := strings.TrimSuffix(filepath.Base(p), ext)
	for i := 1; ; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

func UndoAll(dir string) error {
	h, err := LoadHistory(dir)
	if err != nil {
		return err
	}
	if len(h.Moves) == 0 {
		return fmt.Errorf("no moves to undo")
	}
	for i := len(h.Moves) - 1; i >= 0; i-- {
		m := h.Moves[i]
		if err := os.Rename(m.To, m.From); err != nil {
			if err := copyFile(m.To, m.From); err != nil {
				continue
			}
			os.Remove(m.To)
		}
		parent := filepath.Dir(m.To)
		os.Remove(parent)
	}
	os.Remove(historyPath(dir))
	return nil
}
