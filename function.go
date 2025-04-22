package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func healthCheck(response http.ResponseWriter, r *http.Request) {
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("OK"))
	log.Println("Health check OK")
}

func validateChirp(response http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	checkedChirp := chirp{}
	err := decoder.Decode(&checkedChirp)
	if err != nil {
		log.Printf("Internal Server Error: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(checkedChirp.Body) > 140 {
		errorResponse(response, http.StatusBadRequest, "Chirp is too long")
		return
	}
	log.Println("Chirp validated")

	if len(checkedChirp.Body) <= 140 {
		checkProfanity(checkedChirp, response)
	}

}

func checkProfanity(chirp chirp, response http.ResponseWriter) {
	type cleanedChirp struct {
		CleanedBody string `json:"cleaned_body"`
	}

	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	listToCheck := strings.Split(chirp.Body, " ")
	for i, word := range listToCheck {
		for _, badWord := range profaneWords {
			if strings.ToLower(word) == badWord {
				listToCheck[i] = "****"
			}
		}
	}
	newChirpBody := strings.Join(listToCheck, " ")
	newChirp := cleanedChirp{}
	newChirp.CleanedBody = newChirpBody
	log.Println("Profanity cleaned from chirp")
	jsonResponse(response, http.StatusOK, newChirp)
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
