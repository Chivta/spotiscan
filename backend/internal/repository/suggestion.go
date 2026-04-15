package repository

import (
	"context"
	"database/sql"
	"strings"

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

// checks if suggestion exists and pending and lock it
// returns ErrNotFound, ErrSuggestionNotPending where appropriate or ErrDatabaseFailure with logging
func lockPendingSuggestion(ctx context.Context, tx *sql.Tx, table string, id int) error {
	var state string
	err := tx.QueryRowContext(
		ctx,
		`SELECT state FROM `+table+` WHERE id = $1 FOR UPDATE`,
		id,
	).Scan(&state)
	if err != nil {
		if err == sql.ErrNoRows {
			return appErrors.ErrNotFound
		}
		log.Error().Err(err).Msg("failed to query suggestion for locking")
		return appErrors.ErrDatabaseFailure
	}
	if state != "pending" {
		return appErrors.ErrSuggestionNotPending
	}
	return nil
}


func (r *SuggestionRepo) CreateArtistInsertSuggestion(ctx context.Context, name, description string, creatorID int) (models.ArtistInsertSuggestion, error) {
	var suggestion models.ArtistInsertSuggestion

	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO artist_insert_suggestions (artist_name, description, creator_id) 
		 VALUES ($1, $2, $3) 
		 RETURNING id, created_at, updated_at`,
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
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, artist_name, description, state, creator_id, created_at 
		 FROM artist_insert_suggestions 
		 WHERE creator_id = $1 
		 ORDER BY created_at DESC`,
		creatorID)
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to begin transaction for deleting artist suggestion")
		return appErrors.ErrDatabaseFailure
	}
	defer tx.Rollback()
	err = lockPendingSuggestion(ctx, tx, "artist_insert_suggestions", id)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(
		ctx,
		`DELETE FROM artist_insert_suggestions
		 WHERE id = $1 AND creator_id = $2`,
		id,
		creatorID,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete artist suggestion from database")
		return appErrors.ErrDatabaseFailure
	}
	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Msg("failed to commit transaction for deleting artist suggestion")
		return appErrors.ErrDatabaseFailure
	}
	return nil
}

func (r *SuggestionRepo) UpdateArtistInsertSuggestion(ctx context.Context, id, creatorID int, description string) (models.ArtistInsertSuggestion, error) {
	var suggestion models.ArtistInsertSuggestion

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to begin transaction for updating artist insert suggestion")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}
	defer tx.Rollback()
	err = lockPendingSuggestion(ctx, tx, "artist_insert_suggestions", id)
	if err != nil {
		return models.ArtistInsertSuggestion{}, err
	}

	err = tx.QueryRowContext(
		ctx,
		`UPDATE artist_insert_suggestions
		 SET description = $1, updated_at = NOW()
		 WHERE id = $2 AND creator_id = $3
		 RETURNING id, artist_name, description, state, creator_id, created_at, updated_at`,
		description,
		id,
		creatorID,
	).Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.CreatorID, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err != nil {
		log.Error().Err(err).Msg("failed to update artist insert suggestion in database")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Msg("failed to commit transaction for updating artist insert suggestion")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}

	return suggestion, nil
}

func (r *SuggestionRepo) CreateArtistDeleteSuggestion(ctx context.Context, artistName, description string, creatorID int) (models.ArtistDeleteSuggestion, error) {
	var suggestion models.ArtistDeleteSuggestion

	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO artist_delete_suggestions (artist_name, description, creator_id) 
		 VALUES ($1, $2, $3) 
		 RETURNING id, created_at, updated_at`,
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
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, artist_name, description, state, creator_id, created_at 
		 FROM artist_delete_suggestions 
		 WHERE creator_id = $1 
		 ORDER BY created_at DESC`, 
		creatorID)
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to begin transaction for deleting artist delete suggestion")
		return appErrors.ErrDatabaseFailure
	}
	defer tx.Rollback()
	err = lockPendingSuggestion(ctx, tx, "artist_delete_suggestions", id)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		`DELETE FROM artist_delete_suggestions WHERE id = $1 AND creator_id = $2`,
		id,
		creatorID,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete artist delete suggestion from database")
		return appErrors.ErrDatabaseFailure
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Msg("failed to commit transaction for deleting artist delete suggestion")
		return appErrors.ErrDatabaseFailure
	}

	return nil
}

