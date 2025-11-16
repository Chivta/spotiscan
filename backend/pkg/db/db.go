package db

import (
	"database/sql"
	"log"
	"spotiscan/models"

	_ "github.com/lib/pq"
)

func NewDBConnection(DatabaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", DatabaseURL)

	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

type DB struct {
	conn *sql.DB
}

func (db *DB) EmailUsed(email string) (bool, error) {
	rows, err := db.conn.Query(`SELECT 1 FROM users WHERE email=$1`, email)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		return true, nil
	}
	return false, nil
}

func (db *DB) UsernameExists(username string) (bool, error) {
	rows, err := db.conn.Query(`SELECT 1 FROM users WHERE username=$1`, username)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		return true, nil
	}
	return false, nil
}

func (db *DB) CreateUserWithSession(username, email, passwordHash string, sessionToken string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create user
	var userID int
	err = tx.QueryRow(
		`INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING user_id`,
		username, email, passwordHash,
	).Scan(&userID)
	if err != nil {
		return err
	}

	// TODO: set session expiration time properly
	_, err = tx.Exec(
		`INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, NOW() + INTERVAL '30 days')
		ON CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at`,
		userID, sessionToken,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (db *DB) GetUserIDBySessionToken(token string) (int, error) {
	var userID int
	err := db.conn.QueryRow(`
		SELECT user_id FROM sessions 
		WHERE token=$1 AND expires_at > NOW()
	`, token).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (db *DB) GetUserByID(id int) (*models.User, error) {
	var username, email string
	err := db.conn.QueryRow(`
		SELECT username, email FROM users 
		WHERE user_id=$1
	`, id).Scan(&username, &email)
	if err != nil {
		return nil, err
	}
	return &models.User{
		ID:       id,
		Username: username,
		Email:    email,
	}, nil
}

func (db *DB) DeleteSession(token string) error {
	_, err := db.conn.Exec(`
		DELETE FROM sessions WHERE token=$1
	`, token)
	return err
}

func (db *DB) Close() error {
	err := db.conn.Close()
	if err != nil {
		log.Println("Error closing database connection:", err)
	}
	return err
}

func (db *DB) GetUserIDByEmailOrUsername(emailOrUsername string) (int, error) {
	var userID int
	err := db.conn.QueryRow(`
		SELECT user_id FROM users 
		WHERE email=$1 OR username=$1
	`, emailOrUsername).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

// Creates or replaces a session for the given user ID with the provided session token.
func (db *DB) CreateSession(userID int, sessionToken string) error {
	_, err := db.conn.Exec(
		`INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, NOW() + INTERVAL '30 days')
		ON CONFLICT (user_id) DO UPDATE SET token = EXCLUDED.token, expires_at = EXCLUDED.expires_at`,
		userID, sessionToken,
	)
	return err
}


func (db *DB) GetPasswordHashByUserID(userID int) (string, error) {
	var passwordHash string
	err := db.conn.QueryRow(`
		SELECT password_hash FROM users 
		WHERE user_id=$1
	`, userID).Scan(&passwordHash)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}