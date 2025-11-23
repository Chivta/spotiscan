package spotify

import (
	"context"
	"log"
	"spotiscan/models"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

func NewSpotifyClient(spotifyId, spotifySecret, redirectURI string) *SpotifyClient {
	return &SpotifyClient{
		spotifyId:     spotifyId,
		spotifySecret: spotifySecret,
		redirectURI:   redirectURI,
		auth: spotifyauth.New(
			spotifyauth.WithRedirectURL(redirectURI),
			spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate, spotifyauth.ScopeUserLibraryRead, spotifyauth.ScopePlaylistModifyPublic, spotifyauth.ScopePlaylistModifyPrivate),
			spotifyauth.WithClientID(spotifyId),
			spotifyauth.WithClientSecret(spotifySecret),
		),
	}
}

type SpotifyClient struct {
	spotifyId     string
	spotifySecret string
	redirectURI   string
	auth          *spotifyauth.Authenticator
}

func (c *SpotifyClient) GetAuthURL(state string) string {
	return c.auth.AuthURL(state)
}

func (c *SpotifyClient) DeletePlaylistRuContent(token *oauth2.Token, playlistId string, tracks []models.Track) error {
	ctx := context.Background()
	client := spotify.New(c.auth.Client(ctx, token))

	trackIDs := make([]spotify.ID, 0, len(tracks))
	for _, track := range tracks {
		trackIDs = append(trackIDs, spotify.ID(track.ID))
	}

	_, err := client.RemoveTracksFromPlaylist(ctx, spotify.ID(playlistId), trackIDs...)
	if err != nil {
		log.Printf("couldn't remove tracks from playlist: %v", err)
		return err
	}

	return nil
}

func (c *SpotifyClient) GetUserSavedTracks(token *oauth2.Token) ([]models.Track, error) {
	ctx := context.Background()
	client := spotify.New(c.auth.Client(ctx, token))

	var limit = 50
	var offset = 0
	var allTracks []models.Track
	for {
		page, err := client.CurrentUsersTracks(ctx, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			log.Printf("couldn't fetch user library: %v", err)
			return nil, err
		}

		for _, item := range page.Tracks {
			track := item.FullTrack.SimpleTrack
			var trackArtists []models.Artist
			for _, artist := range track.Artists {
				artist := models.Artist{
					ID:   string(artist.ID),
					Name: artist.Name,
				}

				trackArtists = append(trackArtists, artist)
			}
			allTracks = append(allTracks, models.Track{
				ID:      string(track.ID),
				Name:    track.Name,
				Artists: trackArtists,
			})
		}

		if len(page.Tracks) < limit {
			break
		}

		offset += limit
	}
	log.Println("Total liked songs fetched:", len(allTracks))

	return allTracks, nil
}

func (c *SpotifyClient) getAllPlaylistTracks(client *spotify.Client, playlistId spotify.ID) ([]models.Track, error) {
	var limit = 50
	var offset = 0
	var allTracks []models.Track
	for {
		page, err := client.GetPlaylistItems(context.Background(), playlistId, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			log.Printf("couldn't fetch playlist items: %v", err)
			return nil, err
		}

		for _, item := range page.Items {
			track := item.Track.Track
			var trackArtists []models.Artist
			for _, artist := range track.Artists {
				artist := models.Artist{
					ID:   string(artist.ID),
					Name: artist.Name,
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

func (c *SpotifyClient) GetPlaylistWithTracks(playlistId string, token *oauth2.Token) (*models.Playlist, error) {
	ctx := context.Background()
	client := spotify.New(c.auth.Client(ctx, token))

	userId, err := c.FetchSpotifyUserId(token)
	if err != nil {
		log.Println("couldn't fetch user id:", err)
		userId = "" // ignoring error, will set Owned to false
	}
	spotifyPlaylist, err := client.GetPlaylist(ctx, spotify.ID(playlistId))
	if err != nil {
		log.Printf("couldn't fetch playlist: %v", err)
		return nil, err
	}
	var playlist models.Playlist
	playlist.ID = string(spotifyPlaylist.ID)
	playlist.Name = spotifyPlaylist.Name
	playlist.Owned = spotifyPlaylist.Owner.ID == userId
	// Fill in tracks
	tracks, err := c.getAllPlaylistTracks(client, spotifyPlaylist.ID)
	if err != nil {
		log.Printf("couldn't fetch playlist tracks: %v", err)
		return nil, err
	}
	// remove duplicate tracks
	trackMap := make(map[string]models.Track)
	for _, track := range tracks {
		trackMap[track.ID] = track
	}
	uniqueTracks := make([]models.Track, 0, len(trackMap))
	for _, track := range trackMap {
		uniqueTracks = append(uniqueTracks, track)
	}
	playlist.Tracks = tracks

	return &playlist, nil
}

func (c *SpotifyClient) FetchSpotifyUserId(token *oauth2.Token) (string, error) {
	ctx := context.Background()
	client := spotify.New(c.auth.Client(ctx, token))

	user, err := client.CurrentUser(ctx)
	if err != nil {
		log.Printf("couldn't fetch user: %v", err)
		return "", err
	}

	return user.ID, nil
}

func (c *SpotifyClient) GetUserPlaylists(token *oauth2.Token) ([]models.Playlist, error) {
	ctx := context.Background()
	client := spotify.New(c.auth.Client(ctx, token))

	limit := 50
	offset := 0
	var playlists []models.Playlist
	for {
		page, err := client.CurrentUsersPlaylists(ctx, spotify.Limit(limit), spotify.Offset(offset))
		if err != nil {
			log.Printf("couldn't fetch user playlists: %v", err)
			return nil, err
		}

		for _, p := range page.Playlists {
			imageURL := ""
			if len(p.Images) > 0 {
				imageURL = p.Images[0].URL
			}
			playlists = append(playlists, models.Playlist{
				ID:          string(p.ID),
				Name:        p.Name,
				Description: p.Description,
				ImageURL:    imageURL,
				TrackCount:  int(p.Tracks.Total),
			})
		}

		if len(page.Playlists) < limit {
			break
		}
		offset += limit
	}

	return playlists, nil
}
