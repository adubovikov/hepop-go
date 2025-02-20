package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sipcapture/hepop-go/internal/writer"
	"github.com/sirupsen/logrus"
)

type API struct {
	config  *Config
	writer  writer.Writer
	router  *chi.Mux
	metrics *Metrics
	server  *http.Server
}

type Config struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	EnableMetrics bool   `yaml:"enable_metrics"`
	EnablePprof   bool   `yaml:"enable_pprof"`
	AuthToken     string `yaml:"auth_token"`
}

func NewAPI(config *Config, writer writer.Writer) *API {
	api := &API{
		config:  config,
		writer:  writer,
		router:  chi.NewRouter(),
		metrics: NewMetrics(),
		server: &http.Server{
			Addr: fmt.Sprintf("%s:%d", config.Host, config.Port),
		},
	}

	api.setupRoutes()
	return api
}

func (a *API) setupRoutes() {
	// Middleware
	a.router.Use(middleware.RequestID)
	a.router.Use(middleware.RealIP)
	a.router.Use(middleware.Logger)
	a.router.Use(middleware.Recoverer)
	a.router.Use(a.authMiddleware)

	// API routes
	a.router.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/health", a.handleHealth)

		// Stats
		r.Get("/stats", a.handleStats)

		// Search
		r.Get("/search", a.handleSearch)
		r.Post("/search", a.handleSearch)

		// Debug
		if a.config.EnablePprof {
			r.Mount("/debug", middleware.Profiler())
		}
	})

	// Metrics
	if a.config.EnableMetrics {
		a.router.Mount("/metrics", a.metrics.Handler())
	}
}

func (a *API) Start() error {
	logrus.Infof("Starting HTTP API on %s", a.server.Addr)
	return a.server.ListenAndServe()
}

func (a *API) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

func (a *API) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.config.AuthToken != "" {
			token := r.Header.Get("Authorization")
			if token != "Bearer "+a.config.AuthToken {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (a *API) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := a.writer.Stats()
	json.NewEncoder(w).Encode(stats)
}

type SearchRequest struct {
	Query     string    `json:"query"`
	FromTime  time.Time `json:"from_time"`
	ToTime    time.Time `json:"to_time"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
	OrderBy   string    `json:"order_by"`
	OrderDesc bool      `json:"order_desc"`
}

func (a *API) handleSearch(w http.ResponseWriter, r *http.Request) {
	var req SearchRequest

	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Parse GET parameters
		req.Query = r.URL.Query().Get("q")
		req.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
		req.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
		req.OrderBy = r.URL.Query().Get("order_by")
		req.OrderDesc = r.URL.Query().Get("order_desc") == "true"

		// Parse time range
		if from := r.URL.Query().Get("from"); from != "" {
			req.FromTime, _ = time.Parse(time.RFC3339, from)
		}
		if to := r.URL.Query().Get("to"); to != "" {
			req.ToTime, _ = time.Parse(time.RFC3339, to)
		}
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 100
	}
	if req.OrderBy == "" {
		req.OrderBy = "timestamp"
	}

	// Execute search
	results, err := a.writer.Search(r.Context(), writer.SearchParams{
		Query:     req.Query,
		FromTime:  req.FromTime,
		ToTime:    req.ToTime,
		Limit:     req.Limit,
		Offset:    req.Offset,
		OrderBy:   req.OrderBy,
		OrderDesc: req.OrderDesc,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}
