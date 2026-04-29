package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"sangehassan/back/internal/domain"
	"sangehassan/back/internal/ports"
)

var (
	ErrEmailExists    = errors.New("email already exists")
	ErrPhoneExists    = errors.New("phone already exists")
	ErrPhoneRequired  = errors.New("phone is required")
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
	deals      ports.DealRequestRepository
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewUserAuthService(users ports.UserRepository, refresh ports.RefreshTokenRepository, deals ports.DealRequestRepository, secret string, accessMinutes, refreshDays int) *UserAuthService {
	return &UserAuthService{
		users:      users,
		refresh:    refresh,
		deals:      deals,
		jwtSecret:  []byte(secret),
		accessTTL:  time.Duration(accessMinutes) * time.Minute,
		refreshTTL: time.Duration(refreshDays) * 24 * time.Hour,
	}
}

func (s *UserAuthService) SignUp(ctx context.Context, phone, password string) (domain.UserInfo, TokenPair, error) {
	normalizedPhone := normalizePhone(phone)
	if normalizedPhone == "" {
		return domain.UserInfo{}, TokenPair{}, ErrPhoneRequired
	}

	_, err := s.users.GetByPhone(ctx, normalizedPhone)
	if err == nil {
		return domain.UserInfo{}, TokenPair{}, ErrPhoneExists
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.UserInfo{}, TokenPair{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.UserInfo{}, TokenPair{}, err
	}

	user, err := s.users.Create(ctx, domain.User{
		Email:        phoneAliasEmail(normalizedPhone),
		PasswordHash: string(hash),
		Phone:        &normalizedPhone,
		Role:         "user",
		IsActive:     true,
	})
	if err != nil {
		if conflictErr := resolveUniqueConflict(err); conflictErr != nil {
			return domain.UserInfo{}, TokenPair{}, conflictErr
		}
		return domain.UserInfo{}, TokenPair{}, err
	}

	pair, err := s.issueTokenPair(ctx, user.ID, user.Role)
	if err != nil {
		return domain.UserInfo{}, TokenPair{}, err
	}

	return user.SafeInfo(), pair, nil
}

func (s *UserAuthService) Login(ctx context.Context, phone, password string) (domain.UserInfo, TokenPair, error) {
	normalizedPhone := normalizePhone(phone)
	if normalizedPhone == "" {
		return domain.UserInfo{}, TokenPair{}, ErrInvalidCredentials
	}

	user, err := s.users.GetByPhone(ctx, normalizedPhone)
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

func (s *UserAuthService) UpdateMe(ctx context.Context, userID string, fullName, phone, email *string) (domain.UserInfo, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return domain.UserInfo{}, err
	}

	trimmedFullName := trimOptional(fullName)
	user.FullName = trimmedFullName

	if phone != nil {
		normalizedPhone := normalizePhone(*phone)
		if normalizedPhone == "" {
			return domain.UserInfo{}, ErrPhoneRequired
		}
		if currentPhone := trimOptional(user.Phone); currentPhone == nil || *currentPhone != normalizedPhone {
			existing, lookupErr := s.users.GetByPhone(ctx, normalizedPhone)
			if lookupErr == nil && existing.ID != userID {
				return domain.UserInfo{}, ErrPhoneExists
			}
			if lookupErr != nil && !errors.Is(lookupErr, sql.ErrNoRows) {
				return domain.UserInfo{}, lookupErr
			}
		}
		user.Phone = &normalizedPhone
	}

	if email != nil {
		trimmedEmail := strings.TrimSpace(*email)
		if trimmedEmail != "" && trimmedEmail != user.Email {
			existing, lookupErr := s.users.GetByEmail(ctx, trimmedEmail)
			if lookupErr == nil && existing.ID != userID {
				return domain.UserInfo{}, ErrEmailExists
			}
			if lookupErr != nil && !errors.Is(lookupErr, sql.ErrNoRows) {
				return domain.UserInfo{}, lookupErr
			}
			user.Email = trimmedEmail
		}
	}

	updated, err := s.users.UpdateProfile(ctx, user)
	if err != nil {
		if conflictErr := resolveUniqueConflict(err); conflictErr != nil {
			return domain.UserInfo{}, conflictErr
		}
		return domain.UserInfo{}, err
	}
	return updated.SafeInfo(), nil
}

func (s *UserAuthService) ListRequests(ctx context.Context, userID string) ([]domain.DealRequest, error) {
	if s.deals == nil {
		return nil, nil
	}
	return s.deals.ListByBuyer(ctx, userID, 100, 0)
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

func normalizePhone(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}

	var cleaned strings.Builder
	cleaned.Grow(len(value))
	for _, r := range value {
		if unicode.IsSpace(r) || r == '-' || r == '_' || r == '(' || r == ')' {
			continue
		}
		cleaned.WriteRune(r)
	}
	normalized := cleaned.String()
	if strings.HasPrefix(normalized, "00") {
		normalized = "+" + strings.TrimPrefix(normalized, "00")
	}
	return normalized
}

func phoneAliasEmail(phone string) string {
	var digits strings.Builder
	digits.Grow(len(phone))
	for _, r := range phone {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		}
	}
	value := digits.String()
	if value == "" {
		value = "user"
	}
	return fmt.Sprintf("u%s@phone.sangehassan.local", value)
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func resolveUniqueConflict(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if string(pqErr.Code) != "23505" {
			return nil
		}
		info := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%s %s", pqErr.Constraint, pqErr.Detail)))
		if strings.Contains(info, "email") {
			return ErrEmailExists
		}
		if strings.Contains(info, "phone") {
			return ErrPhoneExists
		}
		return ErrPhoneExists
	}
	return nil
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
