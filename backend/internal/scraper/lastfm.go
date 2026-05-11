package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/chivta/ruscan/internal/shared/domain"
	"github.com/rs/zerolog/log"
)

type artistsRepo interface {
	InsertArtists(ctx context.Context, artists []domain.Artist) error
	GetRuTags(ctx context.Context) ([]string, error)
	GetRuRegionIds(ctx context.Context) ([]string, error)
}

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

func lastFMArtistToModel(a LastFMArtist, tag string) domain.Artist {
	return domain.Artist{
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

func scrapeLastFMArtists(ctx context.Context, method string, tag string, apiKey string) ([]domain.Artist, error) {
	first, err := fetchLastFMPage(ctx, method, tag, apiKey, 1)
	if err != nil {
		return nil, err
	}

	totalPages, err := strconv.Atoi(first.TopArtists.Attr.TotalPages)
	if err != nil {
		return nil, fmt.Errorf("failed to parse totalPages: %w", err)
	}

	artists := make([]domain.Artist, 0)
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

func ScrapeLastFMTopArtistsForAllTags(ctx context.Context, repo artistsRepo, apiKey string) error {
	method := "tag.getTopArtists"

	ruTags, err := repo.GetRuTags(ctx)
	if err != nil {
		return fmt.Errorf("failed to get ru tags: %w", err)
	}

	for _, tag := range ruTags {
		log.Info().Str("tag", tag).Msg("Scraping artists for tag")
		artists, err := scrapeLastFMArtists(ctx, method, tag, apiKey)
		if err != nil {
			log.Error().Err(err).Str("tag", tag).Msg("Failed to scrape artists for tag")
			continue
		}
		if err = repo.InsertArtists(ctx, artists); err != nil {
			log.Error().Err(err).Str("tag", tag).Msg("Failed to insert artists for tag")
			continue
		}
		log.Info().Str("tag", tag).Int("count", len(artists)).Msg("LastFM tag: inserted artists")
	}

	return nil
}
