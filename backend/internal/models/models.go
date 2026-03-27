package models

const (
	JWTDuration           = 15 * 60           // 15 minutes
	RefreshTokenDuration  = 30 * 24 * 60 * 60 // 30 days
	JWTCookieAge          = 30 * 24 * 60 * 60 // 30 days
	RefreshTokenCookieAge = 30 * 24 * 60 * 60 // 30 days
	CookieJWT             = "jwt"
	CookieRefreshToken    = "refresh_token"
)

const (
	AnonSessionDuration = 15 * 60 // 15 minutes
	AnonRequestLimit    = 5      // 5 requests per AnonSessionDuration
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
	ID       string   `json:"ID"`
	Name     string   `json:"Name"`
	ImageURL string   `json:"ImageURL"`
	Artists  []Artist `json:"Artists"`
}

type Artist struct {
	ID            string `json:"ID"`
	Name          string `json:"Name"`
	SpotifyURL    string `json:"URL"`
	DescriptionUA string `json:"DescriptionUA"`
	DescriptionEN string `json:"DescriptionEN"`
	Source        string `json:"Source"`
	SourceURL     string `json:"SourceURL"`
	Country       string `json:"Country"`
	Confirmed     bool   `json:"Confirmed"`
}

type Playlist struct {
	ID          string  `json:"ID"`
	Name        string  `json:"Name"`
	Description string  `json:"Description"`
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
