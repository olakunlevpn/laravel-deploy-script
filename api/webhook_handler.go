package api

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"laravel-deploy-panel/config"
)

// POST /api/webhook/deploy — pull latest code + run post-deploy commands
func handleWebhookDeploy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		writeActionResult(w, ActionResult{Success: false, Output: "failed to load config"})
		return
	}

	siteRoot := config.DeriveSiteRoot(cfg.Domain, cfg.SiteUser)
	phpBin := fmt.Sprintf("php%s", cfg.PHPVersion)

	steps := []struct {
		name string
		cmd  string
		args []string
	}{
		{"git pull", "sudo", []string{"-u", cfg.SiteUser, "git", "-C", siteRoot, "pull", "--ff-only"}},
		{"composer install", "sudo", []string{"-u", cfg.SiteUser, "composer", "install", "--no-dev", "--optimize-autoloader", "--working-dir=" + siteRoot}},
		{"migrate", phpBin, []string{siteRoot + "/artisan", "migrate", "--force"}},
		{"config:cache", phpBin, []string{siteRoot + "/artisan", "config:cache"}},
		{"route:cache", phpBin, []string{siteRoot + "/artisan", "route:cache"}},
		{"view:cache", phpBin, []string{siteRoot + "/artisan", "view:cache"}},
	}

	var outputs []string
	for _, s := range steps {
		out, err := exec.Command(s.cmd, s.args...).CombinedOutput()
		outputs = append(outputs, fmt.Sprintf("[%s] %s", s.name, strings.TrimSpace(string(out))))
		if err != nil {
			writeActionResult(w, ActionResult{
				Success: false,
				Output:  strings.Join(outputs, "\n"),
			})
			return
		}
	}

	writeActionResult(w, ActionResult{
		Success: true,
		Output:  strings.Join(outputs, "\n"),
	})
}
