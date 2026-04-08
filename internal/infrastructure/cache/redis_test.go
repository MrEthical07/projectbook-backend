package cache

import (
	"testing"
	"time"

	"github.com/MrEthical07/superapi/internal/core/config"
)

func TestResolveRedisConfigUsesRedisURLFallback(t *testing.T) {
	t.Setenv("REDIS_URL", "redis://:secret@localhost:6380/2")

	resolved, err := resolveRedisConfig(config.RedisConfig{
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     10,
	})
	if err != nil {
		t.Fatalf("resolveRedisConfig() error = %v", err)
	}

	if got, want := resolved.Addr, "localhost:6380"; got != want {
		t.Fatalf("Addr=%q want=%q", got, want)
	}
	if got, want := resolved.Password, "secret"; got != want {
		t.Fatalf("Password=%q want=%q", got, want)
	}
	if got, want := resolved.DB, 2; got != want {
		t.Fatalf("DB=%d want=%d", got, want)
	}
}

func TestResolveRedisConfigPrefersExplicitAddr(t *testing.T) {
	t.Setenv("REDIS_URL", "redis://:secret@localhost:6380/2")

	resolved, err := resolveRedisConfig(config.RedisConfig{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	if err != nil {
		t.Fatalf("resolveRedisConfig() error = %v", err)
	}

	if got, want := resolved.Addr, "127.0.0.1:6379"; got != want {
		t.Fatalf("Addr=%q want=%q", got, want)
	}
	if got, want := resolved.DB, 0; got != want {
		t.Fatalf("DB=%d want=%d", got, want)
	}
}

func TestResolveRedisConfigParsesURLAddr(t *testing.T) {
	resolved, err := resolveRedisConfig(config.RedisConfig{
		Addr: "redis://:secret@localhost:6381/4",
	})
	if err != nil {
		t.Fatalf("resolveRedisConfig() error = %v", err)
	}

	if got, want := resolved.Addr, "localhost:6381"; got != want {
		t.Fatalf("Addr=%q want=%q", got, want)
	}
	if got, want := resolved.Password, "secret"; got != want {
		t.Fatalf("Password=%q want=%q", got, want)
	}
	if got, want := resolved.DB, 4; got != want {
		t.Fatalf("DB=%d want=%d", got, want)
	}
}

func TestResolveRedisConfigRejectsInvalidURL(t *testing.T) {
	t.Setenv("REDIS_URL", "://not-valid")

	_, err := resolveRedisConfig(config.RedisConfig{})
	if err == nil {
		t.Fatalf("expected parse error for invalid REDIS_URL")
	}
}
