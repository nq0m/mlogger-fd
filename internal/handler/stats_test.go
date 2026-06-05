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
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS bonus_claims (
		bonus_id TEXT PRIMARY KEY,
		claimed INTEGER NOT NULL DEFAULT 0,
		count INTEGER NOT NULL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	if err != nil {
		t.Fatalf("failed to create bonus_claims table: %v", err)
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

func TestGetStats_BonusPointsField(t *testing.T) {
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

	bonusPoints, ok := stats["bonus_points"]
	if !ok {
		t.Fatal("expected bonus_points field in stats response")
	}
	bp, _ := bonusPoints.(float64)
	if bp != 0 {
		t.Errorf("expected bonus_points=0 with no claims, got %v", bp)
	}
}

func TestGetStats_BonusPointsWithClaims(t *testing.T) {
	db := setupStatsTestDB(t)

	// Insert some claimed bonuses
	db.Exec(`INSERT OR REPLACE INTO bonus_claims (bonus_id, claimed, count, updated_at)
		VALUES ('media_publicity', 1, 0, datetime('now'))`)
	db.Exec(`INSERT OR REPLACE INTO bonus_claims (bonus_id, claimed, count, updated_at)
		VALUES ('emergency_power', 1, 3, datetime('now'))`)

	now := time.Now().UTC().Format(time.RFC3339)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 2)`, now)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	rec := httptest.NewRecorder()
	GetStats(db, rec, req)

	var stats map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &stats)

	bonusPoints, ok := stats["bonus_points"]
	if !ok {
		t.Fatal("expected bonus_points field in stats response")
	}
	bp, _ := bonusPoints.(float64)
	// media_publicity = 100, emergency_power = 3 * 100 = 300, total = 400
	if bp != 400 {
		t.Errorf("expected bonus_points=400 (100+300), got %v", bp)
	}
}

func TestGetStats_ScoreIncludesBonus(t *testing.T) {
	db := setupStatsTestDB(t)

	// Add a QSO for raw points and multiplier
	now := time.Now().UTC().Format(time.RFC3339)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 2)`, now)

	// Add a claimed bonus
	db.Exec(`INSERT OR REPLACE INTO bonus_claims (bonus_id, claimed, count, updated_at)
		VALUES ('safety_officer', 1, 0, datetime('now'))`)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	rec := httptest.NewRecorder()
	GetStats(db, rec, req)

	var stats map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &stats)

	score, _ := stats["score"].(float64)
	rawPoints, _ := stats["raw_points"].(float64)
	multiplier, _ := stats["multiplier"].(float64)
	bonusPoints, _ := stats["bonus_points"].(float64)

	// ARRL formula: score = (rawPoints * multiplier) + bonusPoints
	// rawPoints=2, multiplier=1, bonusPoints=100 → score=102
	expectedScore := (rawPoints * multiplier) + bonusPoints
	if score != expectedScore {
		t.Errorf("expected score=%v ((%.0f*%.0f)+%.0f), got %v", expectedScore, rawPoints, multiplier, bonusPoints, score)
	}
	// Bonus must NOT be multiplied: score should be 102, not (2+100)*1 = 102... hmm both are 102 here
	// Let's use a better case: rawPoints=2, multiplier=2, bonusPoints=100
	// Correct: (2*2)+100 = 104. Wrong: (2+100)*2 = 204
}

func TestGetStats_BonusNotMultiplied(t *testing.T) {
	db := setupStatsTestDB(t)

	now := time.Now().UTC().Format(time.RFC3339)
	// Two QSOs on different band/mode = multiplier of 2
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 2)`, now)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'W1AW', '40M', 'SSB', '1D EMA', '1D RI', 1)`, now)

	db.Exec(`INSERT OR REPLACE INTO bonus_claims (bonus_id, claimed, count, updated_at)
		VALUES ('media_publicity', 1, 0, datetime('now'))`)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	rec := httptest.NewRecorder()
	GetStats(db, rec, req)

	var stats map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &stats)

	score, _ := stats["score"].(float64)
	// rawPoints=3, multiplier=2, bonusPoints=100
	// Correct ARRL: (3*2)+100 = 106
	// Wrong (bonus multiplied): (3+100)*2 = 206
	if score != 106 {
		t.Errorf("expected score=106 (bonus added after multiplier), got %v", score)
	}
}
