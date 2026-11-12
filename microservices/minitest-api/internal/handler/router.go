package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/GuiJeff0/minitest/microservices/minitest-api/internal/domain"
)

// ReadinessCheck verifies external dependencies such as PostgreSQL before serving traffic.
type ReadinessCheck func(ctx context.Context) error

// Handler encapsulates HTTP endpoints and dependencies for minitest-api.
type Handler struct {
	service   domain.ItemService
	readyFn   ReadinessCheck
	appEnv    string
	logger    *slog.Logger
}

// NewRouter constructs the HTTP handler tree wrapped with recovery and request logging middlewares.
func NewRouter(svc domain.ItemService, readyFn ReadinessCheck, appEnv string, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	h := &Handler{
		service: svc,
		readyFn: readyFn,
		appEnv:  appEnv,
		logger:  logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", h.root)
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET /readyz", h.readyz)
	mux.HandleFunc("GET /api/v1/items", h.listItems)
	mux.HandleFunc("POST /api/v1/items", h.createItem)

	return h.recoverMiddleware(h.loggingMiddleware(mux))
}

func (h *Handler) root(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "minitest-api",
		"env":     h.appEnv,
		"status":  "running",
	})
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) readyz(w http.ResponseWriter, r *http.Request) {
	if h.readyFn == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not ready",
			"error":  "readiness check not configured",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.readyFn(ctx); err != nil {
		h.logger.Error("readiness check failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "not ready",
			"error":  "database unreachable",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}

func (h *Handler) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		h.logger.Error("failed to list items", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list items"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) createItem(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}

	item, err := h.service.Create(r.Context(), payload.Name)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidItem) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		h.logger.Error("failed to create item", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create item"})
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		h.logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func (h *Handler) recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				h.logger.Error("panic recovered", "panic", rec)
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
