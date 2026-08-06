package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"lazy-organizer/internal/core"
)

//go:embed web/*
var webFS embed.FS

type server struct {
	cfg     *core.Config
	cfgPath string
}

type scanReq struct {
	Dir string `json:"dir"`
}

type organizeReq struct {
	Dir   string     `json:"dir"`
	Files []filePick `json:"files"`
}

type filePick struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type configBody struct {
	Categories map[string]core.CategoryRule `json:"categories"`
	Fallback   string                       `json:"fallback"`
}

func main() {
	port := flag.String("port", "17320", "port to listen on (0 = auto)")
	noOpen := flag.Bool("no-open", false, "do not open browser automatically")
	flag.Parse()

	cfgPath := core.DefaultConfigPath()
	cfg, err := core.LoadConfig(cfgPath)
	if err != nil {
		cfg = core.DefaultConfig()
		core.GenerateConfig(cfgPath)
	}

	s := &server{cfg: cfg, cfgPath: cfgPath}

	mux := http.NewServeMux()
	sub, _ := fs.Sub(webFS, "web")
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/api/scan", s.handleScan)
	mux.HandleFunc("/api/pick-folder", s.handlePickFolder)
	mux.HandleFunc("/api/organize", s.handleOrganize)
	mux.HandleFunc("/api/undo", s.handleUndo)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/config/reset", s.handleConfigReset)

	ln, err := net.Listen("tcp", "127.0.0.1:"+*port)
	if err != nil {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			fmt.Fprintln(os.Stderr, "listen:", err)
			os.Exit(1)
		}
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/", ln.Addr().(*net.TCPAddr).Port)
	fmt.Println("lazy-organizer GUI:", url)

	if !*noOpen {
		go func() {
			time.Sleep(300 * time.Millisecond)
			openBrowser(url)
		}()
	}

	http.Serve(ln, mux)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func decodeBody(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeJSON(w, 400, map[string]string{"error": "bad request: " + err.Error()})
		return false
	}
	return true
}

func (s *server) handleScan(w http.ResponseWriter, r *http.Request) {
	var req scanReq
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Dir == "" {
		writeJSON(w, 400, map[string]string{"error": "dir required"})
		return
	}
	files, err := core.Scan(req.Dir, s.cfg)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"files": files})
}

func (s *server) handlePickFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, 405, map[string]string{"error": "method not allowed"})
		return
	}
	path, err := pickFolder()
	if err != nil || path == "" {
		writeJSON(w, 200, map[string]any{"path": ""})
		return
	}
	writeJSON(w, 200, map[string]any{"path": path})
}

// pickFolder opens a native folder picker on the user's desktop.
func pickFolder() (string, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		script := "Add-Type -AssemblyName System.Windows.Forms; $f=New-Object System.Windows.Forms.FolderBrowserDialog; $f.Description='Select Folder'; if($f.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK){$f.SelectedPath}"
		cmd = exec.Command("powershell", "-NoProfile", "-STA", "-Command", script)
	case "darwin":
		cmd = exec.Command("osascript", "-e", `POSIX path of (choose folder with prompt "Select Folder")`)
	default:
		cmd = exec.Command("zenity", "--file-selection", "--directory", "--title=Select Folder")
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(strings.TrimSpace(string(out)), `/\`), nil
}

func (s *server) handleOrganize(w http.ResponseWriter, r *http.Request) {
	var req organizeReq
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Dir == "" || len(req.Files) == 0 {
		writeJSON(w, 400, map[string]string{"error": "dir and files required"})
		return
	}
	moved := 0
	for _, f := range req.Files {
		fi := core.FileInfo{
			Name:     f.Name,
			Path:     filepath.Join(req.Dir, f.Name),
			Ext:      filepath.Ext(f.Name),
			Category: f.Category,
		}
		if err := core.Move(req.Dir, fi); err == nil {
			moved++
		}
	}
	writeJSON(w, 200, map[string]any{"moved": moved, "total": len(req.Files)})
}

func (s *server) handleUndo(w http.ResponseWriter, r *http.Request) {
	var req scanReq
	if !decodeBody(w, r, &req) {
		return
	}
	if err := core.UndoAll(req.Dir); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, configBody{Categories: s.cfg.Categories, Fallback: s.cfg.Fallback})
	case http.MethodPut:
		var body configBody
		if !decodeBody(w, r, &body) {
			return
		}
		s.cfg.Categories = body.Categories
		s.cfg.Fallback = body.Fallback
		if s.cfg.Fallback == "" {
			s.cfg.Fallback = "Others"
		}
		s.cfg.BuildOrder()
		if err := s.saveConfig(); err != nil {
			writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]string{"status": "ok"})
	default:
		writeJSON(w, 405, map[string]string{"error": "method not allowed"})
	}
}

func (s *server) handleConfigReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, 405, map[string]string{"error": "method not allowed"})
		return
	}
	s.cfg = core.DefaultConfig()
	if err := s.saveConfig(); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (s *server) saveConfig() error {
	data, err := yaml.Marshal(s.cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.cfgPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(s.cfgPath, data, 0644)
}
