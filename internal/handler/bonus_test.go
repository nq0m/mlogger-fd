package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jeremy/mlogger-fd/internal/model"

	_ "modernc.org/sqlite"
)

func setupBonusTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db := setupHandlerTestDB(t)
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS bonus_claims (
		bonus_id TEXT PRIMARY KEY,
		claimed INTEGER NOT NULL DEFAULT 0,
		count INTEGER NOT NULL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	if err != nil {
		t.Fatalf("failed to create bonus_claims table: %v", err)
	}
	return db
}

func TestGetBonuses_EmptyTable(t *testing.T) {
	db := setupBonusTestDB(t)

	req := httptest.NewRequest("GET", "/api/bonuses", nil)
	rec := httptest.NewRecorder()

	GetBonuses(db, rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]model.BonusClaim
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Should return all 18 bonus items with claimed=false, count=0
	if len(resp) != 18 {
		t.Errorf("expected 18 bonus items, got %d", len(resp))
	}

	for _, id := range []string{"emergency_power", "media_publicity", "gota_bonus"} {
		claim, ok := resp[id]
		if !ok {
			t.Errorf("missing bonus key %q in response", id)
			continue
		}
		if claim.Claimed {
			t.Errorf("bonus %q Claimed should be false on empty table", id)
		}
		if claim.Count != 0 {
			t.Errorf("bonus %q Count should be 0 on empty table, got %d", id, claim.Count)
		}
	}
}

func TestGetBonuses_WithData(t *testing.T) {
	db := setupBonusTestDB(t)

	// Insert some claims
	_, err := db.Exec(`INSERT OR REPLACE INTO bonus_claims (bonus_id, claimed, count, updated_at) VALUES
		('emergency_power', 1, 3, datetime('now')),
		('media_publicity', 1, 0, datetime('now'))`)
	if err != nil {
		t.Fatalf("failed to seed data: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/bonuses", nil)
	rec := httptest.NewRecorder()

	GetBonuses(db, rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]model.BonusClaim
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Claimed items should show their values
	ep, ok := resp["emergency_power"]
	if !ok {
		t.Fatal("emergency_power missing from response")
	}
	if !ep.Claimed {
		t.Error("emergency_power Claimed should be true")
	}
	if ep.Count != 3 {
		t.Errorf("emergency_power Count = %d, want 3", ep.Count)
	}

	mp, ok := resp["media_publicity"]
	if !ok {
		t.Fatal("media_publicity missing from response")
	}
	if !mp.Claimed {
		t.Error("media_publicity Claimed should be true")
	}

	// Non-claimed items should still be present with defaults
	sp, ok := resp["social_media"]
	if !ok {
		t.Fatal("social_media missing from response")
	}
	if sp.Claimed {
		t.Error("social_media Claimed should be false")
	}
	if sp.Count != 0 {
		t.Errorf("social_media Count = %d, want 0", sp.Count)
	}
}

func TestPutBonuses_Valid(t *testing.T) {
	db := setupBonusTestDB(t)

	body := `{"emergency_power":{"bonus_id":"emergency_power","claimed":true,"count":3},"media_publicity":{"bonus_id":"media_publicity","claimed":true,"count":0}}`
	req := httptest.NewRequest("PUT", "/api/bonuses", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	PutBonuses(db, rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]model.BonusClaim
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	ep, ok := resp["emergency_power"]
	if !ok {
		t.Fatal("emergency_power missing from response")
	}
	if !ep.Claimed || ep.Count != 3 {
		t.Errorf("emergency_power = {claimed:%v, count:%d}, want {true, 3}", ep.Claimed, ep.Count)
	}

	mp, ok := resp["media_publicity"]
	if !ok {
		t.Fatal("media_publicity missing from response")
	}
	if !mp.Claimed {
		t.Error("media_publicity Claimed should be true")
	}
}

func TestPutBonuses_PersistenceAndGet(t *testing.T) {
	db := setupBonusTestDB(t)

	// PUT claims
	putBody := `{"emergency_power":{"bonus_id":"emergency_power","claimed":true,"count":5}}`
	putReq := httptest.NewRequest("PUT", "/api/bonuses", strings.NewReader(putBody))
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()
	PutBonuses(db, putRec, putReq)

	if putRec.Code != http.StatusOK {
		t.Fatalf("PUT failed: %d: %s", putRec.Code, putRec.Body.String())
	}

	// GET should return the persisted claims
	getReq := httptest.NewRequest("GET", "/api/bonuses", nil)
	getRec := httptest.NewRecorder()
	GetBonuses(db, getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GET failed: %d: %s", getRec.Code, getRec.Body.String())
	}

	var resp map[string]model.BonusClaim
	if err := json.Unmarshal(getRec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	ep, ok := resp["emergency_power"]
	if !ok {
		t.Fatal("emergency_power missing from GET after PUT")
	}
	if !ep.Claimed {
		t.Error("emergency_power Claimed should be true after PUT")
	}
	if ep.Count != 5 {
		t.Errorf("emergency_power Count = %d, want 5", ep.Count)
	}
}

func TestPutBonuses_UnknownBonusID(t *testing.T) {
	db := setupBonusTestDB(t)

	body := `{"invalid_bonus":{"bonus_id":"invalid_bonus","claimed":true,"count":0},"emergency_power":{"bonus_id":"emergency_power","claimed":true,"count":2}}`
	req := httptest.NewRequest("PUT", "/api/bonuses", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	PutBonuses(db, rec, req)

	// Should still succeed (200) — unknown keys are silently ignored
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]model.BonusClaim
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// invalid_bonus should NOT be in response (silently excluded)
	if _, exists := resp["invalid_bonus"]; exists {
		t.Error("invalid_bonus should not appear in response")
	}

	// emergency_power should still have been saved
	ep, ok := resp["emergency_power"]
	if !ok {
		t.Fatal("emergency_power missing from response")
	}
	if !ep.Claimed || ep.Count != 2 {
		t.Errorf("emergency_power = {claimed:%v, count:%d}, want {true, 2}", ep.Claimed, ep.Count)
	}
}

func TestPutBonuses_NegativeCount(t *testing.T) {
	db := setupBonusTestDB(t)

	body := `{"emergency_power":{"bonus_id":"emergency_power","claimed":true,"count":-1}}`
	req := httptest.NewRequest("PUT", "/api/bonuses", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	PutBonuses(db, rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"] == "" {
		t.Error("expected error message for negative count")
	}
}

func TestPutBonuses_ExceedsMaxCount(t *testing.T) {
	db := setupBonusTestDB(t)

	body := `{"emergency_power":{"bonus_id":"emergency_power","claimed":true,"count":21}}`
	req := httptest.NewRequest("PUT", "/api/bonuses", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	PutBonuses(db, rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"] == "" {
		t.Error("expected error message for exceeding MaxCount")
	}
}

func TestPutBonuses_MalformedJSON(t *testing.T) {
	db := setupBonusTestDB(t)

	body := `not json at all`
	req := httptest.NewRequest("PUT", "/api/bonuses", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	PutBonuses(db, rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"] != "invalid JSON" {
		t.Errorf("expected error 'invalid JSON', got '%s'", resp["error"])
	}
}
