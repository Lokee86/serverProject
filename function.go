package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Lokee86/serverProject/internal/database"
	"github.com/google/uuid"
)

// readiness end point response - 200 OK
func healthCheck(response http.ResponseWriter, r *http.Request) {
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("OK"))
	log.Println("Health check OK")
}

// check and filter profanity, no return, modifies at memory address
func checkProfanity(chirp *string) {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	listToCheck := strings.Split(*chirp, " ")
	profanity := false
	for i, word := range listToCheck {
		for _, badWord := range profaneWords {
			if strings.ToLower(word) == badWord {
				listToCheck[i] = "****"
				profanity = true
			}
		}
	}
	*chirp = strings.Join(listToCheck, " ")
	if profanity {
		log.Println("Profanity cleaned from chirp")
	}
}

// send 200 range code response to client with json payload
func jsonResponse(response http.ResponseWriter, code int, payload interface{}, mesg string) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Internal Server Error: %v", err)
		response.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("Status: %v JSON successfully sent. %v", code, mesg)
	response.WriteHeader(code)
	response.Write(jsonData)
}

// send error code and message to client
func errorResponse(response http.ResponseWriter, code int, mesg string) {
	type error struct {
		Error string `json:"error"`
	}

	jsonResponse(response, code, error{Error: mesg}, mesg)
}

// internal server error, server error and log event
func internalError(response http.ResponseWriter, err error) {
	log.Printf("Internal Server Error: %v", err)
	response.Write([]byte(err.Error()))
}

// Parse generated Chirp struct into local json controlled Chirp struct
func jsonSafeChirp(chirp database.Chirp) Chirp {
	jsonSafeChirp := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	return jsonSafeChirp
}

// Parse generated User struct into local json controled User struct - no HashedPassword field transferred
func jsonReturnUser(user database.User) User {
	newUserJson := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	return newUserJson
}

// parse ID string from URL
func extractIDString(response http.ResponseWriter, path string) uuid.UUID {
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[3] == "" {
		errorResponse(response, http.StatusBadRequest, "chirp ID missing")
		return uuid.Nil
	}
	idStr, err := uuid.Parse(parts[3])
	if err != nil {
		internalError(response, err)
		return uuid.Nil
	}
	return idStr
}

// return a 204 No Content response
func noContentResponse(response http.ResponseWriter, mesg string) {
	response.WriteHeader(http.StatusNoContent)
	log.Printf("Status: %v Reponse successfuly sent. %v", http.StatusNoContent, mesg)
}
