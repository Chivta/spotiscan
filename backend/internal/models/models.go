package models

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

type LoginDTO struct {
	Email    string `json:"Email" validate:"required,email"`
	Password string `json:"Password" validate:"required,min=8,max=64"`
}

type SignupDTO struct {
	Email    string `json:"Email" validate:"required,email"`
	Password string `json:"Password" validate:"required,min=8,max=64"`
}

type Track struct {
	SpotifyID string          `json:"SpotifyID"`
	Name      string          `json:"Name"`
	ImageURL  string          `json:"ImageURL"`
	Artists   []SpotifyArtist `json:"Artists"`
}

type SpotifyArtist struct {
	SpotifyID string `json:"SpotifyID"`
	Name      string `json:"Name"`
}

type Artist struct {
	ID            int    `json:"ID"`
	Name          string `json:"Name"`
	SpotifyID     string `json:"SpotifyID"`
	DescriptionUA string `json:"DescriptionUA"`
	DescriptionEN string `json:"DescriptionEN"`
	Source        string `json:"Source"`
	SourceURL     string `json:"SourceURL"`
	Country       string `json:"Country"`
	Confirmed     bool   `json:"Confirmed"`
}

type Playlist struct {
	SpotifyID string  `json:"SpotifyID"`
	Tracks    []Track `json:"Tracks"`
}

type Album struct {
	SpotifyID string  `json:"SpotifyID"`
	Tracks    []Track `json:"Tracks"`
}

type RuContent struct {
	Tracks  []Track  `json:"Tracks"`
	Artists []Artist `json:"Artists"`
}

type User struct {
	ID           int    `json:"ID"`
	Email        string `json:"Email"`
	Role         Role   `json:"Role"`
	PasswordHash string `json:"-"`
}

type ArtistInsertSuggestion struct {
	ID            int    `json:"ID"`
	CreatorID     int    `json:"CreatorID"`
	ArtistName    string `json:"ArtistName"`
	Description   string `json:"Description"`
	State         string `json:"State"`
	DeclineReason string `json:"DeclineReason,omitempty"`
	CreatedAt     string `json:"CreatedAt"`
	UpdatedAt     string `json:"UpdatedAt"`
}

type ArtistDeleteSuggestion struct {
	ID            int    `json:"ID"`
	CreatorID     int    `json:"CreatorID"`
	ArtistName    string `json:"ArtistName"`
	Description   string `json:"Description"`
	State         string `json:"State"`
	DeclineReason string `json:"DeclineReason,omitempty"`
	CreatedAt     string `json:"CreatedAt"`
	UpdatedAt     string `json:"UpdatedAt"`
}
