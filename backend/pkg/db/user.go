package db

import (
	"spotiscan/models"
	"database/sql"
)

func (db *DB) GetUserByID(id int) (*models.User, error) {
	var spotify_id string
	err := db.conn.QueryRow(`
		       SELECT spotify_id FROM users 
		       WHERE user_id=$1
	       `, id).Scan(&spotify_id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &models.User{
		ID:       id,
		SpotifyID: spotify_id,
	}, nil
}

func (db *DB) GetUserIdBySpotifyId(spotifyId string) (int, error) {
	var userID int
	err := db.conn.QueryRow(`
		      SELECT user_id FROM users 
		      WHERE spotify_id=$1
	      `, spotifyId).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return userID, nil
}

func (db *DB) CreateUserWithSession(spotifyId, sessionToken string) (int,error) {
	tx, err := db.conn.Begin()
	if err != nil {
		return 0,err
	}
	defer tx.Rollback()

	// Create user
	var userID int
	err = tx.QueryRow(
		`INSERT INTO users (spotify_id) VALUES ($1) RETURNING user_id`,
		spotifyId,
	).Scan(&userID)
	if err != nil {
		return 0,err
	}

	// TODO: set session expiration time properly
	_, err = tx.Exec(
		`INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, NOW() + INTERVAL '30 days')
		ON CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at`,
		userID, sessionToken,
	)
	if err != nil {
		return 0, err
	}

	return userID, tx.Commit()
}