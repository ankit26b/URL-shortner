package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv          string
	AppPort         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration

	PostgresHost            string
	PostgresPort            string
	PostgresUser            string
	PostgresPassword        string
	PostgresDB              string
	PostgresSSLMode         string
	PostgresMaxConns        int32
	PostgresMinConns        int32
	PostgresMaxConnLifetime time.Duration

	RedisHost         string
	RedisPort         string
	RedisPassword     string
	RedisDB           int
	RedisDialTimeout  time.Duration
	RedisReadTimeout  time.Duration
	RedisWriteTimeout time.Duration
	RedisCacheTTL     time.Duration

	RateLimitRequests int64
	RateLimitWindow   time.Duration

	FeatureCacheEnabled      bool
	FeatureAnalyticsEnabled  bool
	FeatureRateLimitEnabled  bool
	FeatureMonitoringEnabled bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		AppPort:         getEnv("APP_PORT", "8080"),
		ReadTimeout:     getDurationEnv("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
		ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 10*time.Second),

		PostgresHost:            getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:            getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:            getEnv("POSTGRES_USER", "shortener"),
		PostgresPassword:        getEnv("POSTGRES_PASSWORD", "shortener"),
		PostgresDB:              getEnv("POSTGRES_DB", "shortener"),
		PostgresSSLMode:         getEnv("POSTGRES_SSLMODE", "disable"),
		PostgresMaxConns:        getInt32Env("POSTGRES_MAX_CONNS", 10),
		PostgresMinConns:        getInt32Env("POSTGRES_MIN_CONNS", 2),
		PostgresMaxConnLifetime: getDurationEnv("POSTGRES_MAX_CONN_LIFETIME", time.Hour),

		RedisHost:         getEnv("REDIS_HOST", "localhost"),
		RedisPort:         getEnv("REDIS_PORT", "6379"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		RedisDB:           getIntEnv("REDIS_DB", 0),
		RedisDialTimeout:  getDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second),
		RedisReadTimeout:  getDurationEnv("REDIS_READ_TIMEOUT", 3*time.Second),
		RedisWriteTimeout: getDurationEnv("REDIS_WRITE_TIMEOUT", 3*time.Second),
		RedisCacheTTL:     getDurationEnv("REDIS_CACHE_TTL", 10*time.Minute),

		RateLimitRequests: int64(getIntEnv("RATE_LIMIT_REQUESTS", 100)),
		RateLimitWindow:   getDurationEnv("RATE_LIMIT_WINDOW", time.Minute),

		FeatureCacheEnabled:      getBoolEnv("FEATURE_CACHE_ENABLED", true),
		FeatureAnalyticsEnabled:  getBoolEnv("FEATURE_ANALYTICS_ENABLED", true),
		FeatureRateLimitEnabled:  getBoolEnv("FEATURE_RATE_LIMIT_ENABLED", true),
		FeatureMonitoringEnabled: getBoolEnv("FEATURE_MONITORING_ENABLED", true),
	}

	if cfg.AppPort == "" {
		return nil, fmt.Errorf("APP_PORT cannot be empty")
	}

	return cfg, nil
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
		c.PostgresSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return fallback
}

func getInt32Env(key string, fallback int32) int32 {
	if v, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(v)
		if err == nil {
			return int32(i)
		}
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}
