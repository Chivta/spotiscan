package repository

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/appErrors"
	"github.com/chivta/ruscan/internal/models"
)

func NewSuggestionRepo(db *sql.DB) *SuggestionRepo {
	return &SuggestionRepo{
		db:            db,
	}
}

type SuggestionRepo struct {
	db    *sql.DB
}

func (r *SuggestionRepo) CreateArtistInsertSuggestion(ctx context.Context, name, description string, creatorID int) (models.ArtistInsertSuggestion, error) {
	var suggestion models.ArtistInsertSuggestion

	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO artist_insert_suggestions (artist_name, description, creator_id) VALUES ($1, $2, $3) RETURNING id, created_at`,
		name,
		description,
		creatorID,
	).Scan(&suggestion.ID, &suggestion.CreatedAt)

	if err != nil {
		log.Error().Err(err).Msg("failed to insert artist suggestion into database")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}

	suggestion.ArtistName = name
	suggestion.Description = description
	suggestion.CreatorID = creatorID

	return suggestion, nil
}

func (r *SuggestionRepo) GetArtistInsertSuggestions(ctx context.Context, creatorID int) ([]models.ArtistInsertSuggestion, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, artist_name, description, approved, creator_id, created_at FROM artist_insert_suggestions WHERE creator_id = $1`, creatorID)
	if err != nil {
		log.Error().Err(err).Msg("failed to query artist suggestions from database")
		return nil, appErrors.ErrDatabaseFailure
	}
	defer rows.Close()

	var suggestions []models.ArtistInsertSuggestion
	for rows.Next() {
		var suggestion models.ArtistInsertSuggestion
		if err := rows.Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.Approved, &suggestion.CreatorID, &suggestion.CreatedAt); err != nil {
			log.Error().Err(err).Msg("failed to scan artist suggestion row")
			return nil, appErrors.ErrDatabaseFailure
		}
		suggestions = append(suggestions, suggestion)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("error iterating over artist suggestion rows")
		return nil, appErrors.ErrDatabaseFailure
	}

	return suggestions, nil
}

func (r *SuggestionRepo) DeleteArtistInsertSuggestion(ctx context.Context, id, creatorID int) error {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM artist_insert_suggestions WHERE id = $1 AND creator_id = $2`,
		id,
		creatorID,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete artist suggestion from database")
		return appErrors.ErrDatabaseFailure
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows affected after delete")
		return appErrors.ErrDatabaseFailure
	}
	if rows == 0 {
		return appErrors.ErrNotFound
	}

	return nil
}

func (r *SuggestionRepo) UpdateArtistInsertSuggestion(ctx context.Context, id, creatorID int, name, description string) (models.ArtistInsertSuggestion, error) {
	var suggestion models.ArtistInsertSuggestion

	err := r.db.QueryRowContext(
		ctx,
		`UPDATE artist_insert_suggestions
		 SET artist_name = $1, description = $2, updated_at = NOW()
		 WHERE id = $3 AND creator_id = $4
		 RETURNING id, artist_name, description, creator_id, created_at, updated_at`,
		name,
		description,
		id,
		creatorID,
	).Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.CreatorID, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err == sql.ErrNoRows {
		return models.ArtistInsertSuggestion{}, appErrors.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to update artist suggestion in database")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}

	return suggestion, nil
}


func (r *SuggestionRepo) IsArtistInsertSuggestionApproved(ctx context.Context, id, creatorID int) (bool, error) {
	var approved bool
	err := r.db.QueryRowContext(ctx, `SELECT approved FROM artist_insert_suggestions WHERE id = $1 AND creator_id = $2`, id, creatorID).Scan(&approved)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, appErrors.ErrNotFound
		}
		log.Error().Err(err).Msg("failed to check if artist suggestion is approved")
		return false, appErrors.ErrDatabaseFailure
	}
	return approved, nil
}