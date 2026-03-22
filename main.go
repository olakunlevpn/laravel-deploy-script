package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path"

	"laravel-deploy-panel/api"
)

func main() {
	port := flag.Int("port", 4432, "Port to listen on")
	flag.Parse()

	mux := http.NewServeMux()

	// Register API routes
	api.RegisterRoutes(mux)

	// Serve embedded React SPA for all non-API routes
	distFS, err := fs.Sub(frontendDist, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}
	fileServer := http.FileServer(http.FS(distFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve static file; fall back to index.html for SPA routing
		cleaned := path.Clean("/" + r.URL.Path)
		_, err := distFS.Open(cleaned[1:])
		if err != nil {
			// Serve index.html for unknown paths (client-side routing)
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Laravel Deploy Panel running on http://0.0.0.0%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
