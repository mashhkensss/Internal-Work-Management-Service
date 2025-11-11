package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
    envHTTPAddr        = "HTTP_ADDR"
    envReadTimeout     = "HTTP_READ_TIMEOUT"
    envWriteTimeout    = "HTTP_WRITE_TIMEOUT"
    envShutdownTimeout = "HTTP_SHUTDOWN_TIMEOUT"
    envPostgresDSN     = "POSTGRES_DSN"
    envJWTSecret       = "JWT_SECRET"
)

var (
	ErrInvalidConfig = errors.New("invalid config")
)

type Config struct {
    HTTPAddr            string
    HTTPReadTimeout     time.Duration
    HTTPWriteTimeout    time.Duration
    HTTPShutdownTimeout time.Duration
    PostgresDSN         string
    JWTSecret           string
}

// Load reads configuration from env
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:            ":8080",
		HTTPReadTimeout:     5 * time.Second,
		HTTPWriteTimeout:    10 * time.Second,
		HTTPShutdownTimeout: 10 * time.Second,
	}

	if raw, ok := os.LookupEnv(envHTTPAddr); ok && strings.TrimSpace(raw) != "" {
		addr := strings.TrimSpace(raw)

		_, portStr, err := net.SplitHostPort(addr)
		if err != nil {
			return Config{}, fmt.Errorf("%w: %s must be in host:port format: %v", ErrInvalidConfig, envHTTPAddr, err)
		}

		port, err := strconv.Atoi(portStr)
		if err != nil || port <= 0 || port > 65535 {
			return Config{}, fmt.Errorf("%w: %s port must be between 1 and 65535", ErrInvalidConfig, envHTTPAddr)
		}

		cfg.HTTPAddr = addr
	}

	var err error
	if cfg.HTTPReadTimeout, err = durationFromEnv(envReadTimeout, cfg.HTTPReadTimeout); err != nil {
		return Config{}, err
	}
	if cfg.HTTPWriteTimeout, err = durationFromEnv(envWriteTimeout, cfg.HTTPWriteTimeout); err != nil {
		return Config{}, err
	}
	if cfg.HTTPShutdownTimeout, err = durationFromEnv(envShutdownTimeout, cfg.HTTPShutdownTimeout); err != nil {
		return Config{}, err
	}

    if raw, ok := os.LookupEnv(envPostgresDSN); ok {
        cfg.PostgresDSN = strings.TrimSpace(raw)
    }
    if cfg.PostgresDSN == "" {
        return Config{}, fmt.Errorf("%w: %s must be provided", ErrInvalidConfig, envPostgresDSN)
    }

    if raw, ok := os.LookupEnv(envJWTSecret); ok {
        cfg.JWTSecret = strings.TrimSpace(raw)
    }
    if cfg.JWTSecret == "" {
        return Config{}, fmt.Errorf("%w: %s must be provided", ErrInvalidConfig, envJWTSecret)
    }

    return cfg, nil
}

func durationFromEnv(key string, def time.Duration) (time.Duration, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return def, nil
	}
	value, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%w: %s must be a valid duration", ErrInvalidConfig, key)
	}
	if value < 0 {
		return 0, fmt.Errorf("%w: %s must be positive", ErrInvalidConfig, key)
	}
	return value, nil
}
