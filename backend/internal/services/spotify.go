package services

import (
	"log"
	"spotiscan/models"
	"spotiscan/pkg/db"
	spotifyClient "spotiscan/pkg/spotify"
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
	log.Println("Spotify token expired:", expired)
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

func isRussianArtist(artistName string) bool {
	// TODO: Implement actual logic to determine if an artist is Russian.
	return len(artistName)%2 == 0
}

func (s *SpotifyService) GetUserLikedSongsRuContent(oathToken *oauth2.Token) (*models.RuContent, error) {
	tracks, err := s.spotifyClient.GetUserSavedTracks(oathToken)
	if err != nil {
		log.Println(err)
		return nil, ErrSpotifyAPIError
	}

	var ruContent models.RuContent
	var ruTracks []models.Track
	var ruArtists []models.Artist

	for _, track := range tracks {
		isRuTrack := false
		for _, artist := range track.Artists {
			if isRussianArtist(artist.Name) {
				ruArtists = append(ruArtists, artist)
				isRuTrack = true
			}
		}
		if isRuTrack {
			ruTracks = append(ruTracks, track)
		}
	}

	// remove duplicate artists
	artistMap := make(map[string]models.Artist)
	for _, artist := range ruArtists {
		artistMap[artist.ID] = artist
	}
	uniqueArtists := make([]models.Artist, 0, len(artistMap))
	for _, artist := range artistMap {
		uniqueArtists = append(uniqueArtists, artist)
	}
	ruContent.RuArtists = uniqueArtists

	// remove duplicate tracks
	trackMap := make(map[string]models.Track)
	for _, track := range ruTracks {
		trackMap[track.ID] = track
	}
	uniqueTracks := make([]models.Track, 0, len(trackMap))
	for _, track := range trackMap {
		uniqueTracks = append(uniqueTracks, track)
	}
	ruContent.RuTracks = uniqueTracks

	return &ruContent, nil
}

func (s *SpotifyService) GetPlaylistRuContent(playlistId string, oathToken *oauth2.Token) (*models.RuContent, error) {
	playlist, err := s.spotifyClient.GetPlaylist(playlistId, oathToken)
	if err != nil {
		log.Println(err)
		return nil, ErrSpotifyAPIError
	}

	var ruContent models.RuContent

	var ruArtists []models.Artist
	var ruTracks []models.Track

	for _, track := range playlist.Tracks {
		for _, artist := range track.Artists {
			if isRussianArtist(artist.Name) {
				ruArtists = append(ruArtists, artist)
				ruTracks = append(ruTracks, track)
			}
		}
	}

	// remove duplicate artists
	artistMap := make(map[string]models.Artist)
	for _, artist := range ruArtists {
		artistMap[artist.ID] = artist
	}
	uniqueArtists := make([]models.Artist, 0, len(artistMap))
	for _, artist := range artistMap {
		uniqueArtists = append(uniqueArtists, artist)
	}
	ruContent.RuArtists = uniqueArtists

	// remove duplicate tracks
	trackMap := make(map[string]models.Track)
	for _, track := range ruTracks {
		trackMap[track.ID] = track
	}
	uniqueTracks := make([]models.Track, 0, len(trackMap))
	for _, track := range trackMap {
		uniqueTracks = append(uniqueTracks, track)
	}
	ruContent.RuTracks = uniqueTracks

	return &ruContent, nil
}
