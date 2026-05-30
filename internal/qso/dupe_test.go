package qso

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
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

func TestCheckDupe_ExactMatch(t *testing.T) {
	db := setupTestDB(t)

	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange)
		VALUES ('2026-06-27T18:00:00Z', 'K1ABC', '20M', 'CW', '1D EMA', '2A NH')`)

	isDupe, _, err := CheckDupe(db, "K1ABC", "20M", "CW")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isDupe {
		t.Error("expected isDupe=true for same callsign+band+mode")
	}
}

func TestCheckDupe_DifferentMode(t *testing.T) {
	db := setupTestDB(t)

	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange)
		VALUES ('2026-06-27T18:00:00Z', 'K1ABC', '20M', 'CW', '1D EMA', '2A NH')`)

	isDupe, _, err := CheckDupe(db, "K1ABC", "20M", "SSB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isDupe {
		t.Error("expected isDupe=false for different mode")
	}
}

func TestCheckDupe_DifferentBand(t *testing.T) {
	db := setupTestDB(t)

	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange)
		VALUES ('2026-06-27T18:00:00Z', 'K1ABC', '20M', 'SSB', '1D EMA', '2A NH')`)

	isDupe, _, err := CheckDupe(db, "K1ABC", "40M", "SSB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isDupe {
		t.Error("expected isDupe=false for different band")
	}
}

func TestCheckDupe_EmptyDB(t *testing.T) {
	db := setupTestDB(t)

	isDupe, calls, err := CheckDupe(db, "K1ABC", "20M", "CW")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isDupe {
		t.Error("expected isDupe=false for empty database")
	}
	if len(calls) != 0 {
		t.Errorf("expected empty similar calls, got %v", calls)
	}
}

func TestCheckDupe_SkipsAlreadyDuped(t *testing.T) {
	db := setupTestDB(t)

	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, is_dupe, points)
		VALUES ('2026-06-27T18:00:00Z', 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 1, 0)`)

	isDupe, _, err := CheckDupe(db, "K1ABC", "20M", "CW")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isDupe {
		t.Error("expected isDupe=false when prior QSO is already a dupe")
	}
}

func TestSimilarCalls_PrefixMatch(t *testing.T) {
	db := setupTestDB(t)

	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange)
		VALUES ('2026-06-27T18:00:00Z', 'K1XX', '20M', 'SSB', '1D EMA', '2A NH')`)

	calls, err := CheckSimilarCall(db, "K1X")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) == 0 {
		t.Error("expected similar calls for prefix match")
	}
	found := false
	for _, c := range calls {
		if c == "K1XX" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected K1XX in similar calls, got %v", calls)
	}
}

func TestSimilarCalls_NoMatch(t *testing.T) {
	db := setupTestDB(t)

	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange)
		VALUES ('2026-06-27T18:00:00Z', 'W1AW', '20M', 'SSB', '1D EMA', '2A NH')`)

	calls, err := CheckSimilarCall(db, "K1X")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 0 {
		t.Errorf("expected empty similar calls, got %v", calls)
	}
}

func TestSimilarCalls_ExcludesSelf(t *testing.T) {
	db := setupTestDB(t)

	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange)
		VALUES ('2026-06-27T18:00:00Z', 'K1XX', '20M', 'SSB', '1D EMA', '2A NH')`)

	calls, err := CheckSimilarCall(db, "K1XX")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range calls {
		if c == "K1XX" {
			t.Error("similar calls should not include self")
		}
	}
}
