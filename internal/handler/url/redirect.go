package url

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"url-shortener/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"url-shortener/internal/cache"
	"url-shortener/internal/monitoring"
	"url-shortener/internal/service"
)

type RedirectHandler struct {
	service        *service.ShortenerService
	cache          cache.URLCache
	analyticsQueue AnalyticsPublisher
	hooks          monitoring.Hooks
}

type AnalyticsPublisher interface {
	Publish(event model.ClickEvent) error
}

func NewRedirectHandler(service *service.ShortenerService, cache cache.URLCache, analyticsQueue AnalyticsPublisher, hooks monitoring.Hooks) *RedirectHandler {
	if hooks == nil {
		hooks = monitoring.NoopHooks{}
	}
	return &RedirectHandler{service: service, cache: cache, analyticsQueue: analyticsQueue, hooks: hooks}
}

func (h *RedirectHandler) Redirect(c *gin.Context) {
	shortCode := strings.TrimSpace(c.Param("short_code"))
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "short_code is required"})
		return
	}

	if h.cache != nil {
		if cachedURL, err := h.cache.Get(c.Request.Context(), shortCode); err == nil {
			h.hooks.Observe("redirect_cache_hit")
			h.trackRedirectAsync(c, shortCode)
			c.Redirect(http.StatusFound, cachedURL)
			return
		} else if !errors.Is(err, redis.Nil) {
			log.Printf("redis get failed for short_code=%s: %v", shortCode, err)
		}
	}

	longURL, err := h.service.Resolve(c.Request.Context(), shortCode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrShortCodeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "short link not found"})
		case errors.Is(err, service.ErrLinkExpired):
			c.JSON(http.StatusGone, gin.H{"error": "short link expired"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve short url"})
		}
		return
	}
	h.hooks.Observe("redirect_db_hit")

	if h.cache != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			if err := h.cache.Set(ctx, shortCode, longURL); err != nil {
				log.Printf("redis set failed for short_code=%s: %v", shortCode, err)
			}
		}()
	}

	h.trackRedirectAsync(c, shortCode)
	c.Redirect(http.StatusFound, longURL)
}

func (h *RedirectHandler) trackRedirectAsync(c *gin.Context, shortCode string) {
	if h.analyticsQueue == nil {
		return
	}

	event := model.ClickEvent{
		ShortCode: shortCode,
		Timestamp: time.Now().UTC(),
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	go func() {
		if err := h.analyticsQueue.Publish(event); err != nil {
			log.Printf("failed to queue analytics event for short_code=%s: %v", shortCode, err)
		}
	}()
}
