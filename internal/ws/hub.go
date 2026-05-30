package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	started    bool
}

// Client represents a single WebSocket connection.
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

// NewHub creates a new Hub with initialized channels and client map.
func NewHub() *Hub {
	return &Hub{}
}

// Run starts the Hub's main loop. Must be called in a goroutine.
func (h *Hub) Run() {
}

// Broadcast marshals v to JSON and pushes it onto the broadcast channel.
func (h *Hub) Broadcast(v interface{}) error {
	return nil
}

// ReadPump reads messages from the WebSocket connection.
// Exits when the connection is closed or an error occurs.
func (c *Client) ReadPump() {
}

// WritePump writes messages from the Send channel to the WebSocket connection.
// Exits when the Send channel is closed.
func (c *Client) WritePump() {
}
