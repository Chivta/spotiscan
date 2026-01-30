package services

import (
	"log"
	"strings"
	"time"
	"golang.org/x/oauth2"

	"github.com/chivta/spotiscan/internal/repository"
	"github.com/chivta/spotiscan/internal/models"
)

func NewSpotifyService(repo repository.Repo) *SpotifyService {
	return &SpotifyService{
		repo:            repo,
	}
}

type SpotifyService struct {
	repo            repository.Repo
}

func (s *SpotifyService) GetValidSpotifyToken() (*oauth2.Token, error) {
	spotifyToken, err := s.repo.GetSpotifyTokens()
	if err == repository.ErrNotFound || spotifyToken.Expiry.UTC().Before(time.Now().UTC()) {
		newToken, err := s.repo.GetToken()
		if err != nil {
			log.Println("Failed to refresh tokens:", err)
			return nil, ErrSpotifyAPIError
		}
		err = s.repo.StoreSpotifyTokens(newToken)
		if err != nil {
			log.Println("Failed to store refreshed tokens:", err)
			return nil, ErrDatabaseFailure
		}
		spotifyToken = newToken
	} else if err != nil {
		log.Println("Failed to get tokens from DB:", err)
		return nil, ErrDatabaseFailure
	}

	return spotifyToken, nil
}

func (s *SpotifyService) formRuContent(tracks []models.Track) (*models.RuContent, error) {
	var ruContent models.RuContent

	artistMap := make(map[string]models.Artist)
	for _, track := range tracks {
		for _, artist := range track.Artists {
			artistMap[strings.ToLower(artist.Name)] = artist
		}
	}

	ruArtistsMap, err := s.repo.FilterRussian(artistMap)
	if err != nil {
		log.Println("Failed to filter Russian artists:", err)
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

	return &ruContent, nil
}

func (s *SpotifyService) GetPlaylistRuContent(playlistId string, oathToken *oauth2.Token) (*models.RuContent, error) {
	playlist, err := s.repo.GetPlaylistWithTracks(playlistId, oathToken)
	if err != nil {
		log.Println(err)
		return nil, ErrSpotifyAPIError
	}

	ruContent, err := s.formRuContent(playlist.Tracks)
	if err != nil {
		log.Println(err)
		return nil, ErrSpotifyAPIError
	}
	ruContent.AbleToDelete = playlist.Owned

	return ruContent, nil
}
