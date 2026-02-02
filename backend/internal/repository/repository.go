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
	Close() error

	FilterRussian(ctx context.Context, names []string) ([]string, error)
	GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error)
	SetSpotifyToken(ctx context.Context, newToken *oauth2.Token) error
	GetStoredSpotifyToken(ctx context.Context) (*oauth2.Token, error)
	GetRefreshedSpotifyToken(ctx context.Context) (*oauth2.Token, error)
	LoadRussianArtistsToRedis(ctx context.Context)
}

func NewRepo(logger *logger.Logger, db db_client.DBClient, redis redis_client.RedisClient, spotify spotify_client.SpotifyClient) Repo {
	return &repo{
		logger: logger,
		db:     db,
		redis:  redis,
		spotify: spotify,
	}
}

type repo struct {
	logger        *logger.Logger
	db            db_client.DBClient
	spotify spotify_client.SpotifyClient
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
		return ErrSpotifyAPIError
	}
	r.logger.Errorf("unknown spotify error: %T: %v", err, err)
	return ErrSpotifyAPIError
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

func (r *repo) LoadRussianArtistsToRedis(ctx context.Context) {
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

func (r *repo) FilterRussian(ctx context.Context, names []string) ([]string, error) {
	ruNames, err := r.redis.FilterRussianArtistNames(ctx, names)
	if err != nil {
		r.logger.Warnf("redis error: %T: %v", err, err)

		// fallback to db
		ruNames, err = r.filterRussianWithDB(ctx, names)
		if err != nil {
			r.logger.Errorf("db error: %T: %v", err, err)
			return nil, ErrDatabaseError
		}
	}
	return ruNames, nil
}

func (r *repo) filterRussianWithDB(ctx context.Context, names []string) ([]string, error) {
	ruNames, err := r.db.GetRussianArtistNames(ctx, names)
	if err != nil {
		r.logger.Errorf("db error: %T: %v", err, err)
		return nil, ErrDatabaseError
	}
	return ruNames, nil
}

func (r *repo) GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error) {
	playlist, err := r.spotify.GetPlaylistWithTracks(ctx, playlistId)
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
	if r.token != nil {
		return r.token, nil
	}

	token, err := r.db.GetSpotifyToken(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Infof("no stored spotify token found: %T: %v", err, err)
			return nil, ErrNotFound
		}
		r.logger.Errorf("db error: %T: %v", err, err)
		return nil, ErrDatabaseError
	}
	r.token = token
	return token, nil
}

func (r *repo) GetRefreshedSpotifyToken(ctx context.Context) (*oauth2.Token, error) {
	token, err := r.spotify.GetRefreshedSpotifyToken(ctx)
	if err != nil {
		return nil, r.translateSpotifyError(err)
	}
	return token, nil
}
