package scanner

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/shared/appErrors"
	"github.com/chivta/ruscan/internal/shared/models"
)

type (
	filterArtistsRepo interface {
		GetRussianWithInfo(ctx context.Context, names []string) ([]models.Artist, error)
	}
	jobRepo interface {
		PostJobResult(ctx context.Context, jobId string, ruContent *models.RuContent) error
	}
)

func NewSpotifyService(artistRepo filterArtistsRepo, jobRepo jobRepo) *SpotifyService {
	return &SpotifyService{
		artistRepo: artistRepo,
		jobRepo:    jobRepo,
	}
}

type SpotifyService struct {
	artistRepo filterArtistsRepo
	jobRepo    jobRepo
}

func (s *SpotifyService) ProcessContentScanJob(ctx context.Context, content *models.Content, jobId string) bool {
	ruContent, err := s.filterContent(ctx, content)
	if err != nil {
		log.Error().Err(err).Msg("failed to filter content")
		return false
	}
	err = s.jobRepo.PostJobResult(ctx, jobId, ruContent)
	if err != nil {
		log.Error().Err(err).Msg("failed to post job result")
		return false
	}
	return true
}

func (s *SpotifyService) filterContent(ctx context.Context, content *models.Content) (*models.RuContent, error) {
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

	// going over all tracks and checking for ru artists
	ruTracks := make([]models.Track, 0)
	for _, track := range content.Tracks {
		// detecting any ru artists in the track
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
