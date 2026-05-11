package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/domain"
)

const (
	spotifyTokenURL   = "https://accounts.spotify.com/api/token"
	spotifyLimit      = 50
	requestsPerWindow = 175
	slidingWindowSize = 30 * time.Second
	maxConcurrent     = 5
)

func NewSpotifyClient(spotifyId, spotifySecret string) *SpotifyClient {
	return &SpotifyClient{
		httpClient:    http.DefaultClient,
		spotifyId:     spotifyId,
		spotifySecret: spotifySecret,
		sem:           make(chan struct{}, maxConcurrent),
		window:        NewSlidingWindow(slidingWindowSize, requestsPerWindow),
	}
}

type SpotifyClient struct {
	httpClient    *http.Client
	spotifyId     string
	spotifySecret string

	spotifyBlockedUntil time.Time
	blockMu             sync.RWMutex
	sem                 chan struct{}
	window              *SlidingWindow

	tokenExpiry time.Time
	accessToken string
	tokenMu     sync.RWMutex
}

type SpotifyError struct {
	Status int
}

func (e SpotifyError) Error() string {
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
		return "", time.Time{}, &SpotifyError{Status: resp.StatusCode}
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

func (c *SpotifyClient) isSpotifyBlocked() bool {
	c.blockMu.RLock()
	defer c.blockMu.RUnlock()
	return c.spotifyBlockedUntil.After(time.Now())
}

// doRequest executes an HTTP request, handling 429 blocking. On error the body
// is closed internally; callers must close the body only on success.
func (c *SpotifyClient) doRequest(req *http.Request) (*http.Response, error) {
	if c.isSpotifyBlocked() {
		return nil, &SpotifyError{Status: http.StatusTooManyRequests}
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			if secs, err := strconv.Atoi(retryAfter); err == nil {
				log.Warn().Int("retrySeconds", secs).Msg("Spotify API rate limit hit, blocking requests")
				c.blockSpotifyRequests(time.Duration(secs) * time.Second)
			}
		}
		resp.Body.Close()
		return nil, &SpotifyError{Status: http.StatusTooManyRequests}
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, &SpotifyError{Status: resp.StatusCode}
	}
	return resp, nil
}

