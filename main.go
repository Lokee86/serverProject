package main

import (
	"net/http"
)

// Create router and server
func createServer() *http.Server {
	router := http.NewServeMux()
	handler := http.FileServer(http.Dir("."))
	router.Handle("/", handler)

	return &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
}

func main() {
	server := createServer()
	server.ListenAndServe()
}
