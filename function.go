package main

import (
	"log"
	"net/http"
)

func healthCheck(response http.ResponseWriter, r *http.Request) {
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("OK"))
	log.Println("Health check OK")
}
