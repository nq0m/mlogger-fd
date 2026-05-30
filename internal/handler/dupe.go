package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/jeremy/mlogger-fd/internal/qso"
)

func CheckDupeHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	callsign := r.URL.Query().Get("callsign")
	band := r.URL.Query().Get("band")
	mode := r.URL.Query().Get("mode")

	if callsign == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"is_dupe":       false,
			"similar_calls": []string{},
		})
		return
	}

	isDupe, similarCalls, err := qso.CheckDupe(db, callsign, band, mode)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	if similarCalls == nil {
		similarCalls = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"is_dupe":       isDupe,
		"similar_calls": similarCalls,
	})
}
