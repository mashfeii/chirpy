package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	var id uuid.UUID

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(*jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return id, err
	}

	stringID, _ := token.Claims.GetSubject()

	return uuid.Parse(stringID)
}

func GetAuthorizationToken(headers http.Header, apiType string) (string, error) {
	token := headers.Get("Authorization")

	trimedToken := strings.TrimPrefix(token, apiType+" ")

	if token == trimedToken {
		return "", fmt.Errorf("token is empty")
	}

	return trimedToken, nil
}

func MakeRefreshToken() string {
	buffer := make([]byte, 32)

	_, _ = rand.Read(buffer)

	return hex.EncodeToString(buffer)
}