func (r *SuggestionRepo) UpdateArtistDeleteSuggestion(ctx context.Context, id, creatorID int, description string) (models.ArtistDeleteSuggestion, error) {
	var suggestion models.ArtistDeleteSuggestion

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to begin transaction for updating artist delete suggestion")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}
	defer tx.Rollback()
	err = lockPendingSuggestion(ctx, tx, "artist_delete_suggestions", id)
	if err != nil {
		return models.ArtistDeleteSuggestion{}, err
	}

	err = tx.QueryRowContext(
		ctx,
		`UPDATE artist_delete_suggestions
		 SET description = $1, updated_at = NOW()
		 WHERE id = $2 AND creator_id = $3
		 RETURNING id, artist_name, description, state, creator_id, created_at, updated_at`,
		description,
		id,
		creatorID,
	).Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.CreatorID, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err != nil {
		log.Error().Err(err).Msg("failed to update artist delete suggestion in database")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Msg("failed to commit transaction for updating artist delete suggestion")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}

	return suggestion, nil
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
	err = lockPendingSuggestion(ctx, tx, "artist_insert_suggestions", id)
	if err != nil {
		return models.ArtistInsertSuggestion{}, err
	}

	var suggestion models.ArtistInsertSuggestion
	err = tx.QueryRowContext(
		ctx,
		`UPDATE artist_insert_suggestions SET state = 'approved'
		 WHERE id = $1
		 RETURNING id, artist_name, description, state, creator_id, created_at, updated_at`,
		id,
	).Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.CreatorID, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err != nil {
		log.Error().Err(err).Msg("failed to update artist insert suggestion state in database")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}
	nameLower := strings.ToLower(suggestion.ArtistName)

	// if the artist is on the blocklist, remove them from it when approving the insert suggestion
	_, err = tx.ExecContext(
		ctx,
		`DELETE FROM artist_blocklist WHERE name = $1`,
		nameLower,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove artist from blocklist when approving insert suggestion")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO ru_artists (name, description_ua, description_en, source) VALUES ($1, $2, $2, 'crowdsourced') ON CONFLICT (name) DO NOTHING`,
		nameLower, suggestion.Description,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert artist from approved suggestion")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}
	err = tx.Commit()
	if err != nil {
		log.Error().Err(err).Msg("failed to commit artist insert suggestion approval")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}

	err = r.redis.SAdd(ctx, ruArtistsRedisKey, nameLower).Err()
	if err != nil {
		log.Warn().Err(err).Msg("failed to add artist to redis set after approving suggestion")
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
	err = lockPendingSuggestion(ctx, tx, "artist_delete_suggestions", id)
	if err != nil {
		return models.ArtistDeleteSuggestion{}, err
	}

	var suggestion models.ArtistDeleteSuggestion
	err = tx.QueryRowContext(
		ctx,
		`UPDATE artist_delete_suggestions 
		 SET state = 'approved'
		 WHERE id = $1
		 RETURNING id, artist_name, description, state, creator_id, created_at, updated_at`,
		id,
	).Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.CreatorID, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err != nil {
		log.Error().Err(err).Msg("failed to update artist delete suggestion state in database")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}
	nameLower := strings.ToLower(suggestion.ArtistName)

	_, err = tx.ExecContext(ctx, `DELETE FROM ru_artists WHERE name = $1`, nameLower)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete artist from approved suggestion")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO artist_blocklist (name, reason) VALUES ($1,$2) ON CONFLICT (name) DO NOTHING`, strings.ToLower(suggestion.ArtistName), suggestion.Description)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert artist into blocklist after approving delete suggestion")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("failed to commit artist delete suggestion approval")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}

	err = r.redis.SRem(ctx, ruArtistsRedisKey, nameLower).Err()
	if err != nil {
		log.Warn().Err(err).Msg("failed to remove artist from redis set after approving delete suggestion")
	}

	return suggestion, nil
}

func (r *SuggestionRepo) DeclineArtistInsertSuggestion(ctx context.Context, id int, reason string) (models.ArtistInsertSuggestion, error) {
	var suggestion models.ArtistInsertSuggestion
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to begin transaction for declining artist insert suggestion")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}
	defer tx.Rollback()
	err = lockPendingSuggestion(ctx, tx, "artist_insert_suggestions", id)
	if err != nil {
		return models.ArtistInsertSuggestion{}, err
	}
	err = tx.QueryRowContext(
		ctx,
		`UPDATE artist_insert_suggestions 
		 SET state = 'declined', decline_reason = $2
		 WHERE id = $1
		 RETURNING id, artist_name, description, state, decline_reason, creator_id, created_at, updated_at`,
		id,
		reason,
	).Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.DeclineReason, &suggestion.CreatorID, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err != nil {
		log.Error().Err(err).Msg("failed to decline artist insert suggestion in database")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("failed to commit transaction for declining artist insert suggestion")
		return models.ArtistInsertSuggestion{}, appErrors.ErrDatabaseFailure
	}

	return suggestion, nil
}

func (r *SuggestionRepo) DeclineArtistDeleteSuggestion(ctx context.Context, id int, reason string) (models.ArtistDeleteSuggestion, error) {
	var suggestion models.ArtistDeleteSuggestion

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to begin transaction for declining artist delete suggestion")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}
	defer tx.Rollback()
	err = lockPendingSuggestion(ctx, tx, "artist_delete_suggestions", id)
	if err != nil {
		return models.ArtistDeleteSuggestion{}, err
	}
	err = tx.QueryRowContext(
		ctx,
		`UPDATE artist_delete_suggestions
		 SET state = 'declined', decline_reason = $2
		 WHERE id = $1
		 RETURNING id, artist_name, description, state, decline_reason, creator_id, created_at, updated_at`,
		id,
		reason,
	).Scan(&suggestion.ID, &suggestion.ArtistName, &suggestion.Description, &suggestion.State, &suggestion.DeclineReason, &suggestion.CreatorID, &suggestion.CreatedAt, &suggestion.UpdatedAt)
	if err != nil {
		log.Error().Err(err).Msg("failed to decline artist delete suggestion in database")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("failed to commit transaction for declining artist delete suggestion")
		return models.ArtistDeleteSuggestion{}, appErrors.ErrDatabaseFailure
	}

	return suggestion, nil
}
