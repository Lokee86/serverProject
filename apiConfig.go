package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Lokee86/serverProject/internal/auth"
	"github.com/Lokee86/serverProject/internal/database"
)

type apiConfig struct {
	fileServerHits  atomic.Int32
	databaseQueries *database.Queries
	platform        string
}

type token struct {
	Token string `json:"token"`
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
		http.Error(response, "Unauthorized Reset Access", http.StatusForbidden)
		log.Println("Unauthorized Reset Access")
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

// handle requests for access tokens through refresh endpoint
func (a *apiConfig) refreshHandler(response http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		internalError(response, err)
		return
	}
	fullRefreshToken, err := a.databaseQueries.GetRefreshToken(r.Context(), refreshToken)
	if err == sql.ErrNoRows {
		errorResponse(response, http.StatusUnauthorized, "Unauthorized: Refresh token not found")
		return
	} else if fullRefreshToken.RevokedAt.Valid {
		errorResponse(response, http.StatusUnauthorized, "Unauthorized: Refresh token revoked")
		return
	} else if err != nil {
		internalError(response, err)
		return
	}
	if time.Now().After(fullRefreshToken.ExpiresAt) {
		errorResponse(response, http.StatusUnauthorized, "Unauthorize: Refresh token expired")
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
	jsonResponse(response, http.StatusOK, newAccessToken, "Access token refresehd.")
}

// revoke refresh tokens
func (a *apiConfig) revokeHandler(response http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		internalError(response, err)
		return
	}
	err = a.databaseQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		errorResponse(response, http.StatusUnauthorized, "Invalid refresh token")
		return
	}
	noContentResponse(response, "Refresh token revoked")
}
