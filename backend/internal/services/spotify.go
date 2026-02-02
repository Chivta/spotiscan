package services

import (
	"context"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/chivta/spotiscan/internal/logger"
	"github.com/chivta/spotiscan/internal/models"
	"github.com/chivta/spotiscan/internal/repository"
)

func NewSpotifyService(logger *logger.Logger, repo repository.Repo) *SpotifyService {
	return &SpotifyService{
		log:  logger,
		repo: repo,
	}
}

type SpotifyService struct {
	log  *logger.Logger
	repo repository.Repo
}

func (s *SpotifyService) GetValidSpotifyToken(ctx context.Context) (*oauth2.Token, error) {
	s.log.Debugf("getting valid spotify token")
	token, err := s.repo.GetStoredSpotifyToken(ctx)
	if err != nil && err != repository.ErrNotFound {
		s.log.Errorf("failed to get stored spotify token: %v", err)
		return nil, s.translateRepositoryError(err)
	}

	if token != nil && token.Expiry.UTC().After(time.Now().UTC()) {
		return token, nil
	}

	s.log.Debugf("token expired or not found, refreshing")
	newToken, err := s.repo.GetRefreshedSpotifyToken(ctx)
	if err != nil || newToken == nil {
		s.log.Errorf("failed to refresh spotify token: %v", err)
		return nil, s.translateRepositoryError(err)
	}

	err = s.repo.SetSpotifyToken(ctx, newToken)
	if err != nil {
		s.log.Errorf("failed to store refreshed spotify token: %v", err)
		return nil, s.translateRepositoryError(err)
	}

	s.log.Debugf("successfully refreshed and stored new token")
	return newToken, nil
}

// formRuContent filters the provided tracks and returns rusContent containing only Russian artists and tracks
// that have at least one Russian artist.
func (s *SpotifyService) formRuContent(ctx context.Context, tracks []models.Track) (*models.RuContent, error) {
	s.log.Debugf("forming RU content for %d tracks", len(tracks))
	var ruContent models.RuContent

	if len(tracks) == 0 {
		s.log.Debugf("no tracks provided, returning empty RU content")
		return &ruContent, nil
	}

	// fill artists map
	artistsMap := make(map[string]models.Artist)
	for _, track := range tracks {
		for _, artist := range track.Artists {
			artistsMap[strings.ToLower(artist.Name)] = artist
		}
	}

	// create list of unique artist names 
	artistNames := make([]string, 0, len(artistsMap))
	for name := range artistsMap {
		artistNames = append(artistNames, name)
	}

	// filter Russian artists using the repository
	ruArtistNames, err := s.repo.FilterRussian(ctx, artistNames)
	if err != nil {
		s.log.Errorf("failed to filter Russian artists: %v", err)
		return nil, ErrDatabaseFailure
	}

	// names should all be in lowercase at this point
	// fill ruArtistsMap
	ruArtistsMap := make(map[string]models.Artist, len(ruArtistNames))
	for _, name := range ruArtistNames {
		ruArtistsMap[name] = artistsMap[name]
	}

	// filter tracks that have at least one Russian artist
	ruTracksMap := make(map[string]models.Track)
	for _, track := range tracks {
		for _, artist := range track.Artists {
			if _, isRu := ruArtistsMap[strings.ToLower(artist.Name)]; isRu {
				ruTracksMap[track.ID] = track
				break
			}
		}
	}
	
	// fill ruContent.Tracks and ruContent.Artists
	ruContent.Tracks = make([]models.Track, 0, len(ruTracksMap))
	for _, track := range ruTracksMap {
		ruContent.Tracks = append(ruContent.Tracks, track)
	}

	ruContent.Artists = make([]models.Artist, 0, len(ruArtistsMap))
	for _, artist := range ruArtistsMap {
		ruContent.Artists = append(ruContent.Artists, artist)
	}

	s.log.Debugf("formed RU content with %d tracks and %d artists", len(ruContent.Tracks), len(ruContent.Artists))
	return &ruContent, nil
}

func (s *SpotifyService) translateRepositoryError(err error) error {
	switch err {
	case nil:
		return nil
	case repository.ErrNotFound:
		return ErrResourceNotFound
	case repository.ErrBadRequest:
		return ErrBadRequest
	case repository.ErrSpotifyAPIError:
		return ErrSpotifyAPIError
	case repository.ErrDatabaseError:
		return ErrDatabaseFailure
	default:
		return ErrInternal
	}
}

func (s *SpotifyService) GetPlaylistRuContent(ctx context.Context, playlistId string) (*models.RuContent, error) {
	s.log.Debugf("getting RU content for playlist %s", playlistId)
	playlist, err := s.repo.GetPlaylistWithTracks(ctx, playlistId)
	if err != nil {
		return nil, s.translateRepositoryError(err)
	}

	ruContent, err := s.formRuContent(ctx, playlist.Tracks)
	if err != nil {
		return nil, s.translateRepositoryError(err)
	}

	s.log.Debugf("successfully retrieved RU content for playlist %s", playlistId)
	return ruContent, nil
}
