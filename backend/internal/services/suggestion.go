package services

import (
	"context"

	"github.com/chivta/ruscan/internal/appErrors"
	"github.com/chivta/ruscan/internal/models"
	"github.com/rs/zerolog/log"
)

type (
	suggestionRepo interface {
		CreateArtistInsertSuggestion(ctx context.Context, name, description string, creatorID int) (models.ArtistInsertSuggestion, error)
		GetArtistInsertSuggestions(ctx context.Context, creatorID int) ([]models.ArtistInsertSuggestion, error)
		DeleteArtistInsertSuggestion(ctx context.Context, id, creatorID int) error
		UpdateArtistInsertSuggestion(ctx context.Context, id, creatorID int, description string) (models.ArtistInsertSuggestion, error)
		IsArtistInsertSuggestionPending(ctx context.Context, id, creatorID int) (bool, error)
		CreateArtistDeleteSuggestion(ctx context.Context, artistName, description string, creatorID int) (models.ArtistDeleteSuggestion, error)
		GetArtistDeleteSuggestions(ctx context.Context, creatorID int) ([]models.ArtistDeleteSuggestion, error)
		DeleteArtistDeleteSuggestion(ctx context.Context, id, creatorID int) error
		UpdateArtistDeleteSuggestion(ctx context.Context, id, creatorID int, description string) (models.ArtistDeleteSuggestion, error)
		IsArtistDeleteSuggestionPending(ctx context.Context, id, creatorID int) (bool, error)
		GetAllArtistInsertSuggestions(ctx context.Context) ([]models.ArtistInsertSuggestion, error)
		ApproveArtistInsertSuggestion(ctx context.Context, id int) (models.ArtistInsertSuggestion, error)
		DeclineArtistInsertSuggestion(ctx context.Context, id int, reason string) (models.ArtistInsertSuggestion, error)
		GetAllArtistDeleteSuggestions(ctx context.Context) ([]models.ArtistDeleteSuggestion, error)
		ApproveArtistDeleteSuggestion(ctx context.Context, id int) (models.ArtistDeleteSuggestion, error)
		DeclineArtistDeleteSuggestion(ctx context.Context, id int, reason string) (models.ArtistDeleteSuggestion, error)
	}
	artistRepo interface {
		ArtistExists(ctx context.Context, name string) (bool, error)
	}
)

func NewSuggestionService(suggestionRepo suggestionRepo, artistRepo artistRepo) *SuggestionService {
	return &SuggestionService{
		suggestionRepo: suggestionRepo,
		artistRepo:     artistRepo,
	}
}

type SuggestionService struct {
	suggestionRepo suggestionRepo
	artistRepo     artistRepo
}

func (s *SuggestionService) CreateArtistInsertSuggestion(ctx context.Context, name, description string, creatorID int) (models.ArtistInsertSuggestion, error) {
	exists, err := s.artistRepo.ArtistExists(ctx, name)
	if err != nil {
		log.Error().Err(err).Msg("failed to check if artist exists")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}
	if exists {
		return models.ArtistInsertSuggestion{}, appErrors.ErrArtistExists
	}

	suggestion, err := s.suggestionRepo.CreateArtistInsertSuggestion(ctx, name, description, creatorID)
	if err != nil {
		return models.ArtistInsertSuggestion{}, err
	}

	return suggestion, nil
}

func (s *SuggestionService) GetArtistInsertSuggestions(ctx context.Context, creatorID int) ([]models.ArtistInsertSuggestion, error) {
	return s.suggestionRepo.GetArtistInsertSuggestions(ctx, creatorID)
}

func (s *SuggestionService) DeleteArtistInsertSuggestion(ctx context.Context, id, creatorID int) error {
	pending, err := s.suggestionRepo.IsArtistInsertSuggestionPending(ctx, id, creatorID)
	if err != nil {
		return err
	}
	if !pending {
		return appErrors.ErrSuggestionNotPending
	}
	return s.suggestionRepo.DeleteArtistInsertSuggestion(ctx, id, creatorID)
}

