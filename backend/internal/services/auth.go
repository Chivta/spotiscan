package services

import (
	"log"
	"spotiscan/pkg/db"
	"spotiscan/pkg/spotify"
	"net/http"
)

func NewAuthService(db *db.DB,spotify *spotify.SpotifyClient) *AuthService {
	return &AuthService{
		db: db,
		spotify: spotify,
	}
}

type AuthService struct {
	db *db.DB
	spotify *spotify.SpotifyClient
}


func (s *AuthService) InitializeSpotifyAuth() (string,error) {
	state := generateRandomString()
	err := s.db.CreateOathState(state)
	if err != nil {
		log.Println(err)
		return "", ErrDatabaseFailure
	}
	authUrl := s.spotify.GetAuthURL(state)
	return authUrl, nil
}

func (s *AuthService) AcceptSpotifyAuthCallback(r *http.Request, state string) (string,error) {
	exists, err := s.db.StateExists(state)
	if err != nil {
		log.Println(err)
		return "",ErrDatabaseFailure
	}
	if !exists {
		log.Println("invalid oauth state")
		return "",ErrInvalidState
	}

	spotifyToken, err := s.spotify.GetToken(r,state)
	if err != nil {
		log.Println(err)
		return "",ErrSpotifyAPIError
	}

	spotifyId, err := s.spotify.FetchSpotifyUserId(spotifyToken)
	if err != nil {
		log.Println(err)
		return "",ErrSpotifyAPIError
	}

	sessionToken := generateRandomString()
	userId, err := s.db.GetUserIdBySpotifyId(spotifyId)
	if err == db.ErrNotFound {
		userId, err = s.db.CreateUserWithSession(spotifyId,sessionToken)
		if err != nil {
			log.Println(err)
			return "",ErrDatabaseFailure
		}
	} else if err != nil {
		log.Println(err)
		return "",ErrDatabaseFailure
	} else {
		err = s.db.CreateSession(userId, sessionToken)
	}

	err = s.db.StoreSpotifyTokens(userId, spotifyToken)
	if err != nil {
		log.Println(err)
		return "",ErrDatabaseFailure
	}

	err = s.db.DeleteOathState(state)
	if err != nil {
		log.Println(err)
		return "",ErrDatabaseFailure
	}

	return sessionToken, nil
}

func (s *AuthService) Logout(sessionToken string) error {
	err := s.db.DeleteSession(sessionToken)
	if err != nil {
		log.Println(err)
		return ErrDatabaseFailure
	}

	return nil
}