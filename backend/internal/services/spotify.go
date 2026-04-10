package services

import (
	"context"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/appErrors"
	"github.com/chivta/ruscan/internal/metrics"
	"github.com/chivta/ruscan/internal/models"
)

type (
	filterArtistsRepo interface {
		FilterRussian(ctx context.Context, names []string) ([]string, error)
		GetArtistsInfo(ctx context.Context, names []string) ([]models.Artist, error)
	}
	getPlaylistRepo interface {
		GetPlaylistWithTracks(ctx context.Context, playlistId string) (*models.Playlist, error)
		GetTrack(ctx context.Context, trackId string) (*models.Track, error)
		GetAlbum(ctx context.Context, albumId string) (*models.Album, error)
		GetArtist(ctx context.Context, artistId string) (*models.Artist, error)
	}
)

func NewSpotifyService(artistRepo filterArtistsRepo, playlistRepo getPlaylistRepo) *SpotifyService {
	return &SpotifyService{
		artistRepo:   artistRepo,
		playlistRepo: playlistRepo,
	}
}

type SpotifyService struct {
	artistRepo   filterArtistsRepo
	playlistRepo getPlaylistRepo
}

// formRuContent filters the provided tracks and returns rusContent containing only Russian artists and tracks
// that have at least one Russian artist.
func (s *SpotifyService) formRuContent(ctx context.Context, tracks []models.Track) (*models.RuContent, error) {
	var ruContent models.RuContent

	if len(tracks) == 0 {
		return &ruContent, nil
	}

	// fill artists map with lowercase names as keys for easier matching later
	artistsMap := make(map[string]models.Artist)
	for _, track := range tracks {
		for _, artist := range track.Artists {
			artistsMap[strings.ToLower(artist.Name)] = models.Artist{
				Name:      artist.Name,
				SpotifyID: artist.SpotifyID,
			}
		}
	}

	// create slice of unique artist names
	artistNames := make([]string, 0, len(artistsMap))
	for name := range artistsMap {
		artistNames = append(artistNames, name)
	}

	// filter Russian artists using the repository
	ruArtistNames, err := s.artistRepo.FilterRussian(ctx, artistNames)
	if err != nil {
		log.Error().Err(err).Msg("failed to filter Russian artists")
		return nil, appErrors.ErrDatabaseFailure
	}

	ruArtistsWithInfo, err := s.artistRepo.GetArtistsInfo(ctx, ruArtistNames)
	if err != nil {
		log.Error().Err(err).Msg("failed to get Russian artists info")
		return nil, appErrors.ErrDatabaseFailure
	}

	// update artistsMap with info from the database for Russian artists
	for _, artistInfo := range ruArtistsWithInfo {
		artist := artistsMap[strings.ToLower(artistInfo.Name)]
		artist.ID = artistInfo.ID
		artist.DescriptionUA = artistInfo.DescriptionUA
		artist.DescriptionEN = artistInfo.DescriptionEN
		artist.Source = artistInfo.Source
		artist.SourceURL = artistInfo.SourceURL
		artist.Country = artistInfo.Country
		artist.Confirmed = artistInfo.Confirmed
		artistsMap[strings.ToLower(artistInfo.Name)] = artist
	}

	// names should all be in lowercase at this point
	// fill ruArtistsMap for faster lookup when filtering tracks later
	ruArtistsMap := make(map[string]models.Artist, len(ruArtistNames))
	for _, name := range ruArtistNames {
		ruArtistsMap[name] = artistsMap[name]
	}

	// filter tracks that have at least one Russian artist
	ruTracksMap := make(map[string]models.Track)
	for _, track := range tracks {
		for _, artist := range track.Artists {
			if _, isRu := ruArtistsMap[strings.ToLower(artist.Name)]; isRu {
				ruTracksMap[track.SpotifyID] = track
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

	return &ruContent, nil
}

func (s *SpotifyService) GetTrackRuContent(ctx context.Context, trackId string) (*models.RuContent, error) {
	start := time.Now()

	track, err := s.playlistRepo.GetTrack(ctx, trackId)
	if err != nil {
		log.Error().Err(err).Msg("failed to get track with artists")
		metrics.ErrorsTotal.WithLabelValues(metrics.ErrorTypeLabel(err)).Inc()
		return nil, err
	}

	ruContent, err := s.formRuContent(ctx, []models.Track{*track})
	if err != nil {
		log.Error().Err(err).Msg("failed to form RU content")
		metrics.ErrorsTotal.WithLabelValues(metrics.ErrorTypeLabel(err)).Inc()
		return nil, err
	}

	elapsed := time.Since(start).Seconds()
	metrics.ScansTotal.Inc()
	metrics.ScanDuration.Observe(elapsed)
	if role, _ := ctx.Value(models.UserRoleKey).(models.Role); role == models.RoleAnon {
		metrics.AnonScansTotal.Inc()
		metrics.AnonScanDuration.Observe(elapsed)
	}
	return ruContent, nil
}

func (s *SpotifyService) GetPlaylistRuContent(ctx context.Context, playlistId string) (*models.RuContent, error) {
	start := time.Now()

	playlist, err := s.playlistRepo.GetPlaylistWithTracks(ctx, playlistId)
	if err != nil {
		log.Error().Err(err).Msg("failed to get playlist with tracks")
		metrics.ErrorsTotal.WithLabelValues(metrics.ErrorTypeLabel(err)).Inc()
		return nil, err
	}

	ruContent, err := s.formRuContent(ctx, playlist.Tracks)
	if err != nil {
		log.Error().Err(err).Msg("failed to form RU content")
		metrics.ErrorsTotal.WithLabelValues(metrics.ErrorTypeLabel(err)).Inc()
		return nil, err
	}

	elapsed := time.Since(start).Seconds()
	metrics.ScansTotal.Inc()
	metrics.ScanDuration.Observe(elapsed)
	if role, _ := ctx.Value(models.UserRoleKey).(models.Role); role == models.RoleAnon {
		metrics.AnonScansTotal.Inc()
		metrics.AnonScanDuration.Observe(elapsed)
	}
	return ruContent, nil
}

func (s *SpotifyService) GetAlbumRuContent(ctx context.Context, albumId string) (*models.RuContent, error) {
	album, err := s.playlistRepo.GetAlbum(ctx, albumId)
	if err != nil {
		log.Error().Err(err).Msg("failed to get album with tracks")
		metrics.ErrorsTotal.WithLabelValues(metrics.ErrorTypeLabel(err)).Inc()
		return nil, err
	}

	ruContent, err := s.formRuContent(ctx, album.Tracks)
	if err != nil {
		log.Error().Err(err).Msg("failed to form RU content")
		metrics.ErrorsTotal.WithLabelValues(metrics.ErrorTypeLabel(err)).Inc()
		return nil, err
	}

	return ruContent, nil
}

func (s *SpotifyService) GetArtistRuContent(ctx context.Context, artistId string) (*models.RuContent, error) {
	artist, err := s.playlistRepo.GetArtist(ctx, artistId)
	if err != nil {
		log.Error().Err(err).Msg("failed to get artist")
		metrics.ErrorsTotal.WithLabelValues(metrics.ErrorTypeLabel(err)).Inc()
		return nil, appErrors.ErrSpotifyAPIError
	}

	// TODO: maybe add singular artist check method
	artistInfo, err := s.artistRepo.GetArtistsInfo(ctx, []string{strings.ToLower(artist.Name)})
	if err != nil {
		log.Error().Err(err).Msg("failed to get Russian artist info")
		metrics.ErrorsTotal.WithLabelValues(metrics.ErrorTypeLabel(err)).Inc()
		return nil, appErrors.ErrDatabaseFailure
	}

	if len(artistInfo) == 0 {
		return &models.RuContent{}, nil
	}

	return &models.RuContent{
		Artists: []models.Artist{artistInfo[0]},
	}, nil
}
