package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestNewHub verifies that NewHub returns an initialized Hub.
func TestNewHub(t *testing.T) {
	h := NewHub()

	if h.clients == nil {
		t.Error("Hub.clients map should be initialized")
	}
	if len(h.clients) != 0 {
		t.Error("Hub.clients should be empty")
	}
	if h.broadcast == nil {
		t.Error("Hub.broadcast channel should be initialized")
	}
	if h.Register == nil {
		t.Error("Hub.Register channel should be initialized")
	}
	if h.Unregister == nil {
		t.Error("Hub.Unregister channel should be initialized")
	}
	if cap(h.broadcast) != 256 {
		t.Errorf("broadcast channel buffer should be 256, got %d", cap(h.broadcast))
	}
}

// TestHubRegister verifies clients are added via register channel.
func TestHubRegister(t *testing.T) {
	h := NewHub()
	go h.Run()

	// Wait for Run to start
	time.Sleep(10 * time.Millisecond)

	client := &Client{
		Hub:  h,
		Send: make(chan []byte, 64),
	}
	h.Register <- client

	// Give Run() time to process
	time.Sleep(10 * time.Millisecond)

	h.mu.RLock()
	_, ok := h.clients[client]
	h.mu.RUnlock()

	if !ok {
		t.Error("client should be registered in clients map")
	}
	if len(h.clients) != 1 {
		t.Errorf("expected 1 client, got %d", len(h.clients))
	}
}

// TestHubUnregister verifies clients are removed and send channel closed.
func TestHubUnregister(t *testing.T) {
	h := NewHub()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	client := &Client{
		Hub:  h,
		Send: make(chan []byte, 64),
	}

	// Register first
	h.Register <- client
	time.Sleep(10 * time.Millisecond)

	// Then unregister
		h.Unregister <- client
	time.Sleep(10 * time.Millisecond)

	h.mu.RLock()
	_, ok := h.clients[client]
	h.mu.RUnlock()

	if ok {
		t.Error("client should be removed from clients map after unregister")
	}
	if len(h.clients) != 0 {
		t.Errorf("expected 0 clients, got %d", len(h.clients))
	}

	// Verify send channel was closed
	select {
	case _, open := <-client.Send:
		if open {
			t.Error("client send channel should be closed after unregister")
		}
	default:
		// Channel might block if not closed — that's a failure
		t.Error("client send channel should be closed (non-blocking read not possible with unbuffered read on closed chan)")
	}
}

// TestHubBroadcast verifies Broadcast marshals JSON and delivers to registered clients.
func TestHubBroadcast(t *testing.T) {
	h := NewHub()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	// Register a client to receive the broadcast
	c := &Client{Hub: h, Send: make(chan []byte, 64)}
	h.Register <- c
	time.Sleep(10 * time.Millisecond)

	msg := map[string]string{"type": "test", "data": "hello"}
	err := h.Broadcast(msg)
	if err != nil {
		t.Fatalf("Broadcast failed: %v", err)
	}

	// The message should be fanned out to the registered client
	select {
	case data := <-c.Send:
		var decoded map[string]string
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal broadcast data: %v", err)
		}
		if decoded["type"] != "test" {
			t.Errorf("expected type='test', got %q", decoded["type"])
		}
		if decoded["data"] != "hello" {
			t.Errorf("expected data='hello', got %q", decoded["data"])
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("broadcast message not delivered to client")
	}
}

