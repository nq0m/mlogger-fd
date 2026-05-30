package handler

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jeremy/mlogger-fd/internal/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // trusted LAN per AGENTS.md
	},
}

// ServeWS upgrades an HTTP connection to WebSocket and registers with the Hub.
func ServeWS(hub *ws.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}

	client := &ws.Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 64),
	}

	// Register with hub
	hub.Register <- client

	// Start write pump in goroutine
	go client.WritePump()

	// Read pump blocks until disconnect
	client.ReadPump()
}
