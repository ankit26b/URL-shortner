package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"url-shortener/internal/repository"
)

const base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var ErrInvalidURL = errors.New("invalid URL format")
var ErrShortCodeNotFound = errors.New("short code not found")
var ErrLinkExpired = errors.New("link expired")

type ShortenerService struct {
	repo *repository.URLRepository
}

func NewShortenerService(repo *repository.URLRepository) *ShortenerService {
	return &ShortenerService{repo: repo}
}

func (s *ShortenerService) Shorten(ctx context.Context, longURL string) (string, error) {
	normalized, err := validateAndNormalizeURL(longURL)
	if err != nil {
		return "", err
	}

	existing, err := s.repo.FindByLongURL(ctx, normalized)
	if err == nil {
		return existing.ShortCode, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	id, err := s.nextID(ctx)
	if err != nil {
		return "", err
	}

	shortCode := encodeBase62(id)
	mapping, err := s.repo.Insert(ctx, normalized, shortCode)
	if err != nil {
		return "", err
	}

	return mapping.ShortCode, nil
}

func (s *ShortenerService) Resolve(ctx context.Context, shortCode string) (string, error) {
	mapping, err := s.repo.FindByShortCode(ctx, shortCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrShortCodeNotFound
	}
	if err != nil {
		return "", err
	}

	if mapping.ExpiresAt != nil && mapping.ExpiresAt.Before(time.Now().UTC()) {
		return "", ErrLinkExpired
	}

	return mapping.LongURL, nil
}

func (s *ShortenerService) nextID(ctx context.Context) (int64, error) {
	for {
		const candidateStartID int64 = 100000
		candidate, err := randomID(candidateStartID)
		if err != nil {
			return 0, err
		}

		_, err = s.repo.FindByShortCode(ctx, encodeBase62(candidate))
		if errors.Is(err, pgx.ErrNoRows) {
			return candidate, nil
		}
		if err != nil {
			return 0, err
		}
	}
}

func validateAndNormalizeURL(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", ErrInvalidURL
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ErrInvalidURL
	}

	return parsed.String(), nil
}

func encodeBase62(num int64) string {
	if num == 0 {
		return string(base62Alphabet[0])
	}

	result := make([]byte, 0)
	for num > 0 {
		rem := num % 62
		result = append(result, base62Alphabet[rem])
		num /= 62
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}
