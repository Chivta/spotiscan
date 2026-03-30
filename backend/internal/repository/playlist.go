package repository

import (
	"context"

	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/chivta/spotiscan/internal/spotify"

	"database/sql"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewPlaylistRepo(log *logger.Logger, db *sql.DB, redis *redis.Client, spotifyClient *spotify.SpotifyClient) *PlaylistRepo {
	return &PlaylistRepo{
		log:   log,
		redis: redis,
		spotifyClient: spotifyClient,
	}
}

type PlaylistRepo struct {
	log           *logger.Logger
	redis         *redis.Client
	spotifyClient *spotify.SpotifyClient

	spotifyBlockedUntil time.Time
}

func (r *PlaylistRepo) GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error) {
	playlist, err := r.spotifyClient.GetSpotifyPlaylist(ctx, playlistId)
	if err != nil {
		r.log.Errorf("error fetching playlist %s: %v", playlistId, err)
		return nil, translateSpotifyError(err)
	}

	return playlist, nil
}
