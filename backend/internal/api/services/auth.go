package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/domain"
)

type (
	sessionTokenRepo interface {
		StoreRefreshTokenHash(ctx context.Context, userID int, refreshTokenHash string, expiresAt time.Time) error
		GetRefreshTokenByUserID(ctx context.Context, userID int) (string, time.Time, error)
		DeleteRefreshTokenHash(ctx context.Context, userID int) error
		IncrementAnonRequestCounter(ctx context.Context, anonID, path string, expiration time.Duration) (int, error)
	}
	userRepo interface {
		GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
		GetUserByID(ctx context.Context, id int) (*domain.User, error)
		CreateUser(ctx context.Context, user *domain.User) (int, error)
	}
)

func NewAuthService(jwtSecret []byte, tokenRepo sessionTokenRepo, userRepo userRepo) *AuthService {
	return &AuthService{
		jwtSecret: jwtSecret,
		tokenRepo: tokenRepo,
		userRepo:  userRepo,
	}
}

type AuthService struct {
	tokenRepo sessionTokenRepo
	userRepo  userRepo
	jwtSecret []byte
}

type JWTClaims struct {
	UserID string      `json:"user_id"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}

type Session struct {
	JWT          string
	RefreshToken string
	UserID       string
	Role         domain.Role
}

type AnonymousSession struct {
	JWT    string
	UserID string
	Role   domain.Role
}

func (s *AuthService) Login(ctx context.Context, loginDTO domain.LoginDTO) (*Session, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, loginDTO.Email)
	if err != nil {
		return nil, err
	}

	if err := comparePasswordHashWithSalt(user.PasswordHash, loginDTO.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	session, err := s.createSession(ctx, user)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *AuthService) Signup(ctx context.Context, signupDTO domain.SignupDTO) (*Session, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signupDTO.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := domain.User{
		Email:        signupDTO.Email,
		PasswordHash: string(hashedPassword),
		Role:         domain.RoleUser,
	}
	id, err := s.userRepo.CreateUser(ctx, &user)
	if err != nil {
		if err != domain.ErrEmailExists {
			log.Error().Err(err).Msg("failed to create user")
		}
		return nil, err
	}
	user.ID = id

	session, err := s.createSession(ctx, &user)
	if err != nil {
		log.Error().Err(err).Msg("failed to create session after user creation")
		return nil, err
	}
	return session, nil
}

// createSession issues a new JWT + refresh token, stores the refresh token hash, and returns session.
func (s *AuthService) createSession(ctx context.Context, user *domain.User) (*Session, error) {
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshTokenHash := hashRefreshToken(refreshToken)
	refreshExpiresAt := time.Now().Add(time.Duration(domain.RefreshTokenDuration) * time.Second)
	if err := s.tokenRepo.StoreRefreshTokenHash(ctx, user.ID, refreshTokenHash, refreshExpiresAt); err != nil {
		return nil, err
	}

	claims := JWTClaims{
		UserID: strconv.Itoa(user.ID),
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(domain.JWTDuration) * time.Second)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrUnauthorized, err)
	}

	return &Session{
		JWT:          signed,
		RefreshToken: refreshToken,
		UserID:       strconv.Itoa(user.ID),
		Role:         user.Role,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID int) error {
	return s.tokenRepo.DeleteRefreshTokenHash(ctx, userID)
}

func (s *AuthService) ParseJWT(jwtStr string) (JWTClaims, error) {
	var claims JWTClaims
	_, err := jwt.ParseWithClaims(jwtStr, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return claims, fmt.Errorf("%w: %w", domain.ErrUnauthorized, err)
	}
	return claims, nil
}

func (s *AuthService) CreateAnonymousSession(ctx context.Context) (*AnonymousSession, error) {
	anonID, err := generateAnonUserID()
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate anon user ID")
		return nil, domain.ErrInternal
	}

	claims := JWTClaims{
		UserID: anonID,
		Role:   domain.RoleAnon,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(domain.AnonSessionDuration) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrUnauthorized, err)
	}

	return &AnonymousSession{
		JWT:    signed,
		UserID: anonID,
		Role:   domain.RoleAnon,
	}, nil
}

func (s *AuthService) IncrementAnonQuota(ctx context.Context, anonID, path string) (int, error) {
	return s.tokenRepo.IncrementAnonRequestCounter(ctx, anonID, path, time.Duration(domain.AnonSessionDuration)*time.Second)
}

// ExchangeRefreshToken validates the refresh token against the DB, revokes it, and returns a new session.
func (s *AuthService) ExchangeRefreshToken(ctx context.Context, expiredJWTStr, refreshToken string) (*Session, error) {
	var claims JWTClaims
	_, err := jwt.NewParser(jwt.WithoutClaimsValidation()).ParseWithClaims(expiredJWTStr, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	userID, err := strconv.Atoi(claims.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	incomingHash := hashRefreshToken(refreshToken)

	storedHash, expiresAt, err := s.tokenRepo.GetRefreshTokenByUserID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	if subtle.ConstantTimeCompare([]byte(storedHash), []byte(incomingHash)) != 1 {
		return nil, domain.ErrUnauthorized
	}

	if !time.Now().Before(expiresAt) {
		return nil, domain.ErrUnauthorized
	}

	if err := s.tokenRepo.DeleteRefreshTokenHash(ctx, userID); err != nil {
		return nil, domain.ErrUnauthorized
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.createSession(ctx, user)
}

func comparePasswordHashWithSalt(storedHash string, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plainPassword))
}

func generateAnonUserID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func generateRefreshToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
