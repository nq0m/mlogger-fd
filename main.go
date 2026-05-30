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
		r.Route("/qso", func(r chi.Router) {
			r.Post("/", func(w http.ResponseWriter, r *http.Request) {
				handler.CreateQSO(database, w, r)
			})
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				handler.ListQSOs(database, w, r)
			})
		})
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
