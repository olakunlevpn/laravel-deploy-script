package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"laravel-deploy-panel/config"
)

type ActionResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

func runCommand(name string, args ...string) ActionResult {
	out, err := exec.Command(name, args...).CombinedOutput()
	return ActionResult{
		Success: err == nil,
		Output:  strings.TrimSpace(string(out)),
	}
}

func writeActionResult(w http.ResponseWriter, result ActionResult) {
	w.Header().Set("Content-Type", "application/json")
	if !result.Success {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(result)
}

// POST /api/actions/nginx/:action — reload, restart
func handleNginxAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	action := strings.TrimPrefix(r.URL.Path, "/api/actions/nginx/")
	switch action {
	case "reload":
		writeActionResult(w, runCommand("systemctl", "reload", "nginx"))
	case "restart":
		writeActionResult(w, runCommand("systemctl", "restart", "nginx"))
	default:
		http.Error(w, "unknown action: "+action, http.StatusBadRequest)
	}
}

// POST /api/actions/supervisor/:action — start, stop, restart
func handleSupervisorAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	action := strings.TrimPrefix(r.URL.Path, "/api/actions/supervisor/")
	switch action {
	case "start", "stop", "restart":
		writeActionResult(w, runCommand("systemctl", action, "supervisor"))
	default:
		http.Error(w, "unknown action: "+action, http.StatusBadRequest)
	}
}

// POST /api/actions/queue-worker/:action — start, stop, restart
func handleQueueWorkerAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	action := strings.TrimPrefix(r.URL.Path, "/api/actions/queue-worker/")
	cfg, err := config.Load(configPath)
	if err != nil {
		http.Error(w, "failed to load config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	workerName := config.DeriveSupervisorName(cfg.Domain)
	switch action {
	case "start", "stop", "restart":
		writeActionResult(w, runCommand("supervisorctl", action, workerName))
	default:
		http.Error(w, "unknown action: "+action, http.StatusBadRequest)
	}
}

// POST /api/actions/ssl/renew
func handleSSLRenew(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeActionResult(w, runCommand("certbot", "renew", "--nginx", "--non-interactive"))
}

// POST /api/actions/permissions
func handlePermissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		http.Error(w, "failed to load config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	siteRoot := config.DeriveSiteRoot(cfg.Domain, cfg.SiteUser)
	result := runCommand("chown", "-R", fmt.Sprintf("%s:%s", cfg.SiteUser, cfg.SiteGroup), siteRoot)
	if result.Success {
		result = runCommand("chmod", "-R", "755", siteRoot)
	}
	if result.Success {
		result = runCommand("chmod", "-R", "ug+rwx", siteRoot+"/storage")
	}
	writeActionResult(w, result)
}

var allowedLaravelActions = map[string]string{
	"cache:clear":      "cache:clear",
	"config:clear":     "config:clear",
	"route:clear":      "route:clear",
	"view:clear":       "view:clear",
	"migrate":          "migrate --force",
	"migrate:rollback": "migrate:rollback --force",
	"optimize":         "optimize",
	"storage:link":     "storage:link",
}

// POST /api/actions/laravel/:action
func handleLaravelAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	action := strings.TrimPrefix(r.URL.Path, "/api/actions/laravel/")
	artisanArgs, ok := allowedLaravelActions[action]
	if !ok {
		http.Error(w, "unknown action: "+action, http.StatusBadRequest)
		return
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		http.Error(w, "failed to load config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Validate PHPVersion to prevent injection via phpBin
	matched, _ := regexp.MatchString(`^[0-9]+\.[0-9]+$`, cfg.PHPVersion)
	if !matched {
		http.Error(w, "invalid php version in config", http.StatusBadRequest)
		return
	}
	siteRoot := config.DeriveSiteRoot(cfg.Domain, cfg.SiteUser)
	phpBin := fmt.Sprintf("php%s", cfg.PHPVersion)
	artisanPath := siteRoot + "/artisan"
	args := append([]string{artisanPath}, strings.Fields(artisanArgs)...)
	cmd := exec.Command(phpBin, args...)
	cmd.Dir = siteRoot
	out, err := cmd.CombinedOutput()
	writeActionResult(w, ActionResult{
		Success: err == nil,
		Output:  strings.TrimSpace(string(out)),
	})
}
