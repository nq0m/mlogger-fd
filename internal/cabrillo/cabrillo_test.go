package cabrillo

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupCabrilloTestDB(t *testing.T) *sql.DB {
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
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS station_config (
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

func setupCabrilloTestDB_NoConfig(t *testing.T) *sql.DB {
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

func TestGenerate_EmptyDB(t *testing.T) {
	db := setupCabrilloTestDB(t)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "START-OF-LOG: 3.0") {
		t.Error("missing START-OF-LOG header")
	}
	if !strings.Contains(result, "CONTEST: ARRL-FIELD-DAY") {
		t.Error("missing CONTEST header")
	}
	if !strings.Contains(result, "CALLSIGN: N0CALL") {
		t.Error("missing CALLSIGN header")
	}
	if !strings.Contains(result, "END-OF-LOG:") {
		t.Error("missing END-OF-LOG footer")
	}
}

func TestGenerate_WithQSOs(t *testing.T) {
	db := setupCabrilloTestDB(t)

	ts := "2026-06-27T18:00:00Z"
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 2)`, ts)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "QSO:") {
		t.Error("expected at least one QSO line")
	}
	if !strings.Contains(result, "K1ABC") {
		t.Error("expected K1ABC in output")
	}
}

func TestGenerate_DupeQSO(t *testing.T) {
	db := setupCabrilloTestDB(t)

	ts := "2026-06-27T18:00:00Z"
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points, is_dupe)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 0, 1)`, ts)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "dupe") {
		t.Error("dupe QSO should be flagged as dupe")
	}
}

func TestGenerate_ModeMapping(t *testing.T) {
	db := setupCabrilloTestDB(t)

	ts := "2026-06-27T18:00:00Z"
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '20M', 'SSB', '1D EMA', '2A NH', 1)`, ts)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "PH") {
		t.Error("SSB should map to PH in Cabrillo")
	}
}

func TestGenerate_BandToFreq(t *testing.T) {
	db := setupCabrilloTestDB(t)

	ts := "2026-06-27T18:00:00Z"
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '40M', 'CW', '1D EMA', '2A NH', 2)`, ts)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "7000") {
		t.Error("40M should map to 7000 kHz")
	}
}

func TestGenerate_ScoreCalculation(t *testing.T) {
	db := setupCabrilloTestDB(t)

	ts := "2026-06-27T18:00:00Z"
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 2)`, ts)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'W1AW', '40M', 'SSB', '1D EMA', '1D RI', 1)`, ts)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "CLAIMED-SCORE:") {
		t.Error("missing CLAIMED-SCORE header")
	}
}

func TestGenerate_ScoringWithBonus(t *testing.T) {
	db := setupCabrilloTestDB(t)

	ts := "2026-06-27T18:00:00Z"
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 2)`, ts)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'W1AW', '40M', 'SSB', '1D EMA', '1D RI', 1)`, ts)

	// Add a claimed bonus
	db.Exec(`INSERT OR REPLACE INTO bonus_claims (bonus_id, claimed, count, updated_at)
		VALUES ('media_publicity', 1, 0, datetime('now'))`)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "CLAIMED-SCORE:") {
		t.Error("missing CLAIMED-SCORE header")
	}

	// rawPoints=3, multiplier=2, bonus=100
	// ARRL: (3*2)+100 = 106
	if !strings.Contains(result, "CLAIMED-SCORE: 106") {
		t.Error("expected CLAIMED-SCORE: 106 (bonus added after multiplier)")
	}
}

func TestGenerate_BonusSOAPBOX(t *testing.T) {
	db := setupCabrilloTestDB(t)

	ts := "2026-06-27T18:00:00Z"
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 2)`, ts)

	db.Exec(`INSERT OR REPLACE INTO bonus_claims (bonus_id, claimed, count, updated_at)
		VALUES ('media_publicity', 1, 0, datetime('now'))`)
	db.Exec(`INSERT OR REPLACE INTO bonus_claims (bonus_id, claimed, count, updated_at)
		VALUES ('safety_officer', 1, 0, datetime('now'))`)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "SOAPBOX: Bonus:") {
		t.Error("expected SOAPBOX: Bonus: lines in Cabrillo output")
	}
	if !strings.Contains(result, "SOAPBOX: Total Bonus Points") {
		t.Error("expected SOAPBOX: Total Bonus Points line in Cabrillo output")
	}
	// media_publicity=100 + safety_officer=100 = 200 total bonus
	if !strings.Contains(result, "SOAPBOX: Total Bonus Points = 200") {
		t.Error("expected SOAPBOX: Total Bonus Points = 200")
	}
}

func TestGenerate_HeaderFormat(t *testing.T) {
	db := setupCabrilloTestDB(t)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "CREATED-BY: FDLogger") {
		t.Error("missing CREATED-BY header")
	}
	if !strings.Contains(result, "CATEGORY-OPERATOR:") {
		t.Error("missing CATEGORY-OPERATOR header")
	}
	if !strings.Contains(result, "CATEGORY-POWER:") {
		t.Error("missing CATEGORY-POWER header")
	}
	if !strings.Contains(result, "CATEGORY-STATION: PORTABLE") {
		t.Error("missing CATEGORY-STATION: PORTABLE header")
	}
}

func TestGenerate_DateFormat(t *testing.T) {
	db := setupCabrilloTestDB(t)

	ts := "2026-06-27T18:30:00Z"
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 2)`, ts)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "2026-06-27") {
		t.Error("missing YYYY-MM-DD date format")
	}
	if !strings.Contains(result, "1830") {
		t.Error("missing HHMM time format for 18:30 UTC")
	}
}

