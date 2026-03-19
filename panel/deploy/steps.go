package deploy

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"laravel-deploy-panel/config"
)

type StepResult struct {
	Step   int    `json:"step"`
	Name   string `json:"name"`
	Status string `json:"status"` // "running", "success", "failed", "skipped"
	Output string `json:"output"`
}

type Runner struct {
	Cfg *config.Config
}

func NewRunner(cfg *config.Config) *Runner {
	return &Runner{Cfg: cfg}
}

func (r *Runner) run(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func (r *Runner) siteRoot() string {
	return config.DeriveSiteRoot(r.Cfg.Domain, r.Cfg.SiteUser)
}

func escapeSQLString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	return s
}

func (r *Runner) phpBin() string {
	return fmt.Sprintf("php%s", r.Cfg.PHPVersion)
}

// Step1: Create project directory
func (r *Runner) Step1CreateDirectory() StepResult {
	root := r.siteRoot()
	if err := os.MkdirAll(root, 0755); err != nil {
		return StepResult{Step: 1, Name: "Create Directory", Status: "failed", Output: err.Error()}
	}
	out, err := r.run("chown", fmt.Sprintf("%s:%s", r.Cfg.SiteUser, r.Cfg.SiteGroup), root)
	if err != nil {
		return StepResult{Step: 1, Name: "Create Directory", Status: "failed", Output: out}
	}
	return StepResult{Step: 1, Name: "Create Directory", Status: "success", Output: "Created " + root}
}

// Step2: Clone repository
func (r *Runner) Step2CloneRepo() StepResult {
	root := r.siteRoot()

	// If directory exists and has files, back it up
	entries, _ := os.ReadDir(root)
	if len(entries) > 0 {
		backup := root + ".backup-" + time.Now().Format("20060102-150405")
		if err := os.Rename(root, backup); err != nil {
			return StepResult{Step: 2, Name: "Clone Repository", Status: "failed", Output: "Failed to backup existing directory: " + err.Error()}
		}
		// Re-create the directory
		os.MkdirAll(root, 0755)
	}

	out, err := r.run("sudo", "-u", r.Cfg.SiteUser, "git", "clone",
		"--branch", r.Cfg.GithubBranch, r.Cfg.GithubRepo, root)
	if err != nil {
		return StepResult{Step: 2, Name: "Clone Repository", Status: "failed", Output: out}
	}
	return StepResult{Step: 2, Name: "Clone Repository", Status: "success", Output: out}
}

// Step3: Setup database
func (r *Runner) Step3SetupDatabase() StepResult {
	if r.Cfg.DBType == "postgresql" {
		return r.setupPostgresql()
	}
	return r.setupMysql()
}

func (r *Runner) setupMysql() StepResult {
	dbName := r.Cfg.DBName
	dbUser := escapeSQLString(r.Cfg.DBUser)
	dbPass := escapeSQLString(r.Cfg.DBPassword)
	cmds := []string{
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", dbName),
		fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'localhost' IDENTIFIED BY '%s';", dbUser, dbPass),
		fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO '%s'@'localhost';", dbName, dbUser),
		"FLUSH PRIVILEGES;",
	}
	sql := strings.Join(cmds, " ")
	out, err := r.run("mysql", "-u", "root", "-e", sql)
	if err != nil {
		return StepResult{Step: 3, Name: "Setup Database", Status: "failed", Output: out}
	}
	return StepResult{Step: 3, Name: "Setup Database", Status: "success", Output: "MySQL database and user created"}
}

