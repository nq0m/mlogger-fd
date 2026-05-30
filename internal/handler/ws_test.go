package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jeremy/mlogger-fd/internal/ws"
)

// TestServeWSUpgrade verifies HTTP GET /ws upgrades to WebSocket.
func TestServeWSUpgrade(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeWS(hub, w, r)
	}))
	defer srv.Close()

	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("WebSocket upgrade failed: %v", err)
	}
	defer conn.Close()

	// Verify client is registered
	time.Sleep(20 * time.Millisecond)

	// Send a ping to verify connection is alive
	conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		t.Fatalf("failed to send ping: %v", err)
	}
}

// TestCreateQSOBroadcast verifies qso_created is broadcast to WebSocket clients.
func TestCreateQSOBroadcast(t *testing.T) {
	db := setupHandlerTestDB(t)
	hub := ws.NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	// Create test server that handles /ws and /api/qso
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/ws") {
			ServeWS(hub, w, r)
			return
		}
		// QSO handler
		CreateQSO(db, hub, w, r)
	}))
	defer srv.Close()

	// Connect WebSocket client
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket connect failed: %v", err)
	}
	defer wsConn.Close()

	time.Sleep(20 * time.Millisecond)

	// Submit a QSO via HTTP POST
	body := `{"callsign":"K1ABC","band":"20M","mode":"CW","recv_exchange":"2A NH"}`
	resp, err := http.Post(srv.URL+"/api/qso", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("HTTP POST failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	// Read the broadcast from WebSocket
	wsConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msgBytes, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read broadcast message: %v", err)
	}

	var msg map[string]interface{}
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		t.Fatalf("failed to unmarshal broadcast: %v", err)
	}

	// Verify the message structure
	if msg["type"] != "qso_created" {
		t.Errorf("expected type='qso_created', got %v", msg["type"])
	}
	if msg["callsign"] != "K1ABC" {
		t.Errorf("expected callsign='K1ABC', got %v", msg["callsign"])
	}
	if msg["band"] != "20M" {
		t.Errorf("expected band='20M', got %v", msg["band"])
	}
	if msg["mode"] != "CW" {
		t.Errorf("expected mode='CW', got %v", msg["mode"])
	}
	if msg["recv_exchange"] != "2A NH" {
		t.Errorf("expected recv_exchange='2A NH', got %v", msg["recv_exchange"])
	}
	if _, ok := msg["id"]; !ok {
		t.Error("expected 'id' field in broadcast")
	}
	if _, ok := msg["is_dupe"]; !ok {
		t.Error("expected 'is_dupe' field in broadcast")
	}
	if _, ok := msg["points"]; !ok {
		t.Error("expected 'points' field in broadcast")
	}
}

// TestCreateQSOWithNilHub verifies backward compatibility when hub is nil.
func TestCreateQSOWithNilHub(t *testing.T) {
	db := setupHandlerTestDB(t)

	body := `{"callsign":"K1ABC","band":"20M","mode":"CW","recv_exchange":"2A NH"}`
	req := httptest.NewRequest("POST", "/api/qso", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	CreateQSO(db, nil, rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["is_dupe"] != false {
		t.Error("first QSO should not be dupe")
	}
}

// TestCreateQSOConcurrent verifies multiple concurrent inserts succeed.
func TestCreateQSOConcurrent(t *testing.T) {
	db := setupHandlerTestDB(t)
	hub := ws.NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	errChan := make(chan error, 3)

	for i := 0; i < 3; i++ {
		go func(callsign string) {
			body := `{"callsign":"` + callsign + `","band":"20M","mode":"SSB","recv_exchange":"2A NH"}`
			req := httptest.NewRequest("POST", "/api/qso", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			CreateQSO(db, hub, rec, req)
			if rec.Code != http.StatusCreated {
				errChan <- fmt.Errorf("expected 201, got %d", rec.Code)
				return
			}
			errChan <- nil
		}(string(rune('A' + i)) + "1ZZZ")
	}

	for i := 0; i < 3; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("concurrent QSO insert failed: %v", err)
		}
	}
}

// TestWebSocketMultipleMessages verifies multiple QSOs are broadcast.
func TestWebSocketMultipleMessages(t *testing.T) {
	db := setupHandlerTestDB(t)
	hub := ws.NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/ws") {
			ServeWS(hub, w, r)
			return
		}
		CreateQSO(db, hub, w, r)
	}))
	defer srv.Close()

	// Connect WebSocket
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket connect failed: %v", err)
	}
	defer wsConn.Close()

	time.Sleep(20 * time.Millisecond)

	// Submit two QSOs
	for _, callsign := range []string{"W1AW", "K1ZZ"} {
		body := `{"callsign":"` + callsign + `","band":"40M","mode":"LSB","recv_exchange":"2A NH"}`
		resp, err := http.Post(srv.URL+"/api/qso", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatalf("HTTP POST failed: %v", err)
		}
		resp.Body.Close()
	}

	// Read both messages
	wsConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for i := 0; i < 2; i++ {
		_, msgBytes, err := wsConn.ReadMessage()
		if err != nil {
			t.Fatalf("failed to read broadcast %d: %v", i, err)
		}
		var msg map[string]interface{}
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			t.Fatalf("failed to unmarshal broadcast %d: %v", i, err)
		}
		if msg["type"] != "qso_created" {
			t.Errorf("message %d: expected type='qso_created'", i)
		}
	}
}
