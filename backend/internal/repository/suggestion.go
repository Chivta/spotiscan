package repository

import (
	"context"
	"strings"
	"database/sql"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/chivta/ruscan/internal/appErrors"
	"github.com/chivta/ruscan/internal/models"
)

func NewSuggestionRepo(db *sql.DB, redis *redis.Client) *SuggestionRepo {
	return &SuggestionRepo{
		db:    db,
		redis: redis, // for updating cached set
	}
}

type SuggestionRepo struct {
	db    *sql.DB
	redis *redis.Client
}

func (r *SuggestionRepo) CreateArtistInsertSuggestion(ctx context.Context, name, description string, creatorID int) (models.ArtistInsertSuggestion, error) {
	var suggestion models.ArtistInsertSuggestion

	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO artist_insert_suggestions (artist_name, description, creator_id) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`,
		name,
		description,
		creatorID,
	).Scan(&suggestion.ID, &suggestion.CreatedAt, &suggestion.UpdatedAt)

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
	rows, err := r.db.QueryContext(ctx, `SELECT id, artist_name, description, state, creator_id, created_at FROM artist_insert_suggestions WHERE creator_id = $1 ORDER BY created_at DESC`, creatorID)
	if err != nil {
		log.Error().Err(err).Msg("failed to query artist suggestions from database")
		return nil, appErrors.ErrDatabaseFailure
	}
	defer rows.Close()

	var suggestions []models.ArtistInsertSuggestion
	for rows.Next() {
		var suggestion models.ArtistInsertSuggestion
		if err := rows.Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.CreatorID, &suggestion.CreatedAt); err != nil {
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
		 RETURNING id, artist_name, description, state, creator_id, created_at, updated_at`,
		name,
		description,
		id,
		creatorID,
	).Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.CreatorID, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err == sql.ErrNoRows {
		return models.ArtistInsertSuggestion{}, appErrors.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to update artist suggestion in database")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}

	return suggestion, nil
}

func (r *SuggestionRepo) CreateArtistDeleteSuggestion(ctx context.Context, artistName, description string, creatorID int) (models.ArtistDeleteSuggestion, error) {
	var suggestion models.ArtistDeleteSuggestion

	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO artist_delete_suggestions (artist_name, description, creator_id) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`,
		artistName,
		description,
		creatorID,
	).Scan(&suggestion.ID, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert artist delete suggestion into database")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}

	suggestion.ArtistName = artistName
	suggestion.Description = description
	suggestion.CreatorID = creatorID

	return suggestion, nil
}

func (r *SuggestionRepo) GetArtistDeleteSuggestions(ctx context.Context, creatorID int) ([]models.ArtistDeleteSuggestion, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, artist_name, description, state, creator_id, created_at FROM artist_delete_suggestions WHERE creator_id = $1 ORDER BY created_at DESC`, creatorID)
	if err != nil {
		log.Error().Err(err).Msg("failed to query artist delete suggestions from database")
		return nil, appErrors.ErrDatabaseFailure
	}
	defer rows.Close()

	var suggestions []models.ArtistDeleteSuggestion
	for rows.Next() {
		var suggestion models.ArtistDeleteSuggestion
		if err := rows.Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.CreatorID, &suggestion.CreatedAt); err != nil {
			log.Error().Err(err).Msg("failed to scan artist delete suggestion row")
			return nil, appErrors.ErrDatabaseFailure
		}
		suggestions = append(suggestions, suggestion)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("error iterating over artist delete suggestion rows")
		return nil, appErrors.ErrDatabaseFailure
	}

	return suggestions, nil
}

