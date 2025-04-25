package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Lokee86/serverProject/internal/auth"
	"github.com/Lokee86/serverProject/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileServerHits  atomic.Int32
	databaseQueries *database.Queries
	platform        string
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

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

type handleUser struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresinSeconds int32  `json:"expires_in_seconds"`
}

type token struct {
	Token string `json:"token"`
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

// increment file server hit counter
func (a *apiConfig) serverHitCounter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// metrics handler serves metrics to admins
func (a *apiConfig) metricsHandler(response http.ResponseWriter, r *http.Request) {
	log.Printf("Hits: %v", a.fileServerHits.Load())
	response.WriteHeader(http.StatusOK)
	response.Header().Set("Content-Type", "text/html")
	response.Write(fmt.Appendf(nil,
		`<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
</html>`, a.fileServerHits.Load()))
}

// reset file server hit counter
func (a *apiConfig) resetCounter(response http.ResponseWriter, r *http.Request) {
	if a.platform != "dev" {
		http.Error(response, "Unauthorized Access", http.StatusForbidden)
		return
	}
	a.fileServerHits.Store(0)
	err := a.databaseQueries.ResetUsers(r.Context())
	if err != nil {
		log.Printf("Error resetting 'users': %v", err)
		http.Error(response, "failed to create user", http.StatusInternalServerError)
	}
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("Hits counter reset to 0"))
	log.Println("Hit counter reset to 0")

}

// create new user in the database
func (a *apiConfig) createUserHandler(response http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := handleUser{}
	err := decoder.Decode(&params)
	if err != nil {
		internalError(response, err)
		return
	}
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		internalError(response, err)
		return
	}
	compatibleParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}
	newUser, err := a.databaseQueries.CreateUser(r.Context(), compatibleParams)
	if err != nil {
		internalError(response, err)
		return
	}
	newUserJson := jsonReturnUser(newUser)

	jsonResponse(response, http.StatusCreated, newUserJson)

}

// handle login requests
func (a *apiConfig) loginHandler(response http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := handleUser{}
	err := decoder.Decode(&params)
	if err != nil {
		internalError(response, err)
		return
	}
	user, err := a.databaseQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		internalError(response, err)
		return
	}
	err = auth.CheckPasswordHash(user.HashedPassword, params.Password)
	if err != nil {
		errorResponse(response, http.StatusUnauthorized, "Unauthorized")
		return
	}
	loggedInUser := jsonReturnUser(user)
	token, err := auth.MakeJWT(loggedInUser.ID, 3600*time.Second)
	if err != nil {
		internalError(response, err)
		return
	}
	loggedInUser.Token = token
	loggedInUser.RefreshToken, err = auth.MakeRefreshToken()
	if err != nil {
		internalError(response, err)
		return
	}
	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     loggedInUser.RefreshToken,
		UserID:    loggedInUser.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	}
	err = a.databaseQueries.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		internalError(response, err)
		return
	}

	jsonResponse(response, http.StatusOK, loggedInUser)
}

// handle requests for access tokens through refresh endpoint
func (a *apiConfig) refreshHandler(response http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		internalError(response, err)
		return
	}
	fullRefreshToken, err := a.databaseQueries.GetRefreshToken(r.Context(), refreshToken)
	if err == sql.ErrNoRows {
		errorResponse(response, http.StatusUnauthorized, "Unauthorized")
		return
	} else if err != nil {
		internalError(response, err)
		return
	}
	if time.Now().After(fullRefreshToken.ExpiresAt) {
		errorResponse(response, http.StatusUnauthorized, "Unauthorize")
		return
	}
	newAccessTokenValue, err := auth.MakeJWT(fullRefreshToken.UserID, 60*time.Minute)
	if err != nil {
		internalError(response, err)
		return
	}
	newAccessToken := token{
		Token: newAccessTokenValue,
	}
	jsonResponse(response, http.StatusNoContent, newAccessToken)
}
