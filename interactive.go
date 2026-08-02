package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func InteractiveClassify(files []FileInfo, cfg *Config) []FileInfo {
	cats := cfg.ListCategories()
	sort.Strings(cats)

	scanner := bufio.NewScanner(os.Stdin)

	for i := range files {
		f := &files[i]
		fmt.Printf("\n[%d/%d] %s\n", i+1, len(files), f.Name)
		fmt.Printf("  suggested: %s\n", f.Category)
		fmt.Println("  categories:")
		for j, c := range cats {
			marker := " "
			if c == f.Category {
				marker = ">"
			}
			fmt.Printf("    %s %d) %s\n", marker, j+1, c)
		}
		fmt.Printf("  enter number, name, or Enter to accept [%s]: ", f.Category)

		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		for j, c := range cats {
			if fmt.Sprintf("%d", j+1) == input {
				f.Category = c
				goto done
			}
		}
		for _, c := range cats {
			if strings.EqualFold(c, input) {
				f.Category = c
				goto done
			}
		}
		f.Category = input
	done:
	}
	return files
}
