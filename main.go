package main

import (
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

// API hit counter increments
func (apiCfg *apiConfig) incrementCoutner(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCfg.fileserverHits.Add(1)
		handler.ServeHTTP(w, r)
	})
}

// Create router and server
func createServer(apiCfg *apiConfig) *http.Server {
	router := http.NewServeMux()
	handler := apiCfg.incrementCoutner(http.FileServer(http.Dir(".")))
	router.Handle("/app/", http.StripPrefix("/app/", handler))
	router.HandleFunc("/healthz", func(response http.ResponseWriter, r *http.Request) {
		response.WriteHeader(http.StatusOK)
		response.Write([]byte("OK"))
	})
	router.HandleFunc("/metrics", func(response http.ResponseWriter, r *http.Request) {
		response.WriteHeader(http.StatusOK)
		response.Write([]byte("OK"))
	})

	return &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
}

func main() {
	apiCfg := &apiConfig{}
	server := createServer(apiCfg)
	server.ListenAndServe()
}
