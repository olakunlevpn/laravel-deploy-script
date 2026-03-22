package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"laravel-deploy-panel/config"
)

type ServerInfo struct {
	// System
	Hostname    string `json:"hostname"`
	OS          string `json:"os"`
	Kernel      string `json:"kernel"`
	Uptime      string `json:"uptime"`
	CPUCores    int    `json:"cpu_cores"`
	LoadAverage string `json:"load_average"`

	// Memory
	MemoryTotal string `json:"memory_total"`
	MemoryUsed  string `json:"memory_used"`
	MemoryFree  string `json:"memory_free"`
	MemoryPct   string `json:"memory_pct"`

	// Disk
	DiskTotal string `json:"disk_total"`
	DiskUsed  string `json:"disk_used"`
	DiskFree  string `json:"disk_free"`
	DiskPct   string `json:"disk_pct"`

	// Software versions
	PHPVersion      string `json:"php_version"`
	NginxVersion    string `json:"nginx_version"`
	DBVersion       string `json:"db_version"`
	ComposerVersion string `json:"composer_version"`

	// Config info
	ServerIP    string `json:"server_ip"`
	Domain      string `json:"domain"`
	SiteRoot    string `json:"site_root"`
	SiteUser    string `json:"site_user"`
	DBType      string `json:"db_type"`
	DBName      string `json:"db_name"`
	DBUser      string `json:"db_user"`
	GithubRepo  string `json:"github_repo"`
	GithubBranch string `json:"github_branch"`
	PHPConfig   string `json:"php_config"`
	QueueWorker bool   `json:"queue_worker"`
	Scheduler   bool   `json:"scheduler"`
}

func cmdOutput(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return cmdOutput("hostname")
	}
	return h
}

func getOSInfo() string {
	// Try /etc/os-release first
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				name := strings.TrimPrefix(line, "PRETTY_NAME=")
				name = strings.Trim(name, "\"")
				return name
			}
		}
	}
	return cmdOutput("uname", "-s", "-r")
}

func getKernel() string {
	return cmdOutput("uname", "-r")
}

func getUptime() string {
	out := cmdOutput("uptime", "-p")
	if out != "" {
		return strings.TrimPrefix(out, "up ")
	}
	return cmdOutput("uptime")
}

func getLoadAverage() string {
	out := cmdOutput("cat", "/proc/loadavg")
	if out != "" {
		parts := strings.Fields(out)
		if len(parts) >= 3 {
			return fmt.Sprintf("%s, %s, %s", parts[0], parts[1], parts[2])
		}
	}
	return ""
}

func getMemoryInfo() (total, used, free, pct string) {
	out := cmdOutput("free", "-h", "--si")
	if out == "" {
		return "—", "—", "—", "—"
	}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Mem:") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				total = fields[1]
				used = fields[2]
				free = fields[3]
			}
			break
		}
	}
	// Calculate percentage from /proc/meminfo for accuracy
	data, err := os.ReadFile("/proc/meminfo")
	if err == nil {
		var memTotal, memAvail int64
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "MemTotal:") {
				fmt.Sscanf(strings.TrimPrefix(line, "MemTotal:"), "%d", &memTotal)
			}
			if strings.HasPrefix(line, "MemAvailable:") {
				fmt.Sscanf(strings.TrimPrefix(line, "MemAvailable:"), "%d", &memAvail)
			}
		}
		if memTotal > 0 {
			usedPct := float64(memTotal-memAvail) / float64(memTotal) * 100
			pct = fmt.Sprintf("%.0f%%", usedPct)
		}
	}
	if pct == "" {
		pct = "—"
	}
	return
}

func getDiskInfo() (total, used, free, pct string) {
	out := cmdOutput("df", "-h", "/")
	if out == "" {
		return "—", "—", "—", "—"
	}
	lines := strings.Split(out, "\n")
	if len(lines) >= 2 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 5 {
			total = fields[1]
			used = fields[2]
			free = fields[3]
			pct = fields[4]
		}
	}
	return
}

func getVersionCmd(name string, args ...string) string {
	out := cmdOutput(name, args...)
	if out == "" {
		return "Not installed"
	}
	// Take first line only
	lines := strings.Split(out, "\n")
	return lines[0]
}

func handleGetServerInfo(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(configPath)
	if err != nil {
		http.Error(w, "failed to load config", http.StatusInternalServerError)
		return
	}

	phpVersion := cfg.PHPVersion
	if phpVersion == "" {
		phpVersion = "8.3"
	}
	phpBin := fmt.Sprintf("php%s", phpVersion)

	memTotal, memUsed, memFree, memPct := getMemoryInfo()
	diskTotal, diskUsed, diskFree, diskPct := getDiskInfo()

	// Get DB version
	var dbVersion string
	if cfg.DBType == "postgresql" {
		dbVersion = getVersionCmd("psql", "--version")
	} else {
		dbVersion = getVersionCmd("mysql", "--version")
	}

	info := ServerInfo{
		Hostname:    getHostname(),
		OS:          getOSInfo(),
		Kernel:      getKernel(),
		Uptime:      getUptime(),
		CPUCores:    runtime.NumCPU(),
		LoadAverage: getLoadAverage(),

		MemoryTotal: memTotal,
		MemoryUsed:  memUsed,
		MemoryFree:  memFree,
		MemoryPct:   memPct,

		DiskTotal: diskTotal,
		DiskUsed:  diskUsed,
		DiskFree:  diskFree,
		DiskPct:   diskPct,

		PHPVersion:      getVersionCmd(phpBin, "-v"),
		NginxVersion:    getVersionCmd("nginx", "-v"),
		DBVersion:       dbVersion,
		ComposerVersion: getVersionCmd("composer", "--version"),

		ServerIP:     getServerIP(),
		Domain:       cfg.Domain,
		SiteRoot:     config.DeriveSiteRoot(cfg.Domain, cfg.SiteUser),
		SiteUser:     cfg.SiteUser,
		DBType:       cfg.DBType,
		DBName:       cfg.DBName,
		DBUser:       cfg.DBUser,
		GithubRepo:   cfg.GithubRepo,
		GithubBranch: cfg.GithubBranch,
		PHPConfig:    phpVersion,
		QueueWorker:  cfg.EnableQueueWorker,
		Scheduler:    cfg.EnableScheduler,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