func (r *SuggestionRepo) DeleteArtistDeleteSuggestion(ctx context.Context, id, creatorID int) error {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM artist_delete_suggestions WHERE id = $1 AND creator_id = $2`,
		id,
		creatorID,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete artist delete suggestion from database")
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

func (r *SuggestionRepo) UpdateArtistDeleteSuggestion(ctx context.Context, id, creatorID int, artistName, description string) (models.ArtistDeleteSuggestion, error) {
	var suggestion models.ArtistDeleteSuggestion

	err := r.db.QueryRowContext(
		ctx,
		`UPDATE artist_delete_suggestions
		 SET artist_name = $1, description = $2, updated_at = NOW()
		 WHERE id = $3 AND creator_id = $4
		 RETURNING id, artist_name, description, state, creator_id, created_at, updated_at`,
		artistName,
		description,
		id,
		creatorID,
	).Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.CreatorID, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err == sql.ErrNoRows {
		return models.ArtistDeleteSuggestion{}, appErrors.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to update artist delete suggestion in database")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}

	return suggestion, nil
}

func (r *SuggestionRepo) IsArtistDeleteSuggestionPending(ctx context.Context, id, creatorID int) (bool, error) {
	var pending bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM artist_delete_suggestions WHERE id = $1 AND creator_id = $2 AND state='pending')`, id, creatorID).Scan(&pending)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, appErrors.ErrNotFound
		}
		log.Error().Err(err).Msg("failed to check if artist delete suggestion is pending")
		return false, appErrors.ErrDatabaseFailure
	}
	return pending, nil
}

func (r *SuggestionRepo) IsArtistInsertSuggestionPending(ctx context.Context, id, creatorID int) (bool, error) {
	var pending bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM artist_insert_suggestions WHERE id = $1 AND creator_id = $2 AND state='pending')`, id, creatorID).Scan(&pending)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, appErrors.ErrNotFound
		}
		log.Error().Err(err).Msg("failed to check if artist insert suggestion is pending")
		return false, appErrors.ErrDatabaseFailure
	}
	return pending, nil
}

func (r *SuggestionRepo) GetAllArtistInsertSuggestions(ctx context.Context) ([]models.ArtistInsertSuggestion, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, artist_name, description, state, decline_reason, creator_id, created_at FROM artist_insert_suggestions ORDER BY created_at DESC`)
	if err != nil {
		log.Error().Err(err).Msg("failed to query all artist insert suggestions from database")
		return nil, appErrors.ErrDatabaseFailure
	}
	defer rows.Close()

	var suggestions []models.ArtistInsertSuggestion
	for rows.Next() {
		var suggestion models.ArtistInsertSuggestion
		var declineReason sql.NullString
		if err := rows.Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &declineReason, &suggestion.CreatorID, &suggestion.CreatedAt); err != nil {
			log.Error().Err(err).Msg("failed to scan artist insert suggestion row")
			return nil, appErrors.ErrDatabaseFailure
		}
		suggestion.DeclineReason = declineReason.String
		suggestions = append(suggestions, suggestion)
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("error iterating over artist insert suggestion rows")
		return nil, appErrors.ErrDatabaseFailure
	}
	return suggestions, nil
}

