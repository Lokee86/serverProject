package auth

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var secret string

func init() {
	godotenv.Load()
	secret = os.Getenv("JWT_SECRET")
}

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	expires := time.Minute

	token, err := MakeJWT(userID, expires)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}
	log.Println(token)
	parsedID, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}

	if parsedID != userID {
		t.Errorf("Expected %v, got %v", userID, parsedID)
	}
}

func TestValidateJWT_Expired(t *testing.T) {
	userID := uuid.New()
	expires := -1 * time.Minute // already expired

	token, err := MakeJWT(userID, expires)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	_, err = ValidateJWT(token)
	if err == nil {
		t.Fatal("expected error for expired token, got none")
	}

	// Check if the error is actually a token expiry error
	_, err = ValidateJWT(token)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			// Token is expired (expected case)
		} else {
			t.Fatalf("expected token expired error, got: %v", err)
		}
	}
}

func TestValidateJWT_InvalidSignature(t *testing.T) {
	userID := uuid.New()
	token, err := MakeJWT(userID, time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	// Temporarily override TokenSecret for this test
	oldSecret := TokenSecret
	TokenSecret = "wrong-secret"
	defer func() { TokenSecret = oldSecret }() // restore after test

	_, err = ValidateJWT(token)
	if err == nil {
		t.Fatal("expected error for invalid signature, got none")
	}
}
