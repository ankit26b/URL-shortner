package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"

	"url-shortener/internal/config"
	"url-shortener/internal/handler"
)

func New(cfg *config.Config, db *pgxpool.Pool, rdb *redis.Client) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	_ = db
	_ = rdb

	h := handler.NewHealthHandler()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", h.Health)
	}

	return r
}
