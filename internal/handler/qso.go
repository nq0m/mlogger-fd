package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jeremy/mlogger-fd/internal/model"
	"github.com/jeremy/mlogger-fd/internal/qso"
	"github.com/jeremy/mlogger-fd/internal/ws"
)

func CreateQSO(db *sql.DB, hub *ws.Hub, w http.ResponseWriter, r *http.Request) {
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

	isDupe, similarCalls, err := qso.CheckDupe(db, input.Callsign, input.Band, input.Mode)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	points := qso.CalculatePoints(input.Mode, isDupe)
	isDupeInt := 0
	if isDupe {
		isDupeInt = 1
	}

	result, err := db.Exec(
		`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, operator, is_dupe, points, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now, input.Callsign, input.Band, input.Mode,
		input.SentExchange, input.RecvExchange, input.Operator,
		isDupeInt, points, now,
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
		"is_dupe":       isDupe,
		"similar_calls": similarCalls,
		"points":        points,
		"timestamp":     now,
	})
}

func ListQSOs(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	searchStr := r.URL.Query().Get("search")

	limit := 50
	offset := 0

	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
		limit = l
	}
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
		offset = o
	}

	var rows *sql.Rows
	var err error

	if searchStr != "" {
		rows, err = db.Query(
			`SELECT id, timestamp, callsign, band, mode, recv_exchange, is_dupe, points
			 FROM qsos WHERE callsign LIKE ? ORDER BY timestamp DESC LIMIT ? OFFSET ?`,
			searchStr+"%", limit, offset,
		)
	} else {
		rows, err = db.Query(
			`SELECT id, timestamp, callsign, band, mode, recv_exchange, is_dupe, points
			 FROM qsos ORDER BY timestamp DESC LIMIT ? OFFSET ?`,
			limit, offset,
		)
	}
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

func UpdateQSO(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid QSO ID"})
		return
	}

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

	points := qso.CalculatePoints(input.Mode, false)

	result, err := db.Exec(
		`UPDATE qsos SET callsign=?, band=?, mode=?, recv_exchange=?, sent_exchange=?, operator=?, points=? WHERE id=?`,
		input.Callsign, input.Band, input.Mode, input.RecvExchange,
		input.SentExchange, input.Operator, points, id,
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "QSO not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":            id,
		"callsign":      input.Callsign,
		"band":          input.Band,
		"mode":          input.Mode,
		"recv_exchange": input.RecvExchange,
		"points":        points,
	})
}
