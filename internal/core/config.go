package core

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type CategoryRule struct {
	Extensions []string `yaml:"extensions" json:"extensions"`
	Keywords   []string `yaml:"keywords" json:"keywords"`
}

type Config struct {
	Categories map[string]CategoryRule `yaml:"categories"`
	Fallback   string                 `yaml:"fallback"`
	orderedCats []string
}

func (c *Config) Classify(ext, name string) string {
	ext = strings.ToLower(ext)
	base := strings.ToLower(strings.TrimSuffix(name, ext))

	for _, cat := range c.orderedCats {
		rule := c.Categories[cat]
		for _, e := range rule.Extensions {
			if strings.ToLower(e) == ext {
				return cat
			}
		}
	}

	for _, cat := range c.orderedCats {
		rule := c.Categories[cat]
		for _, kw := range rule.Keywords {
			if strings.Contains(base, strings.ToLower(kw)) {
				return cat
			}
		}
	}

	return c.Fallback
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Fallback == "" {
		cfg.Fallback = "Others"
	}
	cfg.BuildOrder()
	return &cfg, nil
}

func DefaultConfig() *Config {
	cfg := &Config{
		Categories: map[string]CategoryRule{
			"Documents":   {Extensions: []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".csv", ".odt", ".rtf"}, Keywords: []string{"invoice", "receipt", "report", "resume", "cv"}},
			"Images":      {Extensions: []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp", ".ico", ".tiff"}, Keywords: []string{"screenshot", "img", "photo", "wallpaper", "capture"}},
			"Videos":      {Extensions: []string{".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm"}, Keywords: []string{"video", "film", "movie", "recording"}},
			"Music":       {Extensions: []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a"}, Keywords: []string{"music", "song", "audio", "soundtrack", "album"}},
			"Archives":    {Extensions: []string{".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz"}, Keywords: []string{"backup", "dump"}},
			"Code":        {Extensions: []string{".py", ".js", ".go", ".java", ".c", ".cpp", ".rs", ".html", ".css", ".json", ".yaml", ".yml", ".xml", ".sh", ".sql"}, Keywords: []string{"src", "source", "main", "app", "lib", "test", "spec"}},
			"Executables": {Extensions: []string{".exe", ".msi", ".deb", ".rpm", ".dmg", ".appimage", ".snap"}, Keywords: []string{"installer", "setup", "install"}},
			"Torrents":    {Extensions: []string{".torrent"}},
			"Fonts":       {Extensions: []string{".ttf", ".otf", ".woff", ".woff2"}, Keywords: []string{"font"}},
			"ISOs":        {Extensions: []string{".iso"}},
		},
		Fallback: "Others",
	}
	cfg.BuildOrder()
	return cfg
}

func (c *Config) BuildOrder() {
	c.orderedCats = make([]string, 0, len(c.Categories))
	for cat := range c.Categories {
		c.orderedCats = append(c.orderedCats, cat)
	}
	sort.Strings(c.orderedCats)
}

func DefaultConfigPath() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "lazy-organizer", "categories.yaml")
}

func GenerateConfig(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	content := `# lazy-organizer categories
# Edit as needed.
#
# Each category has:
#   extensions: [.ext1, .ext2, ...]  — match by file extension
#   keywords: [word1, word2, ...]    — match by filename substring
#
# Priority: extension > keyword > fallback (Others)

categories:
  Documents:
    extensions: [.pdf, .doc, .docx, .xls, .xlsx, .ppt, .pptx, .txt, .csv, .odt, .rtf]
    keywords: [invoice, receipt, report, resume, cv]
  Images:
    extensions: [.jpg, .jpeg, .png, .gif, .bmp, .svg, .webp, .ico, .tiff]
    keywords: [screenshot, img, photo, wallpaper, capture]
  Videos:
    extensions: [.mp4, .mkv, .avi, .mov, .wmv, .flv, .webm]
    keywords: [video, film, movie, recording]
  Music:
    extensions: [.mp3, .wav, .flac, .aac, .ogg, .wma, .m4a]
    keywords: [music, song, audio, soundtrack, album]
  Archives:
    extensions: [.zip, .rar, .7z, .tar, .gz, .bz2, .xz]
    keywords: [backup, dump]
  Code:
    extensions: [.py, .js, .go, .java, .c, .cpp, .rs, .html, .css, .json, .yaml, .yml, .xml, .sh, .sql]
    keywords: [src, source, main, app, lib, test, spec]
  Executables:
    extensions: [.exe, .msi, .deb, .rpm, .dmg, .appimage, .snap]
    keywords: [installer, setup, install]
  Torrents:
    extensions: [.torrent]
    keywords: []
  Fonts:
    extensions: [.ttf, .otf, .woff, .woff2]
    keywords: [font]
  ISOs:
    extensions: [.iso]
    keywords: []

fallback: Others
`
	return os.WriteFile(path, []byte(content), 0644)
}

func (c *Config) ListCategories() []string {
	return append(c.orderedCats, c.Fallback)
}