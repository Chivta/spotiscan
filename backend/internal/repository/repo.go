package repository

import (
	"context"
	"golang.org/x/oauth2"

	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/chivta/spotiscan/internal/repository/db_client"
	"github.com/chivta/spotiscan/internal/repository/redis_client"
	"github.com/chivta/spotiscan/internal/repository/spotify_client"
)

type Repo interface {
	InitDB(DatabaseURL string) error
	InitRedis(redisURL string) error
	InitSpotifyClient(clientID, clientSecret string)
	Close() error

	FilterRussian(artists map[string]models.Artist) (map[string]models.Artist, error)
	StoreSpotifyTokens(token *oauth2.Token) error
	GetSpotifyTokens() (*oauth2.Token, error)
	GetToken() (*oauth2.Token, error)
	GetPlaylistWithTracks(playlistId string, token *oauth2.Token) (*models.Playlist, error)
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
}

func (r *repo) InitDB(DatabaseURL string) error {
	database, err := db_client.NewDBConnection(DatabaseURL)
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

func (r *repo) FilterRussian(artists map[string]models.Artist) (map[string]models.Artist, error) {
	return r.db.FilterRussian(context.Background(), artists)
}

func (r *repo) StoreSpotifyTokens(token *oauth2.Token) error {
	return r.db.StoreSpotifyTokens(context.Background(), token)
}

func (r *repo) GetSpotifyTokens() (*oauth2.Token, error) {
	return r.db.GetSpotifyTokens(context.Background())
}

func (r *repo) GetPlaylistWithTracks(playlistId string, token *oauth2.Token) (*models.Playlist, error) {
	return r.spotifyClient.GetPlaylistWithTracks(context.Background(), playlistId, token)
}

func (r *repo) GetToken() (*oauth2.Token, error) {
	return r.spotifyClient.GetToken(context.Background())
}