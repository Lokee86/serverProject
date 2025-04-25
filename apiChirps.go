package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Lokee86/serverProject/internal/auth"
	"github.com/Lokee86/serverProject/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"create_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type handleChirp struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

// fetches all chirps from table 'chirps' in database
func (a *apiConfig) fetchChirps(response http.ResponseWriter, r *http.Request) {
	chirps, err := a.databaseQueries.GetAllChirps(r.Context())
	if err != nil {
		internalError(response, err)
		return
	}
	var jsonSafeChirps []Chirp
	for _, chirp := range chirps {
		jsonSafeChirp := jsonSafeChirp(chirp)
		jsonSafeChirps = append(jsonSafeChirps, jsonSafeChirp)
	}
	jsonResponse(response, http.StatusOK, jsonSafeChirps)
}

// fetches a single chirp by id from table 'chirps' in database
func (a *apiConfig) fetchSingleChirp(response http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[3] == "" {
		errorResponse(response, http.StatusBadRequest, "chirp ID missing")
		return
	}
	idStr, err := uuid.Parse(parts[3])
	if err != nil {
		internalError(response, err)
		return
	}

	chirp, err := a.databaseQueries.SelectSingleChirp(r.Context(), idStr)
	if err != nil {
		internalError(response, err)
		return
	}
	jsonSafeChirp := jsonSafeChirp(chirp)
	jsonResponse(response, http.StatusOK, jsonSafeChirp)
}

// validates length of submitted chirp
func (a *apiConfig) validateChirp(response http.ResponseWriter, r *http.Request) {

	checkedChirp := handleChirp{}

	userToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		internalError(response, err)
		return
	}
	checkedChirp.UserID, err = auth.ValidateJWT(userToken)
	if err != nil {
		errorResponse(response, http.StatusUnauthorized, "Unauthorized")
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&checkedChirp)
	if err != nil {
		internalError(response, err)
		return
	}

	if len(checkedChirp.Body) > 140 {
		errorResponse(response, http.StatusBadRequest, "Chirp is too long")
		return
	}
	log.Println("Chirp validated")
	checkProfanity(&checkedChirp.Body)
	a.addChirp(response, checkedChirp, r)
}

// adds chirp to table 'chirps' in database
func (a *apiConfig) addChirp(response http.ResponseWriter, checkedChirp handleChirp, r *http.Request) {
	compatibleChirp := database.CreateChirpParams{
		Body:   checkedChirp.Body,
		UserID: checkedChirp.UserID,
	}
	chirp, err := a.databaseQueries.CreateChirp(r.Context(), compatibleChirp)
	if err != nil {
		internalError(response, err)
		log.Println("Database Insertion Error")
		return
	}
	jsonSafeChirp := jsonSafeChirp(chirp)
	jsonResponse(response, http.StatusCreated, jsonSafeChirp)
}
