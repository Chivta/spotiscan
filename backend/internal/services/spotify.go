package services

import (
	"log"
	"net/http"
	"spotiscan/pkg/db"
	"spotiscan/pkg/spotify"
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

func (s *SpotifyService) GetRuArtistsFromPlaylist(playlistId string, userId int) ([]string, error) {
	oathToken, err := s.db.GetSpotifyTokensByUserId(userId)
	if err != nil {
		log.Println(err)
		return nil, ErrDatabaseFailure
	}

	ruArtists, err := s.spotify.GetRuArtistsFromPlaylist(playlistId,oathToken)
	if err != nil {
		log.Println(err)
		return nil, ErrSpotifyAPIError
	}

	return ruArtists, nil
}

func (s *SpotifyService) InitializeSpotifyAuth(userId int) (string,error) {
	state := generateRandomString()
	err := s.db.CreateOathState(state, userId)
	log.Println(userId)
	if err != nil {
		log.Println(err)
		return "", ErrDatabaseFailure
	}
	authUrl := s.spotify.GetAuthURL(state)
	return authUrl, nil
}

func (s *SpotifyService) AcceptCallback(r *http.Request, state string, userId int) error {
	stateUserId, err := s.db.GetUserIdByState(state)
	if err != nil {
		log.Println(err)
		return ErrDatabaseFailure
	}
	if stateUserId == 0 {
		log.Println("invalid oauth state")
		return ErrInvalidState
	}
	if stateUserId != userId {
		log.Println("state user id does not match callback user id")
		return ErrInvalidState
	}

	token, err := s.spotify.AcceptRequest(r,state)
	if err != nil {
		log.Println(err)
		return ErrSpotifyAPIError
	}

	err = s.db.StoreSpotifyTokens(userId, token)
	if err != nil {
		log.Println(err)
		return ErrDatabaseFailure
	}

	err = s.db.DeleteOathState(state)
	if err != nil {
		log.Println(err)
		return ErrDatabaseFailure
	}

	return nil
}
