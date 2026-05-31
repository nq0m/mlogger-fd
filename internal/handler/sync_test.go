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

func setupSyncTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file::memory:?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&cache=shared")
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS qsos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		client_id TEXT UNIQUE,
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

func TestSyncQSOs_EmptyBody(t *testing.T) {
	db := setupSyncTestDB(t)

	req := httptest.NewRequest("POST", "/api/sync", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	SyncQSOs(db, nil, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSyncQSOs_ValidBatch(t *testing.T) {
	db := setupSyncTestDB(t)

	body := `{"qsos":[{"client_id":"test-uuid-1","callsign":"K1ABC","band":"20M","mode":"CW","recv_exchange":"2A NH","sent_exchange":"1D EMA","operator":"OP1"}]}`
	req := httptest.NewRequest("POST", "/api/sync", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	SyncQSOs(db, nil, rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	synced, _ := resp["synced"].(float64)
	if synced != 1 {
		t.Errorf("expected 1 synced, got %v", synced)
	}

	mappings, ok := resp["mappings"].([]interface{})
	if !ok || len(mappings) != 1 {
		t.Errorf("expected 1 mapping, got %v", mappings)
	}
}

func TestSyncQSOs_ValidationFailure(t *testing.T) {
	db := setupSyncTestDB(t)

	body := `{"qsos":[{"client_id":"test-uuid-2","band":"20M","mode":"CW","recv_exchange":"2A NH"}]}`
	req := httptest.NewRequest("POST", "/api/sync", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	SyncQSOs(db, nil, rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing callsign, got %d", rec.Code)
	}
}

func TestSyncQSOs_DedupClientID(t *testing.T) {
	db := setupSyncTestDB(t)

	body := `{"qsos":[{"client_id":"test-uuid-3","callsign":"K1ABC","band":"20M","mode":"CW","recv_exchange":"2A NH","sent_exchange":"1D EMA","operator":"OP1"}]}`
	req := httptest.NewRequest("POST", "/api/sync", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	SyncQSOs(db, nil, rec, req)

	req2 := httptest.NewRequest("POST", "/api/sync", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	SyncQSOs(db, nil, rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 on replay, got %d: %s", rec2.Code, rec2.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rec2.Body.Bytes(), &resp)
	synced, _ := resp["synced"].(float64)
	if synced != 0 {
		t.Errorf("expected 0 synced on replay (dedup), got %v", synced)
	}
}
