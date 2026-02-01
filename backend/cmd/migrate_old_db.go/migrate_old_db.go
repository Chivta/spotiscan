package main

import (
	"github.com/chivta/spotiscan/internal/repository/db_client/sqlite_migration"
	
)

func main() {
	sqlite_migration.Migrate()
}