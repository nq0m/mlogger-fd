package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeremy/mlogger-fd/internal/model"
)

func setupConfigTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db := setupHandlerTestDB(t)
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS station_config (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		callsign TEXT NOT NULL DEFAULT 'N0CALL',
		class TEXT NOT NULL DEFAULT '1D',
		arrl_section TEXT NOT NULL DEFAULT 'EMA',
		transmitter_count INTEGER NOT NULL DEFAULT 1,
		power_level TEXT NOT NULL DEFAULT 'LOW',
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	if err != nil {
		t.Fatalf("failed to create station_config table: %v", err)
	}
	return db
}

func TestPutStationConfig_Valid(t *testing.T) {
	tdb := setupConfigTestDB(t)

	body := `{"callsign":"K1ABC","class":"1D","arrl_section":"EMA","transmitter_count":2,"power_level":"LOW"}`
	req := httptest.NewRequest("PUT", "/api/station-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	PutStationConfig(tdb, rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var cfg model.StationConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if cfg.Callsign != "K1ABC" {
		t.Errorf("expected callsign K1ABC, got %s", cfg.Callsign)
	}
	if cfg.TransmitterCount != 2 {
		t.Errorf("expected transmitter_count 2, got %d", cfg.TransmitterCount)
	}
}

func TestGetStationConfig_AfterPut(t *testing.T) {
	tdb := setupConfigTestDB(t)

	// First PUT a config
	putBody := `{"callsign":"K1ABC","class":"1D","arrl_section":"EMA","transmitter_count":2,"power_level":"LOW"}`
	req := httptest.NewRequest("PUT", "/api/station-config", strings.NewReader(putBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	PutStationConfig(tdb, rec, req)

	// Then GET it
	getReq := httptest.NewRequest("GET", "/api/station-config", nil)
	getRec := httptest.NewRecorder()
	GetStationConfig(tdb, getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", getRec.Code, getRec.Body.String())
	}

	var cfg model.StationConfig
	if err := json.Unmarshal(getRec.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if cfg.Callsign != "K1ABC" {
		t.Errorf("expected callsign K1ABC, got %s", cfg.Callsign)
	}
	if cfg.Class != "1D" {
		t.Errorf("expected class 1D, got %s", cfg.Class)
	}
	if cfg.ARRLSection != "EMA" {
		t.Errorf("expected arrl_section EMA, got %s", cfg.ARRLSection)
	}
	if cfg.TransmitterCount != 2 {
		t.Errorf("expected transmitter_count 2, got %d", cfg.TransmitterCount)
	}
	if cfg.PowerLevel != "LOW" {
		t.Errorf("expected power_level LOW, got %s", cfg.PowerLevel)
	}
}

func TestPutStationConfig_InvalidCallsign(t *testing.T) {
	tdb := setupConfigTestDB(t)

	body := `{"callsign":"","class":"1D","arrl_section":"EMA","transmitter_count":1,"power_level":"LOW"}`
	req := httptest.NewRequest("PUT", "/api/station-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	PutStationConfig(tdb, rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"] != "callsign is required" {
		t.Errorf("expected error 'callsign is required', got '%s'", resp["error"])
	}
}

func TestPutStationConfig_InvalidPowerLevel(t *testing.T) {
	tdb := setupConfigTestDB(t)

	body := `{"callsign":"K1ABC","class":"1D","arrl_section":"EMA","transmitter_count":1,"power_level":"MEDIUM"}`
	req := httptest.NewRequest("PUT", "/api/station-config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	PutStationConfig(tdb, rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"] != "power_level must be LOW, HIGH, or QRP" {
		t.Errorf("expected error about power_level, got '%s'", resp["error"])
	}
}

func TestGetStationConfig_EmptyDatabase(t *testing.T) {
	tdb := setupConfigTestDB(t)

	req := httptest.NewRequest("GET", "/api/station-config", nil)
	rec := httptest.NewRecorder()
	GetStationConfig(tdb, rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on empty DB, got %d: %s", rec.Code, rec.Body.String())
	}

	var cfg model.StationConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Should return defaults
	def := model.DefaultStationConfig()
	if cfg.Callsign != def.Callsign {
		t.Errorf("expected default callsign %s, got %s", def.Callsign, cfg.Callsign)
	}
	if cfg.Class != def.Class {
		t.Errorf("expected default class %s, got %s", def.Class, cfg.Class)
	}
	if cfg.ARRLSection != def.ARRLSection {
		t.Errorf("expected default section %s, got %s", def.ARRLSection, cfg.ARRLSection)
	}
	if cfg.TransmitterCount != def.TransmitterCount {
		t.Errorf("expected default tx count %d, got %d", def.TransmitterCount, cfg.TransmitterCount)
	}
	if cfg.PowerLevel != def.PowerLevel {
		t.Errorf("expected default power %s, got %s", def.PowerLevel, cfg.PowerLevel)
	}
}

func TestStationConfigPersistence(t *testing.T) {
	// Create a single DB, PUT config, then open a second connection and GET
	tdb := setupConfigTestDB(t)

	putBody := `{"callsign":"W1AW","class":"2A","arrl_section":"CT","transmitter_count":3,"power_level":"HIGH"}`
	req := httptest.NewRequest("PUT", "/api/station-config", strings.NewReader(putBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	PutStationConfig(tdb, rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("PUT failed: %d: %s", rec.Code, rec.Body.String())
	}

	// Verify via GET on same connection
	getReq := httptest.NewRequest("GET", "/api/station-config", nil)
	getRec := httptest.NewRecorder()
	GetStationConfig(tdb, getRec, getReq)

	var cfg model.StationConfig
	json.Unmarshal(getRec.Body.Bytes(), &cfg)
	if cfg.Callsign != "W1AW" {
		t.Errorf("persistence check failed: expected W1AW, got %s", cfg.Callsign)
	}
	if cfg.TransmitterCount != 3 {
		t.Errorf("persistence check failed: expected tx 3, got %d", cfg.TransmitterCount)
	}
}

func TestPutStationConfig_Overwrite(t *testing.T) {
	tdb := setupConfigTestDB(t)

	body1 := `{"callsign":"W1AW","class":"2A","arrl_section":"CT","transmitter_count":3,"power_level":"HIGH"}`
	req1 := httptest.NewRequest("PUT", "/api/station-config", strings.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	PutStationConfig(tdb, rec1, req1)

	// Second PUT with different values
	body2 := `{"callsign":"K2XYZ","class":"3A","arrl_section":"NH","transmitter_count":5,"power_level":"QRP"}`
	req2 := httptest.NewRequest("PUT", "/api/station-config", strings.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	PutStationConfig(tdb, rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("second PUT failed: %d: %s", rec2.Code, rec2.Body.String())
	}

	var cfg model.StationConfig
	json.Unmarshal(rec2.Body.Bytes(), &cfg)

	if cfg.Callsign != "K2XYZ" {
		t.Errorf("overwrite failed: expected K2XYZ, got %s", cfg.Callsign)
	}
	if cfg.Class != "3A" {
		t.Errorf("overwrite failed: expected 3A, got %s", cfg.Class)
	}
	if cfg.PowerLevel != "QRP" {
		t.Errorf("overwrite failed: expected QRP, got %s", cfg.PowerLevel)
	}
}
