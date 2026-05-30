package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func setupHandlerTestDB(t *testing.T) *sql.DB {
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

func TestCreateQSO_DupeMarking(t *testing.T) {
	db := setupHandlerTestDB(t)

	body := `{"callsign":"K1ABC","band":"20M","mode":"CW","recv_exchange":"2A NH"}`
	req := httptest.NewRequest("POST", "/api/qso", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	CreateQSO(db, rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var first map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &first)
	if first["is_dupe"] != false {
		t.Error("first QSO should not be dupe")
	}
	pts, _ := first["points"].(float64)
	if pts != 2 {
		t.Errorf("CW should be 2 points, got %v", pts)
	}

	req2 := httptest.NewRequest("POST", "/api/qso", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()

	CreateQSO(db, rec2, req2)
	if rec2.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var second map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &second)
	if second["is_dupe"] != true {
		t.Error("second QSO with same callsign+band+mode should be dupe")
	}
	pts2, _ := second["points"].(float64)
	if pts2 != 0 {
		t.Errorf("dupe QSO should have 0 points, got %v", pts2)
	}
	if ids, ok := second["similar_calls"].([]interface{}); !ok || len(ids) == 0 {
		t.Error("dupe QSO should include similar_calls")
	}
}

func TestCreateQSO_Validation(t *testing.T) {
	db := setupHandlerTestDB(t)

	body := `{"band":"20M","mode":"CW","recv_exchange":"2A NH"}`
	req := httptest.NewRequest("POST", "/api/qso", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	CreateQSO(db, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing callsign, got %d", rec.Code)
	}
}
