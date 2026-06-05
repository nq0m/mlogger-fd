package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/jeremy/mlogger-fd/internal/model"
)

// GetBonuses returns all bonus claims as a JSON map keyed by bonus_id.
// If no claims exist in the database, returns all 18 default bonus items
// with claimed=false and count=0.
func GetBonuses(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	// Build the response map initialized from default bonuses
	resp := make(map[string]model.BonusClaim, len(model.DefaultBonuses))
	for _, item := range model.DefaultBonuses {
		resp[item.ID] = model.BonusClaim{
			BonusID: item.ID,
			Claimed: false,
			Count:   0,
		}
	}

	// Overlay any persisted claims from the database
	rows, err := db.Query("SELECT bonus_id, claimed, count, updated_at FROM bonus_claims")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var claim model.BonusClaim
		var claimedInt int
		if err := rows.Scan(&claim.BonusID, &claimedInt, &claim.Count, &claim.UpdatedAt); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
			return
		}
		claim.Claimed = claimedInt != 0
		resp[claim.BonusID] = claim
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// PutBonuses accepts a JSON map of bonus claims, validates each entry,
// and persists valid claims in a single database transaction.
// Unknown bonus IDs are silently skipped. Invalid counts return 400.
func PutBonuses(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var body map[string]model.BonusClaim
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}

	// Validate all entries first (fail-fast on invalid counts)
	for id, claim := range body {
		claim.BonusID = id

		// Check if bonus_id is a recognized type
		found := false
		for _, item := range model.DefaultBonuses {
			if item.ID == id {
				found = true
				break
			}
		}
		if !found {
			continue // silently skip unknown bonus IDs
		}

		if errMsg := model.ValidateBonusClaim(id, claim.Claimed, claim.Count); errMsg != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": errMsg})
			return
		}
	}

	// Begin transaction for atomic persistence
	tx, err := db.Begin()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO bonus_claims (bonus_id, claimed, count, updated_at) VALUES (?, ?, ?, datetime('now'))`)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}
	defer stmt.Close()

	for id, claim := range body {
		// Skip unknown bonus IDs (validated above)
		found := false
		for _, item := range model.DefaultBonuses {
			if item.ID == id {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		claimedInt := 0
		if claim.Claimed {
			claimedInt = 1
		}
		if _, err := stmt.Exec(id, claimedInt, claim.Count); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	// Return full state (same as GET)
	GetBonuses(db, w, r)
}
