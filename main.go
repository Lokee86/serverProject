package main

import (
	"net/http"
)

// Create router and server
func createServer() *http.Server {
	router := http.NewServeMux()
	router.Handle("/", rootFileHandler())

	return &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
}

// create file handler
func rootFileHandler() http.Handler {
	return http.FileServer(http.Dir("."))
}

func main() {
	server := createServer()
	server.ListenAndServe()
}
