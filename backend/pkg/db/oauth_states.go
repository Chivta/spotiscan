package db

func (db *DB) CreateOathState(state string) error {
	_, err := db.conn.Exec(
		`INSERT INTO oauth_states (state, created_at, expires_at) VALUES ($1, NOW(), NOW() + INTERVAL '10 minutes')`,
		state,
	)
	return err
}

func (db *DB) DeleteOathState(state string) error {
	_, err := db.conn.Exec(`
		DELETE FROM oauth_states WHERE state=$1
	`, state)
	return err
}

func (db *DB) StateExists(state string) (bool, error) {
	rows,err := db.conn.Query(`
			SELECT 1 FROM oauth_states 
			WHERE state=$1 AND expires_at > NOW()
		`, state)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		return true, nil
	}
	return false, nil
}