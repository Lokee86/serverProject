package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func healthCheck(response http.ResponseWriter, r *http.Request) {
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("OK"))
	log.Println("Health check OK")
}

func jsonResponse(response http.ResponseWriter, code int, payload interface{}) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Internal Server Error: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("Status: %v JSON successfully sent", code)
	response.WriteHeader(code)
	response.Write(jsonData)
}

func errorResponse(response http.ResponseWriter, code int, mesg string) {
	type error struct {
		Error string `json:"error"`
	}

	jsonResponse(response, code, error{Error: mesg})
}
