package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"url-shortener/internal/model"
)

type mockURLStore struct {
	findByLongURLFn    func(ctx context.Context, longURL string) (*model.URLMapping, error)
	findByShortCodeFn  func(ctx context.Context, shortCode string) (*model.URLMapping, error)
	insertFn           func(ctx context.Context, longURL, shortCode string, expiresAt *time.Time) (*model.URLMapping, error)
}

func (m *mockURLStore) FindByLongURL(ctx context.Context, longURL string) (*model.URLMapping, error) {
	return m.findByLongURLFn(ctx, longURL)
}

func (m *mockURLStore) FindByShortCode(ctx context.Context, shortCode string) (*model.URLMapping, error) {
	return m.findByShortCodeFn(ctx, shortCode)
}

func (m *mockURLStore) Insert(ctx context.Context, longURL, shortCode string, expiresAt *time.Time) (*model.URLMapping, error) {
	return m.insertFn(ctx, longURL, shortCode, expiresAt)
}

func TestEncodeBase62(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   int64
		want string
	}{
		{name: "zero", in: 0, want: "0"},
		{name: "single digit", in: 61, want: "Z"},
		{name: "base rollover", in: 62, want: "10"},
		{name: "multi digit", in: 3843, want: "ZZ"},
		{name: "large", in: 100000, want: "q0U"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := encodeBase62(tc.in); got != tc.want {
				t.Fatalf("encodeBase62(%d) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestShortenReturnsExistingCode(t *testing.T) {
	svc := NewShortenerService(&mockURLStore{
		findByLongURLFn: func(_ context.Context, longURL string) (*model.URLMapping, error) {
			if longURL != "https://example.com/path" {
				t.Fatalf("got normalized url %q", longURL)
			}
			return &model.URLMapping{ShortCode: "abc123"}, nil
		},
		findByShortCodeFn: func(context.Context, string) (*model.URLMapping, error) {
			t.Fatal("FindByShortCode should not be called when mapping exists")
			return nil, nil
		},
		insertFn: func(context.Context, string, string, *time.Time) (*model.URLMapping, error) {
			t.Fatal("Insert should not be called when mapping exists")
			return nil, nil
		},
	})

	got, err := svc.Shorten(context.Background(), " https://example.com/path ", nil)
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}
	if got != "abc123" {
		t.Fatalf("Shorten() = %q, want %q", got, "abc123")
	}
}

func TestShortenInsertsWhenNotFound(t *testing.T) {
	store := &mockURLStore{}
	store.findByLongURLFn = func(context.Context, string) (*model.URLMapping, error) { return nil, pgx.ErrNoRows }
	store.findByShortCodeFn = func(context.Context, string) (*model.URLMapping, error) { return nil, pgx.ErrNoRows }
	store.insertFn = func(_ context.Context, longURL, shortCode string, _ *time.Time) (*model.URLMapping, error) {
		if longURL != "https://example.com" {
			t.Fatalf("insert longURL = %q", longURL)
		}
		if shortCode == "" {
			t.Fatal("insert shortCode should not be empty")
		}
		return &model.URLMapping{ShortCode: shortCode}, nil
	}

	svc := NewShortenerService(store)
	got, err := svc.Shorten(context.Background(), "https://example.com", nil)
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}
	if got == "" {
		t.Fatal("Shorten() returned empty short code")
	}
}

func TestShortenValidationAndErrors(t *testing.T) {
	svc := NewShortenerService(&mockURLStore{
		findByLongURLFn: func(context.Context, string) (*model.URLMapping, error) { return nil, pgx.ErrNoRows },
		findByShortCodeFn: func(context.Context, string) (*model.URLMapping, error) { return nil, pgx.ErrNoRows },
		insertFn: func(context.Context, string, string, *time.Time) (*model.URLMapping, error) {
			return nil, errors.New("db down")
		},
	})

	if _, err := svc.Shorten(context.Background(), "ftp://example.com", nil); !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}

	past := time.Now().UTC().Add(-1 * time.Minute)
	if _, err := svc.Shorten(context.Background(), "https://example.com", &past); !errors.Is(err, ErrInvalidExpiryTime) {
		t.Fatalf("expected ErrInvalidExpiryTime, got %v", err)
	}

	if _, err := svc.Shorten(context.Background(), "https://example.com", nil); err == nil {
		t.Fatal("expected insert error")
	}
}

func TestResolveScenarios(t *testing.T) {
	now := time.Now().UTC()
	tests := []struct {
		name    string
		ret     *model.URLMapping
		err     error
		wantErr error
	}{
		{name: "not found", err: pgx.ErrNoRows, wantErr: ErrShortCodeNotFound},
		{name: "expired", ret: &model.URLMapping{LongURL: "https://example.com", ExpiresAt: ptrTime(now.Add(-time.Second))}, wantErr: ErrLinkExpired},
		{name: "success", ret: &model.URLMapping{LongURL: "https://example.com", ExpiresAt: ptrTime(now.Add(time.Hour))}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc := NewShortenerService(&mockURLStore{
				findByLongURLFn: func(context.Context, string) (*model.URLMapping, error) { return nil, nil },
				findByShortCodeFn: func(context.Context, string) (*model.URLMapping, error) { return tc.ret, tc.err },
				insertFn: func(context.Context, string, string, *time.Time) (*model.URLMapping, error) { return nil, nil },
			})

			got, err := svc.Resolve(context.Background(), "abc")
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Resolve() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}
			if got != "https://example.com" {
				t.Fatalf("Resolve() = %q", got)
			}
		})
	}
}

func ptrTime(t time.Time) *time.Time { return &t }
