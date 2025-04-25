package auth

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	godotenv.Load()
	secret := os.Getenv("JWT_SECRET")
	expires := time.Minute

	token, err := MakeJWT(userID, secret, expires)
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
