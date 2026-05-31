package handler

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jeremy/mlogger-fd/internal/model"
	"github.com/jeremy/mlogger-fd/internal/qso"
	"github.com/jeremy/mlogger-fd/internal/ws"
)

func SyncQSOs(db *sql.DB, hub *ws.Hub, w http.ResponseWriter, r *http.Request) {
	var input struct {
		QSOs []struct {
			ClientID     string `json:"client_id"`
			Callsign     string `json:"callsign"`
			Band         string `json:"band"`
			Mode         string `json:"mode"`
			RecvExchange string `json:"recv_exchange"`
			SentExchange string `json:"sent_exchange"`
			Operator     string `json:"operator"`
		} `json:"qsos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}

	for _, q := range input.QSOs {
		inputModel := model.CreateQSOInput{
			ClientID:     q.ClientID,
			Callsign:     q.Callsign,
			Band:         q.Band,
			Mode:         q.Mode,
			RecvExchange: q.RecvExchange,
			SentExchange: q.SentExchange,
			Operator:     q.Operator,
		}
		if msg := model.ValidateRequired(inputModel); msg != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": msg})
			return
		}
	}

	type mapping struct {
		ClientID string `json:"client_id"`
		ServerID int64  `json:"server_id"`
	}

	now := time.Now().UTC().Format(time.RFC3339)
	var mappings []mapping
	synced := 0

	for _, q := range input.QSOs {
		sentExchange := q.SentExchange
		if sentExchange == "" {
			sentExchange = "1D EMA"
		}

		isDupe, _, err := qso.CheckDupe(db, q.Callsign, q.Band, q.Mode)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
			return
		}

		points := qso.CalculatePoints(q.Mode, isDupe)
		isDupeInt := 0
		if isDupe {
			isDupeInt = 1
		}

		result, err := db.Exec(
			`INSERT INTO qsos (client_id, timestamp, callsign, band, mode, sent_exchange, recv_exchange, operator, is_dupe, points, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(client_id) DO NOTHING`,
			q.ClientID, now, q.Callsign, q.Band, q.Mode,
			sentExchange, q.RecvExchange, q.Operator,
			isDupeInt, points, now,
		)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "database error"})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			id, _ := result.LastInsertId()
			synced++
			mappings = append(mappings, mapping{ClientID: q.ClientID, ServerID: id})

			if hub != nil {
				if err := hub.Broadcast(map[string]interface{}{
					"type":          "qso_created",
					"id":            id,
					"client_id":     q.ClientID,
					"timestamp":     now,
					"callsign":      q.Callsign,
					"band":          q.Band,
					"mode":          q.Mode,
					"recv_exchange": q.RecvExchange,
					"sent_exchange": sentExchange,
					"operator":      q.Operator,
					"is_dupe":       isDupe,
					"points":        points,
				}); err != nil {
					slog.Warn("failed to broadcast synced QSO", "error", err)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"synced":   synced,
		"mappings": mappings,
	})
}
