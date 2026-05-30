package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupStatsTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_pragma=journal_mode(WAL)")
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS qsos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT NOT NULL,
		callsign TEXT NOT NULL,
		band TEXT NOT NULL,
		mode TEXT NOT NULL,
		sent_exchange TEXT NOT NULL,
		recv_exchange TEXT NOT NULL,
		operator TEXT,
		is_dupe INTEGER NOT NULL DEFAULT 0,
		points INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestGetStats_EmptyDB(t *testing.T) {
	db := setupStatsTestDB(t)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	rec := httptest.NewRecorder()
	GetStats(db, rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var stats map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &stats); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	total, _ := stats["total"].(float64)
	if total != 0 {
		t.Errorf("expected total=0, got %v", total)
	}
	score, _ := stats["score"].(float64)
	if score != 0 {
		t.Errorf("expected score=0, got %v", score)
	}
	multiplier, _ := stats["multiplier"].(float64)
	if multiplier != 1 {
		t.Errorf("expected multiplier=1, got %v", multiplier)
	}
}

func TestGetStats_WithQSOs(t *testing.T) {
	db := setupStatsTestDB(t)

	now := time.Now().UTC().Format(time.RFC3339)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 2)`, now)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'W1AW', '40M', 'SSB', '1D EMA', '1D RI', 1)`, now)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	rec := httptest.NewRecorder()
	GetStats(db, rec, req)

	var stats map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &stats)

	total, _ := stats["total"].(float64)
	if total != 2 {
		t.Errorf("expected total=2, got %v", total)
	}
	rawPoints, _ := stats["raw_points"].(float64)
	if rawPoints != 3 {
		t.Errorf("expected raw_points=3 (2+1), got %v", rawPoints)
	}
	multiplier, _ := stats["multiplier"].(float64)
	if multiplier != 2 {
		t.Errorf("expected multiplier=2 (20M_CW + 40M_SSB), got %v", multiplier)
	}
	score, _ := stats["score"].(float64)
	if score != 6 {
		t.Errorf("expected score=6 (3×2), got %v", score)
	}
}

func TestGetStats_DupesExcluded(t *testing.T) {
	db := setupStatsTestDB(t)

	now := time.Now().UTC().Format(time.RFC3339)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points, is_dupe)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 2, 0)`, now)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points, is_dupe)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 0, 1)`, now)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	rec := httptest.NewRecorder()
	GetStats(db, rec, req)

	var stats map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &stats)

	total, _ := stats["total"].(float64)
	if total != 1 {
		t.Errorf("expected total=1 (excluding dupe), got %v", total)
	}
	rawPoints, _ := stats["raw_points"].(float64)
	if rawPoints != 2 {
		t.Errorf("expected raw_points=2 (excluding dupe), got %v", rawPoints)
	}
}

func TestGetStats_RateWindows(t *testing.T) {
	db := setupStatsTestDB(t)

	now := time.Now().UTC()
	recent := now.Add(-5 * time.Minute).Format(time.RFC3339)
	old := now.Add(-2 * time.Hour).Format(time.RFC3339)

	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 2)`, recent)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'W1AW', '40M', 'SSB', '1D EMA', '1D RI', 1)`, old)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	rec := httptest.NewRecorder()
	GetStats(db, rec, req)

	var stats map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &stats)

	rate10min, _ := stats["rate_10min"].(float64)
	if rate10min == 0 {
		t.Error("expected rate_10min > 0 for recent QSO")
	}
}

func TestGetStats_Breakdown(t *testing.T) {
	db := setupStatsTestDB(t)

	now := time.Now().UTC().Format(time.RFC3339)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 2)`, now)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1DEF', '20M', 'CW', '1D EMA', '2A NH', 2)`, now)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'W1AW', '40M', 'SSB', '1D EMA', '1D RI', 1)`, now)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	rec := httptest.NewRecorder()
	GetStats(db, rec, req)

	var stats map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &stats)

	breakdown, ok := stats["breakdown"].(map[string]interface{})
	if !ok {
		t.Fatal("expected breakdown to be a map")
	}

	cw20, _ := breakdown["20M_CW"].(float64)
	if cw20 != 2 {
		t.Errorf("expected 20M_CW=2, got %v", cw20)
	}
	ssb40, _ := breakdown["40M_SSB"].(float64)
	if ssb40 != 1 {
		t.Errorf("expected 40M_SSB=1, got %v", ssb40)
	}
}
