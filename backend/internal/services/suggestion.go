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
		UpdateArtistInsertSuggestion(ctx context.Context, id, creatorID int, name, description string) (models.ArtistInsertSuggestion, error)
		IsArtistInsertSuggestionApproved(ctx context.Context, id, creatorID int) (bool, error)
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
	approved, err := s.suggestionRepo.IsArtistInsertSuggestionApproved(ctx, id, creatorID)
	if err != nil {
		return err
	}
	if approved {
		return appErrors.ErrSuggestionApproved
	}
	return s.suggestionRepo.DeleteArtistInsertSuggestion(ctx, id, creatorID)
}

func (s *SuggestionService) UpdateArtistInsertSuggestion(ctx context.Context, id, creatorID int, name, description string) (models.ArtistInsertSuggestion, error) {
	exists, err := s.artistRepo.ArtistExists(ctx, name)
	if err != nil {
		log.Error().Err(err).Msg("failed to check if artist exists")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}
	if exists {
		return models.ArtistInsertSuggestion{}, appErrors.ErrArtistExists
	}

	return s.suggestionRepo.UpdateArtistInsertSuggestion(ctx, id, creatorID, name, description)
}
