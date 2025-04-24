package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Lokee86/serverProject/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileServerHits  atomic.Int32
	databaseQueries *database.Queries
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

// increment hit counter
func (a *apiConfig) serverHitCounter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// metrics handler
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

// reset counter
func (a *apiConfig) resetCounter(response http.ResponseWriter, r *http.Request) {
	a.fileServerHits.Store(0)
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("Hits counter reset to 0"))
	log.Println("Hit counter reset to 0")

}

// create user
func (a *apiConfig) createUserHandler(response http.ResponseWriter, r *http.Request) {
	type createNewUser struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := createNewUser{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding Json: %v", err)
		http.Error(response, "failed to decode Json", http.StatusInternalServerError)
		return
	}
	newUser, err := a.databaseQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(response, "failed to create user", http.StatusInternalServerError)
		return
	}
	newUserJson := User{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
	}

	response.WriteHeader(http.StatusCreated)
	newUserData, err := json.Marshal(newUserJson)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(response, "failed to encode response", http.StatusInternalServerError)
		return
	}
	response.Write(newUserData)

}
