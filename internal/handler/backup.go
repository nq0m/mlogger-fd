package handler

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DownloadBackup streams the SQLite database file as an HTTP download with
// a timestamped filename. It uses io.Copy for efficient streaming (WAL mode
// guarantees the .db file is always consistent, even while writers are active).
func DownloadBackup(db *sql.DB, dbPath string, w http.ResponseWriter, r *http.Request) {
	// Flush committed WAL pages to the main .db file (best-effort)
	db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")

	now := time.Now().UTC().Format("20060102_150405")
	filename := fmt.Sprintf("fdlogger_backup_%s.db", now)

	f, err := os.Open(dbPath)
	if err != nil {
		http.Error(w, "Failed to open database", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	if _, err := io.Copy(w, f); err != nil {
		// Client may have disconnected — not a server error
		return
	}
}
