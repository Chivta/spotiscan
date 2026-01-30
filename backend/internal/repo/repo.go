package repo

import (
	"golang.org/x/oauth2"

	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/chivta/spotiscan/internal/repo/db"
	"github.com/chivta/spotiscan/internal/repo/redis"
)

type Repo interface {
	InitDB(DatabaseURL string) error
	InitRedis(redisURL string) error
	Close() error
	FilterRussian(artists map[string]models.Artist) (map[string]models.Artist, error)
	StoreSpotifyTokens(token *oauth2.Token) error
	GetSpotifyTokens() (*oauth2.Token, error)
}

func NewRepo(logger *logger.Logger) Repo {
	return &repo{
		logger: logger,
	}
}

type repo struct {
	logger *logger.Logger
	db db.DB
	redis redis.RedisClient
}

func (r *repo) InitDB(DatabaseURL string) error {
	database, err := db.NewDBConnection(DatabaseURL)
	if err != nil {
		return err
	}
	r.db = database
	return nil
}

func (r *repo) InitRedis(redisURL string) error {
	redisClient, err := redis.NewRedisClient(redisURL)
	if err != nil {
		return err
	}
	r.redis = redisClient
	return nil
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
	return r.db.FilterRussian(artists)
}

func (r *repo) StoreSpotifyTokens(token *oauth2.Token) error {
	return r.db.StoreSpotifyTokens(token)
}

func (r *repo) GetSpotifyTokens() (*oauth2.Token, error) {
	return r.db.GetSpotifyTokens()
}