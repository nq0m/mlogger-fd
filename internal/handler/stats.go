package handler

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"time"
)

func GetStats(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()

	var total int
	var rawPoints int
	if err := db.QueryRow("SELECT COUNT(*), COALESCE(SUM(points), 0) FROM qsos WHERE is_dupe = 0").Scan(&total, &rawPoints); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	tenMinAgo := now.Add(-10 * time.Minute).Format(time.RFC3339)
	var q10min int
	if err := db.QueryRow("SELECT COUNT(*) FROM qsos WHERE timestamp >= ?", tenMinAgo).Scan(&q10min); err != nil {
		q10min = 0
	}
	rate10min := math.Round(float64(q10min) / 10.0 * 60.0)

	oneHourAgo := now.Add(-1 * time.Hour).Format(time.RFC3339)
	var q1hr int
	if err := db.QueryRow("SELECT COUNT(*) FROM qsos WHERE timestamp >= ?", oneHourAgo).Scan(&q1hr); err != nil {
		q1hr = 0
	}
	rate1hr := math.Round(float64(q1hr))

	var multiplier int
	if err := db.QueryRow("SELECT COUNT(DISTINCT band || '_' || mode) FROM qsos WHERE is_dupe = 0").Scan(&multiplier); err != nil {
		multiplier = 1
	}
	if multiplier < 1 {
		multiplier = 1
	}

	score := rawPoints * multiplier

	rows, err := db.Query("SELECT band, mode, COUNT(*) FROM qsos WHERE is_dupe = 0 GROUP BY band, mode ORDER BY band, mode")
	if err != nil {
		rows = nil
	}

	breakdown := make(map[string]int)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var band, mode string
			var count int
			if err := rows.Scan(&band, &mode, &count); err != nil {
				continue
			}
			breakdown[band+"_"+mode] = count
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":      total,
		"raw_points": rawPoints,
		"multiplier": multiplier,
		"score":      score,
		"rate_10min": rate10min,
		"rate_1hr":   rate1hr,
		"breakdown":  breakdown,
	})
}
