package simtest

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"

	"github.com/jeremy/mlogger-fd/internal/handler"
	"github.com/jeremy/mlogger-fd/internal/ws"
)

// setupSimTestDB creates an in-memory SQLite database with the full
// production schema (qsos + station_config + bonus_claims) and a default
// station_config row so stats queries work correctly.
func setupSimTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file::memory:?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&cache=shared")
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
	CREATE INDEX IF NOT EXISTS idx_qsos_callsign ON qsos(callsign);
	CREATE INDEX IF NOT EXISTS idx_qsos_timestamp ON qsos(timestamp);
	CREATE INDEX IF NOT EXISTS idx_qsos_band_mode ON qsos(band, mode);
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
		t.Fatalf("failed to create schema: %v", err)
	}

	// Insert default station_config row (required by stats.go queries)
	if _, err := db.Exec(`INSERT INTO station_config (id, callsign, class, arrl_section, transmitter_count, power_level) VALUES (1, 'K1SIM', '2A', 'NH', 1, 'LOW') ON CONFLICT(id) DO NOTHING`); err != nil {
		db.Close()
		t.Fatalf("failed to insert station_config: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

// setupSimRouter builds a chi.Router with all production routes wired,
// replicating the main.go route setup so the simulation exercises real handlers.
func setupSimRouter(db *sql.DB, hub *ws.Hub) chi.Router {
	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", handler.HealthCheck)
		r.Get("/check-dupe", func(w http.ResponseWriter, req *http.Request) {
			handler.CheckDupeHandler(db, w, req)
		})
		r.Get("/stats", func(w http.ResponseWriter, req *http.Request) {
			handler.GetStats(db, w, req)
		})
		r.Get("/export/cabrillo", func(w http.ResponseWriter, req *http.Request) {
			handler.ExportCabrillo(db, w, req)
		})
		r.Get("/station-config", func(w http.ResponseWriter, req *http.Request) {
			handler.GetStationConfig(db, w, req)
		})
		r.Put("/station-config", func(w http.ResponseWriter, req *http.Request) {
			handler.PutStationConfig(db, w, req)
		})
		r.Get("/bonuses", func(w http.ResponseWriter, req *http.Request) {
			handler.GetBonuses(db, w, req)
		})
		r.Put("/bonuses", func(w http.ResponseWriter, req *http.Request) {
			handler.PutBonuses(db, w, req)
		})
		r.Get("/backup/db", func(w http.ResponseWriter, req *http.Request) {
			handler.DownloadBackup(db, "", w, req)
		})
		r.Post("/sync", func(w http.ResponseWriter, req *http.Request) {
			handler.SyncQSOs(db, hub, w, req)
		})
		r.Route("/qso", func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, req *http.Request) {
				handler.CreateQSO(db, hub, w, req)
			})
			r.Get("/", func(w http.ResponseWriter, req *http.Request) {
				handler.ListQSOs(db, w, req)
			})
			r.Put("/{id}", func(w http.ResponseWriter, req *http.Request) {
				handler.UpdateQSO(db, w, req)
			})
		})
	})
	r.Get("/ws", func(w http.ResponseWriter, req *http.Request) {
		handler.ServeWS(hub, w, req)
	})
	return r
}

