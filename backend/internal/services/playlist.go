package services

import (
	"spotiscan/pkg/db"
)

func NewPlaylistService(db *db.DB) *PlaylistService {
	return &PlaylistService{
		db: db,
	}
}

type PlaylistService struct{
	db *db.DB
}

func (s *PlaylistService) GetRuArtists(id string) ([]string, error) {
	return []string{"artist1", "artist2", "artist3"}, nil
}
