package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
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
	idQuery := r.URL.Query().Get("author_id")
	sortQuery := r.URL.Query().Get("sort")
	var chirps []database.Chirp
	var err error
	var uuidQuery uuid.UUID
	var jsonSafeChirps []Chirp

	if idQuery != "" {
		uuidQuery, err = uuid.Parse(idQuery)
		if err != nil {
			internalError(response, err)
			return
		}

		chirps, err = a.databaseQueries.GetChirpsByID(r.Context(), uuidQuery)

		if err == sql.ErrNoRows {
			errorResponse(response, http.StatusNotFound, "Not Found: No chirps from that ID")
			return
		} else if err != nil {
			internalError(response, err)
			return
		}
		for _, chirp := range chirps {
			jsonSafeChirp := jsonSafeChirp(chirp)
			jsonSafeChirps = append(jsonSafeChirps, jsonSafeChirp)
		}
		jsonResponse(response, http.StatusOK, jsonSafeChirps, "Fetched chirps from provided ID")
	} else {
		if err == sql.ErrNoRows {
			errorResponse(response, http.StatusNotFound, "Not Found: No chirps from that ID")
			return
		} else if err != nil {
			internalError(response, err)
			return
		}
		for _, chirp := range chirps {
			jsonSafeChirp := jsonSafeChirp(chirp)
			jsonSafeChirps = append(jsonSafeChirps, jsonSafeChirp)
		}
		jsonResponse(response, http.StatusOK, jsonSafeChirps, "All chirps fetched")
	}

}

// fetches a single chirp by id from table 'chirps' in database
func (a *apiConfig) fetchSingleChirp(response http.ResponseWriter, r *http.Request) {
	idStr := extractIDString(response, r.URL.Path)

	chirp, err := a.databaseQueries.SelectSingleChirp(r.Context(), idStr)
	if err == sql.ErrNoRows {
		errorResponse(response, http.StatusNotFound, "Chirp not found")
		return
	} else if err != nil {
		internalError(response, err)
		return
	}
	jsonSafeChirp := jsonSafeChirp(chirp)
	jsonResponse(response, http.StatusOK, jsonSafeChirp, "Single chirp query successful")
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
		errorResponse(response, http.StatusUnauthorized, "Unauthorized: Invalid access token")
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
	jsonResponse(response, http.StatusCreated, jsonSafeChirp, "Chirp added successfully")
}

// remove chirp from database if user is autorized
func (a *apiConfig) deleteChirp(response http.ResponseWriter, r *http.Request) {
	userToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorResponse(response, http.StatusUnauthorized, "Unauthorized: Invalid access token")
		return
	}
	userID, err := auth.ValidateJWT(userToken)
	if err != nil {
		errorResponse(response, http.StatusUnauthorized, "Unauthorized: Invalid access token")
		return
	}
	idStr := extractIDString(response, r.URL.Path)
	chirp, err := a.databaseQueries.SelectSingleChirp(r.Context(), idStr)
	if err != nil {
		internalError(response, err)
		return
	}
	if userID != chirp.UserID {
		errorResponse(response, http.StatusForbidden, "Forbidden: Not your chirp")
		return
	}
	err = a.databaseQueries.DeleteChirp(r.Context(), idStr)
	if err != nil {
		internalError(response, err)
		return
	}
	noContentResponse(response, "Chirp deleted successfully")

}
