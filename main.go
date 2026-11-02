package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	if err := runMigrations(ctx, db); err != nil {
		log.Fatal(err)
	}
	store := &Store{DB: db}
	crawler := &Crawler{Store: store, NetworkClient: NewNetworkClient()}
	go crawler.RunHourly(ctx)

	r := chi.NewRouter()
	r.Get("/api/", func(w http.ResponseWriter, req *http.Request) {
		links, err := getDesktopLinks(req.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		known, err := store.ExistingLinks(req.Context(), links)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, difference(links, known))
	})
	r.Get("/api/all", func(w http.ResponseWriter, req *http.Request) {
		rows, err := store.All(req.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, rows)
	})
	r.Get("/api/list-files", func(w http.ResponseWriter, req *http.Request) {
		files := []string{}
		_ = filepath.Walk("./temp", func(path string, info os.FileInfo, err error) error {
			if err == nil && info != nil && !info.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		writeJSON(w, files)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	srv := &http.Server{Addr: ":" + port, Handler: r, ReadHeaderTimeout: 10 * time.Second}
	stop, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	go func() {
		<-stop.Done()
		shutdownCtx, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		_ = srv.Shutdown(shutdownCtx)
	}()
	log.Printf("Dora wallpaper crawler listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write JSON: %v", err)
	}
}
