package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func MakeRefreshToken() (string, error) {
	refreshTokenBytes := make([]byte, 32)
	_, err := rand.Read(refreshTokenBytes)
	if err != nil {
		return "", fmt.Errorf("error generating refresh token: %w", err)
	}
	refreshToken := hex.EncodeToString(refreshTokenBytes)
	return refreshToken, nil
}
