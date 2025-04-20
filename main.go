package main

import (
	"net/http"
)

// Create router and server
func createServer() *http.Server {
	router := http.NewServeMux()
	handler := http.FileServer(http.Dir("."))
	router.Handle("/app/", http.StripPrefix("/app/", handler))
	router.HandleFunc("/healthz", func(response http.ResponseWriter, r *http.Request) {
		response.WriteHeader(http.StatusOK)
		response.Write([]byte("OK"))
	})

	return &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
}

func main() {
	server := createServer()
	server.ListenAndServe()
}
