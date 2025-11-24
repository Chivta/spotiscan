package services

import (
	"log"
	"spotiscan/models"
	"spotiscan/pkg/db"
	spotifyClient "spotiscan/pkg/spotify"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

func NewSpotifyService(db *db.DB, spotifyClient *spotifyClient.SpotifyClient) *SpotifyService {
	return &SpotifyService{
		db:            db,
		spotifyClient: spotifyClient,
	}
}

type SpotifyService struct {
	db            *db.DB
	spotifyClient *spotifyClient.SpotifyClient
}

func (s *SpotifyService) GetValidUserSpotifyToken(userId int) (*oauth2.Token, error) {
	spotifyToken, err := s.db.GetSpotifyTokensByUserId(userId)
	if err != nil {
		log.Println(err)
		return nil, ErrDatabaseFailure
	}

	expired := spotifyToken.Expiry.UTC().Before(time.Now().UTC())
	if expired {
		newToken, err := s.spotifyClient.RefreshToken(spotifyToken)
		if err != nil {
			log.Println("Failed to refresh tokens:", err)
			return nil, ErrSpotifyAPIError
		}
		err = s.db.StoreSpotifyTokens(userId, newToken)
		if err != nil {
			log.Println("Failed to store refreshed tokens:", err)
			return nil, ErrDatabaseFailure
		}
		spotifyToken = newToken
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
	
	ruArtistsMap, err := s.db.FilterRussian(artistMap)
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

func (s *SpotifyService) GetUserLikedSongsRuContent(oathToken *oauth2.Token) (*models.RuContent, error) {
	tracks, err := s.spotifyClient.GetUserSavedTracks(oathToken)
	if err != nil {
		log.Println(err)
		return nil, ErrSpotifyAPIError
	}

	ruContent, err := s.formRuContent(tracks)
	if err != nil {
		log.Println(err)
		return nil, ErrSpotifyAPIError
	}

	ruContent.AbleToDelete = true
	
	return ruContent, nil
}

func (s *SpotifyService) DeletePlaylistRuContent(oathToken *oauth2.Token, playlistId string, tracks []models.Track) error {
	err := s.spotifyClient.DeletePlaylistRuContent(oathToken, playlistId, tracks)
	if err != nil {
		log.Println(err)
		return ErrSpotifyAPIError
	}
	return nil
}

func (s *SpotifyService) GetPlaylistRuContent(playlistId string, oathToken *oauth2.Token) (*models.RuContent, error) {
	playlist, err := s.spotifyClient.GetPlaylistWithTracks(playlistId, oathToken)
	if err != nil {
		log.Println(err)
		return nil, ErrSpotifyAPIError
	}

	ruContent,err := s.formRuContent(playlist.Tracks)
	if err != nil {
		log.Println(err)
		return nil, ErrSpotifyAPIError
	}
	ruContent.AbleToDelete = playlist.Owned

	return ruContent, nil
}

func (s *SpotifyService) GetUserPlaylists(oathToken *oauth2.Token) ([]models.Playlist, error) {
	playlists, err := s.spotifyClient.GetUserPlaylists(oathToken)
	if err != nil {
		log.Println(err)
		return nil, ErrSpotifyAPIError
	}
	return playlists, nil
}
