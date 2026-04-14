package appErrors

type AppError struct {
	HTTPCode int
	Code     string
	Message  string
}

func (e *AppError) Error() string { return e.Code }

// sentinel errors returned by the services layer to translate into HTTP responses
var (
	ErrInternal             = &AppError{500, "INTERNAL_ERROR", "internal server error"}
	ErrDatabaseFailure      = &AppError{500, "DATABASE_ERROR", "database error"}
	ErrSpotifyAPIError      = &AppError{500, "SPOTIFY_API_ERROR", "spotify api error"}
	ErrTooManyRequests      = &AppError{429, "TOO_MANY_REQUESTS", "too many requests"}
	ErrNotFound             = &AppError{404, "NOT_FOUND", "not found"}
	ErrSpotifyNotFound      = &AppError{404, "SPOTIFY_NOT_FOUND", "spotify not found"}
	ErrBadRequest           = &AppError{400, "BAD_REQUEST", "bad request"}
	ErrUnauthorized         = &AppError{401, "UNAUTHORIZED", "unauthorized"}
	ErrForbidden            = &AppError{403, "FORBIDDEN", "forbidden"}
	ErrQuotaExceeded        = &AppError{429, "ANON_QUOTA_EXCEEDED", "anonymous quota exceeded"}
	ErrEmailExists          = &AppError{409, "EMAIL_EXISTS", "email already exists"}
	ErrInvalidCredentials   = &AppError{401, "INVALID_CREDENTIALS", "invalid credentials"}
	ErrArtistExists         = &AppError{409, "ARTIST_EXISTS", "artist already exists"}
	ErrSuggestionNotPending = &AppError{400, "SUGGESTION_NOT_PENDING", "suggestion in not in pending state"}
)
