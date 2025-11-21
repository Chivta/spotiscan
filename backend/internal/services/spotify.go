package services

import (
	"log"
	"spotiscan/pkg/db"
	"spotiscan/pkg/spotify"
	"golang.org/x/oauth2"
	"time"
)

func NewSpotifyService(db *db.DB, spotify *spotify.SpotifyClient) *SpotifyService {
	return &SpotifyService{
		db:      db,
		spotify: spotify,
	}
}

type SpotifyService struct {
	db      *db.DB
	spotify *spotify.SpotifyClient
}

func (s *SpotifyService) GetValidUserSpotifyToken(userId int) (*oauth2.Token, error) {
	spotifyToken, err := s.db.GetSpotifyTokensByUserId(userId)
	if err != nil {
		log.Println(err)
		return nil, ErrDatabaseFailure
	}

	expired := spotifyToken.Expiry.After(time.Now())
	if expired {
		newToken, err := s.spotify.RefreshToken(spotifyToken)
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

func (s *SpotifyService) GetRuArtistsFromPlaylist(playlistId string, userId int, oathToken *oauth2.Token) ([]string, error) {
	ruArtists, err := s.spotify.GetRuArtistsFromPlaylist(playlistId,oathToken)
	if err != nil {
		log.Println(err)
		return nil, ErrSpotifyAPIError
	}

	return ruArtists, nil
}



