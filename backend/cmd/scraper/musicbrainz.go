package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/chivta/spotiscan/internal/models"
)

const (
	mbBaseURL   = "https://musicbrainz.org/ws/2"
	mbPageSize  = 100
	mbUserAgent = "spotiscan/1.0 (https://spotiscan.chivtar.dev)"
)

type MusicBrainzArtistsResponse struct {
	Count   int                 `json:"artist-count"`
	Offset  int                 `json:"artist-offset"`
	Artists []MusicBrainzArtist `json:"artists"`
}

type MusicBrainzArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func musicBrainzArtistToModel(a MusicBrainzArtist) models.Artist {
	return models.Artist{
		Name:          a.Name,
		Source:        "musicbrainz",
		SourceURL:     fmt.Sprintf("https://musicbrainz.org/artist/%s", a.ID),
		DescriptionEN: "Artist is from Russia according to MusicBrainz",
		DescriptionUA: "Виконавець з Росії згідно з MusicBrainz",
		Confirmed:     false,
		Country:       "RU",
	}
}

// doWithRateLimit performs the request and retries on 429/503, honouring Retry-After.
func doWithRateLimit(ctx context.Context, req *http.Request) (*http.Response, error) {
	for {
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != http.StatusTooManyRequests && res.StatusCode != http.StatusServiceUnavailable {
			return res, nil
		}
		res.Body.Close()

		wait := 5 * time.Second
		if retryAfter := res.Header.Get("Retry-After"); retryAfter != "" {
			if secs, err := strconv.Atoi(retryAfter); err == nil {
				wait = time.Duration(secs) * time.Second
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
}

func fetchMusicBrainzPage(ctx context.Context, areaID string, offset int) (*MusicBrainzArtistsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(
		"%s/artist?area=%s&limit=%d&offset=%d&fmt=json",
		mbBaseURL, areaID, mbPageSize, offset,
	), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request at offset %d: %w", offset, err)
	}
	req.Header.Set("User-Agent", mbUserAgent)

	res, err := doWithRateLimit(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch offset %d: %w", offset, err)
	}
	defer res.Body.Close()

	var resp MusicBrainzArtistsResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response at offset %d: %w", offset, err)
	}
	return &resp, nil
}

func scrapeMusicBrainzArtistsByArea(ctx context.Context, areaID string) ([]models.Artist, error) {
	first, err := fetchMusicBrainzPage(ctx, areaID, 0)
	if err != nil {
		return nil, err
	}

	artists := make([]models.Artist, 0, first.Count)
	for _, a := range first.Artists {
		artists = append(artists, musicBrainzArtistToModel(a))
	}

	for offset := mbPageSize; offset < first.Count; offset += mbPageSize {
		resp, err := fetchMusicBrainzPage(ctx, areaID, offset)
		if err != nil {
			return nil, err
		}
		for _, a := range resp.Artists {
			artists = append(artists, musicBrainzArtistToModel(a))
		}
	}

	return artists, nil
}

func scrapeMusicBrainzArtistsForAllRegions(ctx context.Context, repo artistsRepo) error {
	regionIDs, err := repo.GetRuRegionIds(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ru region ids: %w", err)
	}

	var totalInserted int
	for _, id := range regionIDs {
		artists, err := scrapeMusicBrainzArtistsByArea(ctx, id)
		if err != nil {
			log.Error().Err(err).Str("areaId", id).Msg("Failed to scrape MusicBrainz artists for area")
			continue
		}
		if err := repo.InsertArtists(ctx, artists); err != nil {
			log.Error().Err(err).Str("areaId", id).Msg("Failed to insert MusicBrainz artists for area")
			continue
		}
		totalInserted += len(artists)
	}
	log.Info().Int("total", totalInserted).Int("regions", len(regionIDs)).Msg("MusicBrainz: inserted artists")

	return nil
}
