package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Lokee86/serverProject/internal/auth"
	"github.com/Lokee86/serverProject/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

type handleUser struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresinSeconds int32  `json:"expires_in_seconds"`
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

	jsonResponse(response, http.StatusCreated, newUserJson, "New user created successfully")

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
		errorResponse(response, http.StatusUnauthorized, "Unauthorized: Incorrect Password")
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

	jsonResponse(response, http.StatusOK, loggedInUser, "User logged in successfully")
}

// update user password and email
func (a *apiConfig) updateAccount(response http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	newData := handleUser{}
	err := decoder.Decode(&newData)
	if err != nil {
		internalError(response, err)
		return
	}
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
	user, err := a.databaseQueries.GetUserByID(r.Context(), userID)
	if err != nil {
		internalError(response, err)
		return
	}
	user.Email = newData.Email
	user.HashedPassword, err = auth.HashPassword(newData.Password)
	if err != nil {
		internalError(response, err)
		return
	}
	jsonData := jsonReturnUser(user)
	jsonResponse(response, http.StatusOK, jsonData, "Account successfully updated")

}

func (a *apiConfig) upgradeAccount(response http.ResponseWriter, r *http.Request) {
	type Upgrade struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	upgradeData := Upgrade{}
	err := decoder.Decode(&upgradeData)
	if err != nil {
		internalError(response, err)
		return
	}
	if upgradeData.Event != "user.upgraded" {
		noContentResponse(response, "Event not supported")
		return
	}
	userID, err := uuid.Parse(upgradeData.Data.UserID)
	if err != nil {
		internalError(response, err)
		return
	}
	err = a.databaseQueries.ActivateChirpyRed(r.Context(), userID)
	if err == sql.ErrNoRows {
		errorResponse(response, http.StatusNotFound, "User not found")
	} else if err != nil {
		noContentResponse(response, err.Error())
		return
	}
	noContentResponse(response, "User successfully upgraded")
}
