package url

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"

	"url-shortener/internal/cache"
	"url-shortener/internal/model"
	"url-shortener/internal/monitoring"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"
)

type stubURLStore struct {
	findByLongURLFn   func(context.Context, string) (*model.URLMapping, error)
	findByShortCodeFn func(context.Context, string) (*model.URLMapping, error)
	insertFn          func(context.Context, string, string, *time.Time) (*model.URLMapping, error)
}

func (s *stubURLStore) FindByLongURL(ctx context.Context, longURL string) (*model.URLMapping, error) {
	return s.findByLongURLFn(ctx, longURL)
}
func (s *stubURLStore) FindByShortCode(ctx context.Context, shortCode string) (*model.URLMapping, error) {
	return s.findByShortCodeFn(ctx, shortCode)
}
func (s *stubURLStore) Insert(ctx context.Context, longURL, shortCode string, expiresAt *time.Time) (*model.URLMapping, error) {
	return s.insertFn(ctx, longURL, shortCode, expiresAt)
}

type stubCache struct {
	getFn func(context.Context, string) (string, error)
	setFn func(context.Context, string, string) error
}

func (s stubCache) Get(ctx context.Context, shortCode string) (string, error) { return s.getFn(ctx, shortCode) }
func (s stubCache) Set(ctx context.Context, shortCode, longURL string) error { return s.setFn(ctx, shortCode, longURL) }

var _ cache.URLCache = (*stubCache)(nil)

type stubAnalytics struct{ called chan model.ClickEvent }

func (s *stubAnalytics) Publish(event model.ClickEvent) error {
	s.called <- event
	return nil
}

func newTestRouter(shorten *ShortenHandler, redirect *RedirectHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/shorten", shorten.Shorten)
	r.GET("/:short_code", redirect.Redirect)
	return r
}

func TestShortenEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{name: "missing url", body: map[string]any{}, wantStatus: http.StatusBadRequest},
		{name: "invalid url", body: map[string]any{"long_url": "not-a-url"}, wantStatus: http.StatusBadRequest},
		{name: "valid", body: map[string]any{"long_url": "https://example.com"}, wantStatus: http.StatusOK},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo := &stubURLStore{
				findByLongURLFn: func(context.Context, string) (*model.URLMapping, error) { return nil, pgx.ErrNoRows },
				findByShortCodeFn: func(context.Context, string) (*model.URLMapping, error) { return nil, pgx.ErrNoRows },
				insertFn: func(_ context.Context, longURL, shortCode string, _ *time.Time) (*model.URLMapping, error) {
					return &model.URLMapping{LongURL: longURL, ShortCode: shortCode}, nil
				},
			}
			svc := service.NewShortenerService(repo)
			shorten := NewShortenHandler(svc)
			redirect := NewRedirectHandler(svc, nil, nil, monitoring.NoopHooks{})
			r := newTestRouter(shorten, redirect)

			payload, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d body=%s", w.Code, tc.wantStatus, w.Body.String())
			}
		})
	}
}

func TestRedirectEndpoint(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		repo := &stubURLStore{
			findByLongURLFn:   func(context.Context, string) (*model.URLMapping, error) { return nil, nil },
			findByShortCodeFn: func(context.Context, string) (*model.URLMapping, error) { t.Fatal("db should not be queried on cache hit"); return nil, nil },
			insertFn:          func(context.Context, string, string, *time.Time) (*model.URLMapping, error) { return nil, nil },
		}
		svc := service.NewShortenerService(repo)
		analytics := &stubAnalytics{called: make(chan model.ClickEvent, 1)}
		cache := stubCache{
			getFn: func(context.Context, string) (string, error) { return "https://cached.example", nil },
			setFn: func(context.Context, string, string) error { return nil },
		}
		r := newTestRouter(NewShortenHandler(svc), NewRedirectHandler(svc, cache, analytics, monitoring.NoopHooks{}))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/abc", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusFound {
			t.Fatalf("status = %d", w.Code)
		}
		if got := w.Header().Get("Location"); got != "https://cached.example" {
			t.Fatalf("location = %q", got)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &stubURLStore{
			findByLongURLFn:   func(context.Context, string) (*model.URLMapping, error) { return nil, nil },
			findByShortCodeFn: func(context.Context, string) (*model.URLMapping, error) { return nil, pgx.ErrNoRows },
			insertFn:          func(context.Context, string, string, *time.Time) (*model.URLMapping, error) { return nil, nil },
		}
		svc := service.NewShortenerService(repo)
		cache := stubCache{
			getFn: func(context.Context, string) (string, error) { return "", redis.Nil },
			setFn: func(context.Context, string, string) error { return nil },
		}
		r := newTestRouter(NewShortenHandler(svc), NewRedirectHandler(svc, cache, nil, monitoring.NoopHooks{}))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("cache miss falls back to db", func(t *testing.T) {
		repo := &stubURLStore{
			findByLongURLFn: func(context.Context, string) (*model.URLMapping, error) { return nil, nil },
			findByShortCodeFn: func(context.Context, string) (*model.URLMapping, error) {
				return &model.URLMapping{LongURL: "https://db.example"}, nil
			},
			insertFn: func(context.Context, string, string, *time.Time) (*model.URLMapping, error) { return nil, nil },
		}
		svc := service.NewShortenerService(repo)
		cache := stubCache{
			getFn: func(context.Context, string) (string, error) { return "", errors.New("redis unavailable") },
			setFn: func(context.Context, string, string) error { return nil },
		}
		r := newTestRouter(NewShortenHandler(svc), NewRedirectHandler(svc, cache, nil, monitoring.NoopHooks{}))

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusFound {
			t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
		}
	})
}

var _ repository.URLStore = (*stubURLStore)(nil)
