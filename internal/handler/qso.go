package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jeremy/mlogger-fd/internal/model"
	"github.com/jeremy/mlogger-fd/internal/qso"
)

func CreateQSO(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	var input model.CreateQSOInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}

	if msg := model.ValidateRequired(input); msg != "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": msg})
		return
	}

	if input.SentExchange == "" {
		input.SentExchange = "1D EMA"
	}

	now := time.Now().UTC().Format(time.RFC3339)
	points := qso.CalculatePoints(input.Mode, false)
	isDupe := 0

	result, err := db.Exec(
		`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, operator, is_dupe, points, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, input.Callsign, input.Band, input.Mode,
		input.SentExchange, input.RecvExchange, input.Operator,
		isDupe, points, now,
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	id, _ := result.LastInsertId()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":            id,
		"is_dupe":       false,
		"similar_calls": []string{},
		"points":        points,
		"timestamp":     now,
	})
}

func ListQSOs(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
		limit = l
	}
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	rows, err := db.Query(
		`SELECT id, timestamp, callsign, band, mode, recv_exchange, is_dupe, points
		 FROM qsos ORDER BY timestamp DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}
	defer rows.Close()

	var qsos []map[string]interface{}
	for rows.Next() {
		var id int64
		var timestamp, callsign, band, mode, recvExchange string
		var isDupe, points int

		if err := rows.Scan(&id, &timestamp, &callsign, &band, &mode, &recvExchange, &isDupe, &points); err != nil {
			continue
		}

		qsos = append(qsos, map[string]interface{}{
			"id":            id,
			"timestamp":     timestamp,
			"callsign":      callsign,
			"band":          band,
			"mode":          mode,
			"recv_exchange": recvExchange,
			"is_dupe":       isDupe == 1,
			"points":        points,
		})
	}

	if qsos == nil {
		qsos = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(qsos)
}
