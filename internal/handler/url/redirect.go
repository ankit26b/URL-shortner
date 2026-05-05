package url

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"url-shortener/internal/service"
)

type RedirectHandler struct {
	service *service.ShortenerService
	cache   *redis.Client
}

func NewRedirectHandler(service *service.ShortenerService, cache *redis.Client) *RedirectHandler {
	return &RedirectHandler{service: service, cache: cache}
}

func (h *RedirectHandler) Redirect(c *gin.Context) {
	shortCode := strings.TrimSpace(c.Param("short_code"))
	if shortCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "short_code is required"})
		return
	}

	if h.cache != nil {
		if cachedURL, err := h.cache.Get(c.Request.Context(), "url:"+shortCode).Result(); err == nil {
			h.logRequestAsync(c, shortCode, true)
			c.Redirect(http.StatusFound, cachedURL)
			return
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

	if h.cache != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			_ = h.cache.Set(ctx, "url:"+shortCode, longURL, 10*time.Minute).Err()
		}()
	}

	h.logRequestAsync(c, shortCode, false)
	c.Redirect(http.StatusFound, longURL)
}

func (h *RedirectHandler) logRequestAsync(c *gin.Context, shortCode string, cacheHit bool) {
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")
	go func() {
		log.Printf("redirect short_code=%s ip=%s cache_hit=%t user_agent=%q", shortCode, ip, cacheHit, ua)
	}()
}
