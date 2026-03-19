package models

const (
	// TODO: make sure in prod values are correct
	JWTDuration          = 15 // 15 minutes
	RefreshTokenDuration = 30 * 24 * 60 * 60 // 30 days
	JWTCookieAge          = 30 * 24 * 60 * 60 // 30 days
	RefreshTokenCookieAge = 30 * 24 * 60 * 60 // 30 days
	CookieJWT             = "jwt"
	CookieRefreshToken    = "refresh_token"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type LoginDTO struct {
	Email    string `json:"Email" validate:"required,email"`
	Password string `json:"Password" validate:"required,min=8,max=64,alphanum"`
}

type SignupDTO struct {
	Email    string `json:"Email" validate:"required,email"`
	Password string `json:"Password" validate:"required,min=8,max=64,alphanum"`
}

type Track struct {
	ID       string   `json:"ID"`
	Name     string   `json:"Name"`
	ImageURL string   `json:"ImageURL"`
	Artists  []Artist `json:"Artists"`
}

type Artist struct {
	ID   string `json:"ID"`
	URL  string `json:"URL"`
	Name string `json:"Name"`
}

type Playlist struct {
	ID          string  `json:"ID"`
	Name        string  `json:"Name"`
	Description string  `json:"Description"`
	Owned       bool    `json:"Owned"`
	ImageURL    string  `json:"ImageURL"`
	TrackCount  int     `json:"TrackCount"`
	Tracks      []Track `json:"Tracks"`
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
