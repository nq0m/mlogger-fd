package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/jeremy/mlogger-fd/internal/cabrillo"
)

func ExportCabrillo(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	result, err := cabrillo.Generate(db)
	if err != nil {
		http.Error(w, "Failed to generate Cabrillo file", http.StatusInternalServerError)
		return
	}

	// Read callsign from station_config for filename, fall back to n0call
	callsign := "n0call"
	var c string
	if err := db.QueryRow("SELECT callsign FROM station_config WHERE id = 1").Scan(&c); err == nil && c != "" {
		callsign = strings.ToLower(c)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_field_day.cbr\"", callsign))
	w.Write([]byte(result))
}
