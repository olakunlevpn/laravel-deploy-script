package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"laravel-deploy-panel/config"
	"laravel-deploy-panel/deploy"
)

type DeployStepStatus struct {
	Step      int    `json:"step"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Output    string `json:"output"`
	Timestamp string `json:"timestamp"`
}

type DeployStatus struct {
	Running bool               `json:"running"`
	Steps   []DeployStepStatus `json:"steps"`
}

var deployMu sync.Mutex
var lastDeployStatus = &DeployStatus{Steps: []DeployStepStatus{}}

var deployHistoryPath = deployHistoryFilePath()

func deployHistoryFilePath() string {
	exe, err := os.Executable()
	if err != nil {
		return "deploy_history.json"
	}
	return filepath.Join(filepath.Dir(exe), "deploy_history.json")
}

func persistDeployStatus() {
	deployMu.Lock()
	data, err := json.MarshalIndent(lastDeployStatus, "", "  ")
	deployMu.Unlock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "deploy history: failed to marshal: %v\n", err)
		return
	}
	if err := os.WriteFile(deployHistoryPath, data, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "deploy history: failed to write: %v\n", err)
	}
}

func loadDeployHistory() {
	data, err := os.ReadFile(deployHistoryPath)
	if err != nil {
		return
	}
	deployMu.Lock()
	json.Unmarshal(data, lastDeployStatus)
	deployMu.Unlock()
}

func init() {
	loadDeployHistory()
}

// GET /api/deploy/stream — SSE stream, EventSource-compatible
func handleDeployStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(w, "data: {\"error\": \"failed to load config\"}\n\n")
		flusher.Flush()
		return
	}

	deployMu.Lock()
	if lastDeployStatus.Running {
		deployMu.Unlock()
		fmt.Fprintf(w, "data: {\"error\": \"deploy already in progress\"}\n\n")
		flusher.Flush()
		return
	}
	lastDeployStatus = &DeployStatus{Running: true, Steps: []DeployStepStatus{}}
	deployMu.Unlock()

	// Run preflight checks first
	preflight := deploy.RunPreflight(cfg)
	preflightData, _ := json.Marshal(map[string]interface{}{
		"preflight": true,
		"result":    preflight,
	})
	fmt.Fprintf(w, "data: %s\n\n", preflightData)
	flusher.Flush()

	if !preflight.Passed {
		deployMu.Lock()
		lastDeployStatus.Running = false
		deployMu.Unlock()
		fmt.Fprintf(w, "data: {\"error\": \"preflight checks failed\"}\n\n")
		flusher.Flush()
		persistDeployStatus()
		return
	}

	runner := deploy.NewRunner(cfg)
	runner.RunAll(func(result deploy.StepResult) {
		step := DeployStepStatus{
			Step:      result.Step,
			Name:      result.Name,
			Status:    result.Status,
			Output:    result.Output,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		deployMu.Lock()
		lastDeployStatus.Steps = append(lastDeployStatus.Steps, step)
		deployMu.Unlock()

		persistDeployStatus()

		data, _ := json.Marshal(step)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	})

	deployMu.Lock()
	lastDeployStatus.Running = false
	deployMu.Unlock()

	persistDeployStatus()

	fmt.Fprintf(w, "data: {\"done\": true}\n\n")
	flusher.Flush()
}

// GET /api/deploy/status
func handleDeployStatus(w http.ResponseWriter, r *http.Request) {
	deployMu.Lock()
	steps := make([]DeployStepStatus, len(lastDeployStatus.Steps))
	copy(steps, lastDeployStatus.Steps)
	status := DeployStatus{Running: lastDeployStatus.Running, Steps: steps}
	deployMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// POST /api/deploy/step/:n — re-run single step
func handleDeployStep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	n, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil || n < 1 || n > 11 {
		http.Error(w, "invalid step number", http.StatusBadRequest)
		return
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		http.Error(w, "failed to load config", http.StatusInternalServerError)
		return
	}

	deployMu.Lock()
	if lastDeployStatus.Running {
		deployMu.Unlock()
		http.Error(w, "deploy already in progress", http.StatusConflict)
		return
	}
	lastDeployStatus.Running = true
	deployMu.Unlock()

	runner := deploy.NewRunner(cfg)
	result := runner.RunStep(n)

	deployMu.Lock()
	lastDeployStatus.Running = false
	deployMu.Unlock()

	persistDeployStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GET /api/deploy/preflight
func handleDeployPreflight(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(configPath)
	if err != nil {
		http.Error(w, "failed to load config", http.StatusInternalServerError)
		return
	}
	result := deploy.RunPreflight(cfg)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
