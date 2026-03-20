package repository

import (
	"context"
	"database/sql"
	"net/url"
	"time"

	"github.com/lib/pq"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"

	"github.com/chivta/spotiscan/internal/appErrors"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"
)

type (
	CacheClient interface {
		SetRussianArtistNames(ctx context.Context, names []string) error
		FilterRussianArtistNames(ctx context.Context, names []string) ([]string, error)
	}

	DBClient interface {
		GetRussianArtistNames(ctx context.Context, names []string) ([]string, error)
		GetAllRussianArtistNames(ctx context.Context) ([]string, error)
		GetSpotifyToken(ctx context.Context) (*oauth2.Token, error)
		SetSpotifyToken(ctx context.Context, newToken *oauth2.Token) error
		GetUserByID(ctx context.Context, id int) (*models.User, error)
		GetUserByEmail(ctx context.Context, email string) (*models.User, error)
		CreateUser(ctx context.Context, user *models.User) (int, error)
		GetRefreshTokenByUserID(ctx context.Context, userID int) (string, time.Time, error)
		StoreRefreshTokenHash(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error
		DeleteRefreshTokenHash(ctx context.Context, userID int) error
	}

	SpotifyClient interface {
		GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error)
		GetRefreshedSpotifyToken(ctx context.Context) (*oauth2.Token, error)
	}
)

func NewRepo(logger *logger.Logger, db DBClient, redis CacheClient, spotify SpotifyClient) *Repo {
	return &Repo{
		logger:  logger,
		db:      db,
		redis:   redis,
		spotify: spotify,
	}
}

type Repo struct {
	logger  *logger.Logger
	db      DBClient
	spotify SpotifyClient
	redis   CacheClient

	token *oauth2.Token
}

func (r *Repo) translateSpotifyError(err error) error {
	if spotifyErr, ok := err.(spotify.Error); ok {
		r.logger.Debugf("spotify api error: %T: %v, status: %d", spotifyErr, spotifyErr, spotifyErr.Status)
		switch spotifyErr.Status {
		case 404:
			r.logger.Infof("spotify not found: %T: %v", spotifyErr, spotifyErr)
			return appErrors.ErrNotFound
		case 400:
			r.logger.Infof("spotify bad request: %T: %v", spotifyErr, spotifyErr)
			return appErrors.ErrBadRequest
		default:
			r.logger.Errorf("spotify error: %T: %v", spotifyErr, spotifyErr)
			return appErrors.ErrSpotifyAPIError
		}
	}
	if urlErr, ok := err.(*url.Error); ok {
		r.logger.Infof("spotify network error: %T: %v", urlErr, urlErr)
		return appErrors.ErrSpotifyAPIError
	}
	r.logger.Errorf("unknown spotify error: %T: %v", err, err)
	return appErrors.ErrSpotifyAPIError
}

func (r *Repo) LoadRussianArtistsToRedis(ctx context.Context) {
	if r.redis == nil {
		r.logger.Infof("redis not available, skipping artist cache load")
		return
	}
	if r.db == nil {
		r.logger.Warnf("db client is not initialized, cannot load ru artists to redis")
		return
	}

	r.logger.Infof("lazy loading ru_artists set from DB")
	allNames, err := r.db.GetAllRussianArtistNames(ctx)
	if err != nil {
		r.logger.Warnf("failed to load all ru artists from DB: %v", err)
		return
	}
	err = r.redis.SetRussianArtistNames(ctx, allNames)
	if err != nil {
		r.logger.Warnf("failed to set ru artists in redis: %v", err)
		return
	}
	r.logger.Infof("successfully loaded %d ru artists into redis", len(allNames))
}

func (r *Repo) FilterRussian(ctx context.Context, names []string) ([]string, error) {
	if r.redis != nil {
		ruNames, err := r.redis.FilterRussianArtistNames(ctx, names)
		if err != nil {
			r.logger.Warnf("redis error: %T: %v", err, err)
		} else {
			return ruNames, nil
		}
	}
	// fallback to db
	ruNames, err := r.filterRussianWithDB(ctx, names)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		return nil, appErrors.ErrDatabaseFailure
	}
	return ruNames, nil
}

func (r *Repo) filterRussianWithDB(ctx context.Context, names []string) ([]string, error) {
	ruNames, err := r.db.GetRussianArtistNames(ctx, names)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		return nil, appErrors.ErrDatabaseFailure
	}
	return ruNames, nil
}

func (r *Repo) GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error) {
	playlist, err := r.spotify.GetPlaylistWithTracks(ctx, playlistId)
	if err != nil {
		return nil, r.translateSpotifyError(err)
	}
	return playlist, nil
}

func (r *Repo) SetSpotifyToken(ctx context.Context, newToken *oauth2.Token) error {
	err := r.db.SetSpotifyToken(ctx, newToken)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		return appErrors.ErrDatabaseFailure
	}
	r.token = newToken
	return nil
}

func (r *Repo) GetStoredSpotifyToken(ctx context.Context) (*oauth2.Token, error) {
	if r.token != nil {
		return r.token, nil
	}

	token, err := r.db.GetSpotifyToken(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Infof("no stored spotify token found: %T: %v", err, err)
			return nil, appErrors.ErrNotFound
		}
		r.logger.Errorf("db error: %T: %v", err, err)
		return nil, appErrors.ErrDatabaseFailure
	}
	r.token = token
	return token, nil
}

func (r *Repo) GetRefreshedSpotifyToken(ctx context.Context) (*oauth2.Token, error) {
	token, err := r.spotify.GetRefreshedSpotifyToken(ctx)
	if err != nil {
		return nil, r.translateSpotifyError(err)
	}
	return token, nil
}

func (r *Repo) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	user, err := r.db.GetUserByID(ctx, id)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		if err == sql.ErrNoRows {
			return nil, appErrors.ErrNotFound
		}
		return nil, appErrors.ErrDatabaseFailure
	}
	return user, nil
}

func (r *Repo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := r.db.GetUserByEmail(ctx, email)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		if err == sql.ErrNoRows {
			return nil, appErrors.ErrUnauthorized
		}
		return nil, appErrors.ErrDatabaseFailure
	}
	return user, nil
}

func (r *Repo) StoreRefreshTokenHash(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	if err := r.db.StoreRefreshTokenHash(ctx, userID, tokenHash, expiresAt); err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		return appErrors.ErrDatabaseFailure
	}
	return nil
}

func (r *Repo) CreateUser(ctx context.Context, user *models.User) (int, error) {
	id, err := r.db.CreateUser(ctx, user)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return 0, appErrors.ErrEmailExists
		}
		return 0, appErrors.ErrDatabaseFailure
	}
	return id, nil
}

func (r *Repo) GetRefreshTokenByUserID(ctx context.Context, userID int) (string, time.Time, error) {
	tokenHash, expiresAt, err := r.db.GetRefreshTokenByUserID(ctx, userID)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		if err == sql.ErrNoRows {
			return "", time.Time{}, appErrors.ErrUnauthorized
		}
		return "", time.Time{}, appErrors.ErrDatabaseFailure
	}
	return tokenHash, expiresAt, nil
}

func (r *Repo) DeleteRefreshTokenHash(ctx context.Context, userID int) error {
	if err := r.db.DeleteRefreshTokenHash(ctx, userID); err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		return appErrors.ErrDatabaseFailure
	}
	return nil
}
