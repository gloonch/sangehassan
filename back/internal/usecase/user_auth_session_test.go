package usecase

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAccessTokenPreservesMillisecondIssueOrder(t *testing.T) {
	secret := []byte("test-secret-with-at-least-thirty-two-bytes")
	issued := time.Unix(1_800_000_000, 456_000_000).UTC()
	claims := userAccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    userJWTIssuer,
			Subject:   "user-id",
			IssuedAt:  jwt.NewNumericDate(issued),
			ExpiresAt: jwt.NewNumericDate(issued.Add(time.Hour)),
		},
		IssuedAtMillis: issued.UnixMilli(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	service := &UserAuthService{jwtSecret: secret}
	id, parsedIssued, err := service.ParseAccessDetails(token)
	if err != nil {
		t.Fatal(err)
	}
	if id != "user-id" || !parsedIssued.Equal(issued) {
		t.Fatalf("parsed id/time = %s/%s, want user-id/%s", id, parsedIssued, issued)
	}
}
