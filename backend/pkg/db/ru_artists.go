package db

import (
)

func (db *DB) IsRussian(name string) (bool, error) {
	row,err := db.conn.Query(`
			SELECT 1 FROM ru_artists 
			WHERE name=$1
	`, name)
	if err != nil {
		return false, err
	}
	defer row.Close()
	return row.Next(), nil
}