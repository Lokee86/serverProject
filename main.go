package main

import (
	"log"
	"net/http"
)

const pathRoot = "."
const port = ":8080"

type chirp struct {
	Body string `json:"body"`
}

// Create router and server
func createServer(apiCfg *apiConfig) *http.Server {
	router := http.NewServeMux()
	handler := http.StripPrefix("/app/", http.FileServer(http.Dir(pathRoot)))
	router.Handle("/app/", apiCfg.serverHitCounter(handler))
	router.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	router.HandleFunc("GET /api/healthz", healthCheck)
	router.HandleFunc("POST /admin/reset", apiCfg.resetCounter)
	router.HandleFunc("POST /api/validate_chirp", validateChirp)
	return &http.Server{
		Addr:    port,
		Handler: router,
	}
}

// EXECUTE MAIN FUNCTION
func main() {
	apiCfg := &apiConfig{}
	server := createServer(apiCfg)
	log.Printf("Server running on Port%v from %v", port, pathRoot)
	log.Fatal(server.ListenAndServe())
}
