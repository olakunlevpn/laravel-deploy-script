package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"laravel-deploy-panel/config"
)

func laravelLogPath() (string, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return "", err
	}
	return config.DeriveSiteRoot(cfg.Domain, cfg.SiteUser) + "/storage/logs/laravel.log", nil
}

// GET /api/logs/laravel — returns last 100 lines
func handleGetLogs(w http.ResponseWriter, r *http.Request) {
	path, err := laravelLogPath()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"lines": []string{},
			"error": "failed to load config: " + err.Error(),
		})
		return
	}
	lines, err := tailFile(path, 100)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"lines": []string{},
			"error": err.Error(),
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"lines": lines})
}

// POST /api/logs/laravel — truncates laravel.log
func handleClearLogs(w http.ResponseWriter, r *http.Request) {
	path, err := laravelLogPath()
	if err != nil {
		writeActionResult(w, ActionResult{Success: false, Output: "failed to load config: " + err.Error()})
		return
	}
	if err := os.Truncate(path, 0); err != nil {
		writeActionResult(w, ActionResult{Success: false, Output: err.Error()})
		return
	}
	writeActionResult(w, ActionResult{Success: true, Output: "Log cleared"})
}

// GET /api/logs/nginx-access — returns last 100 lines of nginx access log
func handleGetNginxAccessLogs(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(configPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"lines": []string{}, "error": "failed to load config"})
		return
	}
	path := fmt.Sprintf("/var/log/nginx/%s-access.log", cfg.Domain)
	if _, err := os.Stat(path); err != nil {
		path = "/var/log/nginx/access.log"
	}
	serveLogFile(w, path)
}

// GET /api/logs/nginx-error — returns last 100 lines of nginx error log
func handleGetNginxErrorLogs(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(configPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"lines": []string{}, "error": "failed to load config"})
		return
	}
	path := fmt.Sprintf("/var/log/nginx/%s-error.log", cfg.Domain)
	if _, err := os.Stat(path); err != nil {
		path = "/var/log/nginx/error.log"
	}
	serveLogFile(w, path)
}

func serveLogFile(w http.ResponseWriter, path string) {
	lines, err := tailFile(path, 100)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"lines": []string{}, "error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"lines": lines})
}

// GET /api/env — returns .env file content
func handleGetEnv(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(configPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"content": "", "error": "failed to load config"})
		return
	}
	envPath := config.DeriveSiteRoot(cfg.Domain, cfg.SiteUser) + "/.env"
	data, err := os.ReadFile(envPath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"content": "", "error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"content": string(data)})
}

// POST /api/env — saves .env file content
func handlePostEnv(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(configPath)
	if err != nil {
		writeActionResult(w, ActionResult{Success: false, Output: "failed to load config"})
		return
	}
	var body struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeActionResult(w, ActionResult{Success: false, Output: "invalid request body"})
		return
	}
	// Basic safety: reject if content contains null bytes
	if strings.Contains(body.Content, "\x00") {
		writeActionResult(w, ActionResult{Success: false, Output: "invalid content"})
		return
	}
	envPath := config.DeriveSiteRoot(cfg.Domain, cfg.SiteUser) + "/.env"
	// Backup existing .env before overwrite
	if existing, err := os.ReadFile(envPath); err == nil {
		os.WriteFile(envPath+".backup", existing, 0640)
	}
	if err := os.WriteFile(envPath, []byte(body.Content), 0640); err != nil {
		writeActionResult(w, ActionResult{Success: false, Output: err.Error()})
		return
	}
	writeActionResult(w, ActionResult{Success: true, Output: ".env saved"})
}

func tailFile(path string, n int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB per line max
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}
	return lines, scanner.Err()
}