func (r *Runner) setupPostgresql() StepResult {
	dbUser := r.Cfg.DBUser
	dbPass := strings.ReplaceAll(r.Cfg.DBPassword, "'", "''") // PostgreSQL uses '' to escape
	dbName := r.Cfg.DBName

	// Create user (ignore "already exists" errors)
	out, err := r.run("sudo", "-u", "postgres", "psql", "-c",
		fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s';", dbUser, dbPass))
	if err != nil && !strings.Contains(out, "already exists") {
		return StepResult{Step: 3, Name: "Setup Database", Status: "failed", Output: "Failed to create PostgreSQL user: " + out}
	}

	// Create database
	out, err = r.run("sudo", "-u", "postgres", "psql", "-c",
		fmt.Sprintf("CREATE DATABASE %s OWNER %s;", dbName, dbUser))
	if err != nil && !strings.Contains(out, "already exists") {
		return StepResult{Step: 3, Name: "Setup Database", Status: "failed", Output: "Failed to create PostgreSQL database: " + out}
	}

	// Grant privileges
	out, err = r.run("sudo", "-u", "postgres", "psql", "-c",
		fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s;", dbName, dbUser))
	if err != nil {
		return StepResult{Step: 3, Name: "Setup Database", Status: "failed", Output: "Failed to grant privileges: " + out}
	}

	return StepResult{Step: 3, Name: "Setup Database", Status: "success", Output: "PostgreSQL database and user created"}
}

// Step4: Configure .env
func (r *Runner) Step4ConfigureEnv() StepResult {
	root := r.siteRoot()
	envExample := root + "/.env.example"
	envFile := root + "/.env"

	if _, err := os.Stat(envExample); err == nil {
		if _, err := r.run("cp", envExample, envFile); err != nil {
			return StepResult{Step: 4, Name: "Configure .env", Status: "failed", Output: "Failed to copy .env.example"}
		}
	} else {
		if err := os.WriteFile(envFile, []byte(""), 0640); err != nil {
			return StepResult{Step: 4, Name: "Configure .env", Status: "failed", Output: err.Error()}
		}
	}

	subs := map[string]string{
		"APP_ENV=local":            "APP_ENV=production",
		"APP_DEBUG=true":           "APP_DEBUG=false",
		"APP_URL=http://localhost": fmt.Sprintf("APP_URL=https://%s", r.Cfg.Domain),
		"DB_DATABASE=laravel":      fmt.Sprintf("DB_DATABASE=%s", r.Cfg.DBName),
		"DB_USERNAME=root":         fmt.Sprintf("DB_USERNAME=%s", r.Cfg.DBUser),
		"DB_PASSWORD=":             fmt.Sprintf("DB_PASSWORD=%s", r.Cfg.DBPassword),
		"QUEUE_CONNECTION=sync":    "QUEUE_CONNECTION=database",
		"SESSION_DRIVER=database":  "SESSION_DRIVER=file",
	}

	if r.Cfg.DBType == "postgresql" {
		subs["DB_CONNECTION=mysql"] = "DB_CONNECTION=pgsql"
		subs["DB_PORT=3306"] = "DB_PORT=5432"
	}

	for old, new := range subs {
		r.run("sed", "-i", fmt.Sprintf("s|%s|%s|g", old, new), envFile)
	}

	return StepResult{Step: 4, Name: "Configure .env", Status: "success", Output: ".env configured"}
}

// Step5: Install dependencies
func (r *Runner) Step5InstallDependencies() StepResult {
	root := r.siteRoot()
	php := r.phpBin()

	steps := []struct {
		cmd  string
		args []string
	}{
		{"sudo", []string{"-u", r.Cfg.SiteUser, "composer", "install", "--no-dev", "--optimize-autoloader", "--working-dir=" + root}},
		{php, []string{root + "/artisan", "key:generate", "--force"}},
		{php, []string{root + "/artisan", "migrate", "--force"}},
		{php, []string{root + "/artisan", "storage:link"}},
		{php, []string{root + "/artisan", "config:cache"}},
		{php, []string{root + "/artisan", "route:cache"}},
		{php, []string{root + "/artisan", "view:cache"}},
	}

	var outputs []string
	for _, s := range steps {
		out, err := r.run(s.cmd, s.args...)
		outputs = append(outputs, out)
		if err != nil {
			return StepResult{Step: 5, Name: "Install Dependencies", Status: "failed", Output: strings.Join(outputs, "\n")}
		}
	}
	return StepResult{Step: 5, Name: "Install Dependencies", Status: "success", Output: strings.Join(outputs, "\n")}
}

// Step6: Set permissions
func (r *Runner) Step6SetPermissions() StepResult {
	root := r.siteRoot()
	own := fmt.Sprintf("%s:%s", r.Cfg.SiteUser, r.Cfg.SiteGroup)
	cmds := [][]string{
		{"chown", "-R", own, root},
		{"chmod", "-R", "755", root},
		{"chmod", "-R", "ug+rwx", root + "/storage"},
		{"chmod", "-R", "ug+rwx", root + "/bootstrap/cache"},
	}
	for _, c := range cmds {
		if out, err := r.run(c[0], c[1:]...); err != nil {
			return StepResult{Step: 6, Name: "Set Permissions", Status: "failed", Output: out}
		}
	}
	return StepResult{Step: 6, Name: "Set Permissions", Status: "success", Output: "Permissions set"}
}

// Step7: Configure Nginx
func (r *Runner) Step7ConfigureNginx() StepResult {
	root := r.siteRoot()
	php := r.Cfg.PHPVersion
	conf := fmt.Sprintf(`server {
    listen 80;
    server_name %s www.%s;
    root %s/public;
    index index.php index.html;

    add_header X-Frame-Options "SAMEORIGIN";
    add_header X-XSS-Protection "1; mode=block";
    add_header X-Content-Type-Options "nosniff";
    add_header Referrer-Policy "strict-origin-when-cross-origin";
    add_header Permissions-Policy "camera=(), microphone=(), geolocation=()";

    charset utf-8;
    client_max_body_size 100M;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        fastcgi_pass unix:/var/run/php/php%s-fpm.sock;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        include fastcgi_params;
    }

    location ~ /\.(?!well-known).* {
        deny all;
    }
}`, r.Cfg.Domain, r.Cfg.Domain, root, php)

	confPath := fmt.Sprintf("/etc/nginx/sites-available/%s", r.Cfg.Domain)
	if err := os.WriteFile(confPath, []byte(conf), 0644); err != nil {
		return StepResult{Step: 7, Name: "Configure Nginx", Status: "failed", Output: err.Error()}
	}

	linkPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", r.Cfg.Domain)
	os.Remove(linkPath)
	if err := os.Symlink(confPath, linkPath); err != nil {
		return StepResult{Step: 7, Name: "Configure Nginx", Status: "failed", Output: err.Error()}
	}

	if out, err := r.run("nginx", "-t"); err != nil {
		return StepResult{Step: 7, Name: "Configure Nginx", Status: "failed", Output: out}
	}
	if out, err := r.run("systemctl", "reload", "nginx"); err != nil {
		return StepResult{Step: 7, Name: "Configure Nginx", Status: "failed", Output: out}
	}
	return StepResult{Step: 7, Name: "Configure Nginx", Status: "success", Output: "Nginx configured and reloaded"}
}

// Step8: Install SSL (skipped if DNS not confirmed)
func (r *Runner) Step8InstallSSL() StepResult {
	if !r.Cfg.DNSConfirmed {
		return StepResult{Step: 8, Name: "Install SSL", Status: "skipped", Output: "DNS not confirmed — SSL skipped"}
	}

	// Verify DNS resolves before running certbot
	ips, err := net.LookupHost(r.Cfg.Domain)
	if err != nil || len(ips) == 0 {
		return StepResult{Step: 8, Name: "Install SSL", Status: "failed",
			Output: fmt.Sprintf("DNS lookup failed for %s — point your domain to this server first", r.Cfg.Domain)}
	}

	out, err := r.run("certbot", "--nginx", "--redirect", "--non-interactive",
		"--agree-tos", "-m", fmt.Sprintf("admin@%s", r.Cfg.Domain),
		"-d", r.Cfg.Domain, "-d", fmt.Sprintf("www.%s", r.Cfg.Domain))
	if err != nil {
		return StepResult{Step: 8, Name: "Install SSL", Status: "failed", Output: out}
	}
	return StepResult{Step: 8, Name: "Install SSL", Status: "success", Output: out}
}

// Step9: Setup queue worker
func (r *Runner) Step9SetupQueueWorker() StepResult {
	if !r.Cfg.EnableQueueWorker {
		return StepResult{Step: 9, Name: "Setup Queue Worker", Status: "skipped", Output: "Queue worker disabled"}
	}
	workerName := config.DeriveSupervisorName(r.Cfg.Domain)
	root := r.siteRoot()
	php := r.phpBin()
	conf := fmt.Sprintf(`[program:%s]
process_name=%%(program_name)s_%%(process_num)02d
command=%s %s/artisan queue:work --sleep=3 --tries=3 --timeout=120
autostart=true
autorestart=true
stopasgroup=true
killasgroup=true
user=%s
numprocs=1
redirect_stderr=true
stdout_logfile=%s/storage/logs/worker.log
stopwaitsecs=3600
`, workerName, php, root, r.Cfg.SiteUser, root)

	confPath := fmt.Sprintf("/etc/supervisor/conf.d/%s.conf", workerName)
	if err := os.WriteFile(confPath, []byte(conf), 0644); err != nil {
		return StepResult{Step: 9, Name: "Setup Queue Worker", Status: "failed", Output: err.Error()}
	}
	r.run("supervisorctl", "reread")
	r.run("supervisorctl", "update")
	out, err := r.run("supervisorctl", "start", workerName)
	if err != nil {
		return StepResult{Step: 9, Name: "Setup Queue Worker", Status: "failed", Output: out}
	}
	return StepResult{Step: 9, Name: "Setup Queue Worker", Status: "success", Output: "Queue worker started"}
}

// Step10: Setup scheduler cron
func (r *Runner) Step10SetupScheduler() StepResult {
	if !r.Cfg.EnableScheduler {
		return StepResult{Step: 10, Name: "Setup Scheduler", Status: "skipped", Output: "Scheduler disabled"}
	}
	root := r.siteRoot()
	php := r.phpBin()
	cronLine := fmt.Sprintf("* * * * * %s %s/artisan schedule:run >> /dev/null 2>&1", php, root)

	// Check if already exists
	existing, _ := exec.Command("crontab", "-u", r.Cfg.SiteUser, "-l").Output()
	if strings.Contains(string(existing), "artisan schedule:run") {
		return StepResult{Step: 10, Name: "Setup Scheduler", Status: "success", Output: "Cron already exists"}
	}

	newCron := strings.TrimSpace(string(existing)) + "\n" + cronLine + "\n"
	cmd := exec.Command("crontab", "-u", r.Cfg.SiteUser, "-")
	cmd.Stdin = strings.NewReader(newCron)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return StepResult{Step: 10, Name: "Setup Scheduler", Status: "failed", Output: string(out)}
	}
	return StepResult{Step: 10, Name: "Setup Scheduler", Status: "success", Output: "Scheduler cron added"}
}

// HealthCheck pings the deployed site to verify it responds
func (r *Runner) HealthCheck() StepResult {
	// Use localhost with Host header to avoid DNS dependency
	out, err := r.run("curl", "-sS", "-o", "/dev/null", "-w", "%{http_code}",
		"--max-time", "10",
		"-H", fmt.Sprintf("Host: %s", r.Cfg.Domain),
		"http://127.0.0.1")
	if err != nil {
		return StepResult{Step: 11, Name: "Health Check", Status: "failed", Output: "Site unreachable: " + out}
	}
	if out == "200" || out == "302" || out == "301" {
		return StepResult{Step: 11, Name: "Health Check", Status: "success", Output: fmt.Sprintf("Site responded with HTTP %s", out)}
	}
	return StepResult{Step: 11, Name: "Health Check", Status: "failed", Output: fmt.Sprintf("Site responded with HTTP %s — expected 200, 301, or 302", out)}
}

// RunAll runs all 10 steps, calling progress callback after each
func (r *Runner) RunAll(progress func(StepResult)) {
	steps := []func() StepResult{
		r.Step1CreateDirectory,
		r.Step2CloneRepo,
		r.Step3SetupDatabase,
		r.Step4ConfigureEnv,
		r.Step5InstallDependencies,
		r.Step6SetPermissions,
		r.Step7ConfigureNginx,
		r.Step8InstallSSL,
		r.Step9SetupQueueWorker,
		r.Step10SetupScheduler,
	}
	for _, step := range steps {
		result := step()
		progress(result)
		if result.Status == "failed" {
			return
		}
	}
	// Post-deploy health check (non-blocking — doesn't stop on failure)
	progress(r.HealthCheck())
}

// RunStep runs a single step by number (1-11)
func (r *Runner) RunStep(n int) StepResult {
	steps := map[int]func() StepResult{
		1:  r.Step1CreateDirectory,
		2:  r.Step2CloneRepo,
		3:  r.Step3SetupDatabase,
		4:  r.Step4ConfigureEnv,
		5:  r.Step5InstallDependencies,
		6:  r.Step6SetPermissions,
		7:  r.Step7ConfigureNginx,
		8:  r.Step8InstallSSL,
		9:  r.Step9SetupQueueWorker,
		10: r.Step10SetupScheduler,
		11: r.HealthCheck,
	}
	fn, ok := steps[n]
	if !ok {
		return StepResult{Step: n, Status: "failed", Output: "unknown step"}
	}
	return fn()
}
