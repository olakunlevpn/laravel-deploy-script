package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"laravel-deploy-panel/config"
)

var configPath = configFilePath()

func configFilePath() string {
	exe, err := os.Executable()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(filepath.Dir(exe), "config.json")
}

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(configPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

func handlePostConfig(w http.ResponseWriter, r *http.Request) {
	var cfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate config before saving
	if err := cfg.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Before saving, backup existing config
	if _, err := os.Stat(configPath); err == nil {
		backupPath := configPath + ".backup"
		data, _ := os.ReadFile(configPath)
		os.WriteFile(backupPath, data, 0600)
	}

	if err := config.Save(&cfg, configPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}
