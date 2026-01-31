package repository

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"

	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/chivta/spotiscan/internal/repository/db_client"
	"github.com/chivta/spotiscan/internal/repository/redis_client"
	"github.com/chivta/spotiscan/internal/repository/spotify_client"
)

type Repo interface {
	InitDB(databaseURL string) error
	InitRedis(redisURL string) error
	InitSpotifyClient(clientID, clientSecret string)
	Close() error

	FilterRussian(ctx context.Context, artists map[string]models.Artist) (map[string]models.Artist, error)
	GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error)
	SetSpotifyToken(ctx context.Context, newToken *oauth2.Token) error
	GetStoredSpotifyToken(ctx context.Context) (*oauth2.Token, error)
	GetRefreshedSpotifyToken(ctx context.Context) (*oauth2.Token, error)
}

func NewRepo(logger *logger.Logger) Repo {
	return &repo{
		logger: logger,
	}
}

type repo struct {
	logger        *logger.Logger
	db            db_client.DBClient
	spotifyClient spotify_client.SpotifyClient
	redis         redis_client.RedisClient

	token *oauth2.Token
}

func (r *repo) translateSpotifyError(err error) error {
	if spotifyErr, ok := err.(spotify.Error); ok {
		r.logger.Debugf("spotify api error: %T: %v, status: %d", spotifyErr, spotifyErr, spotifyErr.Status)
		switch spotifyErr.Status {
		case 404:
			r.logger.Infof("spotify not found: %T: %v", spotifyErr, spotifyErr)
			return ErrNotFound
		case 400:
			r.logger.Infof("spotify bad request: %T: %v", spotifyErr, spotifyErr)
			return ErrBadRequest
		default:
			r.logger.Errorf("spotify error: %T: %v", spotifyErr, spotifyErr)
			return ErrSpotifyAPIError
		}
	}
	if urlErr, ok := err.(*url.Error); ok {
		r.logger.Infof("spotify network error: %T: %v", urlErr, urlErr)
		return ErrBadRequest
	}
	r.logger.Errorf("unknown spotify error: %T: %v", err, err)
	return ErrSpotifyAPIError
}

func (r *repo) InitDB(databaseURL string) error {
	database, err := db_client.NewDBConnection(databaseURL)
	if err != nil {
		return err
	}
	r.db = database
	return nil
}

func (r *repo) InitRedis(redisURL string) error {
	redisClient, err := redis_client.NewRedisClient(redisURL)
	if err != nil {
		return err
	}
	r.redis = redisClient
	return nil
}

func (r *repo) InitSpotifyClient(clientID, clientSecret string) {
	r.spotifyClient = spotify_client.NewSpotifyClient(clientID, clientSecret)
}

func (r *repo) Close() error {
	if r.db != nil {
		err := r.db.Close()
		if err != nil {
			return err
		}
	}
	if r.redis != nil {
		err := r.redis.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *repo) FilterRussian(ctx context.Context, artists map[string]models.Artist) (map[string]models.Artist, error) {
	result, err := r.db.FilterRussian(ctx, artists)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		return nil, ErrDatabaseError
	}
	return result, nil
}

func (r *repo) GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error) {
	playlist, err := r.spotifyClient.GetPlaylistWithTracks(ctx, playlistId)
	if err != nil {
		return nil, r.translateSpotifyError(err)
	}
	return playlist, nil
}

func (r *repo) SetSpotifyToken(ctx context.Context, newToken *oauth2.Token) error {
	err := r.db.SetSpotifyToken(ctx, newToken)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		return ErrDatabaseError
	}
	r.token = newToken
	return nil
}

func (r *repo) GetStoredSpotifyToken(ctx context.Context) (*oauth2.Token, error) {
	token, err := r.db.GetSpotifyToken(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Infof("no stored spotify token found: %T: %v", err, err)
			return nil, ErrNotFound
		}
		r.logger.Errorf("db error: %T: %v", err, err)
		return nil, ErrDatabaseError
	}
	return token, nil
}

func (r *repo) GetRefreshedSpotifyToken(ctx context.Context) (*oauth2.Token, error) {
	token, err := r.spotifyClient.GetRefreshedSpotifyToken(ctx)
	if err != nil {
		return nil, r.translateSpotifyError(err)
	}
	return token, nil
}
