package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"

	"url-shortener/internal/analytics"
	"url-shortener/internal/cache"
	"url-shortener/internal/config"
	"url-shortener/internal/handler"
	urlhandler "url-shortener/internal/handler/url"
	"url-shortener/internal/middleware"
	"url-shortener/internal/monitoring"
	"url-shortener/internal/repository"
	"url-shortener/internal/service"
)

func New(cfg *config.Config, db *pgxpool.Pool, rdb *redis.Client) (*gin.Engine, *analytics.Queue) {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	healthHandler := handler.NewHealthHandler()
	urlRepo := repository.NewURLRepository(db)
	shortenerService := service.NewShortenerService(urlRepo)
	shortenHandler := urlhandler.NewShortenHandler(shortenerService)

	var rateLimitMiddleware gin.HandlerFunc = func(c *gin.Context) { c.Next() }
	if cfg.FeatureRateLimitEnabled {
		rateLimiter := middleware.NewRateLimiter(rdb, cfg.RateLimitRequests, cfg.RateLimitWindow)
		rateLimitMiddleware = rateLimiter.Middleware()
	}

	var urlCache cache.URLCache
	if cfg.FeatureCacheEnabled && rdb != nil {
		urlCache = cache.NewRedisURLCache(rdb, cfg.RedisCacheTTL)
	}

	clickRepo := repository.NewClickRepository(db)
	analyticsQueue := analytics.NewQueue(2048, clickRepo)
	if !cfg.FeatureAnalyticsEnabled {
		analyticsQueue = nil
	}

	var hooks monitoring.Hooks = monitoring.NoopHooks{}
	if cfg.FeatureMonitoringEnabled {
		hooks = monitoring.LogHooks{}
	}

	redirectHandler := urlhandler.NewRedirectHandler(shortenerService, urlCache, analyticsQueue, hooks)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Health)
		v1.POST("/shorten", rateLimitMiddleware, shortenHandler.Shorten)
	}

	r.GET("/:short_code", redirectHandler.Redirect)

	return r, analyticsQueue
}
