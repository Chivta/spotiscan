package scanner

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/appErrors"
	"github.com/chivta/ruscan/internal/shared/models"
)

type filterArtistsRepo interface {
	GetRussianWithInfo(ctx context.Context, names []string) ([]models.Artist, error)
}

func NewSpotifyService(artistRepo filterArtistsRepo) *SpotifyService {
	return &SpotifyService{artistRepo: artistRepo}
}

type SpotifyService struct {
	artistRepo filterArtistsRepo
}

func (s *SpotifyService) FilterContent(ctx context.Context, content *models.Content) (*models.RuContent, error) {
	if content == nil {
		return nil, fmt.Errorf("content is nil")
	}

	artistNames := make([]string, len(content.Artists))
	artistNamesMap := make(map[string]models.ArtistRef)
	for i, artist := range content.Artists {
		artistNames[i] = strings.ToLower(artist.Name)
		artistNamesMap[strings.ToLower(artist.Name)] = artist
	}

	ruArtists, err := s.artistRepo.GetRussianWithInfo(ctx, artistNames)
	if err != nil {
		log.Error().Err(err).Msg("failed to filter Russian artists and get info")
		return nil, appErrors.ErrDatabaseFailure
	}

	ruArtistIDsMap := make(map[string]struct{}, len(ruArtists))
	for i := range ruArtists {
		externalID := artistNamesMap[strings.ToLower(ruArtists[i].Name)].ExternalID
		ruArtists[i].SpotifyID = externalID
		ruArtistIDsMap[externalID] = struct{}{}
	}

	ruTracks := make([]models.Track, 0)
	for _, track := range content.Tracks {
		for _, a := range track.ArtistRefs {
			if _, isRu := ruArtistIDsMap[a.ExternalID]; isRu {
				ruTracks = append(ruTracks, track)
				break
			}
		}
	}

	return &models.RuContent{
		Artists: ruArtists,
		Tracks:  ruTracks,
	}, nil
}
