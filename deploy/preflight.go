package deploy

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"laravel-deploy-panel/config"
)

type PreflightResult struct {
	Passed bool    `json:"passed"`
	Checks []Check `json:"checks"`
}

type Check struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

func RunPreflight(cfg *config.Config) PreflightResult {
	var checks []Check
	allPassed := true

	// Check required commands exist
	required := []struct{ cmd, name string }{
		{"nginx", "Nginx"},
		{"git", "Git"},
		{"composer", "Composer"},
		{fmt.Sprintf("php%s", cfg.PHPVersion), fmt.Sprintf("PHP %s", cfg.PHPVersion)},
	}

	if cfg.DBType == "" || cfg.DBType == "mysql" {
		required = append(required, struct{ cmd, name string }{"mysql", "MySQL Client"})
	} else if cfg.DBType == "postgresql" {
		required = append(required, struct{ cmd, name string }{"psql", "PostgreSQL Client"})
	}

	if cfg.DNSConfirmed {
		required = append(required, struct{ cmd, name string }{"certbot", "Certbot"})
	}
	if cfg.EnableQueueWorker {
		required = append(required, struct{ cmd, name string }{"supervisorctl", "Supervisor"})
	}

	for _, r := range required {
		_, err := exec.LookPath(r.cmd)
		c := Check{Name: r.name, Passed: err == nil}
		if err != nil {
			c.Message = fmt.Sprintf("%s not found in PATH", r.cmd)
			allPassed = false
		} else {
			c.Message = "Installed"
		}
		checks = append(checks, c)
	}

	// Check services are running
	services := []string{"nginx"}
	if cfg.DBType == "" || cfg.DBType == "mysql" {
		services = append(services, "mysql")
	} else if cfg.DBType == "postgresql" {
		services = append(services, "postgresql")
	}
	for _, svc := range services {
		out, err := exec.Command("systemctl", "is-active", svc).Output()
		active := strings.TrimSpace(string(out)) == "active"
		c := Check{Name: svc + " service", Passed: active}
		if !active || err != nil {
			c.Message = fmt.Sprintf("%s is not running", svc)
			allPassed = false
		} else {
			c.Message = "Running"
		}
		checks = append(checks, c)
	}

	// Check disk space (warn if < 1GB free on /home)
	out, err := exec.Command("df", "--output=avail", "-B1", "/home").Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		if len(lines) > 1 {
			avail := strings.TrimSpace(lines[1])
			// Parse and check > 1GB
			var bytes int64
			fmt.Sscanf(avail, "%d", &bytes)
			c := Check{Name: "Disk space", Passed: bytes > 1073741824}
			if bytes > 1073741824 {
				c.Message = fmt.Sprintf("%d MB free", bytes/1048576)
			} else {
				c.Message = fmt.Sprintf("Only %d MB free — at least 1 GB recommended", bytes/1048576)
				allPassed = false
			}
			checks = append(checks, c)
		}
	}

	// DNS resolution check
	if cfg.DNSConfirmed && cfg.Domain != "" {
		ips, err := net.LookupHost(cfg.Domain)
		c := Check{Name: "DNS resolution", Passed: err == nil && len(ips) > 0}
		if err != nil || len(ips) == 0 {
			// Don't fail preflight for this, just warn
			c.Passed = true
			c.Message = fmt.Sprintf("WARNING: %s does not resolve yet", cfg.Domain)
		} else {
			c.Message = fmt.Sprintf("Resolves to %s", strings.Join(ips, ", "))
		}
		checks = append(checks, c)
	}

	return PreflightResult{Passed: allPassed, Checks: checks}
}