func TestGenerate_ExchangePadding(t *testing.T) {
	db := setupCabrilloTestDB(t)

	tn := time.Now().UTC().Format(time.RFC3339)
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'K1ABC', '20M', 'CW', '1D EMA', '2A NH', 2)`, tn)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "K1ABC") && !strings.Contains(result, "2A NH") {
		t.Error("expected callsign and exchange in QSO line")
	}
}

func TestGenerate_WithConfig(t *testing.T) {
	db := setupCabrilloTestDB(t)

	db.Exec(`INSERT OR REPLACE INTO station_config (id, callsign, class, arrl_section, power_level)
		VALUES (1, 'K1ABC', '2A', 'SNJ', 'HIGH')`)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "CALLSIGN: K1ABC") {
		t.Error("expected CALLSIGN: K1ABC from station_config")
	}
	if !strings.Contains(result, "ARRL-SECTION: SNJ") {
		t.Error("expected ARRL-SECTION: SNJ from station_config")
	}
	if !strings.Contains(result, "CATEGORY-POWER: HIGH") {
		t.Error("expected CATEGORY-POWER: HIGH from station_config")
	}
	if !strings.Contains(result, "CATEGORY-CLASS: 2A") {
		t.Error("expected CATEGORY-CLASS: 2A from station_config")
	}
}

func TestGenerate_ConfigFallback_NoRow(t *testing.T) {
	db := setupCabrilloTestDB(t)
	// station_config table exists but no row inserted — should fall back to defaults

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "CALLSIGN: N0CALL") {
		t.Error("expected fallback CALLSIGN: N0CALL when no config row")
	}
	if !strings.Contains(result, "ARRL-SECTION: NH") {
		t.Error("expected fallback ARRL-SECTION: NH when no config row")
	}
}

func TestGenerate_ConfigFallback_NoTable(t *testing.T) {
	db := setupCabrilloTestDB_NoConfig(t)
	// station_config table does not exist at all — should fall back to defaults

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "CALLSIGN: N0CALL") {
		t.Error("expected fallback CALLSIGN: N0CALL when no station_config table")
	}
	if !strings.Contains(result, "ARRL-SECTION: NH") {
		t.Error("expected fallback ARRL-SECTION: NH when no station_config table")
	}
}

func TestGenerate_ConfigFallback_EmptyClass(t *testing.T) {
	db := setupCabrilloTestDB(t)

	db.Exec(`INSERT OR REPLACE INTO station_config (id, callsign, class, arrl_section, power_level)
		VALUES (1, 'W1AW', '', 'CT', 'LOW')`)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "CATEGORY-CLASS: 1D") {
		t.Error("expected fallback CATEGORY-CLASS: 1D when class is empty string")
	}
}

func TestGenerate_ConfigDoesNotAffectQSOs(t *testing.T) {
	db := setupCabrilloTestDB(t)

	db.Exec(`INSERT OR REPLACE INTO station_config (id, callsign, class, arrl_section, power_level)
		VALUES (1, 'K1ABC', '2A', 'SNJ', 'HIGH')`)

	ts := "2026-06-27T18:00:00Z"
	db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points)
		VALUES (?, 'W1XYZ', '20M', 'CW', '1D EMA', '2A NH', 2)`, ts)

	result, err := Generate(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// QSO line should still contain the QSO callsign, not the station callsign
	if !strings.Contains(result, "QSO:") {
		t.Error("expected QSO lines in output")
	}
	if !strings.Contains(result, "W1XYZ") {
		t.Error("expected W1XYZ QSO in output")
	}
	if !strings.Contains(result, "CLAIMED-SCORE:") {
		t.Error("expected CLAIMED-SCORE in output")
	}
}
