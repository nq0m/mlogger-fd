package qso

import "database/sql"

func CheckDupe(db *sql.DB, callsign, band, mode string) (bool, []string, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM qsos WHERE callsign = ? AND band = ? AND mode = ? AND is_dupe = 0`,
		callsign, band, mode,
	).Scan(&count)
	if err != nil {
		return false, nil, err
	}

	if count > 0 {
		similarCalls, err := CheckSimilarCall(db, callsign)
		return true, similarCalls, err
	}

	return false, nil, nil
}

func CheckSimilarCall(db *sql.DB, callsign string) ([]string, error) {
	rows, err := db.Query(
		`SELECT DISTINCT callsign FROM qsos WHERE callsign != ? AND (callsign LIKE ? OR ? LIKE callsign || '%') LIMIT 5`,
		callsign, callsign+"%", callsign,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		results = append(results, c)
	}
	if results == nil {
		results = []string{}
	}
	return results, nil
}
