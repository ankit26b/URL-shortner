package url

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/service"
)

type ShortenHandler struct {
	service *service.ShortenerService
}

type shortenRequest struct {
	LongURL    string     `json:"long_url" binding:"required"`
	ExpiryTime *time.Time `json:"expiry_time,omitempty"`
}

type shortenResponse struct {
	ShortCode string `json:"short_code"`
}

func NewShortenHandler(service *service.ShortenerService) *ShortenHandler {
	return &ShortenHandler{service: service}
}

func (h *ShortenHandler) Shorten(c *gin.Context) {
	var req shortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "long_url is required"})
		return
	}

	shortCode, err := h.service.Shorten(c.Request.Context(), req.LongURL, req.ExpiryTime)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid URL format"})
			return
		}
		if errors.Is(err, service.ErrInvalidExpiryTime) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expiry_time must be in the future"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create short url"})
		return
	}

	c.JSON(http.StatusOK, shortenResponse{ShortCode: shortCode})
}
