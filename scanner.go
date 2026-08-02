package main

import (
	"os"
	"path/filepath"
	"sort"
)

type FileInfo struct {
	Name     string
	Path     string
	Ext      string
	Category string
}

func Scan(dir string, cfg *Config) ([]FileInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []FileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		cat := cfg.Classify(ext, e.Name())
		files = append(files, FileInfo{
			Name:     e.Name(),
			Path:     filepath.Join(dir, e.Name()),
			Ext:      ext,
			Category: cat,
		})
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].Category != files[j].Category {
			return files[i].Category < files[j].Category
		}
		return files[i].Name < files[j].Name
	})
	return files, nil
}
