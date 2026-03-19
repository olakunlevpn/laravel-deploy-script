package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeriveDBName(t *testing.T) {
	cases := []struct {
		domain string
		want   string
	}{
		{"myapp.com", "myapp"},
		{"my-app.com", "my_app"},
		{"sub.myapp.com", "sub_myapp"},
		{"my--app.com", "my_app"},
		{"myapp.org", "myapp_org"},
	}
	for _, c := range cases {
		got := DeriveDBName(c.domain)
		if got != c.want {
			t.Errorf("DeriveDBName(%q) = %q, want %q", c.domain, got, c.want)
		}
	}
}

func TestDeriveDBUser(t *testing.T) {
	if got := DeriveDBUser("myapp"); got != "myapp_user" {
		t.Errorf("got %q, want %q", got, "myapp_user")
	}
}

func TestDeriveSiteRoot(t *testing.T) {
	if got := DeriveSiteRoot("myapp.com", "forge"); got != "/home/forge/myapp.com" {
		t.Errorf("got %q", got)
	}
}

func TestDeriveSupervisorName(t *testing.T) {
	if got := DeriveSupervisorName("myapp.com"); got != "myapp-com-worker" {
		t.Errorf("got %q", got)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{
		Domain:            "myapp.com",
		GithubRepo:        "https://github.com/user/repo",
		GithubBranch:      "main",
		DBPassword:        "secret",
		DBName:            "myapp",
		DBUser:            "myapp_user",
		PHPVersion:        "8.3",
		SiteUser:          "forge",
		SiteGroup:         "www-data",
		EnableQueueWorker: true,
		EnableScheduler:   true,
		DNSConfirmed:      false,
	}

	if err := Save(cfg, path); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if loaded.Domain != cfg.Domain {
		t.Errorf("Domain mismatch: got %q, want %q", loaded.Domain, cfg.Domain)
	}
	if loaded.DBPassword != cfg.DBPassword {
		t.Errorf("DBPassword mismatch")
	}

	_ = os.Remove(path)
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/tmp/nonexistent-panel-config.json")
	if err != nil {
		t.Fatal("expected no error for missing file, got:", err)
	}
	// Should return defaults
	if cfg.PHPVersion != "8.3" {
		t.Errorf("expected default PHPVersion 8.3, got %q", cfg.PHPVersion)
	}
	if cfg.GithubBranch != "main" {
		t.Errorf("expected default GithubBranch main, got %q", cfg.GithubBranch)
	}
}
