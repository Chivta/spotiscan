package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"
)

// LastFMTopArtistsResponse matches the payload from tag.getTopArtists.
type LastFMTopArtistsResponse struct {
	TopArtists LastFMTopArtists `json:"topartists"`
}

type LastFMTopArtists struct {
	Artist []LastFMArtist       `json:"artist"`
	Attr   LastFMTopArtistsAttr `json:"@attr"`
}

type LastFMArtist struct {
	Name       string         `json:"name"`
	URL        string         `json:"url"`
	Streamable string         `json:"streamable"`
	Image      []LastFMImage  `json:"image"`
	Attr       LastFMRankAttr `json:"@attr"`
}

type LastFMImage struct {
	Text string `json:"#text"`
	Size string `json:"size"`
}

type LastFMRankAttr struct {
	Rank string `json:"rank"`
}

type LastFMTopArtistsAttr struct {
	Tag        string `json:"tag"`
	Page       string `json:"page"`
	PerPage    string `json:"perPage"`
	TotalPages string `json:"totalPages"`
	Total      string `json:"total"`
}

func lastFMArtistToModel(a LastFMArtist, tag string) models.Artist {
	return models.Artist{
		Name:          a.Name,
		Source:        "lastfm",
		SourceURL:     a.URL,
		DescriptionEN: fmt.Sprintf("Artist has \"%s\" tag on LastFM", tag),
		DescriptionUA: fmt.Sprintf("Виконавець має \"%s\" тег на LastFM", tag),
		Confirmed:     false,
		Country:       "RU",
	}
}

func fetchLastFMPage(ctx context.Context, method, tag, apiKey string, page int) (*LastFMTopArtistsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(
		"http://ws.audioscrobbler.com/2.0/?method=%s&tag=%s&api_key=%s&format=json&page=%d",
		method, tag, apiKey, page,
	), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for page %d: %w", page, err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page %d: %w", page, err)
	}
	defer res.Body.Close()

	var resp LastFMTopArtistsResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode page %d response: %w", page, err)
	}
	return &resp, nil
}

func scrapeLastFMArtists(ctx context.Context, method string, tag string, apiKey string) ([]models.Artist, error) {
	first, err := fetchLastFMPage(ctx, method, tag, apiKey, 1)
	if err != nil {
		return nil, err
	}

	totalPages, err := strconv.Atoi(first.TopArtists.Attr.TotalPages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse totalPages: %w", err)
	}

	artists := make([]models.Artist, 0)
	for _, a := range first.TopArtists.Artist {
		artists = append(artists, lastFMArtistToModel(a, tag))
	}

	for page := 2; page <= totalPages; page++ {
		resp, err := fetchLastFMPage(ctx, method, tag, apiKey, page)
		if err != nil {
			return nil, err
		}
		for _, a := range resp.TopArtists.Artist {
			artists = append(artists, lastFMArtistToModel(a, tag))
		}
	}

	return artists, nil
}

func scrapeLastFMTopArtistsForAllTags(ctx context.Context, appLogger *logger.Logger, repo artistsRepo, apiKey string) error {
	method := "tag.getTopArtists"

	ruTags, err := repo.GetRuTags(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ru tags: %w", err)
	}

	for _, tag := range ruTags {
		appLogger.Infof("Scraping artists for tag '%s'", tag)
		artists, err := scrapeLastFMArtists(ctx, method, tag, apiKey)
		if err != nil {
			appLogger.Errorf("Failed to scrape artists for tag '%s': %v", tag, err)
			continue
		}
		if err = repo.InsertArtists(ctx, artists); err != nil {
			appLogger.Errorf("Failed to insert artists for tag '%s': %v", tag, err)
			continue
		}
		appLogger.Infof("LastFM tag '%s': inserted %d artists", tag, len(artists))
	}

	return nil
}
