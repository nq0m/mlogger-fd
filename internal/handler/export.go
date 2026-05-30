package handler

import (
	"database/sql"
	"net/http"

	"github.com/jeremy/mlogger-fd/internal/cabrillo"
)

func ExportCabrillo(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	result, err := cabrillo.Generate(db)
	if err != nil {
		http.Error(w, "Failed to generate Cabrillo file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=\"n0call_field_day.cbr\"")
	w.Write([]byte(result))
}
