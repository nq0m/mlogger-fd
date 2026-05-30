package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

func GetStats(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{})
}
