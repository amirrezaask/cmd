package infra

import "database/sql"

var (
	MainDB    *sql.DB
	MetricsDB *sql.DB
)
