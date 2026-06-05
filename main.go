package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jeremy/mlogger-fd/internal/db"
	"github.com/jeremy/mlogger-fd/internal/handler"
	"github.com/jeremy/mlogger-fd/internal/ws"
)

//go:embed frontend/build/*
var staticFiles embed.FS

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	dbPath := os.Getenv("FDLOGGER_DB_PATH")
	if dbPath == "" {
		dbPath = "fdlogger.db"
	}

	database, err := db.Open(dbPath)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	hub := ws.NewHub()
	go hub.Run()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", handler.HealthCheck)
		r.Get("/check-dupe", func(w http.ResponseWriter, r *http.Request) {
			handler.CheckDupeHandler(database, w, r)
		})
		r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
			handler.GetStats(database, w, r)
		})
		r.Get("/export/cabrillo", func(w http.ResponseWriter, r *http.Request) {
			handler.ExportCabrillo(database, w, r)
		})
		r.Get("/station-config", func(w http.ResponseWriter, r *http.Request) {
			handler.GetStationConfig(database, w, r)
		})
		r.Put("/station-config", func(w http.ResponseWriter, r *http.Request) {
			handler.PutStationConfig(database, w, r)
		})
		r.Get("/bonuses", func(w http.ResponseWriter, r *http.Request) {
			handler.GetBonuses(database, w, r)
		})
		r.Put("/bonuses", func(w http.ResponseWriter, r *http.Request) {
			handler.PutBonuses(database, w, r)
		})
		r.Get("/backup/db", func(w http.ResponseWriter, r *http.Request) {
			handler.DownloadBackup(database, dbPath, w, r)
		})
		r.Post("/sync", func(w http.ResponseWriter, r *http.Request) {
			handler.SyncQSOs(database, hub, w, r)
		})
		r.Route("/qso", func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				handler.CreateQSO(database, hub, w, r)
			})
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				handler.ListQSOs(database, w, r)
			})
			r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) {
				handler.UpdateQSO(database, w, r)
			})
		})
	})

	// WebSocket endpoint
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeWS(hub, w, r)
	})

	r.Get("/*", spaHandler())

	port := os.Getenv("FDLOGGER_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("starting server", "port", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func spaHandler() http.HandlerFunc {
	content, err := fs.Sub(staticFiles, "frontend/build")
	if err != nil {
		panic("embedded frontend/build not found: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(content))

	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		f, err := content.Open(path)
		if err != nil {
			r.URL.Path = "/"
		} else {
			f.Close()
		}
		fileServer.ServeHTTP(w, r)
	}
}
