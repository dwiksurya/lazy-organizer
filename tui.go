package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"lazy-organizer/internal/core"
)

var scanner = bufio.NewScanner(os.Stdin)

func readLine(prompt string) string {
	fmt.Print(prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func RunTUI(cfgPath string) error {
	cfg, err := core.LoadConfig(cfgPath)
	if err != nil {
		cfg = core.DefaultConfig()
		fmt.Println("(config not found, using defaults)")
	}

	for {
		fmt.Println("\n╔══════════════════════════════════════╗")
		fmt.Println("║   lazy-organizer — Config Editor     ║")
		fmt.Println("╠══════════════════════════════════════╣")
		fmt.Println("║  1) List categories                  ║")
		fmt.Println("║  2) Edit category                    ║")
		fmt.Println("║  3) Add new category                 ║")
		fmt.Println("║  4) Delete category                  ║")
		fmt.Println("║  5) Edit fallback                    ║")
		fmt.Println("║  6) Reset to defaults                ║")
		fmt.Println("║  0) Save & exit                      ║")
		fmt.Println("╚══════════════════════════════════════╝")

		choice := readLine("\nSelect: ")

		switch choice {
		case "1":
			showCategories(cfg)
		case "2":
			editCategory(cfg)
		case "3":
			addCategory(cfg)
		case "4":
			deleteCategory(cfg)
		case "5":
			editFallback(cfg)
		case "6":
			cfg = core.DefaultConfig()
			fmt.Println("✓ Reset to defaults")
		case "0":
			return saveConfig(cfgPath, cfg)
		default:
			fmt.Println("Invalid choice")
		}
	}
}

func showCategories(cfg *core.Config) {
	fmt.Printf("\n  Fallback: %s\n\n", cfg.Fallback)
	for _, name := range cfg.ListCategories() {
		rule := cfg.Categories[name]
		fmt.Printf("  [%s]\n", name)
		fmt.Printf("    extensions: %s\n", strings.Join(rule.Extensions, ", "))
		if len(rule.Keywords) > 0 {
			fmt.Printf("    keywords:   %s\n", strings.Join(rule.Keywords, ", "))
		}
		fmt.Println()
	}
}

func editCategory(cfg *core.Config) {
	cats := cfg.ListCategories()
	fmt.Println("\n  Categories:")
	for i, c := range cats {
		fmt.Printf("    %d) %s\n", i+1, c)
	}
	choice := readLine("\n  Select number (0=cancel): ")
	idx := -1
	for i := range cats {
		if fmt.Sprintf("%d", i+1) == choice {
			idx = i
			break
		}
	}
	if idx < 0 || idx >= len(cats) {
		return
	}
	name := cats[idx]
	if name == cfg.Fallback {
		fmt.Println("  Cannot edit fallback here (use option 5)")
		return
	}
	rule := cfg.Categories[name]

	fmt.Printf("\n  Edit [%s]\n", name)
	fmt.Printf("  Current extensions: %s\n", strings.Join(rule.Extensions, ", "))
	newExt := readLine("  New extensions (comma-separated, Enter=skip): ")
	if newExt != "" {
		rule.Extensions = parseCSV(newExt)
	}

	fmt.Printf("  Current keywords: %s\n", strings.Join(rule.Keywords, ", "))
	newKw := readLine("  New keywords (comma-separated, Enter=skip, -=empty): ")
	if newKw == "-" {
		rule.Keywords = nil
	} else if newKw != "" {
		rule.Keywords = parseCSV(newKw)
	}

	cfg.Categories[name] = rule
	cfg.BuildOrder()
	fmt.Printf("  ✓ [%s] updated\n", name)
}

func addCategory(cfg *core.Config) {
	name := readLine("\n  New category name: ")
	if name == "" {
		return
	}
	if _, exists := cfg.Categories[name]; exists {
		fmt.Printf("  ✗ [%s] already exists\n", name)
		return
	}
	ext := readLine("  Extensions (comma-separated): ")
	kw := readLine("  Keywords (comma-separated): ")
	cfg.Categories[name] = core.CategoryRule{
		Extensions: parseCSV(ext),
		Keywords:   parseCSV(kw),
	}
	cfg.BuildOrder()
	fmt.Printf("  ✓ [%s] added\n", name)
}

func deleteCategory(cfg *core.Config) {
	cats := cfg.ListCategories()
	fmt.Println("\n  Categories:")
	for i, c := range cats {
		fmt.Printf("    %d) %s\n", i+1, c)
	}
	choice := readLine("\n  Delete number (0=cancel): ")
	idx := -1
	for i := range cats {
		if fmt.Sprintf("%d", i+1) == choice {
			idx = i
			break
		}
	}
	if idx < 0 || idx >= len(cats) {
		return
	}
	name := cats[idx]
	if name == cfg.Fallback {
		fmt.Println("  Cannot delete fallback")
		return
	}
	confirm := readLine(fmt.Sprintf("  Delete [%s]? (y/N): ", name))
	if confirm == "y" || confirm == "Y" {
		delete(cfg.Categories, name)
		cfg.BuildOrder()
		fmt.Printf("  ✓ [%s] deleted\n", name)
	}
}

func editFallback(cfg *core.Config) {
	fmt.Printf("\n  Current fallback: %s\n", cfg.Fallback)
	new := readLine("  New fallback: ")
	if new != "" {
		cfg.Fallback = new
		fmt.Printf("  ✓ Fallback → %s\n", new)
	}
}

func parseCSV(s string) []string {
	parts := strings.Split(s, ",")
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

func saveConfig(path string, cfg *core.Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data := cfgToYAML(cfg)
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		return err
	}
	fmt.Printf("\n✓ Config saved: %s\n", path)
	return nil
}

func cfgToYAML(cfg *core.Config) string {
	var b strings.Builder
	b.WriteString("# lazy-organizer categories\n")
	b.WriteString("# Priority: extension > keyword > fallback\n\n")
	b.WriteString("categories:\n")
	for _, cat := range cfg.ListCategories() {
		if cat == cfg.Fallback {
			continue
		}
		rule := cfg.Categories[cat]
		fmt.Fprintf(&b, "  %s:\n", cat)
		fmt.Fprintf(&b, "    extensions: [%s]\n", strings.Join(rule.Extensions, ", "))
		if len(rule.Keywords) > 0 {
			fmt.Fprintf(&b, "    keywords: [%s]\n", strings.Join(rule.Keywords, ", "))
		} else {
			fmt.Fprintf(&b, "    keywords: []\n")
		}
	}
	fmt.Fprintf(&b, "\nfallback: %s\n", cfg.Fallback)
	return b.String()
}
