package sqlite_migration

// This function is for migration artists from old sqlite database to postgres database

import (
	"database/sql"
	"log"

	"github.com/chivta/spotiscan/internal/config"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func Migrate() {
	sqlite_db, err := sql.Open("sqlite", "file:bot_data.db")
	if err != nil {
		panic(err)
	}
	err = sqlite_db.Ping()
	if err != nil {
		panic(err)
	}
	defer sqlite_db.Close()

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	postgres_db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}
	err = postgres_db.Ping()
	if err != nil {
		panic(err)
	}
	defer postgres_db.Close()

	artists := make(map[string]struct{}, 25138) // known number of russian artists in old db

	rows, err := sqlite_db.Query("SELECT name FROM artists")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			panic(err)
		}
		artists[name] = struct{}{}
	}

	artistsSlice := make([]string, 0, len(artists))
	for name := range artists {
		artistsSlice = append(artistsSlice, name)
	}
	
	artists = nil // free memory

	log.Println("len artists", len(artistsSlice))
	if len(artistsSlice) > 0 {
		// Use PostgreSQL array and unnest for bulk insert
		query := "INSERT INTO ru_artists (name) SELECT unnest($1::text[])"
		_, err := postgres_db.Exec(query, pq.Array(artistsSlice))
		if err != nil {
			panic(err)
		}
	}
}
