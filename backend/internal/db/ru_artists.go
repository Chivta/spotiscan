package db

import (
	"github.com/lib/pq"
	
	"github.com/chivta/spotiscan/internal/models"
)

func (db *DB) FilterRussian(artists map[string]models.Artist) (map[string]models.Artist, error) {
	var names []string
	for name := range artists {
		names = append(names, name)
	}

	rows, err := db.conn.Query(`
        SELECT name FROM ru_artists WHERE name = ANY($1)
    `, pq.Array(names))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ruNames []string

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		ruNames = append(ruNames, name)
	}
	ruArtists := make(map[string]models.Artist, len(ruNames))

	for _, name := range ruNames {
		ruArtists[name] = artists[name]
	}

	return ruArtists, nil
}