// TestHubRunFanOut verifies broadcast messages are fanned out to all clients.
func TestHubRunFanOut(t *testing.T) {
	h := NewHub()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	// Register two clients
	c1 := &Client{Hub: h, Send: make(chan []byte, 64)}
	c2 := &Client{Hub: h, Send: make(chan []byte, 64)}
	h.Register <- c1
	h.Register <- c2
	time.Sleep(10 * time.Millisecond)

	// Send a broadcast message directly into the broadcast channel
	testMsg := []byte(`{"type":"test","value":42}`)
	h.broadcast <- testMsg

	// Both clients should receive the message
	for i, c := range []*Client{c1, c2} {
		select {
		case msg := <-c.Send:
			if string(msg) != string(testMsg) {
				t.Errorf("client %d received wrong message: %s", i, string(msg))
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("client %d did not receive broadcast message", i)
		}
	}
}

// TestHubNonBlockingSend verifies that a full send channel does not block the hub.
func TestHubNonBlockingSend(t *testing.T) {
	h := NewHub()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	// Create client with small send buffer
	c := &Client{Hub: h, Send: make(chan []byte, 1)}
	h.Register <- c
	time.Sleep(10 * time.Millisecond)

	// Fill the send channel to capacity
	c.Send <- []byte(`full`)

	// Now send broadcast - should not block despite full client send buffer
	msg := []byte(`{"type":"overflow"}`)
	done := make(chan bool, 1)
	go func() {
		h.broadcast <- msg
		done <- true
	}()

	select {
	case <-done:
		// Success — broadcast did not block
	case <-time.After(500 * time.Millisecond):
		t.Error("broadcast blocked on client with full send channel")
	}
}

// TestHubConcurrent verifies no race conditions with concurrent operations.
func TestHubConcurrent(t *testing.T) {
	h := NewHub()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	var wg sync.WaitGroup

	// Register many clients concurrently
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			c := &Client{Hub: h, Send: make(chan []byte, 64)}
			h.Register <- c
			time.Sleep(1 * time.Millisecond)

			// Broadcast a message
			h.broadcast <- []byte(`{"id":` + string(rune('0'+id%10)) + `}`)

			time.Sleep(1 * time.Millisecond)

			// Unregister
			h.Unregister <- c
		}(i)
	}

	wg.Wait()

	// After all operations, check that the hub is still functional
	time.Sleep(50 * time.Millisecond)

	// Should be able to register a new client
	c := &Client{Hub: h, Send: make(chan []byte, 64)}
	h.Register <- c
	time.Sleep(10 * time.Millisecond)

	h.mu.RLock()
	clientCount := len(h.clients)
	h.mu.RUnlock()

	if clientCount != 1 {
		t.Errorf("expected 1 client after concurrent test, got %d", clientCount)
	}
}

// wsTestServer creates a test HTTP server with WebSocket upgrade.
func wsTestServer(t *testing.T, handler func(*Hub)) (*Hub, *httptest.Server) {
	t.Helper()
	h := NewHub()
	go h.Run()
	time.Sleep(10 * time.Millisecond)

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Logf("upgrade error: %v", err)
			return
		}

		client := &Client{
			Hub:  h,
			Conn: conn,
			Send: make(chan []byte, 64),
		}
		h.Register <- client

		go client.WritePump()
		client.ReadPump()
	}))

	t.Cleanup(srv.Close)
	return h, srv
}

// TestClientWritePump verifies messages on Send channel are written to WebSocket.
func TestClientWritePump(t *testing.T) {
	h, srv := wsTestServer(t, nil)

	// Convert http:// to ws://
	url := "ws" + strings.TrimPrefix(srv.URL, "http")

	// Connect a WebSocket client
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("failed to connect WebSocket: %v", err)
	}
	defer ws.Close()

	// Wait for the client to be registered
	time.Sleep(20 * time.Millisecond)

	// Broadcast a message through the hub — this will fan out to all clients
	// The WritePump reads from client.Send and writes to the WebSocket connection
	testMsg := map[string]string{"type": "write_pump_test", "value": "pong"}
	if err := h.Broadcast(testMsg); err != nil {
		t.Fatalf("Broadcast failed: %v", err)
	}

	// Read the message from the WebSocket client — proves WritePump wrote it
	ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msgBytes, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read WebSocket message: %v", err)
	}

	var decoded map[string]string
	if err := json.Unmarshal(msgBytes, &decoded); err != nil {
		t.Fatalf("failed to unmarshal received message: %v", err)
	}
	if decoded["type"] != "write_pump_test" {
		t.Errorf("expected type='write_pump_test', got %q", decoded["type"])
	}
	if decoded["value"] != "pong" {
		t.Errorf("expected value='pong', got %q", decoded["value"])
	}
}
