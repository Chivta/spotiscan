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

	"github.com/chivta/spotiscan/internal/appErrors"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"
)

type AuthRepository interface {
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) (int, error)
	GetRefreshTokenByUserID(ctx context.Context, userID int) (string, time.Time, error)
	StoreRefreshTokenHash(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error
	DeleteRefreshTokenHash(ctx context.Context, userID int) error
	IncrementAnonRequestCounter(ctx context.Context, anonID, path string, ttl time.Duration) (int, error)
}

func NewAuthService(repo AuthRepository, logger *logger.Logger, jwtSecret []byte) *AuthService {
	return &AuthService{
		repo:      repo,
		log:       logger,
		jwtSecret: jwtSecret,
	}
}

type AuthService struct {
	repo      AuthRepository
	log       *logger.Logger
	jwtSecret []byte
}

type JWTClaims struct {
	UserID string      `json:"user_id"`
	Role   models.Role `json:"role"`
	jwt.RegisteredClaims
}

type Session struct {
	JWT          string
	RefreshToken string
	UserID       string
	Role         models.Role
}

type AnonymousSession struct {
	JWT  string
	UserID string
	Role models.Role
}

func (s *AuthService) Login(ctx context.Context, loginDTO models.LoginDTO) (*Session, error) {
	user, err := s.repo.GetUserByEmail(ctx, loginDTO.Email)
	if err != nil {
		return nil, err
	}

	if err := comparePasswordHashWithSalt(user.PasswordHash, loginDTO.Password); err != nil {
		return nil, appErrors.ErrInvalidCredentials
	}

	return s.createSession(ctx, user)
}

func (s *AuthService) Signup(ctx context.Context, signupDTO models.SignupDTO) (*Session, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signupDTO.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", appErrors.ErrInternal, err)
	}

	user := models.User{
		Email:        signupDTO.Email,
		PasswordHash: string(hashedPassword),
		Role:         models.RoleUser,
	}
	id, err := s.repo.CreateUser(ctx, &user)
	if err != nil {
		return nil, err
	}
	user.ID = id

	return s.createSession(ctx, &user)
}

// createSession issues a new JWT + refresh token, stores the refresh token hash, and returns a NewSession.
func (s *AuthService) createSession(ctx context.Context, user *models.User) (*Session, error) {
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	refreshTokenHash := hashRefreshToken(refreshToken)
	refreshExpiresAt := time.Now().Add(time.Duration(models.RefreshTokenDuration) * time.Second)
	if err := s.repo.StoreRefreshTokenHash(ctx, user.ID, refreshTokenHash, refreshExpiresAt); err != nil {
		return nil, err
	}

	claims := JWTClaims{
		UserID: strconv.Itoa(user.ID),
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(models.JWTDuration) * time.Second)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", appErrors.ErrUnauthorized, err)
	}

	return &Session{
		JWT:          signed,
		RefreshToken: refreshToken,
		UserID:       strconv.Itoa(user.ID),
		Role:         user.Role,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID int) error {
	return s.repo.DeleteRefreshTokenHash(ctx, userID)
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
		return JWTClaims{}, fmt.Errorf("%w: %w", appErrors.ErrUnauthorized, err)
	}
	return claims, nil
}

func (s *AuthService) CreateAnonymousSession(ctx context.Context) (*AnonymousSession, error) {
	anonID, err := generateAnonUserID()
	if err != nil {
		s.log.Errorf("Failed to generate anon user ID: %v", err)
		return nil, appErrors.ErrInternal
	}

	claims := JWTClaims{
		UserID: anonID,
		Role:   models.RoleAnon,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(models.AnonSessionDuration) * time.Second)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", appErrors.ErrUnauthorized, err)
	}

	return &AnonymousSession{
		JWT:  signed,
		UserID: anonID,
		Role: models.RoleAnon,
	}, nil
}

func (s *AuthService) IncrementAnonQuota(ctx context.Context, anonID, path string) (int, error) {
	return s.repo.IncrementAnonRequestCounter(ctx, anonID, path, time.Duration(models.AnonSessionDuration)*time.Second)
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
		return nil, appErrors.ErrUnauthorized
	}

	userID, err := strconv.Atoi(claims.UserID)
	if err != nil {
		return nil, appErrors.ErrUnauthorized
	}

	incomingHash := hashRefreshToken(refreshToken)

	storedHash, expiresAt, err := s.repo.GetRefreshTokenByUserID(ctx, userID)
	if err != nil {
		return nil, appErrors.ErrUnauthorized
	}

	if subtle.ConstantTimeCompare([]byte(storedHash), []byte(incomingHash)) != 1 {
		return nil, appErrors.ErrUnauthorized
	}

	if !time.Now().Before(expiresAt) {
		return nil, appErrors.ErrUnauthorized
	}

	if err := s.repo.DeleteRefreshTokenHash(ctx, userID); err != nil {
		return nil, appErrors.ErrUnauthorized
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, appErrors.ErrUnauthorized
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
