package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Config struct {
	Domain            string `json:"domain"`
	GithubRepo        string `json:"github_repo"`
	GithubBranch      string `json:"github_branch"`
	PHPVersion        string `json:"php_version"`
	DBPassword        string `json:"db_password"`
	DBName            string `json:"db_name"`
	DBUser            string `json:"db_user"`
	SiteUser          string `json:"site_user"`
	SiteGroup         string `json:"site_group"`
	DBType            string `json:"db_type"` // "mysql" or "postgresql"
	EnableQueueWorker bool   `json:"enable_queue_worker"`
	EnableScheduler   bool   `json:"enable_scheduler"`
	DNSConfirmed      bool   `json:"dns_confirmed"`
}

var nonAlphaNum    = regexp.MustCompile(`[^a-zA-Z0-9]+`)
var consecutiveUnd = regexp.MustCompile(`_+`)

// DeriveDBName mirrors deploy.sh line 23:
// 1. replace non-alphanumeric chars with _
// 2. strip trailing _com
// 3. collapse consecutive underscores
func DeriveDBName(domain string) string {
	s := nonAlphaNum.ReplaceAllString(domain, "_")
	s = strings.TrimSuffix(s, "_com")
	s = consecutiveUnd.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	return s
}

// DeriveDBUser returns "{dbName}_user"
func DeriveDBUser(dbName string) string {
	return dbName + "_user"
}

// DeriveSiteRoot returns "/home/{siteUser}/{domain}"
func DeriveSiteRoot(domain, siteUser string) string {
	return "/home/" + siteUser + "/" + domain
}

// DeriveSupervisorName returns the supervisor program name for the queue worker.
// Matches deploy.sh: domain dots replaced with hyphens + "-worker"
func DeriveSupervisorName(domain string) string {
	return strings.ReplaceAll(domain, ".", "-") + "-worker"
}

// DetectSiteUser returns SUDO_USER env var if set (panel run via sudo),
// otherwise returns the current user via USER env var, falling back to "www-data".
func DetectSiteUser() string {
	if u := os.Getenv("SUDO_USER"); u != "" {
		return u
	}
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	return "www-data"
}

// Load reads config from path. If the file does not exist, returns defaults.
func Load(path string) (*Config, error) {
	cfg := defaults()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		cfg.SiteUser = DetectSiteUser()
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Save writes config to path as JSON.
func Save(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func defaults() *Config {
	return &Config{
		GithubBranch:      "main",
		PHPVersion:        "8.3",
		DBType:            "mysql",
		SiteGroup:         "www-data",
		EnableQueueWorker: true,
		EnableScheduler:   true,
	}
}

// Validate checks required fields and format constraints, returning combined errors.
func (c *Config) Validate() error {
	var errs []string

	domainRe := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9.\-]*[a-zA-Z0-9])?$`)
	phpRe := regexp.MustCompile(`^[0-9]+\.[0-9]+$`)
	dbNameRe := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	dbUserRe := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	siteUserRe := regexp.MustCompile(`^[a-z_][a-z0-9_\-]*$`)

	if c.Domain == "" {
		errs = append(errs, "domain is required")
	} else if !domainRe.MatchString(c.Domain) {
		errs = append(errs, "domain format is invalid")
	}

	if c.GithubRepo == "" {
		errs = append(errs, "github_repo is required")
	} else if !strings.HasPrefix(c.GithubRepo, "https://") && !strings.HasPrefix(c.GithubRepo, "git@") {
		errs = append(errs, "github_repo must start with https:// or git@")
	}

	if c.GithubBranch == "" {
		errs = append(errs, "github_branch is required")
	}

	if c.PHPVersion != "" && !phpRe.MatchString(c.PHPVersion) {
		errs = append(errs, "php_version must match format X.Y")
	}

	if c.DBPassword == "" {
		errs = append(errs, "db_password is required")
	} else if len(c.DBPassword) < 8 {
		errs = append(errs, "db_password must be at least 8 characters")
	}

	if c.DBName == "" {
		errs = append(errs, "db_name is required")
	} else if !dbNameRe.MatchString(c.DBName) {
		errs = append(errs, "db_name must contain only alphanumeric characters and underscores")
	}

	if c.DBUser == "" {
		errs = append(errs, "db_user is required")
	} else if !dbUserRe.MatchString(c.DBUser) {
		errs = append(errs, "db_user must contain only alphanumeric characters and underscores")
	}

	if c.SiteUser == "" {
		errs = append(errs, "site_user is required")
	} else if !siteUserRe.MatchString(c.SiteUser) {
		errs = append(errs, "site_user format is invalid")
	}

	if c.SiteGroup == "" {
		errs = append(errs, "site_group is required")
	}

	if c.DBType != "" && c.DBType != "mysql" && c.DBType != "postgresql" {
		errs = append(errs, "db_type must be \"mysql\" or \"postgresql\"")
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}
