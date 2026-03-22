package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"laravel-deploy-panel/config"
)

type ServiceStatus struct {
	Name    string `json:"name"`
	Running bool   `json:"running"`
}

type SSLStatus struct {
	ExpiryDate string `json:"expiry_date"`
	DaysLeft   int    `json:"days_left"`
	Valid      bool   `json:"valid"`
}

func isServiceActive(name string) bool {
	err := exec.Command("systemctl", "is-active", "--quiet", name).Run()
	return err == nil
}

func getSSLStatus(domain string) SSLStatus {
	if domain == "" {
		return SSLStatus{}
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", domain+":443", &tls.Config{InsecureSkipVerify: false})
	if err != nil {
		return SSLStatus{Valid: false, ExpiryDate: "unreachable"}
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return SSLStatus{Valid: false}
	}
	expiry := certs[0].NotAfter
	daysLeft := int(time.Until(expiry).Hours() / 24)
	return SSLStatus{
		Valid:      true,
		ExpiryDate: expiry.Format("2006-01-02"),
		DaysLeft:   daysLeft,
	}
}

func getQueueWorkerStatus(domain string) ServiceStatus {
	supervisorName := config.DeriveSupervisorName(domain)
	out, err := exec.Command("supervisorctl", "status", supervisorName).Output()
	if err != nil {
		return ServiceStatus{Name: supervisorName, Running: false}
	}
	running := strings.Contains(string(out), "RUNNING")
	return ServiceStatus{Name: supervisorName, Running: running}
}

func getDBServiceStatus(cfg *config.Config) ServiceStatus {
	if cfg.DBType == "postgresql" {
		// Try common PostgreSQL service names
		for _, name := range []string{"postgresql", "postgres"} {
			if isServiceActive(name) {
				return ServiceStatus{Name: name, Running: true}
			}
		}
		return ServiceStatus{Name: "postgresql", Running: false}
	}
	// MySQL/MariaDB
	for _, name := range []string{"mysql", "mariadb"} {
		if isServiceActive(name) {
			return ServiceStatus{Name: name, Running: true}
		}
	}
	return ServiceStatus{Name: "mysql", Running: false}
}

func getServerIP() string {
	out, err := exec.Command("hostname", "-I").Output()
	if err != nil {
		return "unknown"
	}
	parts := strings.Fields(string(out))
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

type StatusResponse struct {
	Nginx       ServiceStatus `json:"nginx"`
	PHPFPM      ServiceStatus `json:"php_fpm"`
	MySQL       ServiceStatus `json:"mysql"`
	Supervisor  ServiceStatus `json:"supervisor"`
	SSL         SSLStatus     `json:"ssl"`
	QueueWorker ServiceStatus `json:"queue_worker"`
	ServerIP    string        `json:"server_ip"`
}

func handleGetStatus(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(configPath)
	if err != nil {
		http.Error(w, "failed to load config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	phpVersion := cfg.PHPVersion
	if phpVersion == "" {
		phpVersion = "8.3"
	}
	phpService := fmt.Sprintf("php%s-fpm", phpVersion)

	resp := StatusResponse{
		Nginx:       ServiceStatus{Name: "nginx", Running: isServiceActive("nginx")},
		PHPFPM:      ServiceStatus{Name: phpService, Running: isServiceActive(phpService)},
		MySQL:       getDBServiceStatus(cfg),
		Supervisor:  ServiceStatus{Name: "supervisor", Running: isServiceActive("supervisor")},
		SSL:         getSSLStatus(cfg.Domain),
		QueueWorker: getQueueWorkerStatus(cfg.Domain),
		ServerIP:    getServerIP(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