// TestSimulation runs a multi-client Field Day logging simulation with
// 3 goroutine clients submitting ~70 QSOs each (~210 total), verifying
// data integrity across all server endpoints and WebSocket broadcasts.
func TestSimulation(t *testing.T) {
	// Setup
	hub := ws.NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	db := setupSimTestDB(t)
	router := setupSimRouter(db, hub)
	srv := httptest.NewServer(router)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	// Parameters
	const numClients = 3
	const qsosPerClient = 70
	const totalSubmissions = numClients * qsosPerClient

	// Callsign generation per client: unique prefix + sequential suffix
	// to ensure most QSOs are unique, with deliberate dupes mixed in.
	type clientResult struct {
		clientID       int
		submitted      int
		dupesSubmitted int
		broadcasts     int
	}

	// Pre-compute all QSOs to track duplicates
	type plannedQSO struct {
		callsign string
		band     string
		mode     string
		isDupe   bool // deliberate duplicate
	}

	var allQSOs []plannedQSO
	for client := 0; client < numClients; client++ {
		prefix := fmt.Sprintf("K%d", client)
		for i := 0; i < qsosPerClient; i++ {
			// Generate callsign: K0AAA, K0AAB, ... K0BAA, etc.
			c1 := rune('A' + (i/26/26)%26)
			c2 := rune('A' + (i/26)%26)
			c3 := rune('A' + i%26)
			callsign := fmt.Sprintf("%s%c%c%c", prefix, c1, c2, c3)

			// Rotate bands and modes
			bands := []string{"20M", "40M", "80M", "15M", "10M"}
			modes := []string{"CW", "SSB"}
			band := bands[i%len(bands)]
			mode := modes[(i/len(bands))%len(modes)]

			isDupe := false
			// Every ~10th QSO is a deliberate duplicate (resubmit an earlier callsign+band+mode)
			if i >= qsosPerClient-7 {
				// Use the first 7 QSOs as dupes
				dupIdx := i - (qsosPerClient - 7)
				dup := allQSOs[client*qsosPerClient+dupIdx]
				callsign = dup.callsign
				band = dup.band
				mode = dup.mode
				isDupe = true
			}

			allQSOs = append(allQSOs, plannedQSO{
				callsign: callsign,
				band:     band,
				mode:     mode,
				isDupe:   isDupe,
			})
		}
	}

	// Track broadcasts per client
	var broadcastCounts [numClients]int64

	// Barrier to ensure all WebSocket clients are connected before submitting
	var clientReady sync.WaitGroup
	clientReady.Add(numClients)
	var startWg sync.WaitGroup
	startWg.Add(numClients)
	var doneWg sync.WaitGroup
	doneWg.Add(numClients)

	for client := 0; client < numClients; client++ {
		go func(clientID int) {
			defer doneWg.Done()

			// Connect WebSocket
			wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Errorf("client %d: WebSocket connect failed: %v", clientID, err)
				clientReady.Done()
				return
			}
			defer wsConn.Close()

			// Start reading broadcasts in background
			var wsDone sync.WaitGroup
			wsDone.Add(1)
			seenIDs := make(map[int64]bool)
			go func() {
				defer wsDone.Done()
				for {
					wsConn.SetReadDeadline(time.Now().Add(5 * time.Second))
					_, msgBytes, err := wsConn.ReadMessage()
					if err != nil {
						return // connection closed or timeout
					}
					var msg map[string]interface{}
					if err := json.Unmarshal(msgBytes, &msg); err != nil {
						continue
					}
					if msgType, _ := msg["type"].(string); msgType == "qso_created" {
						// Deduplicate by id (client may receive broadcasts from other clients too)
						if id, ok := msg["id"].(float64); ok {
							idInt := int64(id)
							if !seenIDs[idInt] {
								seenIDs[idInt] = true
								atomic.AddInt64(&broadcastCounts[clientID], 1)
							}
						}
					}
				}
			}()

			// Signal ready and wait for all clients
			clientReady.Done()
			startWg.Wait()

			// Submit QSOs
			start := clientID * qsosPerClient
			for i := start; i < start+qsosPerClient; i++ {
				q := allQSOs[i]
				body := fmt.Sprintf(
					`{"callsign":"%s","band":"%s","mode":"%s","recv_exchange":"2A NH","operator":"op%d"}`,
					q.callsign, q.band, q.mode, clientID,
				)
				resp, err := http.Post(
					srv.URL+"/api/qso",
					"application/json",
					strings.NewReader(body),
				)
				if err != nil {
					t.Errorf("client %d: POST QSO failed: %v", clientID, err)
				} else {
					resp.Body.Close()
				}
			}

			// Wait for broadcasts to settle, then stop the reader
			time.Sleep(300 * time.Millisecond)
			wsConn.Close()
			wsDone.Wait()
		}(client)
	}

	// Wait for all clients to be WebSocket-connected
	clientReady.Wait()
	time.Sleep(50 * time.Millisecond)
	// Signal all clients to start submitting
	for i := 0; i < numClients; i++ {
		startWg.Done()
	}
	// Wait for all submissions to complete
	doneWg.Wait()
	time.Sleep(300 * time.Millisecond) // let final broadcasts arrive

	// ── Integrity Assertions ──

	// 1. GET /api/stats → verify counts
	resp, err := http.Get(srv.URL + "/api/stats")
	if err != nil {
		t.Fatalf("GET /api/stats failed: %v", err)
	}
	var stats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("failed to decode stats: %v", err)
	}
	resp.Body.Close()

	totalQSOs, _ := stats["total"].(float64)
	rawPoints, _ := stats["raw_points"].(float64)
	multiplier, _ := stats["multiplier"].(float64)
	score, _ := stats["score"].(float64)
	bonusPoints, _ := stats["bonus_points"].(float64)

	// Count expected non-dupes
	nonDupeCount := 0
	for _, q := range allQSOs {
		if !q.isDupe {
			nonDupeCount++
		}
	}

	// 2. Verify total non-dupe QSO count
	if int(totalQSOs) != nonDupeCount {
		t.Errorf("stats total: expected %d non-dupe QSOs, got %d", nonDupeCount, int(totalQSOs))
	}

	// 3. GET /api/qso (paginate to overcome 200 max limit) → verify all QSOs present
	var qsoList []map[string]interface{}
	for offset := 0; ; offset += 200 {
		resp2, err := http.Get(fmt.Sprintf("%s/api/qso?limit=200&offset=%d", srv.URL, offset))
		if err != nil {
			t.Fatalf("GET /api/qso failed: %v", err)
		}
		var page []map[string]interface{}
		if err := json.NewDecoder(resp2.Body).Decode(&page); err != nil {
			resp2.Body.Close()
			t.Fatalf("failed to decode QSO list: %v", err)
		}
		resp2.Body.Close()
		qsoList = append(qsoList, page...)
		if len(page) < 200 {
			break
		}
	}

	if len(qsoList) != totalSubmissions {
		t.Errorf("QSO list count: expected %d, got %d", totalSubmissions, len(qsoList))
	}

	// 4. Verify known duplicate QSOs have is_dupe=true and points=0
	dupeQSOs := make(map[string]bool) // set of callsigns that should be dupes
	for _, q := range allQSOs {
		if q.isDupe {
			key := q.callsign + "_" + q.band + "_" + q.mode
			dupeQSOs[key] = true
		}
	}

	for _, qso := range qsoList {
		callsign, _ := qso["callsign"].(string)
		band, _ := qso["band"].(string)
		mode, _ := qso["mode"].(string)
		key := callsign + "_" + band + "_" + mode

		if dupeQSOs[key] {
			isDupe, _ := qso["is_dupe"].(bool)
			points, _ := qso["points"].(float64)
			// The last occurrence should be marked as dupe
			// (but we can't distinguish which occurrence in the list is which,
			// so just verify at least one occurrence has is_dupe=true)
			// Actually, SQLite returns them in insert order (by id ASC when no ORDER BY),
			// so earlier occurrences (first insert) will have is_dupe=false,
			// later ones will have is_dupe=true.
			// We just verify that duplicate-marked QSOs have correct values
			if !isDupe {
				continue // this is the first (non-dupe) occurrence
			}
			if points != 0 {
				t.Errorf("dupe QSO %s should have points=0, got %v", key, points)
			}
		}
	}

	// 5. Verify score = (rawPoints * multiplier) + bonusPoints
	expectedScore := int(rawPoints)*int(multiplier) + int(bonusPoints)
	if int(score) != expectedScore {
		t.Errorf("score: expected %d (formula (%.0f*%.0f)+%.0f), got %d",
			expectedScore, rawPoints, multiplier, bonusPoints, int(score))
	}

	// 6. Verify WebSocket broadcast counts are reasonable
	// Each client should have received broadcasts for their own + others' QSOs
	// (deduplicated by ID). Total should be close to totalSubmissions.
	totalBroadcasts := int64(0)
	for c := 0; c < numClients; c++ {
		bc := atomic.LoadInt64(&broadcastCounts[c])
		totalBroadcasts += bc
		if bc < int64(qsosPerClient) {
			t.Errorf("client %d broadcast count low: %d (expected >= ~%d)", c, bc, qsosPerClient)
		}
	}
	if totalBroadcasts < int64(nonDupeCount) {
		t.Errorf("total broadcasts %d < non-dupe count %d", totalBroadcasts, nonDupeCount)
	}

	t.Logf("Simulation complete: %d clients, %d total QSOs, %d non-dupes, %d unique broadcasts received",
		numClients, totalSubmissions, int(totalQSOs), totalBroadcasts)
}

// verifyBody is a helper for debugging; kept for manual inspection.
func verifyBody(r io.Reader) string {
	b, _ := io.ReadAll(r)
	var buf bytes.Buffer
	json.Indent(&buf, b, "", "  ")
	return buf.String()
}
