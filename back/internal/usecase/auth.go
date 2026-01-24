package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

const jwtIssuer = "sangehassan-admin"

// AuthService handles admin authentication.
type AuthService struct {
	repo      ports.AdminRepository
	jwtSecret []byte
	ttl       time.Duration
}

func NewAuthService(repo ports.AdminRepository, secret string, ttlHours int) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(secret),
		ttl:       time.Duration(ttlHours) * time.Hour,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}

	claims := jwt.RegisteredClaims{
		Issuer:    jwtIssuer,
		Subject:   user.Username,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.ttl)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) ParseToken(tokenString string) (string, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := parsed.Claims.(*jwt.RegisteredClaims)
	if !ok || !parsed.Valid {
		return "", errors.New("invalid token")
	}

	if claims.Issuer != jwtIssuer {
		return "", errors.New("invalid token issuer")
	}

	return claims.Subject, nil
}

func (s *AuthService) BuildSessionUser(username string) domain.AdminUser {
	return domain.AdminUser{Username: username}
}
