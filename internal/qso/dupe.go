package qso

import "database/sql"

func CheckDupe(db *sql.DB, callsign, band, mode string) (bool, []string, error) {
	return false, nil, nil
}

func CheckSimilarCall(db *sql.DB, callsign string) ([]string, error) {
	return nil, nil
}
