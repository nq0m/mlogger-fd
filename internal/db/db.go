package db

import (
	"database/sql"
	"log/slog"
	"os"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	var mode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		db.Close()
		return nil, err
	}
	slog.Info("database opened", "path", path, "journal_mode", mode)

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	schema, err := os.ReadFile("internal/db/schema.sql")
	if err != nil {
		db.Close()
		return nil, err
	}

	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