func (c *SpotifyClient) GetSpotifyPlaylist(ctx context.Context, playlistId string) (*domain.Content, error) {
	token, err := c.getValidToken(ctx)
	if err != nil {
		return nil, err
	}

	if err := c.window.Wait(ctx); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/playlists/"+playlistId, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var playlist SpotifyPlaylistResponse
	if err := json.NewDecoder(resp.Body).Decode(&playlist); err != nil {
		return nil, err
	}

	result := &domain.Content{}
	addedTracks := make(map[string]struct{}, len(playlist.Tracks.Items))
	result.Tracks = make([]domain.Track, 0, len(playlist.Tracks.Items))
	addedArtists := make(map[string]struct{}, len(playlist.Tracks.Items))
	result.Artists = make([]domain.ArtistRef, 0, len(playlist.Tracks.Items))

	c.translateItemsToContent(playlist.Tracks.Items, result, addedTracks, addedArtists)

	var (
		total       = playlist.Tracks.Total
		offset      = len(playlist.Tracks.Items)
		fetchErrors []error
		mu          sync.Mutex
		wg          sync.WaitGroup
	)

	for offset < total {
		if err := c.window.Wait(ctx); err != nil {
			fetchErrors = append(fetchErrors, err)
			break
		}
		wg.Add(1)
		go func(offset int) {
			c.sem <- struct{}{}
			defer func() { <-c.sem }()
			defer wg.Done()

			url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks?offset=%d&limit=%d", playlistId, offset, spotifyLimit)
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				mu.Lock()
				fetchErrors = append(fetchErrors, err)
				mu.Unlock()
				return
			}
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := c.doRequest(req)
			if err != nil {
				mu.Lock()
				fetchErrors = append(fetchErrors, err)
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			var page SpotifyPlaylistItemsResponse
			if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
				mu.Lock()
				fetchErrors = append(fetchErrors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			c.translateItemsToContent(page.Items, result, addedTracks, addedArtists)
			mu.Unlock()
		}(offset)
		offset += spotifyLimit
	}
	wg.Wait()
	if len(fetchErrors) > 0 {
		return nil, fetchErrors[0]
	}

	return result, nil
}

func (c *SpotifyClient) translateItemsToContent(items []SpotifyItem, targetContent *domain.Content, addedTracks map[string]struct{}, addedArtists map[string]struct{}) {
	for _, item := range items {
		if item.Track == nil || item.Track.ID == "" || item.IsLocal {
			continue
		}
		if _, exists := addedTracks[item.Track.ID]; exists {
			continue
		}
		var track domain.Track
		track.ExternalID = item.Track.ID
		track.Name = item.Track.Name
		for _, artist := range item.Track.Artists {
			track.ArtistRefs = append(track.ArtistRefs, domain.ArtistRef{
				ExternalID: artist.ID,
				Name:       artist.Name,
			})
			if _, exists := addedArtists[artist.ID]; !exists {
				addedArtists[artist.ID] = struct{}{}
				targetContent.Artists = append(targetContent.Artists, domain.ArtistRef{
					ExternalID: artist.ID,
					Name:       artist.Name,
				})
			}
		}
		if len(item.Track.Album.Images) > 0 {
			track.ImageURL = item.Track.Album.Images[0].URL
		}
		addedTracks[item.Track.ID] = struct{}{}
		targetContent.Tracks = append(targetContent.Tracks, track)
	}
}

func (c *SpotifyClient) GetSpotifyTrack(ctx context.Context, trackId string) (*domain.Content, error) {
	token, err := c.getValidToken(ctx)
	if err != nil {
		return nil, err
	}

	if err := c.window.Wait(ctx); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/tracks/"+trackId, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var trackResp SpotifyTrackResponse
	if err := json.NewDecoder(resp.Body).Decode(&trackResp); err != nil {
		return nil, err
	}

	track := domain.Track{
		ExternalID: trackResp.ID,
		Name:       trackResp.Name,
	}
	artists := make([]domain.ArtistRef, 0, len(trackResp.Artists))
	if len(trackResp.Album.Images) > 0 {
		track.ImageURL = trackResp.Album.Images[0].URL
	}
	for _, artist := range trackResp.Artists {
		track.ArtistRefs = append(track.ArtistRefs, domain.ArtistRef{
			ExternalID: artist.ID,
			Name:       artist.Name,
		})
		artists = append(artists, domain.ArtistRef{
			ExternalID: artist.ID,
			Name:       artist.Name,
		})
	}
	return &domain.Content{
		Tracks:  []domain.Track{track},
		Artists: artists,
	}, nil
}

func (c *SpotifyClient) GetSpotifyAlbum(ctx context.Context, albumId string) (*domain.Content, error) {
	token, err := c.getValidToken(ctx)
	if err != nil {
		return nil, err
	}

	if err := c.window.Wait(ctx); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/albums/"+albumId, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var albumResp SpotifyAlbumResponse
	if err := json.NewDecoder(resp.Body).Decode(&albumResp); err != nil {
		return nil, err
	}

	var imageURL string
	if len(albumResp.Images) > 0 {
		imageURL = albumResp.Images[0].URL
	}

	result := &domain.Content{
		Tracks:  make([]domain.Track, 0, albumResp.Tracks.Total),
		Artists: make([]domain.ArtistRef, 0),
	}
	seenArtists := make(map[string]struct{})
	c.translateAlbumTracksToContent(albumResp.Tracks.Items, result, seenArtists, imageURL)

	var (
		total       = albumResp.Tracks.Total
		offset      = len(albumResp.Tracks.Items)
		fetchErrors []error
		mu          sync.Mutex
		wg          sync.WaitGroup
	)

	for offset < total {
		if err := c.window.Wait(ctx); err != nil {
			fetchErrors = append(fetchErrors, err)
			break
		}
		wg.Add(1)
		go func(offset int) {
			c.sem <- struct{}{}
			defer func() { <-c.sem }()
			defer wg.Done()

			url := fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks?offset=%d&limit=%d", albumId, offset, spotifyLimit)
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				mu.Lock()
				fetchErrors = append(fetchErrors, err)
				mu.Unlock()
				return
			}
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := c.doRequest(req)
			if err != nil {
				mu.Lock()
				fetchErrors = append(fetchErrors, err)
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			var page struct {
				Items []SpotifyAlbumTrack `json:"items"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
				mu.Lock()
				fetchErrors = append(fetchErrors, err)
				mu.Unlock()
				return
			}
			mu.Lock()
			c.translateAlbumTracksToContent(page.Items, result, seenArtists, imageURL)
			mu.Unlock()
		}(offset)
		offset += spotifyLimit
	}
	wg.Wait()
	if len(fetchErrors) > 0 {
		return nil, fetchErrors[0]
	}

	return result, nil
}

func (c *SpotifyClient) translateAlbumTracksToContent(items []SpotifyAlbumTrack, targetContent *domain.Content, seenArtists map[string]struct{}, imageURL string) {
	for _, item := range items {
		if item.ID == "" {
			continue
		}
		track := domain.Track{
			ExternalID: item.ID,
			Name:       item.Name,
			ImageURL:   imageURL,
		}
		for _, artist := range item.Artists {
			track.ArtistRefs = append(track.ArtistRefs, domain.ArtistRef{
				ExternalID: artist.ID,
				Name:       artist.Name,
			})
			if _, seen := seenArtists[artist.ID]; !seen {
				seenArtists[artist.ID] = struct{}{}
				targetContent.Artists = append(targetContent.Artists, domain.ArtistRef{
					ExternalID: artist.ID,
					Name:       artist.Name,
				})
			}
		}
		targetContent.Tracks = append(targetContent.Tracks, track)
	}
}

func (c *SpotifyClient) GetSpotifyArtist(ctx context.Context, artistId string) (*domain.Content, error) {
	token, err := c.getValidToken(ctx)
	if err != nil {
		return nil, err
	}

	if err := c.window.Wait(ctx); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/artists/"+artistId, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var artistResp SpotifyArtistResponse
	if err := json.NewDecoder(resp.Body).Decode(&artistResp); err != nil {
		return nil, err
	}

	return &domain.Content{
		Artists: []domain.ArtistRef{{
			ExternalID: artistResp.ID,
			Name:       artistResp.Name,
		}},
	}, nil
}
