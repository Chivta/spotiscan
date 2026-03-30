package repository

import (
	"context"
	"database/sql"

	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/chivta/spotiscan/internal/spotify"
)

func NewPlaylistRepo(log *logger.Logger, db *sql.DB, spotifyClient *spotify.SpotifyClient) *PlaylistRepo {
	return &PlaylistRepo{
		log:   log,
		spotifyClient: spotifyClient,
	}
}

type PlaylistRepo struct {
	log           *logger.Logger
	spotifyClient *spotify.SpotifyClient
}

func (r *PlaylistRepo) GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error) {
	playlist, err := r.spotifyClient.GetSpotifyPlaylist(ctx, playlistId)
	if err != nil {
		r.log.Errorf("error fetching playlist %s: %v", playlistId, err)
		return nil, translateSpotifyError(err)
	}

	return playlist, nil
}
