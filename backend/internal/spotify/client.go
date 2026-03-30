package spotify

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"
)

const (
	spotifyTokenURL = "https://accounts.spotify.com/api/token"
)

func NewSpotifyClient(spotifyId, spotifySecret string, appLogger *logger.Logger) *SpotifyClient {
	return &SpotifyClient{
		httpClient:    http.DefaultClient,
		spotifyId:     spotifyId,
		spotifySecret: spotifySecret,
		log:           appLogger,
	}
}

type SpotifyClient struct {
	log           *logger.Logger
	httpClient    *http.Client
	spotifyId     string
	spotifySecret string

	spotifyBlockedUntil time.Time
	blockMu             sync.RWMutex

	tokenExpiry time.Time
	accessToken string
	tokenMu     sync.RWMutex
}

type Error struct {
	Status int
}

func (e Error) Error() string {
	return "spotify API error: status code " + strconv.Itoa(e.Status)
}

func (c *SpotifyClient) getValidToken(ctx context.Context) (string, error) {
	c.tokenMu.RLock()
	if c.accessToken != "" && c.tokenExpiry.UTC().After(time.Now().UTC()) {
		token := c.accessToken
		c.tokenMu.RUnlock()
		return token, nil
	}
	c.tokenMu.RUnlock()
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	// Recheck after acquiring write lock in case another goroutine refreshed the token
	if c.accessToken != "" && c.tokenExpiry.UTC().After(time.Now().UTC()) {
		return c.accessToken, nil
	}
	
	token, expiry, err := c.getToken(ctx)
	if err != nil {
		return "", err
	}

	c.accessToken = token
	c.tokenExpiry = expiry
	return c.accessToken, nil
}

func (c *SpotifyClient) getToken(ctx context.Context) (string, time.Time, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", spotifyTokenURL, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", time.Time{}, err
	}
	req.SetBasicAuth(c.spotifyId, c.spotifySecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, &Error{Status: resp.StatusCode}
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenResp.AccessToken, time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second), nil
}

func (c *SpotifyClient) blockSpotifyRequests(duration time.Duration) {
	c.blockMu.Lock()
	defer c.blockMu.Unlock()
	c.spotifyBlockedUntil = time.Now().Add(duration)
}

func (c *SpotifyClient) GetSpotifyPlaylist(ctx context.Context, playlistId string) (*models.Playlist, error) {
	c.blockMu.RLock()
	spotifyBlocked := c.spotifyBlockedUntil.After(time.Now())
	c.blockMu.RUnlock()
	if spotifyBlocked {
		return nil, &Error{Status: http.StatusTooManyRequests}
	}

	token, err := c.getValidToken(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/playlists/"+playlistId, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// block further requests if we hit rate limit, using Retry-After header if available
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				retrySeconds, err := strconv.Atoi(retryAfter)
				if err == nil {
					c.log.Warnf("Spotify API rate limit hit, blocking requests for %d seconds", retrySeconds)
					c.blockSpotifyRequests(time.Duration(retrySeconds) * time.Second)
				}
			}
		}
		return nil, &Error{Status: resp.StatusCode}
	}

	var playlist SpotifyPlaylistResponse
	err = json.NewDecoder(resp.Body).Decode(&playlist)
	if err != nil {
		return nil, err
	}

	var result models.Playlist
	result.ID = playlist.ID
	result.Name = playlist.Name
	result.Description = playlist.Description
	if len(playlist.Images) > 0 {
		result.ImageURL = playlist.Images[0].URL
	}

	addedTracks := make(map[string]struct{}, len(playlist.Tracks.Items))
	result.Tracks = make([]models.Track, 0, len(playlist.Tracks.Items))
	c.translateItemsToTracks(playlist.Tracks.Items, &result.Tracks, addedTracks)

	for playlist.Tracks.Next != nil {
		req, err := http.NewRequestWithContext(ctx, "GET", *playlist.Tracks.Next, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, &Error{Status: resp.StatusCode}
		}

		var page SpotifyPlaylistItemsResponse
		err = json.NewDecoder(resp.Body).Decode(&page)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}
		c.translateItemsToTracks(page.Items, &result.Tracks, addedTracks)

		playlist.Tracks.Next = page.Next
		resp.Body.Close()
	}

	return &result, nil
}

func (c *SpotifyClient) translateItemsToTracks(items []SpotifyItem, targetTracks *[]models.Track, addedTracks map[string]struct{}) {
	for _, item := range items {
		if item.Track == nil || item.Track.ID == "" || item.IsLocal {
			continue // Skip local tracks, invalid data
		}
		_, exists := addedTracks[item.Track.ID]
		if exists {
			continue // Skip duplicate tracks
		}
		var track models.Track
		track.ID = item.Track.ID
		track.Name = item.Track.Name
		for _, artist := range item.Track.Artists {
			track.Artists = append(track.Artists, models.Artist{
				ID:         artist.ID,
				Name:       artist.Name,
				SpotifyURL: "https://open.spotify.com/artist/" + artist.ID,
			})
		}
		if len(item.Track.Album.Images) > 0 {
			track.ImageURL = item.Track.Album.Images[0].URL
		}
		addedTracks[item.Track.ID] = struct{}{}
		*targetTracks = append(*targetTracks, track)
	}
}
