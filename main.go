package main

import (
	"log"
	"net/http"
	"sync/atomic"
)

const pathRoot = "."
const port = ":8080"

type apiConfig struct {
	fileServerHits atomic.Int32
}

// increment hit counter
func (a *apiConfig) serverHitCounter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// metrics handler
func (a *apiConfig) metricsHandler(response http.ResponseWriter, r *http.Request) {
	log.Printf("Hits: %v", a.fileServerHits.Load())
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("metrics"))
}

// reset counter
func (a *apiConfig) resetCounter(response http.ResponseWriter, r *http.Request) {
	a.fileServerHits.Store(0)
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("Hits coutner reset to 0"))
	log.Println("Hit counter reset to 0")

}

// Create router and server
func createServer(apiCfg *apiConfig) *http.Server {
	router := http.NewServeMux()
	handler := http.StripPrefix("/app/", http.FileServer(http.Dir(pathRoot)))
	router.Handle("/app/", apiCfg.serverHitCounter(handler))
	router.HandleFunc("/metrics", apiCfg.metricsHandler)
	router.HandleFunc("/reset", apiCfg.resetCounter)
	router.HandleFunc("/healthz", func(response http.ResponseWriter, r *http.Request) {
		response.WriteHeader(http.StatusOK)
		response.Write([]byte("OK"))
	})

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
