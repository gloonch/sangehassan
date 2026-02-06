package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

var (
	ErrEmailExists    = errors.New("email already exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrRefreshInvalid = errors.New("invalid refresh token")
	ErrRefreshExpired = errors.New("refresh token expired")
	ErrRefreshRevoked = errors.New("refresh token revoked")
	ErrInactiveUser   = errors.New("user inactive")
)

const userJWTIssuer = "sangehassan-user"

type TokenPair struct {
	AccessToken    string
	AccessExpires  time.Time
	RefreshToken   string
	RefreshExpires time.Time
}

type UserAuthService struct {
	users      ports.UserRepository
	refresh    ports.RefreshTokenRepository
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewUserAuthService(users ports.UserRepository, refresh ports.RefreshTokenRepository, secret string, accessMinutes, refreshDays int) *UserAuthService {
	return &UserAuthService{
		users:      users,
		refresh:    refresh,
		jwtSecret:  []byte(secret),
		accessTTL:  time.Duration(accessMinutes) * time.Minute,
		refreshTTL: time.Duration(refreshDays) * 24 * time.Hour,
	}
}

func (s *UserAuthService) SignUp(ctx context.Context, email, password string, fullName, phone *string) (domain.UserInfo, TokenPair, error) {
	_, err := s.users.GetByEmail(ctx, email)
	if err == nil {
		return domain.UserInfo{}, TokenPair{}, ErrEmailExists
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.UserInfo{}, TokenPair{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.UserInfo{}, TokenPair{}, err
	}

	user, err := s.users.Create(ctx, domain.User{
		Email:        email,
		PasswordHash: string(hash),
		FullName:     fullName,
		Phone:        phone,
		Role:         "user",
		IsActive:     true,
	})
	if err != nil {
		return domain.UserInfo{}, TokenPair{}, err
	}

	pair, err := s.issueTokenPair(ctx, user.ID, user.Role)
	if err != nil {
		return domain.UserInfo{}, TokenPair{}, err
	}

	return user.SafeInfo(), pair, nil
}

func (s *UserAuthService) Login(ctx context.Context, email, password string) (domain.UserInfo, TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.UserInfo{}, TokenPair{}, ErrInvalidCredentials
		}
		return domain.UserInfo{}, TokenPair{}, err
	}

	if !user.IsActive {
		return domain.UserInfo{}, TokenPair{}, ErrInactiveUser
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return domain.UserInfo{}, TokenPair{}, ErrInvalidCredentials
	}

	pair, err := s.issueTokenPair(ctx, user.ID, user.Role)
	if err != nil {
		return domain.UserInfo{}, TokenPair{}, err
	}

	_ = s.users.UpdateLastLogin(ctx, user.ID, time.Now())

	return user.SafeInfo(), pair, nil
}

func (s *UserAuthService) Refresh(ctx context.Context, rawRefresh string) (domain.UserInfo, TokenPair, error) {
	hash := hashToken(rawRefresh)
	token, err := s.refresh.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.UserInfo{}, TokenPair{}, ErrRefreshInvalid
		}
		return domain.UserInfo{}, TokenPair{}, err
	}

	if token.RevokedAt != nil {
		return domain.UserInfo{}, TokenPair{}, ErrRefreshRevoked
	}
	if time.Now().After(token.ExpiresAt) {
		return domain.UserInfo{}, TokenPair{}, ErrRefreshExpired
	}

	user, err := s.users.GetByID(ctx, token.UserID)
	if err != nil {
		return domain.UserInfo{}, TokenPair{}, err
	}
	if !user.IsActive {
		return domain.UserInfo{}, TokenPair{}, ErrInactiveUser
	}

	// Rotate refresh token
	_ = s.refresh.Revoke(ctx, token.ID)

	pair, err := s.issueTokenPair(ctx, user.ID, user.Role)
	if err != nil {
		return domain.UserInfo{}, TokenPair{}, err
	}

	return user.SafeInfo(), pair, nil
}

func (s *UserAuthService) Logout(ctx context.Context, userID string) error {
	return s.refresh.DeleteByUser(ctx, userID)
}

func (s *UserAuthService) GetMe(ctx context.Context, userID string) (domain.UserInfo, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.UserInfo{}, ErrUserNotFound
		}
		return domain.UserInfo{}, err
	}
	return user.SafeInfo(), nil
}

func (s *UserAuthService) UpdateMe(ctx context.Context, userID string, fullName, phone *string) (domain.UserInfo, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return domain.UserInfo{}, err
	}
	user.FullName = fullName
	user.Phone = phone
	updated, err := s.users.UpdateProfile(ctx, user)
	if err != nil {
		return domain.UserInfo{}, err
	}
	return updated.SafeInfo(), nil
}

func (s *UserAuthService) ListRequests(ctx context.Context, userID string) ([]interface{}, error) {
	_ = userID
	return []interface{}{}, nil
}

func (s *UserAuthService) issueTokenPair(ctx context.Context, userID, role string) (TokenPair, error) {
	accessExp := time.Now().Add(s.accessTTL)
	refreshExp := time.Now().Add(s.refreshTTL)

	accessClaims := jwt.RegisteredClaims{
		Issuer:    userJWTIssuer,
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(accessExp),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	accessClaims.Audience = jwt.ClaimStrings{role}

	access := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := access.SignedString(s.jwtSecret)
	if err != nil {
		return TokenPair{}, err
	}

	rawRefresh, err := generateSecureToken()
	if err != nil {
		return TokenPair{}, err
	}
	hash := hashToken(rawRefresh)

	_, err = s.refresh.Create(ctx, domain.RefreshToken{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: refreshExp,
	})
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:    accessToken,
		AccessExpires:  accessExp,
		RefreshToken:   rawRefresh,
		RefreshExpires: refreshExp,
	}, nil
}

func (s *UserAuthService) ParseAccess(tokenString string) (string, error) {
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
	if claims.Issuer != userJWTIssuer {
		return "", errors.New("invalid token issuer")
	}
	return claims.Subject, nil
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
