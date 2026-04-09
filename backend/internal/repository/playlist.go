package repository

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/models"
	"github.com/chivta/ruscan/internal/spotify"
)

func NewPlaylistRepo(db *sql.DB, spotifyClient *spotify.SpotifyClient) *PlaylistRepo {
	return &PlaylistRepo{
		spotifyClient: spotifyClient,
	}
}

type PlaylistRepo struct {
	spotifyClient *spotify.SpotifyClient
}

func (r *PlaylistRepo) GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error) {
	playlist, err := r.spotifyClient.GetSpotifyPlaylist(ctx, playlistId)
	if err != nil {
		log.Error().Err(err).Str("playlistId", playlistId).Msg("error fetching playlist")
		return nil, translateSpotifyError(err)
	}

	return playlist, nil
}
