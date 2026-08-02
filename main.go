package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
)

func main() {
	dir := flag.String("dir", ".", "target folder")
	undo := flag.Bool("undo", false, "undo last move")
	dryRun := flag.Bool("dry-run", false, "preview without moving")
	interactive := flag.Bool("interactive", false, "choose category per file")
	yes := flag.Bool("yes", false, "skip confirmation")
	configPath := flag.String("config", "", "path to categories.yaml")
	initCfg := flag.Bool("init-config", false, "generate default config and exit")
	gui := flag.Bool("gui", false, "open TUI config editor")
	flag.Parse()

	cfgPath := *configPath
	if cfgPath == "" {
		cfgPath = DefaultConfigPath()
	}

	if *gui {
		if err := RunTUI(cfgPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *initCfg {
		path := *configPath
		if path == "" {
			path = DefaultConfigPath()
		}
		if err := GenerateConfig(path); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ config generated: %s\n", path)
		return
	}

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		cfg = DefaultConfig()
		fmt.Fprintf(os.Stderr, "warn: config not found (%v), using defaults\n", err)
		fmt.Fprintf(os.Stderr, "  tip: run -init-config to generate %s\n", DefaultConfigPath())
	}

	absDir, _ := filepath.Abs(*dir)

	if *undo {
		if err := UndoAll(absDir); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ undo complete")
		return
	}

	files, err := Scan(absDir, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Println("no files to organize")
		return
	}

	if *interactive {
		files = InteractiveClassify(files, cfg)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FILE\t→ FOLDER")
	fmt.Fprintln(w, "----\t  ------")
	for _, f := range files {
		fmt.Fprintf(w, "%s\t→ %s/\n", f.Name, f.Category)
	}
	w.Flush()

	if *dryRun {
		fmt.Printf("\n(dry-run: %d files would be moved)\n", len(files))
		return
	}

	if !*yes {
		fmt.Printf("\nMove %d files? [y/N] ", len(files))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if scanner.Text() != "y" && scanner.Text() != "Y" {
			fmt.Println("cancelled")
			return
		}
	}

	moved := 0
	for _, f := range files {
		if err := Move(absDir, f); err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", f.Name, err)
			continue
		}
		moved++
	}
	fmt.Printf("✓ moved %d/%d files (undo: -undo)\n", moved, len(files))
}
