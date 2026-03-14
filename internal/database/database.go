package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func Open(dsn string) (*sql.DB, error) {
	// Append pragmas to the DSN so they apply to every connection from the pool.
	// This ensures foreign keys, WAL mode, and busy timeout are always active.
	connector := dsn + "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", connector)
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return db, nil
}