func (r *SuggestionRepo) ApproveArtistInsertSuggestion(ctx context.Context, id int) (models.ArtistInsertSuggestion, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to begin transaction for approving artist insert suggestion")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}
	defer tx.Rollback()

	var suggestion models.ArtistInsertSuggestion
	err = tx.QueryRowContext(
		ctx,
		`UPDATE artist_insert_suggestions SET state = 'approved' WHERE id = $1 AND state = 'pending'
		 RETURNING id, artist_name, description, state, creator_id, created_at, updated_at`,
		id,
	).Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.CreatorID, &suggestion.CreatedAt, &suggestion.UpdatedAt)

	if err == sql.ErrNoRows {
		return models.ArtistInsertSuggestion{}, appErrors.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to approve artist insert suggestion")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}

	// if the artist is on the blocklist, remove them from it when approving the insert suggestion
	_, err = tx.ExecContext(
		ctx,
		`DELETE FROM artist_blocklist WHERE name = $1`,
		strings.ToLower(suggestion.ArtistName),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove artist from blocklist when approving insert suggestion")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO ru_artists (name, description_ua, description_en, source) VALUES ($1, $2, $2, 'crowdsourced') ON CONFLICT (name) DO NOTHING`,
		strings.ToLower(suggestion.ArtistName), suggestion.Description,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert artist from approved suggestion")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}

	err = r.redis.SAdd(ctx, ruArtistsRedisKey, suggestion.ArtistName).Err()
	if err != nil {
		log.Warn().Err(err).Msg("failed to add artist to redis set after approving suggestion")
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("failed to commit artist insert suggestion approval")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}
	return suggestion, nil
}

func (r *SuggestionRepo) GetAllArtistDeleteSuggestions(ctx context.Context) ([]models.ArtistDeleteSuggestion, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, artist_name, description, state, decline_reason, creator_id, created_at FROM artist_delete_suggestions ORDER BY created_at DESC`)
	if err != nil {
		log.Error().Err(err).Msg("failed to query all artist delete suggestions from database")
		return nil, appErrors.ErrDatabaseFailure
	}
	defer rows.Close()

	var suggestions []models.ArtistDeleteSuggestion
	for rows.Next() {
		var suggestion models.ArtistDeleteSuggestion
		var declineReason sql.NullString
		if err := rows.Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &declineReason, &suggestion.CreatorID, &suggestion.CreatedAt); err != nil {
			log.Error().Err(err).Msg("failed to scan artist delete suggestion row")
			return nil, appErrors.ErrDatabaseFailure
		}
		suggestion.DeclineReason = declineReason.String
		suggestions = append(suggestions, suggestion)
	}
	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("error iterating over artist delete suggestion rows")
		return nil, appErrors.ErrDatabaseFailure
	}
	return suggestions, nil
}

func (r *SuggestionRepo) ApproveArtistDeleteSuggestion(ctx context.Context, id int) (models.ArtistDeleteSuggestion, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to begin transaction for approving artist delete suggestion")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}
	defer tx.Rollback()

	var suggestion models.ArtistDeleteSuggestion
	err = tx.QueryRowContext(
		ctx,
		`UPDATE artist_delete_suggestions SET state = 'approved' WHERE id = $1 AND state = 'pending'
		 RETURNING id, artist_name, description, state, creator_id, created_at, updated_at`,
		id,
	).Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.CreatorID, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err == sql.ErrNoRows {
		return models.ArtistDeleteSuggestion{}, appErrors.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to approve artist delete suggestion")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM ru_artists WHERE name = $1`, strings.ToLower(suggestion.ArtistName))
	if err != nil {
		log.Error().Err(err).Msg("failed to delete artist from approved suggestion")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO artist_blocklist (name, reason) VALUES ($1,$2) ON CONFLICT (name) DO NOTHING`, strings.ToLower(suggestion.ArtistName), suggestion.Description)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert artist into blocklist after approving delete suggestion")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}

	err = r.redis.SRem(ctx, ruArtistsRedisKey, strings.ToLower(suggestion.ArtistName)).Err()
	if err != nil {
		log.Warn().Err(err).Msg("failed to remove artist from redis set after approving delete suggestion")
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("failed to commit artist delete suggestion approval")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}
	return suggestion, nil
}

func (r *SuggestionRepo) DeclineArtistInsertSuggestion(ctx context.Context, id int, reason string) (models.ArtistInsertSuggestion, error) {
	var suggestion models.ArtistInsertSuggestion
	err := r.db.QueryRowContext(
		ctx,
		`UPDATE artist_insert_suggestions SET state = 'declined', decline_reason = $2
		 WHERE id = $1 AND state = 'pending'
		 RETURNING id, artist_name, description, state, decline_reason, creator_id, created_at, updated_at`,
		id,
		reason,
	).Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.DeclineReason, &suggestion.CreatorID, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err == sql.ErrNoRows {
		return models.ArtistInsertSuggestion{}, appErrors.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to decline artist insert suggestion")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}
	return suggestion, nil
}

func (r *SuggestionRepo) DeclineArtistDeleteSuggestion(ctx context.Context, id int, reason string) (models.ArtistDeleteSuggestion, error) {
	var suggestion models.ArtistDeleteSuggestion
	err := r.db.QueryRowContext(
		ctx,
		`UPDATE artist_delete_suggestions SET state = 'declined', decline_reason = $2
		 WHERE id = $1 AND state = 'pending'
		 RETURNING id, artist_name, description, state, decline_reason, creator_id, created_at, updated_at`,
		id,
		reason,
	).Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.DeclineReason, &suggestion.CreatorID, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err == sql.ErrNoRows {
		return models.ArtistDeleteSuggestion{}, appErrors.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to decline artist delete suggestion")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}
	return suggestion, nil
}
