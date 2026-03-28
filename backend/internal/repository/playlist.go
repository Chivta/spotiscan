package repository

import (
	"context"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"

	"github.com/chivta/spotiscan/internal/appErrors"
	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"

	"database/sql"

	"github.com/redis/go-redis/v9"
)

func NewPlaylistRepo(logger *logger.Logger, db *sql.DB, redis *redis.Client) *PlaylistRepo {
	return &PlaylistRepo{
		logger:  logger,
		redis:   redis,
	}
}

type PlaylistRepo struct {
	logger  *logger.Logger
	redis   *redis.Client
}

func (r *PlaylistRepo) GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error) {
	token, ok := ctx.Value("spotify_token").(*oauth2.Token)
	if !ok || token == nil {
		r.logger.Errorf("missing or invalid spotify token in context")
		return nil, appErrors.ErrInternal
	}
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)

	spotifyPlaylist, err := client.GetPlaylist(ctx, spotify.ID(playlistId))
	if err != nil {
		return nil, translateSpotifyError(err)
	}
	var playlist models.Playlist
	playlist.ID = string(spotifyPlaylist.ID)
	playlist.Name = spotifyPlaylist.Name
	// Fill in tracks
	tracks, err := r.getAllPlaylistTracks(ctx, client, spotifyPlaylist.ID)
	if err != nil {
		return nil, translateSpotifyError(err)
	}
	// Remove duplicate tracks
	trackMap := make(map[string]models.Track)
	for _, track := range tracks {
		trackMap[track.ID] = track
	}
	uniqueTracks := make([]models.Track, 0, len(trackMap))
	for _, track := range trackMap {
		uniqueTracks = append(uniqueTracks, track)
	}
	playlist.Tracks = uniqueTracks

	return &playlist, nil
}

func (c *PlaylistRepo) getAllPlaylistTracks(ctx context.Context, client *spotify.Client, playlistId spotify.ID) ([]models.Track, error) {
	var limit = 50
	var offset = 0
	var allTracks []models.Track
	for {
		page, err := client.GetPlaylistItems(ctx, playlistId, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			return nil, err
		}

		for _, item := range page.Items {
			track := item.Track.Track
			var trackArtists []models.Artist
			for _, artist := range track.Artists {
				artist := models.Artist{
					ID:         string(artist.ID),
					SpotifyURL: artist.ExternalURLs["spotify"],
					Name:       artist.Name,
				}

				trackArtists = append(trackArtists, artist)
			}
			imageURL := ""
			if len(track.Album.Images) > 0 {
				imageURL = track.Album.Images[0].URL
			}
			allTracks = append(allTracks, models.Track{
				ID:       string(track.ID),
				Name:     track.Name,
				ImageURL: imageURL,
				Artists:  trackArtists,
			})
		}

		if len(page.Items) < limit {
			break
		}

		offset += limit
	}
	return allTracks, nil
}
