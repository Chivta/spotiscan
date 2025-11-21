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

func (db *DB) Close() error {
	err := db.conn.Close()
	if err != nil {
		log.Println("Error closing database connection:", err)
	}
	return err
}

