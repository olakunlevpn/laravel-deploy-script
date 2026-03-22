package api

import "net/http"

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/config", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetConfig(w, r)
		case http.MethodPost:
			handlePostConfig(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/status", corsMiddleware(handleGetStatus))
	mux.HandleFunc("/api/server-info", corsMiddleware(handleGetServerInfo))

	mux.HandleFunc("/api/actions/nginx/", corsMiddleware(handleNginxAction))
	mux.HandleFunc("/api/actions/supervisor/", corsMiddleware(handleSupervisorAction))
	mux.HandleFunc("/api/actions/queue-worker/", corsMiddleware(handleQueueWorkerAction))
	mux.HandleFunc("/api/actions/ssl/renew", corsMiddleware(handleSSLRenew))
	mux.HandleFunc("/api/actions/permissions", corsMiddleware(handlePermissions))
	mux.HandleFunc("/api/actions/laravel/", corsMiddleware(handleLaravelAction))

	mux.HandleFunc("/api/logs/laravel", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetLogs(w, r)
		case http.MethodPost:
			handleClearLogs(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/logs/nginx-access", corsMiddleware(handleGetNginxAccessLogs))
	mux.HandleFunc("/api/logs/nginx-error", corsMiddleware(handleGetNginxErrorLogs))

	mux.HandleFunc("/api/env", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetEnv(w, r)
		case http.MethodPost:
			handlePostEnv(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/api/deploy/stream", handleDeployStream)
	mux.HandleFunc("/api/deploy/status", corsMiddleware(handleDeployStatus))
	mux.HandleFunc("/api/deploy/step/", corsMiddleware(handleDeployStep))
	mux.HandleFunc("/api/deploy/preflight", corsMiddleware(handleDeployPreflight))
	mux.HandleFunc("/api/webhook/deploy", corsMiddleware(handleWebhookDeploy))
}
