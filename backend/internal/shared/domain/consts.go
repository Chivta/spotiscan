package domain

import "time"

const (
	RateLimitRequestLimit  = 50
	RateLimitWindowSeconds = 60
)

const (
	JWTDuration          = 15 * 60           // 15 minutes
	RefreshTokenDuration = 30 * 24 * 60 * 60 // 30 days
	// if jwt cookie ages and gets deleted instead of being expired, then user would be considered anon despite valid refresh token
	JWTCookieAge          = 30 * 24 * 60 * 60 // 30 days
	RefreshTokenCookieAge = 30 * 24 * 60 * 60 // 30 days
	CookieJWT             = "jwt"
	CookieRefreshToken    = "refresh_token"
)

const (
	AnonSessionDuration  = 60 * 60 * 24 // 24 hours
	AnonSessionCookieAge = 60 * 60 * 24 // 24 hours
	AnonRequestLimit     = 5
)

const (
	UserRoleKey = "userRole"
	UserIDKey   = "userID"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
	RoleAnon  Role = "anon"
)

const (
	SpotifyQueueName = "spotify"
	ScannerQueueName = "scanner"
)

const JobTTL = 10 * time.Minute
