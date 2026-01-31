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

func (s *SpotifyService) formRuContent(ctx context.Context, tracks []models.Track) (*models.RuContent, error) {
	s.log.Debugf("forming RU content for %d tracks", len(tracks))
	var ruContent models.RuContent

	artistMap := make(map[string]models.Artist)
	for _, track := range tracks {
		for _, artist := range track.Artists {
			artistMap[strings.ToLower(artist.Name)] = artist
		}
	}

	ruArtistsMap, err := s.repo.FilterRussian(ctx, artistMap)
	if err != nil {
		s.log.Errorf("failed to filter Russian artists: %v", err)
		return nil, ErrDatabaseFailure
	}

	trackMap := make(map[string]models.Track)

	for _, track := range tracks {
		for _, artist := range track.Artists {
			if _, exists := ruArtistsMap[strings.ToLower(artist.Name)]; exists {
				trackMap[track.ID] = track
				break
			}
		}
	}

	for _, track := range trackMap {
		ruContent.Tracks = append(ruContent.Tracks, track)
	}

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
	ruContent.AbleToDelete = playlist.Owned

	s.log.Debugf("successfully retrieved RU content for playlist %s", playlistId)
	return ruContent, nil
}
