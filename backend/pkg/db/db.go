package db

import (
	"database/sql"
	"log"

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
	log.Println(passwordHash)
	err = tx.QueryRow(
		`INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING user_id`,
		username, email, passwordHash,
	).Scan(&userID)
	if err != nil {
		return err
	}


	// TODO: set session expiration time properly
	_, err = tx.Exec(
		`INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, NOW() + INTERVAL '30 days')`,
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