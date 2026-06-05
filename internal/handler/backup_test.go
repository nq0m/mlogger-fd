package handler

import (
	"bytes"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// setupBackupTestDB creates a temporary SQLite file, opens it, creates
// all schema tables, inserts a station_config row so stats queries work,
// then closes the DB and returns the file path. The caller can re-open
// the DB and pass it to DownloadBackup.
func setupBackupTestDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	f, err := os.CreateTemp("", "fdlogger_backup_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	dbPath := f.Name()
	f.Close()

	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	db.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE IF NOT EXISTS qsos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT NOT NULL,
		callsign TEXT NOT NULL,
		band TEXT NOT NULL,
		mode TEXT NOT NULL,
		sent_exchange TEXT NOT NULL,
		recv_exchange TEXT NOT NULL,
		client_id TEXT UNIQUE,
		operator TEXT,
		is_dupe INTEGER NOT NULL DEFAULT 0,
		points INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	);
	CREATE TABLE IF NOT EXISTS station_config (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		callsign TEXT NOT NULL DEFAULT 'N0CALL',
		class TEXT NOT NULL DEFAULT '1D',
		arrl_section TEXT NOT NULL DEFAULT 'EMA',
		transmitter_count INTEGER NOT NULL DEFAULT 1,
		power_level TEXT NOT NULL DEFAULT 'LOW',
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);
	CREATE TABLE IF NOT EXISTS bonus_claims (
		bonus_id TEXT PRIMARY KEY,
		claimed INTEGER NOT NULL DEFAULT 0,
		count INTEGER NOT NULL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);
	`
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		os.Remove(dbPath)
		t.Fatalf("failed to create schema: %v", err)
	}

	// Insert default station config row (required by stats queries)
	if _, err := db.Exec(`INSERT INTO station_config (id, callsign, class, arrl_section) VALUES (1, 'K1TEST', '2A', 'NH') ON CONFLICT(id) DO NOTHING`); err != nil {
		db.Close()
		os.Remove(dbPath)
		t.Fatalf("failed to insert station_config: %v", err)
	}

	// Insert a few QSOs so the backup file has content
	for i := 0; i < 5; i++ {
		_, err := db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points, is_dupe) VALUES (?, ?, ?, ?, ?, ?, ?, 0)`,
			time.Now().UTC().Format(time.RFC3339),
			"K1ABC", "20M", "CW", "2A NH", "2A NH", 2,
		)
		if err != nil {
			db.Close()
			os.Remove(dbPath)
			t.Fatalf("failed to insert test QSO: %v", err)
		}
	}

	db.Close()
	t.Cleanup(func() { os.Remove(dbPath) })
	return nil, dbPath
}

// reopenDB is a helper to reopen a file-based SQLite DB for handler testing.
func reopenDB(t *testing.T, dbPath string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("failed to reopen DB: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	return db
}

// TestBackupDownload verifies the backup handler streams the database file
// with correct HTTP headers and valid SQLite content.
func TestBackupDownload(t *testing.T) {
	_, dbPath := setupBackupTestDB(t)
	db := reopenDB(t, dbPath)

	req := httptest.NewRequest("GET", "/api/backup/db", nil)
	rec := httptest.NewRecorder()

	DownloadBackup(db, dbPath, rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Content-Type must be application/octet-stream
	ct := resp.Header.Get("Content-Type")
	if ct != "application/octet-stream" {
		t.Errorf("expected Content-Type application/octet-stream, got %q", ct)
	}

	// Content-Disposition must contain attachment with timestamped filename
	cd := resp.Header.Get("Content-Disposition")
	if !strings.HasPrefix(cd, "attachment; filename=\"fdlogger_backup_") {
		t.Errorf("expected Content-Disposition attachment with timestamped filename, got %q", cd)
	}

	// Body must start with SQLite format 3 magic header
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	sqliteMagic := []byte("SQLite format 3\x00")
	if len(body) < len(sqliteMagic) || !bytes.Equal(body[:len(sqliteMagic)], sqliteMagic) {
		t.Error("response body does not start with SQLite format 3 magic header")
	}
}

// TestBackupMissingFile verifies the handler returns 500 with a generic
// error message when the database file cannot be opened (ASVS V7).
func TestBackupMissingFile(t *testing.T) {
	_, dbPath := setupBackupTestDB(t)
	db := reopenDB(t, dbPath)

	// Remove the file so os.Open fails
	os.Remove(dbPath)

	req := httptest.NewRequest("GET", "/api/backup/db", nil)
	rec := httptest.NewRecorder()

	DownloadBackup(db, "/nonexistent/path.db", rec, req)

	resp := rec.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Failed to open database") {
		t.Errorf("expected generic error 'Failed to open database', got %q", string(body))
	}
	// ASVS V7: no filesystem path must appear in the error
	if strings.Contains(string(body), dbPath) {
		t.Error("error message must not expose filesystem path (ASVS V7)")
	}
	if strings.Contains(string(body), "/nonexistent") {
		t.Error("error message must not expose filesystem path (ASVS V7)")
	}
}

// TestBackupConcurrent verifies the backup handler does not block when
// QSOs are being created concurrently (WAL mode concurrency).
func TestBackupConcurrent(t *testing.T) {
	_, dbPath := setupBackupTestDB(t)
	db := reopenDB(t, dbPath)

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	// Goroutine: continuously insert QSOs while backup runs
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			_, err := db.Exec(`INSERT INTO qsos (timestamp, callsign, band, mode, sent_exchange, recv_exchange, points, is_dupe) VALUES (?, ?, ?, ?, ?, ?, ?, 0)`,
				time.Now().UTC().Format(time.RFC3339),
				"W1AW", "40M", "LSB", "2A NH", "2A NH", 1,
			)
			if err != nil {
				errCh <- err
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Backup while inserts are happening
	time.Sleep(20 * time.Millisecond) // ensure inserts started
	req := httptest.NewRequest("GET", "/api/backup/db", nil)
	rec := httptest.NewRecorder()

	DownloadBackup(db, dbPath, rec, req)

	resp := rec.Result()
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 during concurrent writes, got %d", resp.StatusCode)
	}

	wg.Wait()
	select {
	case err := <-errCh:
		t.Errorf("concurrent insert failed: %v", err)
	default:
	}
}
