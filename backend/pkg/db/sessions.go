package db

import (
	"database/sql"
)

// Creates or replaces a session for the given user ID with the provided session token.
func (db *DB) CreateSession(userID int, sessionToken string) error {
	_, err := db.conn.Exec(
		`INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, NOW() + INTERVAL '7 days')
		ON CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at`,
		userID, sessionToken,
	)
	return err
}

func (db *DB) GetUserIdBySessionToken(token string) (int, error) {
	var userID int
	err := db.conn.QueryRow(`
		       SELECT user_id FROM sessions 
		       WHERE token=$1 AND expires_at > NOW()
	       `, token).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return userID, nil
}



func (db *DB) DeleteSession(token string) error {
	_, err := db.conn.Exec(`
		DELETE FROM sessions WHERE token=$1
	`, token)
	return err
}