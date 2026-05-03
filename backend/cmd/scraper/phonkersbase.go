package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/models"
)

const (
	pdbUrl       = "https://www.phonkersbase.com/api/artists"
	pdbPageSize  = 50
	pdbUserAgent = "ruscan/1.0 (https://ruscan.chivtar.dev)"
)

type PhonkersDBResponse struct {
	Data   PhonkersDBData `json:"data"`
	Errors []string       `json:"errors"`
}

type PhonkersDBData struct {
	Items []PhonkersDBArtist `json:"items"`
	Info  PhonkersDBInfo     `json:"info"`
}

type PhonkersDBArtist struct {
	Name          string                  `json:"name"`
	Link          string                  `json:"link"`
	DescriptionUA *string                 `json:"description"`
	DescriptionEN *string                 `json:"descriptionEn"`
	Countries     []PhonkersDBCountry     `json:"countries"`
	ListenLabels  []PhonkersDBListenLabel `json:"listenLabels"`
}

type PhonkersDBCountry struct {
	Name string `json:"name"`
}

type PhonkersDBListenLabel struct {
	Name string `json:"name"` // "blocked" for russians
}

type PhonkersDBInfo struct {
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// in phonkersbase, countries are just strings + russia is written as ruzzia
func translateCountry(country string) string {
	switch country {
	case "ruzzia", "russia":
		return "RU"
	case "ukraine":
		return "UA"
	case "kazakstan":
		return "KZ"
	case "kyrgyzstan":
		return "KG"
	case "belarus":
		return "BY"
	default:
		return "RU" // ru by default
	}
}

func phonkersArtistToModel(a PhonkersDBArtist) models.Artist {
	var descEn string
	if a.DescriptionEN == nil {
		descEn = "Artist has \"Don't listen ❌\" label on Phonkersbase"
	} else {
		descEn = *a.DescriptionEN
	}

	var descUa string
	if a.DescriptionUA == nil {
		descUa = "Виконавець має \"Не слухай це ❌\" позначку на Phonkersbase"
	} else {
		descUa = *a.DescriptionUA
	}

	var country string
	if len(a.Countries) > 0 {
		country = translateCountry(a.Countries[0].Name)
	} else {
		country = "RU" // default to RU if no country provided
	}
	return models.Artist{
		Name:          a.Name,
		Source:        "phonkersbase",
		DescriptionEN: descEn,
		DescriptionUA: descUa,
		Confirmed:     true,
		Country:       country,
	}
}

func fetchPhonkersDBPage(ctx context.Context, offset int) (*PhonkersDBResponse, error) {
	url := fmt.Sprintf("%s?limit=%d&offset=%d", pdbUrl, pdbPageSize, offset)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request at offset %d: %w", offset, err)
	}
	req.Header.Set("User-Agent", pdbUserAgent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch offset %d: %w", offset, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body at offset %d: %w", offset, err)
	}

	var resp PhonkersDBResponse
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response at offset %d: %w", offset, err)
	}
	return &resp, nil
}

// checks if artist has "blocked" label, which corresponds to "dont listen it" on phonkersbase
func containsBlockedLabel(labels []PhonkersDBListenLabel) bool {
	for _, l := range labels {
		if l.Name == "blocked" {
			return true
		}
	}
	return false
}

func scrapePhonkersDBartists(ctx context.Context) ([]models.Artist, error) {
	first, err := fetchPhonkersDBPage(ctx, 0)
	if err != nil {
		return nil, err
	}
	log.Info().Int("total", first.Data.Info.Total).Msg("Total artists in PhonkersDB")

	artists := make([]models.Artist, 0, first.Data.Info.Total)
	for _, a := range first.Data.Items {
		if containsBlockedLabel(a.ListenLabels) {
			artists = append(artists, phonkersArtistToModel(a))
		}
	}
	total := first.Data.Info.Total

	for offset := pdbPageSize; offset < total; offset += pdbPageSize {
		resp, err := fetchPhonkersDBPage(ctx, offset)
		if err != nil {
			return nil, err
		}
		for _, a := range resp.Data.Items {
			if containsBlockedLabel(a.ListenLabels) {
				artists = append(artists, phonkersArtistToModel(a))
			}
		}
	}

	return artists, nil
}

func scrapePhonkersDB(ctx context.Context, repo artistsRepo) error {
	artists, err := scrapePhonkersDBartists(ctx)
	if err != nil {
		return fmt.Errorf("failed to scrape PhonkersBase artists: %w", err)
	}
	if err := repo.InsertArtists(ctx, artists); err != nil {
		return fmt.Errorf("failed to insert PhonkersBase artists: %w", err)
	}
	log.Info().Int("count", len(artists)).Msg("Successfully inserted artists from PhonkersBase")
	return nil
}
