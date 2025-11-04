package services

func NewPlaylistService() *PlaylistService {
	return &PlaylistService{}
}

type PlaylistService struct{}

func (s *PlaylistService) GetRuArtists(id string) ([]string, error) {
	return []string{"artist1", "artist2", "artist3"}, nil
}