func (s *SuggestionService) UpdateArtistInsertSuggestion(ctx context.Context, id, creatorID int, description string) (models.ArtistInsertSuggestion, error) {
	pending, err := s.suggestionRepo.IsArtistInsertSuggestionPending(ctx, id, creatorID)
	if err != nil {
		return models.ArtistInsertSuggestion{}, err
	}
	if !pending {
		return models.ArtistInsertSuggestion{}, appErrors.ErrSuggestionNotPending
	}

	return s.suggestionRepo.UpdateArtistInsertSuggestion(ctx, id, creatorID, description)
}

func (s *SuggestionService) CreateArtistDeleteSuggestion(ctx context.Context, artistName, description string, creatorID int) (models.ArtistDeleteSuggestion, error) {
	exists, err := s.artistRepo.ArtistExists(ctx, artistName)
	if err != nil {
		log.Error().Err(err).Msg("failed to check if artist exists")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}
	if !exists {
		return models.ArtistDeleteSuggestion{}, appErrors.ErrNotFound
	}

	return s.suggestionRepo.CreateArtistDeleteSuggestion(ctx, artistName, description, creatorID)
}

func (s *SuggestionService) GetArtistDeleteSuggestions(ctx context.Context, creatorID int) ([]models.ArtistDeleteSuggestion, error) {
	return s.suggestionRepo.GetArtistDeleteSuggestions(ctx, creatorID)
}

func (s *SuggestionService) DeleteArtistDeleteSuggestion(ctx context.Context, id, creatorID int) error {
	pending, err := s.suggestionRepo.IsArtistDeleteSuggestionPending(ctx, id, creatorID)
	if err != nil {
		return err
	}
	if !pending {
		return appErrors.ErrSuggestionNotPending
	}
	return s.suggestionRepo.DeleteArtistDeleteSuggestion(ctx, id, creatorID)
}

func (s *SuggestionService) UpdateArtistDeleteSuggestion(ctx context.Context, id, creatorID int, description string) (models.ArtistDeleteSuggestion, error) {
	pending, err := s.suggestionRepo.IsArtistDeleteSuggestionPending(ctx, id, creatorID)
	if err != nil {
		return models.ArtistDeleteSuggestion{}, err
	}
	if !pending {
		return models.ArtistDeleteSuggestion{}, appErrors.ErrSuggestionNotPending
	}

	return s.suggestionRepo.UpdateArtistDeleteSuggestion(ctx, id, creatorID, description)
}

func (s *SuggestionService) GetAllArtistInsertSuggestions(ctx context.Context) ([]models.ArtistInsertSuggestion, error) {
	return s.suggestionRepo.GetAllArtistInsertSuggestions(ctx)
}

func (s *SuggestionService) ApproveArtistInsertSuggestion(ctx context.Context, id int) (models.ArtistInsertSuggestion, error) {
	return s.suggestionRepo.ApproveArtistInsertSuggestion(ctx, id)
}

func (s *SuggestionService) GetAllArtistDeleteSuggestions(ctx context.Context) ([]models.ArtistDeleteSuggestion, error) {
	return s.suggestionRepo.GetAllArtistDeleteSuggestions(ctx)
}

func (s *SuggestionService) ApproveArtistDeleteSuggestion(ctx context.Context, id int) (models.ArtistDeleteSuggestion, error) {
	return s.suggestionRepo.ApproveArtistDeleteSuggestion(ctx, id)
}

func (s *SuggestionService) DeclineArtistInsertSuggestion(ctx context.Context, id int, reason string) (models.ArtistInsertSuggestion, error) {
	return s.suggestionRepo.DeclineArtistInsertSuggestion(ctx, id, reason)
}

func (s *SuggestionService) DeclineArtistDeleteSuggestion(ctx context.Context, id int, reason string) (models.ArtistDeleteSuggestion, error) {
	return s.suggestionRepo.DeclineArtistDeleteSuggestion(ctx, id, reason)
}
