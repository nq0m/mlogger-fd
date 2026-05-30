package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/jeremy/mlogger-fd/internal/model"
)

func GetStationConfig(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var cfg model.StationConfig
	err := db.QueryRow(`SELECT callsign, class, arrl_section, transmitter_count, power_level
		FROM station_config WHERE id = 1`).Scan(
		&cfg.Callsign, &cfg.Class, &cfg.ARRLSection,
		&cfg.TransmitterCount, &cfg.PowerLevel,
	)
	if err == sql.ErrNoRows {
		cfg = model.DefaultStationConfig()
	} else if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cfg)
}

func PutStationConfig(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var cfg model.StationConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}

	if msg := model.ValidateStationConfig(cfg); msg != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": msg})
		return
	}

	_, err := db.Exec(`INSERT OR REPLACE INTO station_config
		(id, callsign, class, arrl_section, transmitter_count, power_level, updated_at)
		VALUES (1, ?, ?, ?, ?, ?, datetime('now'))`,
		cfg.Callsign, cfg.Class, cfg.ARRLSection,
		cfg.TransmitterCount, cfg.PowerLevel,
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cfg)
}
