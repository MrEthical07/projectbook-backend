package cache

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	corecache "github.com/MrEthical07/superapi/internal/core/cache"
	"github.com/MrEthical07/superapi/internal/core/config"
)

// NewRedisClient creates a Redis client and performs fail-fast startup ping checks.
func NewRedisClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	resolved, err := resolveRedisConfig(cfg)
	if err != nil {
		return nil, err
	}
	return corecache.NewRedisClient(ctx, resolved)
}

// CheckRedisHealth performs a bounded Redis ping for readiness probes.
func CheckRedisHealth(ctx context.Context, client *redis.Client, timeout time.Duration) error {
	return corecache.CheckHealth(ctx, client, timeout)
}

func resolveRedisConfig(cfg config.RedisConfig) (config.RedisConfig, error) {
	resolved := cfg

	// REDIS_URL is used as a fallback when REDIS_ADDR is not provided.
	redisURL := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if strings.TrimSpace(resolved.Addr) == "" {
		if redisURL == "" {
			return resolved, nil
		}

		opts, err := redis.ParseURL(redisURL)
		if err != nil {
			return resolved, fmt.Errorf("parse REDIS_URL: %w", err)
		}

		resolved.Addr = opts.Addr
		if resolved.Password == "" {
			resolved.Password = opts.Password
		}
		resolved.DB = opts.DB
		if opts.DialTimeout > 0 {
			resolved.DialTimeout = opts.DialTimeout
		}
		if opts.ReadTimeout > 0 {
			resolved.ReadTimeout = opts.ReadTimeout
		}
		if opts.WriteTimeout > 0 {
			resolved.WriteTimeout = opts.WriteTimeout
		}
		if opts.PoolSize > 0 {
			resolved.PoolSize = opts.PoolSize
		}
		if opts.MinIdleConns > 0 {
			resolved.MinIdleConns = opts.MinIdleConns
		}
		return resolved, nil
	}

	if strings.Contains(resolved.Addr, "://") {
		opts, err := redis.ParseURL(resolved.Addr)
		if err != nil {
			return resolved, fmt.Errorf("parse redis url addr: %w", err)
		}

		resolved.Addr = opts.Addr
		if resolved.Password == "" {
			resolved.Password = opts.Password
		}
		resolved.DB = opts.DB
	}

	return resolved, nil
}
