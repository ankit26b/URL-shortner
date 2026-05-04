package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"url-shortener/internal/cache/redis"
	"url-shortener/internal/config"
	"url-shortener/internal/logger"
	"url-shortener/internal/router"
	"url-shortener/internal/storage/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	lg := logger.New(cfg.AppEnv)

	ctx := context.Background()

	pg, err := postgres.New(ctx, cfg)
	if err != nil {
		lg.Fatalf("failed to initialize postgres: %v", err)
	}
	defer pg.Close()

	rdb, err := redis.New(cfg)
	if err != nil {
		lg.Fatalf("failed to initialize redis: %v", err)
	}
	defer rdb.Close()

	r := router.New(cfg, pg, rdb)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.AppPort),
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		lg.Printf("server listening on port %s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lg.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		lg.Printf("graceful shutdown failed: %v", err)
	}

	lg.Println("server stopped")
}
