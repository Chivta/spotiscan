package repository

import (
	"context"

	"github.com/chivta/ruscan/internal/shared/domain"
	"github.com/chivta/ruscan/internal/spotify"
)

func NewPlaylistRepo(spotifyClient *spotify.SpotifyClient) *PlaylistRepo {
	return &PlaylistRepo{
		spotifyClient: spotifyClient,
	}
}

type PlaylistRepo struct {
	spotifyClient *spotify.SpotifyClient
}

func (r *PlaylistRepo) GetPlaylistWithTracks(ctx context.Context, playlistId string) (*domain.Content, error) {
	playlist, err := r.spotifyClient.GetSpotifyPlaylist(ctx, playlistId)
	if err != nil {
		return nil, translateSpotifyError(err)
	}

	return playlist, nil
}

func (r *PlaylistRepo) GetTrack(ctx context.Context, trackId string) (*domain.Content, error) {
	track, err := r.spotifyClient.GetSpotifyTrack(ctx, trackId)
	if err != nil {
		return nil, translateSpotifyError(err)
	}

	return track, nil
}

func (r *PlaylistRepo) GetAlbum(ctx context.Context, albumId string) (*domain.Content, error) {
	album, err := r.spotifyClient.GetSpotifyAlbum(ctx, albumId)
	if err != nil {
		return nil, translateSpotifyError(err)
	}

	return album, nil
}

func (r *PlaylistRepo) GetArtist(ctx context.Context, artistId string) (*domain.Content, error) {
	artist, err := r.spotifyClient.GetSpotifyArtist(ctx, artistId)
	if err != nil {
		return nil, translateSpotifyError(err)
	}

	return artist, nil
}
