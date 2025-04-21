package main

import (
	"log"
	"net/http"
)

const pathRoot = "."
const port = ":8080"

// Create router and server
func createServer(apiCfg *apiConfig) *http.Server {
	router := http.NewServeMux()
	handler := http.StripPrefix("/app/", http.FileServer(http.Dir(pathRoot)))
	router.Handle("/app/", apiCfg.serverHitCounter(handler))
	router.HandleFunc("/metrics", apiCfg.metricsHandler)
	router.HandleFunc("/reset", apiCfg.resetCounter)
	router.HandleFunc("/healthz", healthCheck)
	return &http.Server{
		Addr:    port,
		Handler: router,
	}
}

func main() {
	apiCfg := &apiConfig{}
	server := createServer(apiCfg)
	log.Printf("Server running on Port%v from %v", port, pathRoot)
	log.Fatal(server.ListenAndServe())
}
