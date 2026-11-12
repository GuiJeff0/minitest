package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/GuiJeff0/minitest/microservices/minitest-api/internal/domain"
)

type stubItemService struct {
	items []domain.Item
}

func (s *stubItemService) List(ctx context.Context) ([]domain.Item, error) {
	return s.items, nil
}

func (s *stubItemService) Create(ctx context.Context, name string) (domain.Item, error) {
	if strings.TrimSpace(name) == "" {
		return domain.Item{}, domain.ErrInvalidItem
	}
	item := domain.Item{ID: 1, Name: name, CreatedAt: time.Now()}
	s.items = append(s.items, item)
	return item, nil
}

func TestProbesAndBaseRoutes(t *testing.T) {
	svc := &stubItemService{}

	t.Run("healthz returns 200", func(t *testing.T) {
		router := NewRouter(svc, func(ctx context.Context) error { return nil }, "test", nil)
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("readyz returns 200 when dependency is healthy", func(t *testing.T) {
		router := NewRouter(svc, func(ctx context.Context) error { return nil }, "test", nil)
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("readyz returns 503 when dependency fails", func(t *testing.T) {
		router := NewRouter(svc, func(ctx context.Context) error { return errors.New("db down") }, "test", nil)
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status 503, got %d", rec.Code)
		}
	})

	t.Run("root returns service metadata", func(t *testing.T) {
		router := NewRouter(svc, func(ctx context.Context) error { return nil }, "dev", nil)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "minitest-api") {
			t.Fatalf("unexpected body: %s", rec.Body.String())
		}
	})
}
