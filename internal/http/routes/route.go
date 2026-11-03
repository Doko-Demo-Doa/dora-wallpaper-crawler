package routes

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Deps is everything the HTTP layer needs. Handlers depend on these small
// funcs/values instead of concrete types, so routes never import the DB or R2 code.
type Deps struct {
	ListAll       func(ctx context.Context) (any, error)
	ListNew       func(ctx context.Context) (any, error)
	Crawl         func(ctx context.Context) error
	BackdoorToken string
}

// NewRouter builds the router and returns it; the caller owns the http.Server.
func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/api", func(r chi.Router) {
		r.Get("/", jsonHandler(d.ListNew, http.StatusBadGateway))
		r.Get("/all", jsonHandler(d.ListAll, http.StatusInternalServerError))
	})

	r.Route("/backdoor", func(r chi.Router) {
		r.Use(requireToken(d.BackdoorToken))
		r.Post("/start-crawler", func(w http.ResponseWriter, req *http.Request) {
			// Detach from the request so the crawl outlives this response.
			ctx := context.WithoutCancel(req.Context())
			go func() {
				if err := d.Crawl(ctx); err != nil {
					log.Printf("crawl failed: %v", err)
				}
			}()
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("Crawler started"))
		})
	})

	return r
}

func jsonHandler(fn func(context.Context) (any, error), errStatus int) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		v, err := fn(req.Context())
		if err != nil {
			http.Error(w, err.Error(), errStatus)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(w).Encode(v); err != nil {
			log.Printf("write JSON: %v", err)
		}
	}
}
